package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/cagelin/polymarket-go/internal/testutil"
)

func TestPythonStreamEvents(t *testing.T) {
	testStreamEvents(t, "../testdata/streams-events-parity.json")
}
func TestPythonSecureStreamEvents(t *testing.T) {
	testStreamEvents(t, "../testdata/secure-stream-events-parity.json")
}
func testStreamEvents(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path) //nolint:gosec // Paths are fixed repository fixtures passed only by the tests above.
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
		t.Run(test.Name, func(subtest *testing.T) {
			event, parseErr := ParseStreamEvent(test.Service, test.Raw)
			if (parseErr != nil) != test.Expected.Error {
				subtest.Fatalf("error = %v; Python error = %v", parseErr, test.Expected.Error)
			}
			if parseErr != nil {
				return
			}
			testutil.EqualJSON(subtest, event, test.Expected.Value)
		})
	}
}
