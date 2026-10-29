package models

import (
	"encoding/json"
	"os"
	"testing"
)

// BenchmarkParseMarketEvent uses a fixed Python-validated market frame.
func BenchmarkParseMarketEvent(b *testing.B) {
	raw, err := os.ReadFile("../testdata/streams-events-parity.json")
	if err != nil {
		b.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Service  string
			Raw      json.RawMessage
			Expected struct {
				Error bool
			}
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		b.Fatal(err)
	}
	var frame json.RawMessage
	for _, item := range fixture.Cases {
		if item.Service == "market" && !item.Expected.Error {
			frame = item.Raw
			break
		}
	}
	if frame == nil {
		b.Fatal("missing valid market event fixture")
	}
	b.SetBytes(int64(len(frame)))
	b.ReportAllocs()
	for b.Loop() {
		if _, err = ParseStreamEvent("market", frame); err != nil {
			b.Fatal(err)
		}
	}
}
