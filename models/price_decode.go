package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cagelam/polymarket-go/internal/codec"
)

func priceTimestamp(fields map[string]json.RawMessage, key string) (json.RawMessage, error) {
	value, err := codec.Required[int64](fields, key)
	if err != nil {
		return nil, err
	}
	instant := time.UnixMilli(value).UTC()
	if instant.Year() < 1 || instant.Year() > 9999 {
		return nil, fmt.Errorf("price timestamp outside supported calendar")
	}
	return json.Marshal(instant)
}
func pricePoint(raw json.RawMessage) (map[string]json.RawMessage, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	approx := fields["value"]
	if len(approx) == 0 {
		return nil, fmt.Errorf("missing approximate price")
	}
	if approx[0] != '"' {
		var number json.Number
		if err = json.Unmarshal(approx, &number); err != nil {
			return nil, err
		}
		if number == "" {
			return nil, fmt.Errorf("invalid approximate price")
		}
		approx, err = json.Marshal(string(number))
		if err != nil {
			return nil, err
		}
	}
	if _, err = codec.DecimalField(map[string]json.RawMessage{"value": approx}, "value"); err != nil {
		return nil, err
	}
	if _, err = codec.DecimalField(fields, "full_accuracy_value"); err != nil {
		return nil, err
	}
	timestamp, err := priceTimestamp(fields, "timestamp")
	if err != nil {
		return nil, err
	}
	return map[string]json.RawMessage{"timestamp": timestamp, "value": fields["full_accuracy_value"]}, nil
}
func parsePriceEvent(raw []byte) (StreamEvent, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	version, err := codec.Required[int64](fields, "v")
	if err != nil || version != 1 {
		return nil, fmt.Errorf("unsupported price frame version")
	}
	channel, err := codec.Required[string](fields, "channel")
	if err != nil {
		return nil, err
	}
	topic := map[string]string{"price.crypto": "prices.crypto", "price.crypto.twap": "prices.crypto.twap", "price.equity": "prices.equity", "price.equity.twap": "prices.equity.twap"}[channel]
	if topic == "" {
		return nil, fmt.Errorf("unknown price channel")
	}
	for _, key := range []string{"seq", "dropped"} {
		if key == "dropped" && fields[key] == nil {
			continue
		}
		value, parseErr := codec.Required[int64](fields, key)
		if parseErr != nil || value < 0 {
			return nil, fmt.Errorf("invalid price sequence or drop count")
		}
	}
	timestamp, err := priceTimestamp(fields, "ts")
	if err != nil {
		return nil, err
	}
	snapshot := false
	if fields["snapshot"] != nil {
		snapshot, err = codec.Required[bool](fields, "snapshot")
		if err != nil {
			return nil, err
		}
	}
	payload, err := codec.Object(fields["payload"])
	if err != nil {
		return nil, err
	}
	symbol, err := codec.Required[string](payload, "symbol")
	if err != nil {
		return nil, err
	}
	symbolRaw, err := json.Marshal(strings.ToLower(symbol))
	if err != nil {
		return nil, err
	}
	if _, err = codec.Required[string](payload, "source"); err != nil {
		return nil, err
	}
	result := map[string]json.RawMessage{"symbol": symbolRaw, "source": payload["source"]}
	twap := strings.HasSuffix(topic, ".twap")
	if twap {
		window, parseErr := codec.Required[int64](payload, "window_seconds")
		if parseErr != nil || window != 60 {
			return nil, fmt.Errorf("invalid price averaging window")
		}
	}
	if strings.HasPrefix(topic, "prices.equity") && payload["market_status"] != nil {
		status, parseErr := codec.Required[string](payload, "market_status")
		if parseErr != nil || status != "open" && status != "closed" && status != "unknown" {
			return nil, fmt.Errorf("invalid equity market status")
		}
		result["market_status"] = payload["market_status"]
	}
	kind := "update"
	if snapshot {
		kind = "subscribe"
		points, parseErr := codec.ParseList[json.RawMessage](payload["data"])
		if parseErr != nil {
			return nil, parseErr
		}
		values := make([]map[string]json.RawMessage, 0, len(points))
		for _, point := range points {
			value, pointErr := pricePoint(point)
			if pointErr != nil {
				return nil, pointErr
			}
			values = append(values, value)
		}
		result["data"], err = json.Marshal(values)
		if err != nil {
			return nil, err
		}
	} else {
		point, parseErr := pricePoint(fields["payload"])
		if parseErr != nil {
			return nil, parseErr
		}
		result["value"], result["timestamp"] = point["value"], point["timestamp"]
		if !twap {
			if payload["received_at"] != nil {
				result["received_at"], err = priceTimestamp(payload, "received_at")
				if err != nil {
					return nil, err
				}
			}
			if payload["is_carried_forward"] != nil {
				if _, err = codec.Required[bool](payload, "is_carried_forward"); err != nil {
					return nil, err
				}
				result["is_carried_forward"] = payload["is_carried_forward"]
			}
		}
	}
	canonical := map[string]any{"topic": topic, "type": kind, "timestamp": timestamp, "seq": fields["seq"], "payload": result}
	if fields["dropped"] != nil {
		canonical["dropped"] = fields["dropped"]
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return nil, err
	}
	return ParseStreamEvent("price_canonical", encoded)
}
