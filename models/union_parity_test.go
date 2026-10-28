package models

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
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
		t.Run(test.Name, func(t *testing.T) {
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
				t.Fatal(test.Union)
			}
			if (parseErr != nil) != (test.Expected.Error != "") {
				t.Fatalf("error = %v; Python = %s", parseErr, test.Expected.Error)
			}
			if parseErr != nil {
				return
			}
			if fmt.Sprintf("%T", value) != "*models."+test.Expected.Variant {
				t.Fatalf("variant = %T; Python = %s", value, test.Expected.Variant)
			}
			encoded, marshalErr := json.Marshal(value)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			var got, want any
			if err = json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(test.Expected.Value, &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("JSON = %s; Python = %s", encoded, test.Expected.Value)
			}
		})
	}
}
