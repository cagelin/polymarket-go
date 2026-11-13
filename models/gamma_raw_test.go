package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGammaRawFields(t *testing.T) {
	for _, sample := range []struct {
		name, wire string
		prices     *string
	}{
		{name: "string", wire: `" [ \"0.1000\", \"9e-1\" ] "`, prices: new(` [ "0.1000", "9e-1" ] `)},
		{name: "array", wire: `[ 0.1000, 9e-1 ]`, prices: new(`[ 0.1000, 9e-1 ]`)},
		{name: "null", wire: `null`},
	} {
		t.Run(sample.name, func(subtest *testing.T) {
			input := `{"id":"7","closedTime":"2026-11-14T08:30:00+08:00","markets":[{"id":"1","outcomes":["No","Yes"],"outcomePrices":` + sample.wire + `,"closedTime":"2026-11-14 00:30:00+00"}]}`
			var event Event
			if err := json.Unmarshal([]byte(input), &event); err != nil {
				subtest.Fatal(err)
			}
			for range 2 {
				market := event.Markets[0]
				if event.Schedule.ClosedTimeRaw == nil || *event.Schedule.ClosedTimeRaw != "2026-11-14T08:30:00+08:00" || market.State.ClosedTimeRaw == nil || *market.State.ClosedTimeRaw != "2026-11-14 00:30:00+00" {
					subtest.Fatalf("closedTime original lost: %+v / %+v", event.Schedule, market.State)
				}
				if event.Schedule.ClosedTime == nil || event.Schedule.ClosedTime.UTC().Format(time.RFC3339) != "2026-11-14T00:30:00Z" {
					subtest.Errorf("closedTime parsed = %v", event.Schedule.ClosedTime)
				}
				if sample.prices == nil {
					if market.OutcomePricesRaw != nil {
						subtest.Errorf("null prices = %v", market.OutcomePricesRaw)
					}
				} else if market.OutcomePricesRaw == nil || *market.OutcomePricesRaw != *sample.prices {
					subtest.Errorf("prices raw = %v, want %q", market.OutcomePricesRaw, *sample.prices)
				}
				if market.Outcomes.Yes.Label != "No" || market.Outcomes.No.Label != "Yes" {
					subtest.Errorf("original outcome order changed: %+v", market.Outcomes)
				}
				encoded, encodeErr := json.Marshal(event)
				if encodeErr != nil {
					subtest.Fatal(encodeErr)
				}
				if decodeErr := json.Unmarshal(encoded, &event); decodeErr != nil {
					subtest.Fatal(decodeErr)
				}
			}
		})
	}
}

func TestGammaRawFieldsAbsent(t *testing.T) {
	var event Event
	if err := json.Unmarshal([]byte(`{"id":"7","markets":[{"id":"1","outcomes":["Yes","No"]}]}`), &event); err != nil {
		t.Fatal(err)
	}
	if event.Schedule.ClosedTimeRaw != nil || event.Markets[0].State.ClosedTimeRaw != nil || event.Markets[0].OutcomePricesRaw != nil {
		t.Errorf("absent raw fields added: %+v", event)
	}
}
