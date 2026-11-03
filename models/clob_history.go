package models

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cagelin/polymarket-go/internal/codec"
)

// CLOBPriceHistoryPoint preserves CLOB t and p as integer/decimal text.
// These are sampled prices, not trades, order books, or Data API settlement points.
type CLOBPriceHistoryPoint struct {
	Timestamp json.Number `json:"t"`
	Price     json.Number `json:"p"`
}

// UnmarshalJSON rejects missing, noninteger timestamps and nonfinite prices.
func (p *CLOBPriceHistoryPoint) UnmarshalJSON(data []byte) error {
	fields, err := codec.Object(data)
	if err != nil {
		return err
	}
	timestamp, err := codec.Required[json.Number](fields, "t")
	if err != nil {
		return err
	}
	seconds, err := strconv.ParseInt(timestamp.String(), 10, 64)
	if err != nil || seconds < 0 || seconds > 253402300799 {
		return fmt.Errorf("invalid CLOB timestamp")
	}
	price, err := codec.Required[json.Number](fields, "p")
	if err != nil {
		return err
	}
	if _, err = codec.NumericDecimal(fields, "p", false); err != nil {
		return err
	}
	*p = CLOBPriceHistoryPoint{Timestamp: timestamp, Price: price}
	return nil
}
