// Package transactions waits for confirmed EOA or sponsored wallet transactions.
package transactions

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/evm"
	"github.com/cagelam/polymarket-go/internal/transport"
	"github.com/cagelam/polymarket-go/models"
)

// Handle retains submitted identifiers even if waiting is canceled or times out.
// A new Wait may be used to reconcile a previously unknown outcome.
type Handle struct {
	TransactionID   *string
	TransactionHash *string
	rpc             *evm.RPC
	transport       *transport.Client
	headers         transport.HeaderResolver
	maxPolls        int
	pollDelay       time.Duration
}

// NewEOA constructs a handle for a transaction already broadcast by the SDK.
func NewEOA(rpc *evm.RPC, hash string, maxPolls int, delay time.Duration) *Handle {
	return &Handle{TransactionHash: &hash, rpc: rpc, maxPolls: maxPolls, pollDelay: delay}
}

// NewRelayed constructs a handle for an already accepted relayer transaction.
func NewRelayed(t *transport.Client, headers transport.HeaderResolver, id string, hash *string, maxPolls int, delay time.Duration) *Handle {
	var copiedHash *string
	if hash != nil {
		value := *hash
		copiedHash = &value
	}
	return &Handle{TransactionID: &id, TransactionHash: copiedHash, transport: t, headers: headers, maxPolls: maxPolls, pollDelay: delay}
}

// Wait waits for relayer confirmation or a successful on-chain receipt.
func (h *Handle) Wait(ctx context.Context) (*models.TransactionOutcome, error) {
	if h == nil || h.maxPolls <= 0 {
		return nil, &sdkerrors.UserInputError{Message: "invalid transaction handle"}
	}
	for attempt := 0; attempt < h.maxPolls; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var outcome *models.TransactionOutcome
		var err error
		if h.rpc != nil {
			outcome, err = h.receipt(ctx)
		} else {
			outcome, err = h.relayed(ctx)
		}
		if err != nil || outcome != nil {
			return outcome, err
		}
		if attempt+1 == h.maxPolls {
			break
		}
		timer := time.NewTimer(h.pollDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	timeout := &sdkerrors.TransactionTimeoutError{}
	if h.TransactionID != nil {
		timeout.TransactionID = *h.TransactionID
	}
	if h.TransactionHash != nil {
		timeout.TransactionHash = *h.TransactionHash
	}
	return nil, timeout
}
func (h *Handle) receipt(ctx context.Context) (*models.TransactionOutcome, error) {
	if h.TransactionHash == nil {
		return nil, &sdkerrors.UserInputError{Message: "missing transaction hash"}
	}
	raw, err := h.rpc.Request(ctx, "eth_getTransactionReceipt", []any{*h.TransactionHash})
	if err != nil {
		return nil, err
	}
	if string(raw) == "null" {
		return nil, nil
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	switch string(fields["status"]) {
	case `"0x1"`, "1":
		return &models.TransactionOutcome{TransactionHash: *h.TransactionHash}, nil
	case `"0x0"`, "0":
		return nil, &sdkerrors.TransactionFailedError{Message: "EOA transaction reverted: " + *h.TransactionHash}
	default:
		return nil, &sdkerrors.UnexpectedResponseError{Message: "receipt has unrecognized status"}
	}
}
func (h *Handle) relayed(ctx context.Context) (*models.TransactionOutcome, error) {
	if h.TransactionID == nil || h.transport == nil {
		return nil, &sdkerrors.UserInputError{Message: "missing relayer transaction"}
	}
	raw, err := h.transport.AuthenticatedRequest(ctx, h.transport.RelayerURL, "GET", "/v1/account/transactions/"+url.PathEscape(*h.TransactionID), nil, nil, h.headers)
	if err != nil {
		return nil, err
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	state, err := codec.Required[string](fields, "state")
	if err != nil {
		return nil, err
	}
	id, err := codec.Required[string](fields, "transaction_id")
	if err != nil {
		return nil, err
	}
	if id != *h.TransactionID {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "relayer transaction ID mismatch"}
	}
	hashRaw, ok := fields["transaction_hash"]
	if !ok {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "relayer response missing transaction_hash"}
	}
	var hash *string
	if err = json.Unmarshal(hashRaw, &hash); err != nil {
		return nil, err
	}
	switch state {
	case "STATE_CONFIRMED":
		if hash == nil || *hash == "" {
			hash = h.TransactionHash
		}
		if hash == nil || *hash == "" {
			return nil, &sdkerrors.UnexpectedResponseError{Message: "confirmed transaction has no hash"}
		}
		return &models.TransactionOutcome{TransactionHash: *hash, TransactionID: &id}, nil
	case "STATE_FAILED", "STATE_INVALID":
		message := "relayer transaction failed: " + id
		if errorMessage, hasError := fields["error_msg"]; hasError {
			var detail string
			if json.Unmarshal(errorMessage, &detail) == nil && detail != "" {
				message = detail
			}
		}
		return nil, &sdkerrors.TransactionFailedError{Message: message}
	case "STATE_NEW", "STATE_EXECUTED", "STATE_MINED":
		return nil, nil
	default:
		return nil, &sdkerrors.UnexpectedResponseError{Message: "unknown relayer transaction state"}
	}
}
