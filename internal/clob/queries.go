package clob

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

func (q AssetQuery) resolve() (models.ClobAssetID, error) {
	if q.AssetID == "" {
		return "", &sdkerrors.UserInputError{Message: "provide a nonempty asset_id"}
	}
	return q.AssetID, nil
}

type assetRequest struct {
	TokenID models.ClobAssetID `json:"token_id"`
}

type priceRequest struct {
	TokenID models.ClobAssetID `json:"token_id"`
	Side    models.OrderSide   `json:"side"`
}

func (q AssetsQuery) resolve() ([]assetRequest, error) {
	ids := q.AssetIDs
	if len(ids) == 0 {
		return nil, &sdkerrors.UserInputError{Message: "provide a nonempty list of asset identifiers"}
	}
	result := make([]assetRequest, len(ids))
	for i, id := range ids {
		if id == "" {
			return nil, &sdkerrors.UserInputError{Message: fmt.Sprintf("asset identifier %d is empty", i)}
		}
		result[i] = assetRequest{TokenID: id}
	}
	return result, nil
}

func validateSide(side models.OrderSide) error {
	if side != models.Buy && side != models.Sell {
		return &sdkerrors.UserInputError{Message: "side must be BUY or SELL"}
	}
	return nil
}

func (c *Client) singleAsset(ctx context.Context, path string, query AssetQuery, side models.OrderSide) ([]byte, error) {
	id, err := query.resolve()
	if err != nil {
		return nil, err
	}
	params := url.Values{"token_id": {string(id)}}
	if side != "" {
		params.Set("side", string(side))
	}
	return c.request(ctx, http.MethodGet, path, params, nil)
}

func (c *Client) multipleAssets(ctx context.Context, path string, query AssetsQuery) ([]byte, error) {
	body, err := query.resolve()
	if err != nil {
		return nil, err
	}
	return c.request(ctx, http.MethodPost, path, nil, body)
}

func parseDecimalResponse(data []byte, key string) (*decimal.Big, error) {
	fields, err := codec.Object(data)
	if err == nil {
		var value *decimal.Big
		value, err = codec.DecimalField(fields, key)
		if err == nil {
			return value, nil
		}
	}
	return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid decimal response", Cause: err}
}

func parseDecimalMap(data []byte) (map[models.ClobAssetID]*decimal.Big, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid decimal map", Cause: err}
	}
	result := make(map[models.ClobAssetID]*decimal.Big, len(fields))
	for key := range fields {
		value, parseErr := codec.DecimalField(fields, key)
		if parseErr != nil {
			return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid decimal map", Cause: parseErr}
		}
		result[models.ClobAssetID(key)] = value
	}
	return result, nil
}

// GetMidpoint returns the midpoint for an asset.
func (c *Client) GetMidpoint(ctx context.Context, query AssetQuery) (*decimal.Big, error) {
	data, err := c.singleAsset(ctx, "/midpoint", query, "")
	if err != nil {
		return nil, err
	}
	return parseDecimalResponse(data, "mid")
}

// GetMidpoints returns midpoints keyed by asset identifier.
func (c *Client) GetMidpoints(ctx context.Context, query AssetsQuery) (map[models.ClobAssetID]*decimal.Big, error) {
	data, err := c.multipleAssets(ctx, "/midpoints", query)
	if err != nil {
		return nil, err
	}
	return parseDecimalMap(data)
}

// GetPrice returns the price for an asset and side.
func (c *Client) GetPrice(ctx context.Context, query PriceRequest) (*decimal.Big, error) {
	if err := validateSide(query.Side); err != nil {
		return nil, err
	}
	data, err := c.singleAsset(ctx, "/price", query.AssetQuery, query.Side)
	if err != nil {
		return nil, err
	}
	return parseDecimalResponse(data, "price")
}

