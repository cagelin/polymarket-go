// Package wallet implements account deployment and on-chain workflows.
package wallet

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/internal/evm"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/internal/transport"
	"github.com/cagelin/polymarket-go/models"
	"github.com/cagelin/polymarket-go/transactions"
)

const depositFactory = "0x00000000000Fb5C9ADea0298D729A0CB3823Cc07"
const safeMultisend = "0xA238CBeb142c10Ef7Ad8442C6D1f9E89e07e7761"
const proxyFactory = "0xaB45c5A4B0c941a2F231C04C3f49182e1A254052"
const relayHub = "0xD216153c06E857cD7f72665E0aF1d7D82172F494"

// Config controls signing authority and relayer credentials.
type Config struct {
	Signer       *signing.Signer
	Wallet       models.EVMAddress
	Kind         models.WalletType
	Authority    models.SignerType
	BuilderKey   *models.BuilderAPIKey
	RelayerKey   *models.RelayerAPIKey
	PollInterval time.Duration
	MaxPolls     int
	CLOBHeaders  func(string, string, []byte) (http.Header, error)
}

// Client shares a concrete transport and serializes submissions on this account.
type Client struct {
	transport  *transport.Client
	rpc        *evm.RPC
	config     Config
	submitGate chan struct{}
	now        func() time.Time
}

// New constructs wallet operations without making any network requests.
func New(t *transport.Client, config Config) (*Client, error) {
	if config.Signer == nil {
		return nil, &sdkerrors.UserInputError{Message: "wallet signer is required"}
	}
	if config.BuilderKey != nil && config.RelayerKey != nil {
		return nil, &sdkerrors.UserInputError{Message: "choose either BuilderKey or RelayerKey"}
	}
	if config.PollInterval < 0 || config.MaxPolls < 0 {
		return nil, &sdkerrors.UserInputError{Message: "invalid transaction polling configuration"}
	}
	if config.PollInterval == 0 {
		config.PollInterval = 2 * time.Second
	}
	if config.MaxPolls == 0 {
		config.MaxPolls = 100
	}
	if config.BuilderKey != nil {
		value := *config.BuilderKey
		config.BuilderKey = &value
	}
	if config.RelayerKey != nil {
		value := *config.RelayerKey
		config.RelayerKey = &value
	}
	return &Client{transport: t, rpc: evm.NewRPC(t), config: config, now: time.Now, submitGate: make(chan struct{}, 1)}, nil
}
func (c *Client) headers(method, path string, body []byte) (http.Header, error) {
	if key := c.config.RelayerKey; key != nil {
		return http.Header{"RELAYER_API_KEY": {key.Key}, "RELAYER_API_KEY_ADDRESS": {string(key.Address)}}, nil
	}
	key := c.config.BuilderKey
	if key == nil {
		return nil, &sdkerrors.UserInputError{Message: "BuilderKey or RelayerKey is required for sponsored transactions"}
	}
	timestamp := c.now().Unix()
	signature, err := signing.HMAC(key.Secret, timestamp, method, path, body)
	if err != nil {
		return nil, err
	}
	return http.Header{"POLY_BUILDER_API_KEY": {key.Key}, "POLY_BUILDER_PASSPHRASE": {key.Passphrase}, "POLY_BUILDER_TIMESTAMP": {strconv.FormatInt(timestamp, 10)}, "POLY_BUILDER_SIGNATURE": {signature}}, nil
}

// EnsureReady deploys only the current owner-derived Deposit Wallet when absent.
func (c *Client) EnsureReady(ctx context.Context) error {
	if c.config.Kind == models.WalletEOA {
		return nil
	}
	kind := string(c.config.Kind)
	switch c.config.Kind {
	case models.WalletDeposit:
		kind = "WALLET"
	case models.WalletProxy:
		kind = "PROXY"
	case models.WalletSafe:
		kind = "SAFE"
	}
	raw, err := c.transport.Request(ctx, c.transport.RelayerURL, "GET", "/deployed", url.Values{"address": {string(c.config.Wallet)}, "type": {kind}}, nil)
	if err != nil {
		return err
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return err
	}
	ready, err := codec.Required[bool](fields, "deployed")
	if err != nil || ready {
		return err
	}
	expected := c.config.Signer.Wallets()["BEACON"]
	if c.config.Kind != models.WalletDeposit || string(c.config.Wallet) != expected {
		return &sdkerrors.UserInputError{Message: "wallet is not deployed and cannot be created by this signer"}
	}
	handle, err := c.submit(ctx, map[string]any{"type": "WALLET-CREATE", "from": c.config.Signer.Address(), "to": depositFactory, "metadata": "Deploy Deposit Wallet"})
	if err != nil {
		return err
	}
	_, err = handle.Wait(ctx)
	return err
}

