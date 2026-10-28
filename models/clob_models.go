package models

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/ericlagergren/decimal"
)

// ClobAssetID identifies an exchange asset without integer-size restrictions.
type ClobAssetID string

// ConditionID identifies a conditional-token market.
type ConditionID string

// OrderSide selects the buy or sell side.
type OrderSide string

// Supported order sides.
const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

// OrderBookLevel is a price and available share quantity.
type OrderBookLevel struct {
	Price *decimal.Big `json:"price"`
	Size  *decimal.Big `json:"size"`
}

// OrderBook preserves the exchange's level ordering.
// Bids ascend and asks descend; the best level is at the end of each slice.
type OrderBook struct {
	ConditionID    ConditionID      `json:"condition_id"`
	AssetID        ClobAssetID      `json:"asset_id"`
	Timestamp      *time.Time       `json:"timestamp"`
	Bids           []OrderBookLevel `json:"bids"`
	Asks           []OrderBookLevel `json:"asks"`
	MinOrderSize   *decimal.Big     `json:"min_order_size"`
	TickSize       *decimal.Big     `json:"tick_size"`
	NegRisk        bool             `json:"neg_risk"`
	LastTradePrice *decimal.Big     `json:"last_trade_price"`
	Hash           string           `json:"hash"`
}

// LastTradePrice describes the most recent trade.
type LastTradePrice struct {
	Price *decimal.Big `json:"price"`
	Side  OrderSide    `json:"side"`
}

// LastTradePriceForAsset identifies an asset's most recent trade.
type LastTradePriceForAsset struct {
	AssetID ClobAssetID `json:"asset_id"`
	LastTradePrice
}

func assetField(fields map[string]json.RawMessage) (ClobAssetID, error) {
	if _, ok := fields["asset_id"]; ok {
		return codec.Required[ClobAssetID](fields, "asset_id")
	}
	return codec.Required[ClobAssetID](fields, "token_id")
}

// UnmarshalJSON validates required fields and decimal strings.
func (l *OrderBookLevel) UnmarshalJSON(data []byte) error {
	fields, err := codec.Object(data)
	if err != nil {
		return err
	}
	var result OrderBookLevel
	if result.Price, err = codec.DecimalField(fields, "price"); err != nil {
		return err
	}
	if result.Size, err = codec.DecimalField(fields, "size"); err != nil {
		return err
	}
	*l = result
	return nil
}

// UnmarshalJSON maps exchange field names to canonical SDK fields.
func (b *OrderBook) UnmarshalJSON(data []byte) error {
	fields, err := codec.Object(data)
	if err != nil {
		return err
	}
	var result OrderBook
	if result.ConditionID, err = codec.Required[ConditionID](fields, "market"); err != nil {
		return err
	}
	if result.AssetID, err = assetField(fields); err != nil {
		return err
	}
	if result.Bids, err = codec.Required[[]OrderBookLevel](fields, "bids"); err != nil {
		return err
	}
	if result.Asks, err = codec.Required[[]OrderBookLevel](fields, "asks"); err != nil {
		return err
	}
	if result.MinOrderSize, err = codec.DecimalField(fields, "min_order_size"); err != nil {
		return err
	}
	if result.TickSize, err = codec.DecimalField(fields, "tick_size"); err != nil {
		return err
	}
	if result.Hash, err = codec.Required[string](fields, "hash"); err != nil {
		return err
	}
	if result.NegRisk, err = codec.BoolField(fields, "neg_risk"); err != nil {
		return err
	}
	if raw, ok := fields["last_trade_price"]; ok && string(raw) != "null" && string(raw) != `""` {
		if result.LastTradePrice, err = codec.DecimalField(fields, "last_trade_price"); err != nil {
			return err
		}
	}
	if raw, ok := fields["timestamp"]; ok && string(raw) != "null" && string(raw) != `""` {
		value, parseErr := codec.Required[string](fields, "timestamp")
		if parseErr != nil {
			return parseErr
		}
		value = codec.ASCIIDigits(value)
		for _, digit := range value {
			if digit < '0' || digit > '9' {
				return fmt.Errorf("invalid epoch-ms timestamp")
			}
		}
		ms, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr != nil || ms > 253402300799999 {
			return fmt.Errorf("invalid epoch-ms timestamp")
		}
		// Match datetime.fromtimestamp(ms / 1000): Python first converts the
		// timestamp to binary seconds, then rounds to the nearest microsecond.
		seconds, fraction := math.Modf(float64(ms) / 1000)
		stamp := time.Unix(int64(seconds), int64(math.RoundToEven(fraction*1e6))*1000).UTC()
		result.Timestamp = &stamp
	}
	*b = result
	return nil
}

// UnmarshalJSON validates the price and side of a trade.
func (p *LastTradePrice) UnmarshalJSON(data []byte) error {
	fields, err := codec.Object(data)
	if err != nil {
		return err
	}
	var result LastTradePrice
	if result.Price, err = codec.DecimalField(fields, "price"); err != nil {
		return err
	}
	if result.Side, err = codec.Required[OrderSide](fields, "side"); err != nil {
		return err
	}
	if result.Side != Buy && result.Side != Sell {
		return fmt.Errorf("invalid trade side")
	}
	*p = result
	return nil
}

// UnmarshalJSON accepts either asset identifier spelling used by the exchange.
func (p *LastTradePriceForAsset) UnmarshalJSON(data []byte) error {
	fields, err := codec.Object(data)
	if err != nil {
		return err
	}
	var result LastTradePriceForAsset
	if result.AssetID, err = assetField(fields); err != nil {
		return err
	}
	if err = json.Unmarshal(data, &result.LastTradePrice); err != nil {
		return err
	}
	*p = result
	return nil
}
