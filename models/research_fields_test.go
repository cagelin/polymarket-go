package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestResearchEventFields(t *testing.T) {
	source := `{"id":"7","deployingTimestamp":"2026-11-02T09:00:00+08:00","markets":[{"id":"1","outcomes":"[\"Yes\",\"No\"]","createdAt":"2026-11-01T00:00:00Z","updatedAt":"invalid","acceptingOrdersTimestamp":"","eventStartTime":"2026-11-02T00:00:00+08:00","groupItemThreshold":"0002","umaEndDate":"2026-11-03T01:00:00Z","umaResolutionStatuses":"[\"proposed\",\"resolved\"]","automaticallyResolved":false,"resolutionSource":""},{"id":"2","outcomes":["A","B","C"],"extra":123}],"series":[{"id":"3","recurrence":"daily","seriesType":"single"}],"tags":[{"id":"4","forceShow":false,"forceHide":true}]}`
	var event Event
	if err := json.Unmarshal([]byte(source), &event); err != nil {
		t.Fatal(err)
	}
	if len(event.Markets) != 1 || len(event.SkippedMarkets) != 1 {
		t.Fatalf("Event markets/skipped = %d/%d, want 1/1", len(event.Markets), len(event.SkippedMarkets))
	}
	market := event.Markets[0]
	if market.CreatedAt == nil || market.CreatedAt.Time == nil || market.UpdatedAt == nil || market.UpdatedAt.Raw != "invalid" || market.UpdatedAt.Time != nil || market.UpdatedAt.ParseError == "" {
		t.Errorf("Market timestamps = %+v/%+v", market.CreatedAt, market.UpdatedAt)
	}
	if market.State.AcceptingOrdersTimestamp == nil || market.State.AcceptingOrdersTimestamp.Raw != "" || market.State.AcceptingOrdersTimestamp.ParseError == "" {
		t.Errorf("empty acceptingOrdersTimestamp lost: %+v", market.State.AcceptingOrdersTimestamp)
	}
	if market.State.EventStartTime == nil || market.State.EventStartTime.Time == nil || market.State.EventStartTime.Time.Format(time.RFC3339) != "2026-11-01T16:00:00Z" {
		t.Errorf("eventStartTime = %+v, want UTC", market.State.EventStartTime)
	}
	if market.GroupItemIndex == nil || *market.GroupItemIndex != "0002" {
		t.Errorf("groupItemThreshold = %v, want text 0002", market.GroupItemIndex)
	}
	if market.Resolution.UMAEndDate == nil || market.Resolution.UMAEndDate.Time == nil || len(market.Resolution.UMAResolutionStatuses) != 2 || market.Resolution.AutomaticallyResolved == nil || *market.Resolution.AutomaticallyResolved {
		t.Errorf("Market resolution = %+v", market.Resolution)
	}
	if market.Resolution.Source == nil || *market.Resolution.Source != "" {
		t.Errorf("empty resolutionSource = %v", market.Resolution.Source)
	}
	if event.Schedule.DeployingTimestamp == nil || event.Schedule.DeployingTimestamp.Time == nil || event.Schedule.DeployingTimestamp.Time.Hour() != 1 {
		t.Errorf("deployingTimestamp = %+v", event.Schedule.DeployingTimestamp)
	}
	if event.Series[0].Recurrence == nil || *event.Series[0].Recurrence != "daily" || event.Tags[0].ForceShow == nil || *event.Tags[0].ForceShow {
		t.Errorf("relation fields lost: %+v %+v", event.Series, event.Tags)
	}
	var skipped struct {
		ID    string
		Extra int
	}
	if err := json.Unmarshal(event.SkippedMarkets[0].Raw, &skipped); err != nil {
		t.Fatal(err)
	}
	if skipped.ID != "2" || skipped.Extra != 123 || event.SkippedMarkets[0].Reason == "" {
		t.Errorf("skipped market = %+v", event.SkippedMarkets[0])
	}
	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var roundtrip Event
	if err = json.Unmarshal(encoded, &roundtrip); err != nil {
		t.Fatal(err)
	}
	if roundtrip.Markets[0].UpdatedAt.Raw != "invalid" || len(roundtrip.SkippedMarkets) != 1 {
		t.Errorf("Event roundtrip lost research fields")
	}
	if err = json.Unmarshal([]byte(`{"id":"8","markets":[{"id":"9","outcomes":["Yes","No"],"updatedAt":null}]}`), &event); err != nil {
		t.Fatal(err)
	}
	if event.Schedule.DeployingTimestamp != nil || event.Markets[0].UpdatedAt != nil || event.Markets[0].Resolution.Source != nil {
		t.Errorf("absent/null fields should stay nil")
	}
}

func TestResolutionRawPrices(t *testing.T) {
	for _, sample := range []string{`"69"`, `""`, `null`, `"0"`, `"1000000000000000000"`} {
		t.Run(sample, func(subtest *testing.T) {
			raw := `{"status":"posed","extended_review":false,"was_disputed":false,"question_rules_updated":false,"last_updated_at":1700000000,"reproposed_price":` + sample + `}`
			var result Resolution
			if err := json.Unmarshal([]byte(raw), &result); err != nil {
				subtest.Fatal(err)
			}
			if string(result.RawPrices["reproposed_price"]) != sample {
				subtest.Errorf("raw reproposed_price = %s, want %s", result.RawPrices["reproposed_price"], sample)
			}
			if _, exists := result.RawPrices["price"]; exists {
				subtest.Error("absent price must stay absent")
			}
			encoded, encodeErr := json.Marshal(result)
			if encodeErr != nil {
				subtest.Fatal(encodeErr)
			}
			var roundtrip Resolution
			if decodeErr := json.Unmarshal(encoded, &roundtrip); decodeErr != nil {
				subtest.Fatal(decodeErr)
			}
			if string(roundtrip.RawPrices["reproposed_price"]) != sample {
				subtest.Error("resolution roundtrip changed raw value")
			}
			if _, exists := roundtrip.RawPrices["price"]; exists {
				subtest.Error("resolution roundtrip added absent price")
			}
			if sample == `"69"` && result.ReproposedPrice != nil {
				subtest.Error("sentinel interpretation changed")
			}
		})
	}
}

func TestExactHistoryTimestamp(t *testing.T) {
	for _, sample := range []struct{ input, want string }{
		{"1700000000.000001", "2023-11-14T22:13:20.000001Z"},
		{"1700000000.1234565", "2023-11-14T22:13:20.123456Z"},
		{"1700000000.1234575", "2023-11-14T22:13:20.123458Z"},
		{"-0.000001", "1969-12-31T23:59:59.999999Z"},
	} {
		timestamp, err := exactHistoryTime(sample.input)
		if err != nil {
			t.Errorf("exactHistoryTime(%s): %v", sample.input, err)
			continue
		}
		if timestamp.Format(time.RFC3339Nano) != sample.want {
			t.Errorf("exactHistoryTime(%s) = %s, %v; want %s", sample.input, timestamp.Format(time.RFC3339Nano), err, sample.want)
		}
	}
}
