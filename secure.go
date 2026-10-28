package polymarket

import (
	"context"
	"errors"
	"time"

	"github.com/cagelam/polymarket-go/internal/wallet"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/rfq"
	"github.com/cagelam/polymarket-go/streams"

	"github.com/cagelam/polymarket-go/internal/trading"
)

type tradingClient = trading.Client
type walletClient = wallet.Client

// SecureClientConfig supplies signing material and shared transport options.
// Private keys and credentials must never be logged or committed.
type SecureClientConfig struct {
	ClientConfig
	PrivateKey              string
	Wallet                  string
	Credentials             *models.APIKeyCreds
	BuilderKey              *models.BuilderAPIKey
	Nonce                   int64
	RelayerKey              *models.RelayerAPIKey
	TransactionPollInterval time.Duration
	TransactionMaxPolls     int
}

func (c SecureClientConfig) String() string { return "SecureClientConfig(<redacted>)" }

// GoString redacts sensitive fields during Go-syntax formatting.
func (c SecureClientConfig) GoString() string { return c.String() }

// SecureClient adds authenticated account operations to the public client.
type SecureClient struct {
	*PublicClient
	*tradingClient
	*walletClient
	config ClientConfig
}

// NewSecureClient authenticates the signer and deploys its default Deposit Wallet when needed.
func NewSecureClient(ctx context.Context, config SecureClientConfig) (*SecureClient, error) {
	public, err := NewPublicClient(config.ClientConfig)
	if err != nil {
		return nil, err
	}
	account, err := trading.New(ctx, public.transport, trading.Config{PrivateKey: config.PrivateKey, Wallet: config.Wallet, Credentials: config.Credentials, BuilderKey: config.BuilderKey, Nonce: config.Nonce, RelayerKey: config.RelayerKey, TransactionPollInterval: config.TransactionPollInterval, TransactionMaxPolls: config.TransactionMaxPolls})
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
	return &SecureClient{PublicClient: public, tradingClient: account, walletClient: account.WalletClient(), config: config.ClientConfig}, nil
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

// GetTradingApprovalsState reads indexed allowances for this account or an explicit wallet.
func (c *SecureClient) GetTradingApprovalsState(ctx context.Context, address ...models.EVMAddress) (*models.TradingApprovalsState, error) {
	if len(address) > 1 {
		return nil, &sdkerrors.UserInputError{Message: "at most one wallet is accepted"}
	}
	owner := c.Wallet()
	if len(address) == 1 {
		owner = address[0]
	}
	return c.PublicClient.GetTradingApprovalsState(ctx, owner)
}

// OpenRFQSession opens or reuses the account's long-lived quote session.
func (c *SecureClient) OpenRFQSession(ctx context.Context) (*rfq.Session, error) {
	return c.tradingClient.OpenRFQSession(ctx, c.config.RFQWSURL, c.config.HTTPClient, c.config.StreamQueueSize)
}

// Close releases public streams and the authenticated RFQ session.
func (c *SecureClient) Close() error {
	return errors.Join(c.CloseRFQ(), c.PublicClient.Close())
}
