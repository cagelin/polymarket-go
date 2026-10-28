package models

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cagelam/polymarket-go/internal/codec"
)

// gammaField describes only wire coercions; public models remain typed structs.
type gammaField struct {
	name     string
	aliases  []string
	kind     string
	required bool
}

func decodeGamma(data []byte, target any, rules []gammaField) error {
	fields, err := codec.Object(data)
	if err != nil {
		return err
	}
	canonical := make(map[string]json.RawMessage, len(rules))
	for _, rule := range rules {
		value, exists := fields[rule.name]
		for _, alias := range rule.aliases {
			if aliasValue, ok := fields[alias]; ok {
				value, exists = aliasValue, true
				break
			}
		}
		if !exists && rule.kind == "default_sequence" {
			canonical[rule.name] = json.RawMessage("[]")
			continue
		}
		if exists && string(value) == "null" && rule.kind == "default_sequence" {
			return fmt.Errorf("%s must be an array", rule.name)
		}
		if !exists || string(value) == "null" {
			if rule.required {
				return fmt.Errorf("%s is required", rule.name)
			}
			continue
		}
		normalized, normalizeErr := normalizeGammaField(value, rule)
		if normalizeErr != nil {
			return fmt.Errorf("%s: %w", rule.name, normalizeErr)
		}
		canonical[rule.name] = normalized
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, target)
}

func normalizeGammaField(value json.RawMessage, rule gammaField) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{rule.name: value}
	switch rule.kind {
	case "sequence", "default_sequence":
		return codec.DecodeGammaSequence(value)
	case "team_ordering":
		text, err := codec.Required[string](fields, rule.name)
		if err != nil {
			return nil, err
		}
		if text != "home" && text != "away" {
			return nil, fmt.Errorf("invalid team ordering")
		}
		return value, nil
	case "optional_date":
		if string(value) == `""` {
			return json.RawMessage("null"), nil
		}
		text, err := codec.Required[string](fields, rule.name)
		if err != nil {
			return nil, err
		}
		// Preserve the calendar date in the supplied offset, not its UTC date.
		if len(text) >= 10 && text[4] == '-' && text[7] == '-' {
			if _, parseErr := codec.GammaTime(text); parseErr != nil {
				return nil, parseErr
			}
			return json.Marshal(text[:10])
		}
		parsed, err := codec.GammaTime(text)
		if err != nil {
			return nil, err
		}
		return json.Marshal(parsed.Format(time.DateOnly))
	case "epoch_datetime":
		return codec.GammaEpochTime(value)
	case "decimal":
		amount, err := codec.NumericDecimal(fields, rule.name, !rule.required)
		if err != nil {
			return nil, err
		}
		if amount == nil {
			return json.RawMessage("null"), nil
		}
		return json.Marshal(amount.String())
	case "bool":
		v, err := codec.BoolField(fields, rule.name)
		if err != nil {
			return nil, err
		}
		return json.Marshal(v)
	case "int":
		v, err := codec.IntegerField(fields, rule.name)
		if err != nil {
			return nil, err
		}
		return json.Marshal(v)
	case "number", "float":
		text := string(value)
		if value[0] == '"' {
			if err := json.Unmarshal(value, &text); err != nil {
				return nil, err
			}
		}
		if text == "true" {
			text = "1"
		}
		if text == "false" {
			text = "0"
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
		if err != nil || math.IsInf(v, 0) || math.IsNaN(v) {
			return nil, fmt.Errorf("invalid number")
		}
		if rule.kind == "number" {
			if integer, intErr := strconv.ParseInt(text, 10, 64); intErr == nil {
				return json.RawMessage(strconv.FormatInt(integer, 10)), nil
			}
		}
		return json.Marshal(v)
	case "datetime", "date":
		if string(value) == `""` && !rule.required && rule.kind == "datetime" {
			return json.RawMessage("null"), nil
		}
		text, err := codec.Required[string](fields, rule.name)
		if err != nil {
			return nil, err
		}
		if rule.kind == "date" {
			parsed, parseErr := time.Parse(time.DateOnly, text)
			if parseErr != nil {
				return nil, parseErr
			}
			return json.Marshal(parsed.Format(time.DateOnly))
		}
		parsed, err := codec.GammaTime(text)
		if err != nil {
			return nil, err
		}
		return json.Marshal(parsed)
	case "id":
		if value[0] == '"' {
			return value, nil
		}
		text := string(value)
		if strings.ContainsAny(text, ".eE") || text == "true" || text == "false" {
			return nil, fmt.Errorf("expected string or integer identifier")
		}
		var number json.Number
		if err := json.Unmarshal(value, &number); err != nil {
			return nil, err
		}
		return json.Marshal(text)
	case "condition", "event_id", "request_id", "version", "resolution", "uma", "empty_text":
		text, err := codec.Required[string](fields, rule.name)
		if err != nil {
			return nil, err
		}
		if text == "" && rule.kind != "condition" && rule.kind != "version" {
			return json.RawMessage("null"), nil
		}
		switch rule.kind {
		case "condition":
			if !codec.ValidHex(text) {
				return nil, fmt.Errorf("invalid condition ID")
			}
		case "event_id":
			if len(text) != 66 || !codec.ValidHex(text) || !strings.HasSuffix(text, "000000") {
				return nil, fmt.Errorf("invalid onchain event ID")
			}
		case "request_id":
			if len(text) != 66 || !codec.ValidHex(text) || !strings.HasSuffix(text, "00") {
				return nil, fmt.Errorf("invalid request ID")
			}
		case "version":
			if text != "v1" && text != "v2" {
				return nil, fmt.Errorf("invalid protocol version")
			}
		case "resolution":
			if text != "inactive" && text != "active" && text != "resolved" {
				return nil, fmt.Errorf("invalid resolution status")
			}
		case "uma":
			if text != "disputed" && text != "proposed" && text != "requested" && text != "resolved" && text != "settled" {
				return nil, fmt.Errorf("invalid UMA status")
			}
		}
		return value, nil
	default:
		return value, nil
	}
}
