package gamma

import (
	"github.com/cagelam/polymarket-go/internal/transport"
)

// Client implements public gamma queries.
type Client struct{ transport *transport.Client }

// New shares an owned transport.
func New(t *transport.Client) *Client { return &Client{transport: t} }
