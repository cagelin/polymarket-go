package models

import (
	"encoding/json"
	"time"

	"github.com/cagelin/polymarket-go/internal/codec"
)

// TimestampValue preserves a Gamma timestamp string even when parsing fails.
// Time is UTC when parsing succeeds; ParseError explains an invalid or empty value.
// The original field's null or absence is represented by a nil *TimestampValue.
type TimestampValue struct {
	Raw        string     `json:"raw"`
	Time       *time.Time `json:"time"`
	ParseError string     `json:"parse_error,omitempty"`
}

// UnmarshalJSON accepts a wire string or the SDK's own structured representation.
func (v *TimestampValue) UnmarshalJSON(data []byte) error {
	var rawText string
	if len(data) > 0 && data[0] == '{' {
		var saved struct {
			Raw string `json:"raw"`
		}
		if err := json.Unmarshal(data, &saved); err != nil {
			return err
		}
		rawText = saved.Raw
	} else if err := json.Unmarshal(data, &rawText); err != nil {
		return err
	}
	result := TimestampValue{Raw: rawText}
	parsed, err := codec.GammaTime(rawText)
	if err != nil {
		result.ParseError = err.Error()
	} else {
		utc := parsed.UTC()
		result.Time = &utc
	}
	*v = result
	return nil
}

// SkippedMarket retains a child market excluded from the binary Market model.
// Raw is the child JSON value, not the complete original HTTP response body.
type SkippedMarket struct {
	Raw    json.RawMessage `json:"raw"`
	Reason string          `json:"reason"`
}

func resolutionRawPrices(fields map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	if saved, ok := fields["raw_prices"]; ok {
		var original map[string]json.RawMessage
		if err := json.Unmarshal(saved, &original); err != nil {
			return nil, err
		}
		return original, nil
	}
	values := make(map[string]json.RawMessage)
	for _, name := range []string{"proposed_price", "reproposed_price", "price"} {
		if value, ok := fields[name]; ok {
			values[name] = append(json.RawMessage(nil), value...)
		}
	}
	return values, nil
}
