package polymarket

import (
	"context"
	"errors"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/streams"

	"github.com/cagelam/polymarket-go/internal/trading"
)

type tradingClient = trading.Client

// SecureClientConfig supplies signing material and shared transport options.
// Private keys and credentials must never be logged or committed.
type SecureClientConfig struct {
	ClientConfig
	PrivateKey  string
	Wallet      string
	Credentials *models.APIKeyCreds
	BuilderKey  *models.BuilderAPIKey
	Nonce       int64
}

func (c SecureClientConfig) String() string { return "SecureClientConfig(<redacted>)" }

// GoString redacts sensitive fields during Go-syntax formatting.
func (c SecureClientConfig) GoString() string { return c.String() }

// SecureClient adds authenticated account operations to the public client.
type SecureClient struct {
	*PublicClient
	*tradingClient
	config ClientConfig
}

// NewSecureClient authenticates the signer and requires an already-deployed contract wallet.
// Default wallet deployment is part of the following on-chain milestone.
func NewSecureClient(ctx context.Context, config SecureClientConfig) (*SecureClient, error) {
	public, err := NewPublicClient(config.ClientConfig)
	if err != nil {
		return nil, err
	}
	account, err := trading.New(ctx, public.transport, trading.Config{PrivateKey: config.PrivateKey, Wallet: config.Wallet, Credentials: config.Credentials, BuilderKey: config.BuilderKey, Nonce: config.Nonce})
	if err != nil {
		_ = public.Close()
		return nil, err
	}
	credentials := account.Credentials()
	authenticatedStreams, streamErr := streams.New(streams.Config{MarketURL: config.MarketWSURL, UserURL: config.UserWSURL, PriceURL: config.PriceWSURL, SportsURL: config.SportsWSURL, RTDSURL: config.RTDSWSURL, QueueSize: config.StreamQueueSize, HTTPClient: config.HTTPClient, Credentials: &credentials})
	if streamErr != nil {
		_ = public.Close()
		return nil, streamErr
	}
	_ = public.streamClient.Close()
	public.streamClient = authenticatedStreams
	return &SecureClient{PublicClient: public, tradingClient: account, config: config.ClientConfig}, nil
}

// EndAuthentication closes this client and returns a fresh public client.
// It revokes the server-side API key; 401 and 404 are treated as already revoked.
func (c *SecureClient) EndAuthentication(ctx context.Context) (*PublicClient, error) {
	err := c.DeleteAPIKey(ctx)
	_ = c.Close()
	var rejected *sdkerrors.RequestRejectedError
	if err != nil && (!errors.As(err, &rejected) || rejected.Status != 401 && rejected.Status != 404) {
		return nil, err
	}
	return NewPublicClient(c.config)
}