// GetPrices returns batch prices keyed by asset and side.
func (c *Client) GetPrices(ctx context.Context, requests []PriceRequest) (map[models.ClobAssetID]map[models.OrderSide]*decimal.Big, error) {
	if len(requests) == 0 {
		return nil, &sdkerrors.UserInputError{Message: "requests must be nonempty"}
	}
	body := make([]priceRequest, len(requests))
	for i, request := range requests {
		id, err := request.resolve()
		if err != nil {
			return nil, err
		}
		if err = validateSide(request.Side); err != nil {
			return nil, err
		}
		body[i] = priceRequest{TokenID: id, Side: request.Side}
	}
	data, err := c.request(ctx, http.MethodPost, "/prices", nil, body)
	if err != nil {
		return nil, err
	}
	fields, err := codec.Object(data)
	if err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid prices response", Cause: err}
	}
	result := make(map[models.ClobAssetID]map[models.OrderSide]*decimal.Big, len(fields))
	for id, raw := range fields {
		values, parseErr := parseDecimalMap(raw)
		if parseErr != nil {
			return nil, parseErr
		}
		prices := make(map[models.OrderSide]*decimal.Big, len(values))
		for side, price := range values {
			if sideErr := validateSide(models.OrderSide(side)); sideErr != nil {
				return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid price side", Cause: fmt.Errorf("unknown side %q", side)}
			}
			prices[models.OrderSide(side)] = price
		}
		result[models.ClobAssetID(id)] = prices
	}
	return result, nil
}

// GetOrderBook returns an asset's full order book.
func (c *Client) GetOrderBook(ctx context.Context, query AssetQuery) (*models.OrderBook, error) {
	data, err := c.singleAsset(ctx, "/book", query, "")
	if err != nil {
		return nil, err
	}
	book, err := codec.DecodeResponse[models.OrderBook](data)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// GetOrderBooks returns batch order books in the server's response order.
func (c *Client) GetOrderBooks(ctx context.Context, query AssetsQuery) ([]models.OrderBook, error) {
	data, err := c.multipleAssets(ctx, "/books", query)
	if err != nil {
		return nil, err
	}
	return codec.ParseList[models.OrderBook](data)
}

// GetSpread returns the bid-ask spread.
func (c *Client) GetSpread(ctx context.Context, query AssetQuery) (*decimal.Big, error) {
	data, err := c.singleAsset(ctx, "/spread", query, "")
	if err != nil {
		return nil, err
	}
	return parseDecimalResponse(data, "spread")
}

// GetSpreads returns spreads keyed by asset identifier.
func (c *Client) GetSpreads(ctx context.Context, query AssetsQuery) (map[models.ClobAssetID]*decimal.Big, error) {
	data, err := c.multipleAssets(ctx, "/spreads", query)
	if err != nil {
		return nil, err
	}
	return parseDecimalMap(data)
}

// GetLastTradePrice returns nil, nil when the asset has no trades.
func (c *Client) GetLastTradePrice(ctx context.Context, query AssetQuery) (*models.LastTradePrice, error) {
	data, err := c.singleAsset(ctx, "/last-trade-price", query, "")
	if err != nil {
		return nil, err
	}
	fields, err := codec.Object(data)
	if err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid last trade", Cause: err}
	}
	if _, err = codec.DecimalField(fields, "price"); err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid last trade price", Cause: err}
	}
	side, err := codec.Required[models.OrderSide](fields, "side")
	if err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid last trade side", Cause: err}
	}
	if side == "" {
		return nil, nil
	}
	trade, err := codec.DecodeResponse[models.LastTradePrice](data)
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

// GetLastTradePrices returns only assets with trades; match results by AssetID.
func (c *Client) GetLastTradePrices(ctx context.Context, query AssetsQuery) ([]models.LastTradePriceForAsset, error) {
	data, err := c.multipleAssets(ctx, "/last-trades-prices", query)
	if err != nil {
		return nil, err
	}
	return codec.ParseList[models.LastTradePriceForAsset](data)
}

// AssetQuery selects one asset by its canonical identifier.
type AssetQuery struct {
	AssetID models.ClobAssetID
}

// Asset returns a query using the canonical asset identifier.
func Asset(id models.ClobAssetID) AssetQuery { return AssetQuery{AssetID: id} }

// AssetsQuery selects a nonempty list. A nil slice means omitted.
type AssetsQuery struct {
	AssetIDs []models.ClobAssetID
}

// PriceRequest selects an asset and order side for single or batch requests.
type PriceRequest struct {
	AssetQuery
	Side models.OrderSide
}
