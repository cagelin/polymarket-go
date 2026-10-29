package streams

import (
	"encoding/json"
	"time"

	"github.com/cagelin/polymarket-go/models"
)

type priceEnvelope struct {
	Topic     string    `json:"topic"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Seq       *int64    `json:"seq"`
	Dropped   *int64    `json:"dropped"`
	Payload   struct {
		Symbol       string            `json:"symbol"`
		Source       string            `json:"source"`
		Timestamp    time.Time         `json:"timestamp"`
		Value        json.RawMessage   `json:"value"`
		Data         []json.RawMessage `json:"data"`
		MarketStatus *string           `json:"market_status"`
	} `json:"payload"`
}

func priceDetails(event models.StreamEvent) (priceEnvelope, error) {
	raw, err := json.Marshal(event)
	if err != nil {
		return priceEnvelope{}, err
	}
	var value priceEnvelope
	err = json.Unmarshal(raw, &value)
	return value, err
}

// snapshotFor retains two minutes of history and carries equity market-state changes.
func snapshotFor(previous, event models.StreamEvent) models.StreamEvent {
	if event.StreamType() == "subscribe" {
		return event
	}
	current, err := priceDetails(event)
	if err != nil {
		return previous
	}
	var old priceEnvelope
	if previous != nil {
		old, err = priceDetails(previous)
		if err != nil {
			return previous
		}
	}
	var history []json.RawMessage
	if old.Payload.Source == current.Payload.Source && old.Type == "subscribe" {
		history = old.Payload.Data
	}
	if len(history) > 0 {
		var last struct {
			Timestamp time.Time `json:"timestamp"`
		}
		if json.Unmarshal(history[len(history)-1], &last) != nil {
			return previous
		}
		if last.Timestamp.After(current.Payload.Timestamp) {
			if current.Topic != "prices.equity" || current.Payload.MarketStatus == nil {
				return previous
			}
			old.Payload.MarketStatus = current.Payload.MarketStatus
			raw, encodeErr := json.Marshal(old)
			if encodeErr != nil {
				return previous
			}
			updated, parseErr := models.ParseStreamEvent("price_canonical", raw)
			if parseErr != nil {
				return previous
			}
			return updated
		}
	}
	points := make([]json.RawMessage, 0, len(history)+1)
	start := current.Payload.Timestamp.Add(-2 * time.Minute)
	for _, raw := range history {
		var point struct {
			Timestamp time.Time `json:"timestamp"`
		}
		if json.Unmarshal(raw, &point) == nil && point.Timestamp.After(start) && point.Timestamp.Before(current.Payload.Timestamp) {
			points = append(points, raw)
		}
	}
	point, err := json.Marshal(map[string]any{"timestamp": current.Payload.Timestamp, "value": current.Payload.Value})
	if err != nil {
		return previous
	}
	points = append(points, point)
	payload := map[string]any{"symbol": current.Payload.Symbol, "source": current.Payload.Source, "data": points}
	if current.Payload.MarketStatus != nil {
		payload["market_status"] = current.Payload.MarketStatus
	}
	raw, err := json.Marshal(map[string]any{"topic": current.Topic, "type": "subscribe", "timestamp": current.Timestamp, "seq": current.Seq, "dropped": current.Dropped, "payload": payload})
	if err != nil {
		return previous
	}
	result, err := models.ParseStreamEvent("price_canonical", raw)
	if err != nil {
		return previous
	}
	return result
}
