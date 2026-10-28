// Package testutil contains shared assertions for SDK parity tests.
package testutil

import (
	"encoding/json"
	"reflect"
	"testing"
)

// EqualJSON compares a Go value with a Python JSON fixture.
func EqualJSON(t *testing.T, value any, expected json.RawMessage) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var got, want any
	if err = json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(expected, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("JSON = %s; Python = %s", encoded, expected)
	}
}
