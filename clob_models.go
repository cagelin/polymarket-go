package polymarket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cockroachdb/apd/v3"
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

// AssetQuery selects one asset by its canonical identifier.
type AssetQuery struct {
	AssetID ClobAssetID
}

// Asset returns a query using the canonical asset identifier.
func Asset(id ClobAssetID) AssetQuery { return AssetQuery{AssetID: id} }

// AssetsQuery selects a nonempty list. A nil slice means omitted.
type AssetsQuery struct {
	AssetIDs []ClobAssetID
}

// PriceRequest selects an asset and order side for single or batch requests.
type PriceRequest struct {
	AssetQuery
	Side OrderSide
}

// OrderBookLevel is a price and available share quantity.
type OrderBookLevel struct {
	Price *apd.Decimal `json:"price"`
	Size  *apd.Decimal `json:"size"`
}

// OrderBook preserves the exchange's level ordering.
// Bids ascend and asks descend; the best level is at the end of each slice.
type OrderBook struct {
	ConditionID    ConditionID      `json:"condition_id"`
	AssetID        ClobAssetID      `json:"asset_id"`
	Timestamp      *time.Time       `json:"timestamp"`
	Bids           []OrderBookLevel `json:"bids"`
	Asks           []OrderBookLevel `json:"asks"`
	MinOrderSize   *apd.Decimal     `json:"min_order_size"`
	TickSize       *apd.Decimal     `json:"tick_size"`
	NegRisk        bool             `json:"neg_risk"`
	LastTradePrice *apd.Decimal     `json:"last_trade_price"`
	Hash           string           `json:"hash"`
}

// LastTradePrice describes the most recent trade.
type LastTradePrice struct {
	Price *apd.Decimal `json:"price"`
	Side  OrderSide    `json:"side"`
}

// LastTradePriceForAsset identifies an asset's most recent trade.
type LastTradePriceForAsset struct {
	AssetID ClobAssetID `json:"asset_id"`
	LastTradePrice
}

func object(data []byte) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if fields == nil {
		return nil, fmt.Errorf("expected object")
	}
	return fields, nil
}

func required[T any](fields map[string]json.RawMessage, key string) (T, error) {
	var result T
	data, ok := fields[key]
	if !ok || bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return result, fmt.Errorf("%s is required", key)
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("%s: %w", key, err)
	}
	return result, nil
}

func decimalField(fields map[string]json.RawMessage, key string) (*apd.Decimal, error) {
	value, err := required[string](fields, key)
	if err != nil {
		return nil, err
	}
	d, _, err := apd.NewFromString(asciiDigits(strings.ReplaceAll(strings.TrimSpace(value), "_", "")))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	if d.Form != apd.Finite {
		return nil, fmt.Errorf("%s must be finite", key)
	}
	return d, nil
}

func assetField(fields map[string]json.RawMessage) (ClobAssetID, error) {
	if _, ok := fields["asset_id"]; ok {
		return required[ClobAssetID](fields, "asset_id")
	}
	return required[ClobAssetID](fields, "token_id")
}

// Python Decimal and str.isdecimal accept Unicode decimal digits.
func asciiDigits(value string) string {
	return strings.Map(func(r rune) rune {
		if r < 128 {
			return r
		}
		for _, span := range unicode.Nd.R16 {
			if uint32(r) >= uint32(span.Lo) && uint32(r) <= uint32(span.Hi) && (uint32(r)-uint32(span.Lo))%uint32(span.Stride) == 0 {
				return '0' + rune((uint32(r)-uint32(span.Lo))/uint32(span.Stride)%10)
			}
		}
		for _, span := range unicode.Nd.R32 {
			if uint32(r) >= span.Lo && uint32(r) <= span.Hi && (uint32(r)-span.Lo)%span.Stride == 0 {
				return '0' + rune((uint32(r)-span.Lo)/span.Stride%10)
			}
		}
		return r
	}, value)
}

// UnmarshalJSON validates required fields and decimal strings.
func (l *OrderBookLevel) UnmarshalJSON(data []byte) error {
	fields, err := object(data)
	if err != nil {
		return err
	}
	var result OrderBookLevel
	if result.Price, err = decimalField(fields, "price"); err != nil {
		return err
	}
	if result.Size, err = decimalField(fields, "size"); err != nil {
		return err
	}
	*l = result
	return nil
}

// UnmarshalJSON maps exchange field names to canonical SDK fields.
func (b *OrderBook) UnmarshalJSON(data []byte) error {
	fields, err := object(data)
	if err != nil {
		return err
	}
	var result OrderBook
	if result.ConditionID, err = required[ConditionID](fields, "market"); err != nil {
		return err
	}
	if result.AssetID, err = assetField(fields); err != nil {
		return err
	}
	if result.Bids, err = required[[]OrderBookLevel](fields, "bids"); err != nil {
		return err
	}
	if result.Asks, err = required[[]OrderBookLevel](fields, "asks"); err != nil {
		return err
	}
	if result.MinOrderSize, err = decimalField(fields, "min_order_size"); err != nil {
		return err
	}
	if result.TickSize, err = decimalField(fields, "tick_size"); err != nil {
		return err
	}
	if result.Hash, err = required[string](fields, "hash"); err != nil {
		return err
	}
	if result.NegRisk, err = boolField(fields, "neg_risk"); err != nil {
		return err
	}
	if raw, ok := fields["last_trade_price"]; ok && string(raw) != "null" && string(raw) != `""` {
		if result.LastTradePrice, err = decimalField(fields, "last_trade_price"); err != nil {
			return err
		}
	}
	if raw, ok := fields["timestamp"]; ok && string(raw) != "null" && string(raw) != `""` {
		value, parseErr := required[string](fields, "timestamp")
		if parseErr != nil {
			return parseErr
		}
		value = asciiDigits(value)
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

func boolField(fields map[string]json.RawMessage, key string) (bool, error) {
	value, err := required[any](fields, key)
	if err != nil {
		return false, err
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case float64:
		if v == 0 || v == 1 {
			return v == 1, nil
		}
	case string:
		switch strings.ToLower(v) {
		case "0", "off", "f", "false", "n", "no":
			return false, nil
		case "1", "on", "t", "true", "y", "yes":
			return true, nil
		}
	}
	return false, fmt.Errorf("%s must be boolean", key)
}

// UnmarshalJSON validates the price and side of a trade.
func (p *LastTradePrice) UnmarshalJSON(data []byte) error {
	fields, err := object(data)
	if err != nil {
		return err
	}
	var result LastTradePrice
	if result.Price, err = decimalField(fields, "price"); err != nil {
		return err
	}
	if result.Side, err = required[OrderSide](fields, "side"); err != nil {
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
	fields, err := object(data)
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
