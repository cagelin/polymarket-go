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
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds < -62135596800 || seconds >= 253402300800 {
		return fmt.Errorf("invalid epoch seconds")
	}
	whole, fraction := math.Modf(seconds)
	result := PriceHistoryPoint{Timestamp: time.Unix(int64(whole), int64(math.RoundToEven(fraction*1e6))*1000).UTC()}
	if result.Price, err = codec.NumericDecimal(fields, "price", false); err != nil {
		return err
	}
	if result.ResolutionSeconds, err = codec.IntegerField(fields, "resolution_seconds"); err != nil {
		return err
	}
	*p = result
	return nil
}
