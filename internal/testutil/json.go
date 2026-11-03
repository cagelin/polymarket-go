// Package testutil contains shared assertions for SDK parity tests.
package testutil

import (
	"encoding/json"
	"reflect"
	"testing"
)

// EqualJSON compares a Go value with a Python JSON fixture.
func EqualJSON(t *testing.T, value any, expected json.RawMessage, extensions ...string) {
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
	removeExtensions(got, want, extensions)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("JSON = %s; Python = %s", encoded, expected)
	}
}

// Compare pinned Python fields exactly; SDK additions have separate regression tests.
func removeExtensions(actual, expected any, extensions []string) {
	switch got := actual.(type) {
	case map[string]any:
		want, ok := expected.(map[string]any)
		if !ok {
			return
		}
		for _, key := range extensions {
			if _, exists := want[key]; !exists {
				delete(got, key)
			}
		}
		for key, value := range got {
			removeExtensions(value, want[key], extensions)
		}
	case []any:
		want, ok := expected.([]any)
		if !ok || len(got) != len(want) {
			return
		}
		for index, value := range got {
			removeExtensions(value, want[index], extensions)
		}
	}
}
