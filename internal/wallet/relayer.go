package wallet

import (
	"context"
	"errors"
	"math/big"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/evm"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/transactions"
)

var nonceMismatch = regexp.MustCompile(`(?i)batch nonce\s+(\d+)\s+does not match on-chain nonce\s+(\d+)`)

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func rejectedNonce(err error) (string, bool) {
	var rejection *sdkerrors.RequestRejectedError
	if !errors.As(err, &rejection) || rejection.Status != 400 {
		return "", false
	}
	matches := nonceMismatch.FindStringSubmatch(rejection.Message)
	if len(matches) != 3 {
		return "", false
	}
	submitted, _ := new(big.Int).SetString(matches[1], 10)
	onchain, _ := new(big.Int).SetString(matches[2], 10)
	return matches[2], submitted.Cmp(onchain) < 0
}
func retryable(err error) bool {
	if _, ok := errors.AsType[*sdkerrors.RateLimitError](err); ok {
		return true
	}
	var rejected *sdkerrors.RequestRejectedError
	if !errors.As(err, &rejected) || rejected.Status != 400 {
		return false
	}
	_, stale := rejectedNonce(err)
	message := strings.ToLower(rejected.Message)
	return stale || strings.Contains(message, "wallet has in-flight action") || strings.Contains(message, "wallet busy") && strings.Contains(message, "active action")
}
func (c *Client) executeRelayed(ctx context.Context, calls []models.TransactionCall, metadata string) (*transactions.Handle, error) {
	if utf8.RuneCountInString(metadata) > 500 {
		return nil, &sdkerrors.UserInputError{Message: "metadata must be at most 500 characters"}
	}
	if _, err := c.headers("POST", "/submit", nil); err != nil {
		return nil, err
	}
	for attempt := 0; ; attempt++ {
		payload, err := c.relayerPayload(ctx, calls, metadata, "")
		if err != nil {
			return nil, err
		}
		handle, err := c.submit(ctx, payload)
		if nonce, _ := rejectedNonce(err); nonce != "" && c.config.Kind == models.WalletDeposit {
			payload, err = c.relayerPayload(ctx, calls, metadata, nonce)
			if err != nil {
				return nil, err
			}
			handle, err = c.submit(ctx, payload)
		}
		if err == nil || attempt == 10 || !retryable(err) {
			return handle, err
		}
		if err = wait(ctx, c.config.PollInterval); err != nil {
			return nil, err
		}
	}
}
func (c *Client) relayParams(ctx context.Context, kind string) (string, string, error) {
	path := "/v1/account/transactions/params"
	if kind == "PROXY" {
		path = "/relay-payload"
	}
	raw, err := c.transport.AuthenticatedRequest(ctx, c.transport.RelayerURL, "GET", path, url.Values{"address": {c.config.Signer.Address()}, "type": {kind}}, nil, c.headers)
	if err != nil {
		return "", "", err
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return "", "", err
	}
	nonce, err := codec.Required[string](fields, "nonce")
	if err != nil {
		return "", "", err
	}
	value, ok := new(big.Int).SetString(nonce, 10)
	if !ok || value.Sign() < 0 || value.BitLen() > 256 || strings.Trim(nonce, "0123456789") != "" {
		return "", "", &sdkerrors.UnexpectedResponseError{Message: "invalid relayer nonce"}
	}
	address, err := codec.Required[string](fields, "address")
	if err != nil {
		return "", "", err
	}
	normalized, err := evm.Address(address)
	if err != nil {
		return "", "", &sdkerrors.UnexpectedResponseError{Message: "invalid relayer address", Cause: err}
	}
	return nonce, normalized.Hex(), nil
}
func (c *Client) relayerPayload(ctx context.Context, calls []models.TransactionCall, metadata, nonceOverride string) (map[string]any, error) {
	kind := map[models.WalletType]string{models.WalletDeposit: "WALLET", models.WalletProxy: "PROXY", models.WalletSafe: "SAFE"}[c.config.Kind]
	if kind == "" {
		return nil, &sdkerrors.UserInputError{Message: "unsupported relayer wallet type"}
	}
	nonce, relay := nonceOverride, ""
	var err error
	if nonce == "" {
		nonce, relay, err = c.relayParams(ctx, kind)
		if err != nil {
			return nil, err
		}
	}
	body := map[string]any{"type": kind, "from": c.config.Signer.Address(), "nonce": nonce, "metadata": metadata}
	switch c.config.Kind {
	case models.WalletDeposit:
		timestamp := c.now().Unix()
		if c.transport.RPCURL != "https://polygon.drpc.org" {
			raw, requestErr := c.rpc.Request(ctx, "eth_getBlockByNumber", []any{"latest", false})
			if requestErr != nil {
				return nil, requestErr
			}
			fields, decodeErr := codec.Object(raw)
			if decodeErr != nil {
				return nil, decodeErr
			}
			value, decodeErr := codec.Required[string](fields, "timestamp")
			if decodeErr != nil {
				return nil, decodeErr
			}
			blockTime, parseErr := strconv.ParseInt(strings.TrimPrefix(value, "0x"), 16, 64)
			if parseErr != nil {
				return nil, parseErr
			}
			timestamp = max(timestamp, blockTime)
		}
		deadline := strconv.FormatInt(timestamp+600, 10)
		signature, signErr := c.config.Signer.SignDepositBatch(string(c.config.Wallet), calls, nonce, deadline, c.config.Authority)
		if signErr != nil {
			return nil, signErr
		}
		encoded := make([]map[string]string, len(calls))
		for i, call := range calls {
			encoded[i] = map[string]string{"target": string(call.To), "value": call.Value.String(), "data": call.Data}
		}
		body["to"], body["signature"] = depositFactory, signature
		body["depositWalletParams"] = map[string]any{"depositWallet": c.config.Wallet, "deadline": deadline, "calls": encoded}
	case models.WalletProxy:
		data, packErr := evm.ProxyCalls(calls)
		if packErr != nil {
			return nil, packErr
		}
		gas := big.NewInt(200000)
		estimated, estimateErr := c.rpc.Quantity(ctx, "eth_estimateGas", []any{map[string]string{"from": c.config.Signer.Address(), "to": proxyFactory, "data": data}})
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if estimateErr == nil {
			gas = estimated
		}
		signature, signErr := c.config.Signer.SignProxy(proxyFactory, data, gas.String(), nonce, relayHub, relay)
		if signErr != nil {
			return nil, signErr
		}
		body["to"], body["proxyWallet"], body["data"], body["signature"] = proxyFactory, c.config.Wallet, data, signature
		body["signatureParams"] = map[string]string{"gasLimit": gas.String(), "gasPrice": "0", "relay": relay, "relayHub": relayHub, "relayerFee": "0"}
	case models.WalletSafe:
		to, data, value, operation := string(calls[0].To), calls[0].Data, calls[0].Value, 0
		if len(calls) > 1 {
			to, value, operation = safeMultisend, new(big.Int), 1
			data, err = evm.SafeMultisend(calls)
			if err != nil {
				return nil, err
			}
		}
		signature, signErr := c.config.Signer.SignSafe(string(c.config.Wallet), to, data, value, operation, nonce)
		if signErr != nil {
			return nil, signErr
		}
		body["to"], body["proxyWallet"], body["data"], body["signature"] = to, c.config.Wallet, data, signature
		zero := "0x0000000000000000000000000000000000000000"
		body["signatureParams"] = map[string]string{"baseGas": "0", "gasPrice": "0", "gasToken": zero, "operation": strconv.Itoa(operation), "refundReceiver": zero, "safeTxnGas": "0"}
		if value.Sign() > 0 {
			body["value"] = value.String()
		}
	}
	return body, nil
}
