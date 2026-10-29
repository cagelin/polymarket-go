package evm

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"sync/atomic"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/transport"
	"github.com/cagelin/polymarket-go/models"
)

// RPC shares the SDK HTTP transport and assigns unique request IDs.
type RPC struct {
	transport *transport.Client
	sequence  atomic.Uint64
	verified  atomic.Bool
}

// NewRPC borrows an existing SDK transport.
func NewRPC(t *transport.Client) *RPC { return &RPC{transport: t} }

// Request sends exactly one RPC request without automatic write retries.
func (r *RPC) Request(ctx context.Context, method string, params []any) (json.RawMessage, error) {
	id := r.sequence.Add(1)
	raw, err := r.transport.Request(ctx, r.transport.RPCURL, "POST", "", nil, map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	if err != nil {
		return nil, err
	}
	return parseRPC(raw, method, id)
}
func parseRPC(raw []byte, method string, id uint64) (json.RawMessage, error) {
	var envelope struct {
		ID     uint64          `json:"id"`
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int             `json:"code"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid JSON-RPC envelope", Cause: err}
	}
	if envelope.ID != id {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "JSON-RPC response ID mismatch"}
	}
	if envelope.Error != nil {
		return nil, &sdkerrors.RPCError{Method: method, Code: envelope.Error.Code, Message: envelope.Error.Message, Data: envelope.Error.Data}
	}
	if len(envelope.Result) == 0 {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "JSON-RPC response has no result"}
	}
	return envelope.Result, nil
}

// Quantity decodes a nonnegative RPC hex quantity without float conversion.
func (r *RPC) Quantity(ctx context.Context, method string, params []any) (*big.Int, error) {
	raw, err := r.Request(ctx, method, params)
	if err != nil {
		return nil, err
	}
	var text string
	if err = json.Unmarshal(raw, &text); err != nil || !strings.HasPrefix(text, "0x") {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid RPC quantity", Cause: err}
	}
	value, ok := new(big.Int).SetString(text[2:], 16)
	if !ok || value.Sign() < 0 {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid RPC quantity"}
	}
	return value, nil
}

// VerifyChain prevents signing for a mismatched RPC network.
func (r *RPC) VerifyChain(ctx context.Context) error {
	if r.verified.Load() {
		return nil
	}
	chain, err := r.Quantity(ctx, "eth_chainId", []any{})
	if err != nil {
		return err
	}
	if chain.Cmp(big.NewInt(137)) != 0 {
		return &sdkerrors.UserInputError{Message: "RPC chain must be Polygon (137)"}
	}
	r.verified.Store(true)
	return nil
}

// Call reads contract state at the latest block.
func (r *RPC) Call(ctx context.Context, call models.TransactionCall) (string, error) {
	raw, err := r.Request(ctx, "eth_call", []any{map[string]string{"to": string(call.To), "data": call.Data}, "latest"})
	if err != nil {
		return "", err
	}
	return rpcString(raw)
}
func rpcString(raw []byte) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", &sdkerrors.UnexpectedResponseError{Message: "RPC result must be a string", Cause: err}
	}
	if _, err := Hex(value); err != nil {
		return "", &sdkerrors.UnexpectedResponseError{Message: "RPC result must contain hex data", Cause: err}
	}
	return value, nil
}

// CallBatch preserves input order even when the server reorders responses.
func (r *RPC) CallBatch(ctx context.Context, calls []models.TransactionCall) ([]string, error) {
	if len(calls) == 0 {
		return []string{}, nil
	}
	if len(calls) == 1 {
		value, err := r.Call(ctx, calls[0])
		return []string{value}, err
	}
	envelopes := make([]map[string]any, len(calls))
	ids := make([]uint64, len(calls))
	for i, call := range calls {
		ids[i] = r.sequence.Add(1)
		envelopes[i] = map[string]any{"jsonrpc": "2.0", "id": ids[i], "method": "eth_call", "params": []any{map[string]string{"to": string(call.To), "data": call.Data}, "latest"}}
	}
	raw, err := r.transport.Request(ctx, r.transport.RPCURL, "POST", "", nil, envelopes)
	var rejected *sdkerrors.RequestRejectedError
	if errors.As(err, &rejected) && rejected.Status >= 500 {
		middle := (len(calls) + 1) / 2
		left, leftErr := r.CallBatch(ctx, calls[:middle])
		if leftErr != nil {
			return nil, leftErr
		}
		right, rightErr := r.CallBatch(ctx, calls[middle:])
		if rightErr != nil {
			return nil, rightErr
		}
		return append(left, right...), nil
	}
	if err != nil {
		return nil, err
	}
	var rows []json.RawMessage
	if err = json.Unmarshal(raw, &rows); err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid batch RPC response", Cause: err}
	}
	if len(rows) != len(calls) {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "batch RPC result count mismatch"}
	}
	byID := make(map[uint64]json.RawMessage, len(rows))
	for _, row := range rows {
		var item struct {
			ID uint64 `json:"id"`
		}
		if err = json.Unmarshal(row, &item); err != nil {
			return nil, err
		}
		if _, ok := byID[item.ID]; ok {
			return nil, &sdkerrors.UnexpectedResponseError{Message: "duplicate batch RPC ID"}
		}
		byID[item.ID] = row
	}
	results := make([]string, len(calls))
	for i, id := range ids {
		value, parseErr := parseRPC(byID[id], "eth_call", id)
		if parseErr != nil {
			return nil, parseErr
		}
		results[i], err = rpcString(value)
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}
