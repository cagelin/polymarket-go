// Package trading implements authenticated account and order operations.
package trading

import (
	"context"
	"errors"
	"math/big"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/cagelin/polymarket-go/rfq"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/internal/transport"
	walletops "github.com/cagelin/polymarket-go/internal/wallet"
	"github.com/cagelin/polymarket-go/models"
)

// Config supplies credentials and the wallet to authenticate.
type Config struct {
	PrivateKey              string
	Wallet                  string
	Credentials             *models.APIKeyCreds
	BuilderKey              *models.BuilderAPIKey
	Nonce                   int64
	RelayerKey              *models.RelayerAPIKey
	TransactionPollInterval time.Duration
	TransactionMaxPolls     int
}

func (c Config) String() string { return "trading.Config(<redacted>)" }

// GoString redacts sensitive fields during Go-syntax formatting.
func (c Config) GoString() string { return c.String() }

// Client holds immutable authentication state shared by concurrent operations.
type Client struct {
	rfqMu       sync.Mutex
	rfqSession  *rfq.Session
	rfqClosed   bool
	onchain     *walletops.Client
	salt        func(int) (*big.Int, error)
	metadata    metadataCache
	transport   *transport.Client
	signer      *signing.Signer
	wallet      models.EVMAddress
	walletType  models.WalletType
	signerType  models.SignerType
	credentials models.APIKeyCreds
	builderKey  *models.BuilderAPIKey
	now         func() time.Time
}

// New validates credentials, deriving them only when absent or rejected with 401.
func New(ctx context.Context, t *transport.Client, p Config) (*Client, error) {
	if p.Nonce < 0 || p.Credentials != nil && p.Nonce != 0 {
		return nil, input("nonce must be nonnegative and cannot accompany credentials")
	}
	signer, err := signing.New(p.PrivateKey)
	if err != nil {
		return nil, err
	}
	wallet := p.Wallet
	if wallet == "" {
		wallets := signer.Wallets()
		legacy, lookupErr := deployed(ctx, t, wallets["UUPS"], "WALLET")
		if lookupErr != nil {
			return nil, lookupErr
		}
		wallet = wallets["BEACON"]
		if legacy {
			wallet = wallets["UUPS"]
		}
	}
	wallet, err = signing.Address(wallet)
	if err != nil {
		return nil, err
	}
	kind, authority := signer.Classify(wallet)
	c := &Client{transport: t, signer: signer, wallet: models.EVMAddress(wallet), walletType: kind, signerType: authority, now: time.Now}
	if p.BuilderKey != nil {
		value := *p.BuilderKey
		c.builderKey = &value
	}
	valid := false
	if p.Credentials != nil {
		c.credentials = *p.Credentials
		keys, probeErr := c.FetchAPIKeys(ctx)
		var rejected *sdkerrors.RequestRejectedError
		if probeErr != nil && (!errors.As(probeErr, &rejected) || rejected.Status != 401) {
			return nil, probeErr
		}
		valid = probeErr == nil && slices.Contains(keys, c.credentials.Key)
	}
	if !valid {
		headers, signErr := signer.L1(c.now().Unix(), p.Nonce)
		if signErr != nil {
			return nil, signErr
		}
		resolver := func(string, string, []byte) (http.Header, error) { return headers, nil }
		raw, requestErr := t.AuthenticatedRequest(ctx, t.CLOBURL, "POST", "/auth/api-key", nil, nil, resolver)
		var rejected *sdkerrors.RequestRejectedError
		if errors.As(requestErr, &rejected) && rejected.Status == 400 {
			raw, requestErr = t.AuthenticatedRequest(ctx, t.CLOBURL, "GET", "/auth/derive-api-key", nil, nil, resolver)
		}
		if requestErr != nil {
			return nil, requestErr
		}
		c.credentials, err = codec.DecodeResponse[models.APIKeyCreds](raw)
		if err != nil {
			return nil, err
		}
	}
	c.onchain, err = walletops.New(t, walletops.Config{Signer: signer, Wallet: c.wallet, Kind: kind, Authority: authority, BuilderKey: c.builderKey, RelayerKey: p.RelayerKey, PollInterval: p.TransactionPollInterval, MaxPolls: p.TransactionMaxPolls, CLOBHeaders: c.headers})
	if err != nil {
		return nil, err
	}
	if err = c.onchain.EnsureReady(ctx); err != nil {
		return nil, err
	}

	return c, nil
}
func deployed(ctx context.Context, t *transport.Client, wallet, kind string) (bool, error) {
	raw, err := t.Request(ctx, t.RelayerURL, "GET", "/deployed", url.Values{"address": {wallet}, "type": {kind}}, nil)
	if err != nil {
		return false, err
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return false, response(err)
	}
	value, err := codec.Required[bool](fields, "deployed")
	return value, response(err)
}
func input(message string) error { return &sdkerrors.UserInputError{Message: message} }
func response(err error) error {
	if err == nil {
		return nil
	}
	return &sdkerrors.UnexpectedResponseError{Message: "authenticated response did not match expected shape", Cause: err}
}

// Wallet returns the account owning the funds and positions.
func (c *Client) Wallet() models.EVMAddress { return c.wallet }

// Signer returns the key's address, which may differ from the funded wallet.
func (c *Client) Signer() models.EVMAddress { return models.EVMAddress(c.signer.Address()) }

// WalletType returns the account's exchange signature model.
func (c *Client) WalletType() models.WalletType { return c.walletType }

// Credentials returns a copy of the active API credentials.
func (c *Client) Credentials() models.APIKeyCreds { return c.credentials }

// SignerType identifies owner versus session-key signing authority.
func (c *Client) SignerType() models.SignerType { return c.signerType }
func (c *Client) headers(method, path string, body []byte) (http.Header, error) {
	timestamp := c.now().Unix()
	signature, err := signing.HMAC(c.credentials.Secret, timestamp, method, path, body)
	if err != nil {
		return nil, err
	}
	return http.Header{"POLY_ADDRESS": {c.signer.Address()}, "POLY_API_KEY": {c.credentials.Key}, "POLY_PASSPHRASE": {c.credentials.Passphrase}, "POLY_TIMESTAMP": {strconv.FormatInt(timestamp, 10)}, "POLY_SIGNATURE": {signature}}, nil
}
func (c *Client) request(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	return c.transport.AuthenticatedRequest(ctx, c.transport.CLOBURL, method, path, query, body, c.headers)
}

// FetchAPIKeys lists active credentials without exposing their secrets.
func (c *Client) FetchAPIKeys(ctx context.Context) ([]string, error) {
	raw, err := c.request(ctx, "GET", "/auth/api-keys", nil, nil)
	if err != nil {
		return nil, err
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, response(err)
	}
	values, err := codec.Required[[]string](fields, "apiKeys")
	return values, response(err)
}

// WalletClient exposes the shared on-chain implementation to the facade.
func (c *Client) WalletClient() *walletops.Client { return c.onchain }
