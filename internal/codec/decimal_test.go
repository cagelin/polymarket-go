package codec

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestPythonDecimalArithmetic(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/decimal-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct{ Operation, X, Y, Value, Error string }
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, test := range fixture.Cases {
		t.Run(test.Operation+"/"+test.X+"/"+test.Y, func(t *testing.T) {
			x, ok := new(decimal.Big).SetString(test.X)
			if !ok {
				t.Fatal(test.X)
			}
			y, ok := new(decimal.Big).SetString(test.Y)
			if !ok {
				t.Fatal(test.Y)
			}
			result := decimal.WithContext(DecimalContext())
			arithmeticErr := DecimalArithmetic(test.Operation, result, x, y)
			if (arithmeticErr != nil) != (test.Error != "") {
				t.Fatalf("error = %v; Python = %s", arithmeticErr, test.Error)
			}
			if arithmeticErr == nil && result.String() != test.Value {
				t.Errorf("result = %s; Python = %s", result, test.Value)
			}
		})
	}
}
