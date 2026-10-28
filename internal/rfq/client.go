// Package rfq implements public combo catalog queries.
package rfq

import (
	"strings"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/paging"
	"github.com/cagelam/polymarket-go/internal/transport"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
)

// Client shares the SDK transport for public RFQ queries.
type Client struct{ transport *transport.Client }

// New constructs a client with a shared transport.
func New(t *transport.Client) *Client { return &Client{transport: t} }

// ListComboMarketsParams selects excluded condition identifiers and page size.
type ListComboMarketsParams struct {
	Exclude  []string `json:"exclude"`
	PageSize *int     `json:"page_size"`
}

// ListComboMarkets lists markets available for combinations.
func (c *Client) ListComboMarkets(p ListComboMarketsParams) (*pagination.Paginator[models.ComboMarket], error) {
	size := 20
	if p.PageSize != nil {
		size = *p.PageSize
	}
	if size < 1 || size > 100 {
		return nil, &sdkerrors.UserInputError{Message: "page_size must be between 1 and 100"}
	}
	query := map[string]any{}
	if len(p.Exclude) > 0 {
		query["exclude"] = strings.Join(p.Exclude, ",")
	}
	return paging.KeysetPaginator(c.transport, "rfq", c.transport.RFQURL, "/v1/rfq/combo-markets", "cursor", query, size, func(raw []byte) ([]models.ComboMarket, *string, error) {
		fields, err := codec.Object(raw)
		if err != nil {
			return nil, nil, err
		}
		items, err := codec.Required[[]models.ComboMarket](fields, "markets")
		if err != nil {
			return nil, nil, err
		}
		next, err := paging.ServerNextCursor(fields)
		return items, next, err
	}), nil
}
