package polymarket

import (
	"errors"

	"github.com/cagelin/polymarket-go/internal/clob"
	"github.com/cagelin/polymarket-go/internal/data"
	"github.com/cagelin/polymarket-go/internal/gamma"
	"github.com/cagelin/polymarket-go/internal/rfq"
	"github.com/cagelin/polymarket-go/internal/transport"
	"github.com/cagelin/polymarket-go/models"
	"github.com/cagelin/polymarket-go/streams"
)

// ClientConfig configures all public service requests.
type ClientConfig = transport.Config

type clobClient = clob.Client
type gammaClient = gamma.Client
type dataClient = data.Client
type rfqClient = rfq.Client
type streamClient = streams.Client

// PublicClient exposes public APIs backed by one shared transport.
type PublicClient struct {
	*clobClient
	*gammaClient
	*dataClient
	*rfqClient
	*streamClient
	transport *transport.Client
}

// NewPublicClient constructs a client without making a network request.
func NewPublicClient(config ClientConfig) (*PublicClient, error) {
	t, err := transport.New(config)
	if err != nil {
		return nil, err
	}
	s, err := streams.New(streams.Config{MarketURL: config.MarketWSURL, SportsURL: config.SportsWSURL, RTDSURL: config.RTDSWSURL, HTTPClient: config.HTTPClient, QueueSize: config.StreamQueueSize})
	if err != nil {
		_ = t.Close()
		return nil, err
	}
	return &PublicClient{streamClient: s, rfqClient: rfq.New(t), clobClient: clob.New(t), gammaClient: gamma.New(t), dataClient: data.New(t), transport: t}, nil
}

// Close releases owned resources and prevents new requests.
func (c *PublicClient) Close() error { return errors.Join(c.streamClient.Close(), c.transport.Close()) }

// Environment identifies the upstream deployment. URL overrides in ClientConfig
// redirect transports for testing without introducing another upstream environment.
func (c *PublicClient) Environment() models.Environment {
	return models.Environment{Name: "production"}
}
