package data

import (
	"github.com/cagelin/polymarket-go/internal/transport"
)

// Client implements public data queries.
type Client struct{ transport *transport.Client }

// New shares an owned transport.
func New(t *transport.Client) *Client { return &Client{transport: t} }
