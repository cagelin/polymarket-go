package models

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/ericlagergren/decimal"
)

// BalanceAllowance expresses collateral and token quantities in integer base units.
type BalanceAllowance struct {
	Balance    *big.Int            `json:"balance"`
	Allowances map[string]*big.Int `json:"allowances"`
}

// UnmarshalJSON validates integer balances and allowances.
func (m *BalanceAllowance) UnmarshalJSON(raw []byte) error {
	fields, err := codec.Object(raw)
	if err != nil {
		return err
	}
	balance, err := baseUnits(fields["balance"])
	if err != nil {
		return err
	}
	rows, err := codec.Object(fields["allowances"])
	if err != nil {
		return err
	}
	values := make(map[string]*big.Int, len(rows))
	for key, value := range rows {
		parsed, parseErr := baseUnits(value)
		if parseErr != nil {
			return parseErr
		}
		values[key] = parsed
	}
	*m = BalanceAllowance{Balance: balance, Allowances: values}
	return nil
}
func baseUnits(raw json.RawMessage) (*big.Int, error) {
	text := string(raw)
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &text); err != nil {
			return nil, err
		}
		text = codec.ASCIIDigits(strings.TrimSpace(text))
	}
	// Python int permits underscores only between digits.
	digits := strings.TrimPrefix(strings.TrimPrefix(text, "-"), "+")
	if digits == "" || strings.HasPrefix(digits, "_") || strings.HasSuffix(digits, "_") || strings.Contains(digits, "__") || strings.Trim(digits, "0123456789_") != "" {
		return nil, fmt.Errorf("invalid base-unit integer")
	}
	value, ok := new(big.Int).SetString(strings.ReplaceAll(text, "_", ""), 10)
	if !ok {
		return nil, fmt.Errorf("invalid base-unit integer")
	}
	return value, nil
}
func normalizeBuilderFees(raw []byte) ([]byte, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	for _, key := range []string{"builder_maker_fee_rate_bps", "builder_taker_fee_rate_bps"} {
		value, ok := fields[key]
		if !ok || string(value) == "null" || string(value) == "true" || string(value) == "false" {
			continue
		}
		amount, parseErr := codec.NumericDecimal(map[string]json.RawMessage{key: value}, key, false)
		if parseErr != nil {
			return nil, parseErr
		}
		preferredScale := amount.Scale()
		originalZero := amount.Sign() == 0
		amount.SetScale(amount.Scale() + 4)
		if err = codec.RoundDecimal(amount); err != nil {
			return nil, err
		}
		if amount.Sign() != 0 {
			amount.Context.Precision = decimal.UnlimitedPrecision
			amount.Reduce()
		} else if originalZero {
			amount.SetScale(min(max(preferredScale, -999999), 1000026))
		}
		preferredScale = min(preferredScale, 1000026, 28-amount.Precision()+amount.Scale())
		if amount.Scale() < preferredScale {
			ctx := codec.DecimalContext()
			ctx.Quantize(amount, preferredScale)
		}
		fields[key], err = json.Marshal(amount.String())
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(fields)
}
