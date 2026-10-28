package clob

import (
	"context"
	"net/url"

	"github.com/cagelam/polymarket-go/internal/transport"
)

// Client implements public clob queries.
type Client struct{ transport *transport.Client }

// New shares an owned transport.
func New(t *transport.Client) *Client { return &Client{transport: t} }
func (c *Client) request(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	return c.transport.Request(ctx, c.transport.CLOBURL, method, path, query, body)
}
