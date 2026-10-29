package clob

import (
	"context"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

// MarketOrderType controls whether a partial fill is allowed.
type MarketOrderType string

// FOK requires a full fill; FAK permits a partial fill.
const (
	FOK MarketOrderType = "FOK"
	FAK MarketOrderType = "FAK"
)

// EstimateMarketPriceParams specifies collateral Amount when buying or Shares when selling.
type EstimateMarketPriceParams struct {
	AssetQuery
	Side      models.OrderSide
	Amount    *decimal.Big
	Shares    *decimal.Big
	OrderType MarketOrderType // Empty defaults to FOK.
}

// EstimateMarketPrice returns the marginal limit price needed for the requested
// amount, not the average execution price. It never submits an order.
func (c *Client) EstimateMarketPrice(ctx context.Context, params EstimateMarketPriceParams) (*decimal.Big, error) {
	id, err := params.resolve()
	if err != nil {
		return nil, err
	}
	if err = validateSide(params.Side); err != nil {
		return nil, err
	}
	notional := params.Amount
	if params.Side == models.Buy {
		if params.Shares != nil {
			return nil, &sdkerrors.UserInputError{Message: "shares must not be set for BUY"}
		}
	} else {
		if params.Amount != nil {
			return nil, &sdkerrors.UserInputError{Message: "amount must not be set for SELL"}
		}
		notional = params.Shares
	}
	if notional == nil || !notional.IsFinite() || notional.Sign() <= 0 {
		return nil, &sdkerrors.UserInputError{Message: "BUY amount or SELL shares must be positive and finite"}
	}
	orderType := params.OrderType
	if orderType == "" {
		orderType = FOK
	}
	if orderType != FOK && orderType != FAK {
		return nil, &sdkerrors.UserInputError{Message: "order type must be FOK or FAK"}
	}
	book, err := c.GetOrderBook(ctx, Asset(id))
	if err != nil {
		return nil, err
	}
	if book.AssetID != id {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "order book asset differs from requested asset"}
	}
	return EstimatePrice(book, params.Side, notional, orderType)
}

// EstimatePrice calculates the marginal limit price from an existing order book.
func EstimatePrice(book *models.OrderBook, side models.OrderSide, notional *decimal.Big, orderType MarketOrderType) (*decimal.Big, error) {
	levels := book.Bids
	if side == models.Buy {
		levels = book.Asks
	}
	if len(levels) == 0 {
		return nil, &sdkerrors.InsufficientLiquidityError{Message: "no resting liquidity"}
	}
	decimalContext := codec.DecimalContext()
	cumulative, cost := decimal.WithContext(decimalContext), decimal.WithContext(decimalContext)
	var price *decimal.Big
	for i := len(levels) - 1; i >= 0; i-- {
		level := &levels[i]
		quantity := level.Size
		if side == models.Buy {
			if err := codec.DecimalArithmetic("mul", cost, level.Size, level.Price); err != nil {
				return nil, &sdkerrors.UnexpectedResponseError{Message: "order book amount is outside decimal range", Cause: err}
			}
			quantity = cost
		}
		if err := codec.DecimalArithmetic("add", cumulative, cumulative, quantity); err != nil {
			return nil, &sdkerrors.UnexpectedResponseError{Message: "order book size is outside decimal range", Cause: err}
		}
		if cumulative.Cmp(notional) >= 0 {
			price = level.Price
			break
		}
	}
	if price == nil {
		if orderType == FOK {
			return nil, &sdkerrors.InsufficientLiquidityError{Message: "insufficient liquidity for full fill"}
		}
		price = levels[0].Price
	}
	ceiling := decimal.WithContext(decimalContext)
	if err := codec.DecimalArithmetic("sub", ceiling, decimal.New(1, 0), book.TickSize); err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid tick size", Cause: err}
	}
	if price.Cmp(book.TickSize) < 0 || price.Cmp(ceiling) > 0 {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "estimated price is outside the tick-size range"}
	}
	return new(decimal.Big).Copy(price), nil
}
