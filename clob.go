package polymarket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cockroachdb/apd/v3"
)

func (q AssetQuery) resolve() (ClobAssetID, error) {
	if q.AssetID == "" {
		return "", &UserInputError{Message: "provide a nonempty asset_id"}
	}
	return q.AssetID, nil
}

type assetRequest struct {
	TokenID ClobAssetID `json:"token_id"`
}
type priceRequest struct {
	TokenID ClobAssetID `json:"token_id"`
	Side    OrderSide   `json:"side"`
}

func (q AssetsQuery) resolve() ([]assetRequest, error) {
	ids := q.AssetIDs
	if len(ids) == 0 {
		return nil, &UserInputError{Message: "provide a nonempty list of asset identifiers"}
	}
	result := make([]assetRequest, len(ids))
	for i, id := range ids {
		if id == "" {
			return nil, &UserInputError{Message: fmt.Sprintf("asset identifier %d is empty", i)}
		}
		result[i] = assetRequest{TokenID: id}
	}
	return result, nil
}

func validateSide(side OrderSide) error {
	if side != Buy && side != Sell {
		return &UserInputError{Message: "side must be BUY or SELL"}
	}
	return nil
}

func (c *PublicClient) singleAsset(ctx context.Context, path string, query AssetQuery, side OrderSide) ([]byte, error) {
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

func (c *PublicClient) multipleAssets(ctx context.Context, path string, query AssetsQuery) ([]byte, error) {
	body, err := query.resolve()
	if err != nil {
		return nil, err
	}
	return c.request(ctx, http.MethodPost, path, nil, body)
}

func parseDecimalResponse(data []byte, key string) (*apd.Decimal, error) {
	fields, err := object(data)
	if err == nil {
		var value *apd.Decimal
		value, err = decimalField(fields, key)
		if err == nil {
			return value, nil
		}
	}
	return nil, &UnexpectedResponseError{Message: "invalid decimal response", Cause: err}
}

func parseDecimalMap(data []byte) (map[ClobAssetID]*apd.Decimal, error) {
	fields, err := object(data)
	if err != nil {
		return nil, &UnexpectedResponseError{Message: "invalid decimal map", Cause: err}
	}
	result := make(map[ClobAssetID]*apd.Decimal, len(fields))
	for key := range fields {
		value, parseErr := decimalField(fields, key)
		if parseErr != nil {
			return nil, &UnexpectedResponseError{Message: "invalid decimal map", Cause: parseErr}
		}
		result[ClobAssetID(key)] = value
	}
	return result, nil
}

func parseList[T any](data []byte) ([]T, error) {
	result, err := decodeResponse[[]T](data)
	if err == nil && result == nil {
		err = &UnexpectedResponseError{Message: "expected response array"}
	}
	return result, err
}

// GetMidpoint returns the midpoint for an asset.
func (c *PublicClient) GetMidpoint(ctx context.Context, query AssetQuery) (*apd.Decimal, error) {
	data, err := c.singleAsset(ctx, "/midpoint", query, "")
	if err != nil {
		return nil, err
	}
	return parseDecimalResponse(data, "mid")
}

// GetMidpoints returns midpoints keyed by asset identifier.
func (c *PublicClient) GetMidpoints(ctx context.Context, query AssetsQuery) (map[ClobAssetID]*apd.Decimal, error) {
	data, err := c.multipleAssets(ctx, "/midpoints", query)
	if err != nil {
		return nil, err
	}
	return parseDecimalMap(data)
}

// GetPrice returns the price for an asset and side.
func (c *PublicClient) GetPrice(ctx context.Context, query PriceRequest) (*apd.Decimal, error) {
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
func (c *PublicClient) GetPrices(ctx context.Context, requests []PriceRequest) (map[ClobAssetID]map[OrderSide]*apd.Decimal, error) {
	if len(requests) == 0 {
		return nil, &UserInputError{Message: "requests must be nonempty"}
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
	fields, err := object(data)
	if err != nil {
		return nil, &UnexpectedResponseError{Message: "invalid prices response", Cause: err}
	}
	result := make(map[ClobAssetID]map[OrderSide]*apd.Decimal, len(fields))
	for id, raw := range fields {
		values, parseErr := parseDecimalMap(raw)
		if parseErr != nil {
			return nil, parseErr
		}
		prices := make(map[OrderSide]*apd.Decimal, len(values))
		for side, price := range values {
			if sideErr := validateSide(OrderSide(side)); sideErr != nil {
				return nil, &UnexpectedResponseError{Message: "invalid price side", Cause: fmt.Errorf("unknown side %q", side)}
			}
			prices[OrderSide(side)] = price
		}
		result[ClobAssetID(id)] = prices
	}
	return result, nil
}

// GetOrderBook returns an asset's full order book.
func (c *PublicClient) GetOrderBook(ctx context.Context, query AssetQuery) (*OrderBook, error) {
	data, err := c.singleAsset(ctx, "/book", query, "")
	if err != nil {
		return nil, err
	}
	book, err := decodeResponse[OrderBook](data)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// GetOrderBooks returns batch order books in the server's response order.
func (c *PublicClient) GetOrderBooks(ctx context.Context, query AssetsQuery) ([]OrderBook, error) {
	data, err := c.multipleAssets(ctx, "/books", query)
	if err != nil {
		return nil, err
	}
	return parseList[OrderBook](data)
}

// GetSpread returns the bid-ask spread.
func (c *PublicClient) GetSpread(ctx context.Context, query AssetQuery) (*apd.Decimal, error) {
	data, err := c.singleAsset(ctx, "/spread", query, "")
	if err != nil {
		return nil, err
	}
	return parseDecimalResponse(data, "spread")
}

// GetSpreads returns spreads keyed by asset identifier.
func (c *PublicClient) GetSpreads(ctx context.Context, query AssetsQuery) (map[ClobAssetID]*apd.Decimal, error) {
	data, err := c.multipleAssets(ctx, "/spreads", query)
	if err != nil {
		return nil, err
	}
	return parseDecimalMap(data)
}

// GetLastTradePrice returns nil, nil when the asset has no trades.
func (c *PublicClient) GetLastTradePrice(ctx context.Context, query AssetQuery) (*LastTradePrice, error) {
	data, err := c.singleAsset(ctx, "/last-trade-price", query, "")
	if err != nil {
		return nil, err
	}
	fields, err := object(data)
	if err != nil {
		return nil, &UnexpectedResponseError{Message: "invalid last trade", Cause: err}
	}
	if _, err = decimalField(fields, "price"); err != nil {
		return nil, &UnexpectedResponseError{Message: "invalid last trade price", Cause: err}
	}
	side, err := required[OrderSide](fields, "side")
	if err != nil {
		return nil, &UnexpectedResponseError{Message: "invalid last trade side", Cause: err}
	}
	if side == "" {
		return nil, nil
	}
	trade, err := decodeResponse[LastTradePrice](data)
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

// GetLastTradePrices returns only assets with trades; match results by AssetID.
func (c *PublicClient) GetLastTradePrices(ctx context.Context, query AssetsQuery) ([]LastTradePriceForAsset, error) {
	data, err := c.multipleAssets(ctx, "/last-trades-prices", query)
	if err != nil {
		return nil, err
	}
	return parseList[LastTradePriceForAsset](data)
}
