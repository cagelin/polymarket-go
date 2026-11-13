package codec

import (
	"encoding/json"
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestNumericDecimalPreservesJSONPrecision(t *testing.T) {
	for _, literal := range []string{
		"12345678901234567890.123456789",
		"0.12345678901234567890123456789",
		"-0.12345678901234567890123456789",
		"1.2345678901234567890123456789e-100",
		"1e400", "1e-400", "9007199254740993", "0.0100",
	} {
		t.Run(literal, func(subtest *testing.T) {
			expected, valid := new(decimal.Big).SetString(literal)
			if !valid {
				subtest.Fatalf("invalid test literal %q", literal)
			}
			for _, raw := range []string{literal, `"` + literal + `"`} {
				got, err := NumericDecimal(map[string]json.RawMessage{"amount": json.RawMessage(raw)}, "amount", false)
				if err != nil {
					subtest.Fatalf("NumericDecimal(%s): %v", raw, err)
				}
				if got.Cmp(expected) != 0 {
					subtest.Errorf("NumericDecimal(%s)=%s; want %s", raw, got, expected)
				}
			}
		})
	}
}

func TestNumericDecimalRejectsInvalidNumbers(t *testing.T) {
	for _, literal := range []string{"true", "[]", "{}", "1.2.3", `"NaN"`, `"Infinity"`} {
		if _, err := NumericDecimal(map[string]json.RawMessage{"amount": json.RawMessage(literal)}, "amount", false); err == nil {
			t.Errorf("NumericDecimal(%s) succeeded", literal)
		}
	}
}
