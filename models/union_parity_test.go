package models

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/cagelin/polymarket-go/internal/testutil"
)

func TestPythonResponseVariants(t *testing.T) {
	raw, err := os.ReadFile("../testdata/unions-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, Union string
			Response    json.RawMessage
			Expected    struct {
				Value          json.RawMessage
				Variant, Error string
			}
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(subtest *testing.T) {
			var value any
			var parseErr error
			switch test.Union {
			case "Activity":
				value, parseErr = ParseActivity(test.Response)
			case "ComboActivity":
				value, parseErr = ParseComboActivity(test.Response)
			case "BiggestWinner":
				value, parseErr = ParseBiggestWinner(test.Response)
			default:
				subtest.Fatal(test.Union)
			}
			if (parseErr != nil) != (test.Expected.Error != "") {
				subtest.Fatalf("error = %v; Python = %s", parseErr, test.Expected.Error)
			}
			if parseErr != nil {
				return
			}
			if fmt.Sprintf("%T", value) != "*models."+test.Expected.Variant {
				subtest.Fatalf("variant = %T; Python = %s", value, test.Expected.Variant)
			}
			testutil.EqualJSON(subtest, value, test.Expected.Value)
		})
	}
}
