package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cagelam/polymarket-go/internal/codec"
)

func clobTime(raw json.RawMessage, kind string) (json.RawMessage, error) {
	text := string(raw)
	if text == `""` {
		return json.RawMessage("null"), nil
	}
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &text); err != nil {
			return nil, err
		}
		digits := codec.ASCIIDigits(text)
		numeric := strings.Trim(digits, "0123456789") == ""
		if kind == "reward_epoch_ms" {
			numeric = strings.Trim(strings.TrimPrefix(digits, "-"), "0123456789") == ""
		}
		if kind == "stream_epoch_ms" && !numeric {
			return nil, fmt.Errorf("expected unsigned epoch-millisecond string")
		}
		if kind == "game_start_time" && numeric {
			kind = "reward_epoch_ms_unsigned"
		}
		if !numeric {
			parsed, err := codec.GammaTime(text)
			if err != nil {
				return nil, err
			}
			return json.Marshal(parsed)
		}
		text = digits
	}
	if kind == "stream_epoch_ms" && (raw[0] != '"' || strings.HasPrefix(text, "-")) {
		return nil, fmt.Errorf("expected unsigned epoch-millisecond string")
	}
	magnitude, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer epoch timestamp: %w", err)
	}
	seconds := float64(magnitude)
	if kind == "reward_epoch_ms" || kind == "reward_epoch_ms_unsigned" || kind == "stream_epoch_ms" || magnitude >= 100000000000 {
		seconds /= 1000
	}
	return dataTime(json.RawMessage(strconv.FormatFloat(seconds, 'g', -1, 64)), "epoch")
}

func normalizeComboMarket(raw []byte) ([]byte, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	if value, ok := fields["outcomes"]; ok && len(value) > 0 && value[0] == '{' {
		return raw, nil
	}
	names, err := codec.GammaStrings(fields["outcomes"])
	if err != nil {
		return nil, err
	}
	positions, err := codec.GammaStrings(fields["position_ids"])
	if err != nil {
		return nil, err
	}
	prices, err := codec.GammaSequence(fields["outcome_prices"])
	if err != nil {
		return nil, err
	}
	if len(names) != 2 || len(positions) != 2 || len(prices) != 2 {
		return nil, fmt.Errorf("expected two matching combo outcomes, positions and prices")
	}
	outcomes := map[string]any{}
	for i, side := range []string{"yes", "no"} {
		outcomes[side] = map[string]any{"label": names[i], "position_id": positions[i], "price": prices[i]}
	}
	fields["outcomes"], err = json.Marshal(outcomes)
	if err != nil {
		return nil, err
	}
	return json.Marshal(fields)
}
