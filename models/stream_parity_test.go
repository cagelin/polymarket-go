package models

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestPythonStreamEvents(t *testing.T) {
	raw, err := os.ReadFile("../testdata/streams-events-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, Service string
			Raw           json.RawMessage
			Expected      struct {
				Value json.RawMessage
				Error bool
			}
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(t *testing.T) {
			event, parseErr := ParseStreamEvent(test.Service, test.Raw)
			if (parseErr != nil) != test.Expected.Error {
				t.Fatalf("error = %v; Python error = %v", parseErr, test.Expected.Error)
			}
			if parseErr != nil {
				return
			}
			actual, encodeErr := json.Marshal(event)
			if encodeErr != nil {
				t.Fatal(encodeErr)
			}
			var got, want any
			if err = json.Unmarshal(actual, &got); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(test.Expected.Value, &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Go = %s; Python = %s", actual, test.Expected.Value)
			}
		})
	}
}