// ExecuteTransaction executes ordered calls. EOA batches wait between calls.
func (c *Client) ExecuteTransaction(ctx context.Context, calls []models.TransactionCall, metadata *string) (*transactions.Handle, error) {
	if len(calls) == 0 {
		return nil, &sdkerrors.UserInputError{Message: "at least one call is required"}
	}
	copied := make([]models.TransactionCall, len(calls))
	for i, call := range calls {
		to, err := signing.Address(string(call.To))
		if err != nil {
			return nil, err
		}
		if _, err = evm.Hex(call.Data); err != nil {
			return nil, err
		}
		value := call.Value
		if value == nil {
			value = new(big.Int)
		}
		value, err = evm.Uint256(value)
		if err != nil {
			return nil, err
		}
		copied[i] = models.TransactionCall{To: models.EVMAddress(to), Data: call.Data, Value: value}
	}
	if err := c.lock(ctx); err != nil {
		return nil, err
	}
	defer c.unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.config.Kind == models.WalletEOA {
		for i, call := range copied {
			handle, err := c.broadcast(ctx, call)
			if err != nil {
				return nil, err
			}
			if i == len(copied)-1 {
				return handle, nil
			}
			if _, err = handle.Wait(ctx); err != nil {
				return nil, err
			}
		}
	}
	text := "Execute transaction"
	if metadata != nil {
		text = *metadata
	}
	return c.executeRelayed(ctx, copied, text)
}
func (c *Client) broadcast(ctx context.Context, call models.TransactionCall) (*transactions.Handle, error) {
	if err := c.rpc.VerifyChain(ctx); err != nil {
		return nil, err
	}
	nonce, err := c.rpc.Quantity(ctx, "eth_getTransactionCount", []any{c.config.Signer.Address(), "pending"})
	if err != nil {
		return nil, err
	}
	price, err := c.rpc.Quantity(ctx, "eth_gasPrice", []any{})
	if err != nil {
		return nil, err
	}
	gas, err := c.rpc.Quantity(ctx, "eth_estimateGas", []any{map[string]string{"from": c.config.Signer.Address(), "to": string(call.To), "value": "0x" + call.Value.Text(16), "data": call.Data}})
	if err != nil {
		return nil, err
	}
	if !nonce.IsUint64() || !gas.IsUint64() {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "RPC nonce or gas exceeds uint64"}
	}
	signed, err := c.config.Signer.SignLegacyTransaction(call, nonce.Uint64(), gas.Uint64(), price)
	if err != nil {
		return nil, err
	}
	raw, err := c.rpc.Request(ctx, "eth_sendRawTransaction", []any{signed})
	if err != nil {
		return nil, err
	}
	var hash string
	if err = json.Unmarshal(raw, &hash); err != nil {
		return nil, err
	}
	decoded, decodeErr := evm.Hex(hash)
	if decodeErr != nil || len(decoded) != 32 {
		return nil, response("RPC returned an invalid transaction hash")
	}
	return transactions.NewEOA(c.rpc, hash, c.config.MaxPolls, c.config.PollInterval), nil
}
func (c *Client) submit(ctx context.Context, payload map[string]any) (*transactions.Handle, error) {
	raw, err := c.transport.AuthenticatedRequest(ctx, c.transport.RelayerURL, "POST", "/submit", nil, payload, c.headers)
	if err != nil {
		return nil, err
	}
	return c.submitted(raw)
}
func (c *Client) submitted(raw []byte) (*transactions.Handle, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	id, err := codec.Required[string](fields, "transactionID")
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, response("relayer transaction ID cannot be empty")
	}
	state, err := codec.Required[string](fields, "state")
	if err != nil {
		return nil, err
	}
	switch state {
	case "STATE_NEW", "STATE_EXECUTED", "STATE_MINED", "STATE_CONFIRMED", "STATE_INVALID", "STATE_FAILED":
	default:
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid relayer response state"}
	}
	var hash *string
	if value, ok := fields["transactionHash"]; ok {
		if err = json.Unmarshal(value, &hash); err != nil {
			return nil, err
		}
	}
	if hash != nil && *hash == "" {
		hash = nil
	}
	return transactions.NewRelayed(c.transport, c.headers, id, hash, c.config.MaxPolls, c.config.PollInterval), nil
}

func (c *Client) lock(ctx context.Context) error {
	select {
	case c.submitGate <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (c *Client) unlock() { <-c.submitGate }
