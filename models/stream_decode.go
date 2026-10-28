package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/ericlagergren/decimal"
)

func normalizeStreamPayload(raw []byte, kind string) ([]byte, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "SportsGameResult":
		value := string(fields["finishedTimestamp"])
		fallback := fields["finished_timestamp"]
		if (value == "" || value == "null" || value == `""`) && len(fallback) > 0 && string(fallback) != "null" && string(fallback) != `""` {
			fields["finishedTimestamp"] = fallback
		}
	case "EquityPriceUpdatePayload":
		if value, ok := fields["full_accuracy_value"]; ok && string(value) != "null" {
			fields["value"] = value
		}
	case "CryptoPricesChainlinkTwapPayload":
		full, ok := fields["full_accuracy_value"]
		if !ok {
			if _, wire := fields["window_s"]; !wire {
				if _, canonical := fields["window_seconds"]; canonical {
					return raw, nil
				}
			}
			return nil, fmt.Errorf("full_accuracy_value is required")
		}
		if _, err = codec.NumericDecimal(fields, "value", false); err != nil {
			return nil, err
		}
		var integer string
		if err = json.Unmarshal(full, &integer); err != nil {
			return nil, err
		}
		digits := strings.TrimPrefix(integer, "-")
		if digits == "" || strings.Trim(digits, "0123456789") != "" {
			return nil, fmt.Errorf("full_accuracy_value must be a signed integer string")
		}
		amount, ok := new(decimal.Big).SetString(integer)
		if !ok || !amount.IsFinite() {
			return nil, fmt.Errorf("invalid full accuracy value")
		}
		amount.SetScale(amount.Scale() + 18)
		amount.Context.Precision = decimal.UnlimitedPrecision
		amount.Reduce()
		if amount.Sign() == 0 {
			amount.SetSignbit(false)
		}
		fields["value"], err = json.Marshal(amount.String())
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(fields)
}
