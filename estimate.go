package polymarket

import (
	"context"

	"github.com/cockroachdb/apd/v3"
)

// MarketOrderType controls whether a partial fill is allowed.
type MarketOrderType string

// FOK requires a full fill; FAK permits a partial fill.
const (
	FOK MarketOrderType = "FOK"
	FAK MarketOrderType = "FAK"
)

// EstimateMarketPriceParams specifies collateral Amount for BUY or Shares for SELL.
type EstimateMarketPriceParams struct {
	AssetQuery
	Side      OrderSide
	Amount    *apd.Decimal
	Shares    *apd.Decimal
	OrderType MarketOrderType // Empty defaults to FOK.
}

// EstimateMarketPrice returns the marginal limit price needed for the requested
// amount, not the average execution price. It never submits an order.
func (c *PublicClient) EstimateMarketPrice(ctx context.Context, params EstimateMarketPriceParams) (*apd.Decimal, error) {
	id, err := params.resolve()
	if err != nil {
		return nil, err
	}
	if err = validateSide(params.Side); err != nil {
		return nil, err
	}
	notional := params.Amount
	if params.Side == Buy {
		if params.Shares != nil {
			return nil, &UserInputError{Message: "shares must not be set for BUY"}
		}
	} else {
		if params.Amount != nil {
			return nil, &UserInputError{Message: "amount must not be set for SELL"}
		}
		notional = params.Shares
	}
	if notional == nil || notional.Form != apd.Finite || notional.Sign() <= 0 {
		return nil, &UserInputError{Message: "BUY amount or SELL shares must be positive and finite"}
	}
	orderType := params.OrderType
	if orderType == "" {
		orderType = FOK
	}
	if orderType != FOK && orderType != FAK {
		return nil, &UserInputError{Message: "order type must be FOK or FAK"}
	}
	book, err := c.GetOrderBook(ctx, Asset(id))
	if err != nil {
		return nil, err
	}
	if book.AssetID != id {
		return nil, &UnexpectedResponseError{Message: "order book asset differs from requested asset"}
	}
	return estimatePrice(book, params.Side, notional, orderType)
}

func estimatePrice(book *OrderBook, side OrderSide, notional *apd.Decimal, orderType MarketOrderType) (*apd.Decimal, error) {
	levels := book.Bids
	if side == Buy {
		levels = book.Asks
	}
	if len(levels) == 0 {
		return nil, &InsufficientLiquidityError{Message: "no resting liquidity"}
	}
	// Match Python's default decimal precision and half-even rounding. Do not
	// mutate apd.BaseContext or rely on a caller's process-wide decimal settings.
	decimalContext := apd.Context{
		Precision: 28, Rounding: apd.RoundHalfEven,
		MaxExponent: apd.MaxExponent, MinExponent: apd.MinExponent,
		Traps: apd.DefaultTraps,
	}
	var cumulative, cost apd.Decimal
	var price *apd.Decimal
	for i := len(levels) - 1; i >= 0; i-- {
		level := &levels[i]
		quantity := level.Size
		if side == Buy {
			if _, err := decimalContext.Mul(&cost, level.Size, level.Price); err != nil {
				return nil, &UnexpectedResponseError{Message: "order book amount is outside decimal range", Cause: err}
			}
			quantity = &cost
		}
		if _, err := decimalContext.Add(&cumulative, &cumulative, quantity); err != nil {
			return nil, &UnexpectedResponseError{Message: "order book size is outside decimal range", Cause: err}
		}
		if cumulative.Cmp(notional) >= 0 {
			price = level.Price
			break
		}
	}
	if price == nil {
		if orderType == FOK {
			return nil, &InsufficientLiquidityError{Message: "insufficient liquidity for full fill"}
		}
		price = levels[0].Price
	}
	var ceiling apd.Decimal
	if _, err := decimalContext.Sub(&ceiling, apd.New(1, 0), book.TickSize); err != nil {
		return nil, &UnexpectedResponseError{Message: "invalid tick size", Cause: err}
	}
	if price.Cmp(book.TickSize) < 0 || price.Cmp(&ceiling) > 0 {
		return nil, &UnexpectedResponseError{Message: "estimated price is outside the tick-size range"}
	}
	return new(apd.Decimal).Set(price), nil
}
