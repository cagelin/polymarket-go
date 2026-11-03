package models

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/ericlagergren/decimal"
)

// PriceHistoryInterval selects a historical window relative to server time.
type PriceHistoryInterval string

// Supported history windows. Max and all follow the upstream server semantics.
const (
	HistoryMax      PriceHistoryInterval = "max"
	HistoryAll      PriceHistoryInterval = "all"
	HistoryMonth    PriceHistoryInterval = "1m"
	HistoryWeek     PriceHistoryInterval = "1w"
	HistoryDay      PriceHistoryInterval = "1d"
	HistorySixHours PriceHistoryInterval = "6h"
	HistoryHour     PriceHistoryInterval = "1h"
)

// PriceHistoryPoint is a price per share; resolution zero denotes an exact tick.
type PriceHistoryPoint struct {
	Timestamp         time.Time    `json:"timestamp"`
	Price             *decimal.Big `json:"price"`
	ResolutionSeconds int64        `json:"resolution_seconds"`
}

// UnmarshalJSON validates required prices, timestamps and resolutions.
func (p *PriceHistoryPoint) UnmarshalJSON(data []byte) error {
	fields, err := codec.Object(data)
	if err != nil {
		return err
	}
	raw, err := codec.Required[json.RawMessage](fields, "timestamp")
	if err != nil {
		return err
	}
	var value string
	if len(raw) > 0 && raw[0] == '"' {
		if err = json.Unmarshal(raw, &value); err != nil {
			return err
		}
	} else {
		value = string(raw)
	}
	timestamp, err := exactHistoryTime(value)
	if err != nil {
		return err
	}
	result := PriceHistoryPoint{Timestamp: timestamp}
	if result.Price, err = codec.NumericDecimal(fields, "price", false); err != nil {
		return err
	}
	if result.ResolutionSeconds, err = codec.IntegerField(fields, "resolution_seconds"); err != nil {
		return err
	}
	*p = result
	return nil
}

// Match the public microsecond resolution using integer half-even rounding.
func exactHistoryTime(value string) (time.Time, error) {
	var numeric json.Number
	if err := json.Unmarshal([]byte(value), &numeric); err != nil || numeric == "" || numeric.String() != value {
		return time.Time{}, fmt.Errorf("invalid epoch seconds")
	}
	if exponentIndex := strings.IndexAny(value, "eE"); exponentIndex >= 0 {
		exponent, exponentErr := strconv.ParseInt(value[exponentIndex+1:], 10, 64)
		if exponentErr != nil || exponent < -1000 || exponent > 1000 {
			return time.Time{}, fmt.Errorf("epoch exponent out of range")
		}
	}
	seconds, ok := new(big.Rat).SetString(value)
	if !ok || seconds.Cmp(big.NewRat(-62135596800, 1)) < 0 || seconds.Cmp(big.NewRat(253402300800, 1)) >= 0 {
		return time.Time{}, fmt.Errorf("invalid epoch seconds")
	}
	micros := new(big.Rat).Mul(seconds, big.NewRat(1000000, 1))
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(micros.Num(), micros.Denom(), remainder)
	twiceRemainder := new(big.Int).Lsh(new(big.Int).Abs(remainder), 1)
	comparison := twiceRemainder.Cmp(micros.Denom())
	if comparison > 0 || (comparison == 0 && quotient.Bit(0) == 1) {
		quotient.Add(quotient, big.NewInt(int64(micros.Sign())))
	}
	if !quotient.IsInt64() {
		return time.Time{}, fmt.Errorf("invalid epoch seconds")
	}
	timestamp := time.UnixMicro(quotient.Int64()).UTC()
	if timestamp.Year() < 1 || timestamp.Year() > 9999 {
		return time.Time{}, fmt.Errorf("invalid epoch seconds")
	}
	return timestamp, nil
}
