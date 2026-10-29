package trading

import (
	"context"
	"net/http"

	"github.com/cagelin/polymarket-go/rfq"
)

// OpenRFQSession reuses one live authenticated quote session per client.
func (c *Client) OpenRFQSession(ctx context.Context, url string, httpClient *http.Client, queueSize int) (*rfq.Session, error) {
	if err := c.comboAllowed(); err != nil {
		return nil, err
	}
	c.rfqMu.Lock()
	defer c.rfqMu.Unlock()
	if c.rfqClosed {
		return nil, input("client is closed")
	}
	if c.rfqSession != nil && !c.rfqSession.Closed() {
		return c.rfqSession, nil
	}
	session, err := rfq.Open(ctx, rfq.Config{URL: url, HTTPClient: httpClient, QueueSize: queueSize, Signer: c.signer, Credentials: c.credentials, Wallet: c.wallet, WalletType: c.walletType})
	if err != nil {
		return nil, err
	}
	c.rfqSession = session
	return session, nil
}

// CloseRFQ closes the session owned by this client.
func (c *Client) CloseRFQ() error {
	c.rfqMu.Lock()
	defer c.rfqMu.Unlock()
	c.rfqClosed = true
	if c.rfqSession != nil {
		return c.rfqSession.Close()
	}
	return nil
}
