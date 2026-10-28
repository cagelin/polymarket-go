package polymarket

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"
)

// gammaField describes only wire coercions; public models remain typed structs.
type gammaField struct {
	name     string
	aliases  []string
	kind     string
	required bool
}

func decodeGamma(data []byte, target any, rules []gammaField) error {
	fields, err := object(data)
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
	case "decimal":
		amount, err := numericDecimal(fields, rule.name, !rule.required)
		if err != nil {
			return nil, err
		}
		if amount == nil {
			return json.RawMessage("null"), nil
		}
		return json.Marshal(amount.String())
	case "bool":
		v, err := boolField(fields, rule.name)
		if err != nil {
			return nil, err
		}
		return json.Marshal(v)
	case "int":
		v, err := integerField(fields, rule.name)
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
		text, err := required[string](fields, rule.name)
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
		parsed, err := gammaTime(text)
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
		text, err := required[string](fields, rule.name)
		if err != nil {
			return nil, err
		}
		if text == "" && rule.kind != "condition" && rule.kind != "version" {
			return json.RawMessage("null"), nil
		}
		switch rule.kind {
		case "condition":
			if !validHex(text) {
				return nil, fmt.Errorf("invalid condition ID")
			}
		case "event_id":
			if len(text) != 66 || !validHex(text) || !strings.HasSuffix(text, "000000") {
				return nil, fmt.Errorf("invalid onchain event ID")
			}
		case "request_id":
			if len(text) != 66 || !validHex(text) || !strings.HasSuffix(text, "00") {
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

func validHex(value string) bool {
	if !strings.HasPrefix(value, "0x") || len(value) <= 2 {
		return false
	}
	// The upstream response validator accepts odd-length hex strings as well.
	digits := value[2:]
	if len(digits)%2 != 0 {
		digits = "0" + digits
	}
	_, err := hex.DecodeString(digits)
	return err == nil
}

func numericDecimal(fields map[string]json.RawMessage, key string, optional bool) (*apd.Decimal, error) {
	raw, ok := fields[key]
	if optional && (!ok || string(raw) == "null" || string(raw) == `""`) {
		return nil, nil
	}
	if !ok || string(raw) == "null" {
		return nil, fmt.Errorf("%s is required", key)
	}
	var text string
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &text); err != nil {
			return nil, err
		}
	} else {
		var number json.Number
		if err := json.Unmarshal(raw, &number); err != nil {
			return nil, err
		}
		text = number.String()
		if strings.ContainsAny(text, ".eE") {
			value, err := strconv.ParseFloat(text, 64)
			if err != nil {
				return nil, err
			}
			text = pythonFloat(value)
		}
	}
	encoded, err := json.Marshal(text)
	if err != nil {
		return nil, err
	}
	return decimalField(map[string]json.RawMessage{key: encoded}, key)
}

func integerField(fields map[string]json.RawMessage, key string) (int64, error) {
	raw, err := required[json.RawMessage](fields, key)
	if err != nil {
		return 0, err
	}
	value := string(raw)
	if value == "true" {
		return 1, nil
	}
	if value == "false" {
		return 0, nil
	}
	if raw[0] == '"' {
		if err = json.Unmarshal(raw, &value); err != nil {
			return 0, err
		}
	}
	value = asciiDigits(strings.ReplaceAll(strings.TrimSpace(value), "_", ""))
	if result, parseErr := strconv.ParseInt(value, 10, 64); parseErr == nil {
		return result, nil
	}
	d, _, err := apd.NewFromString(value)
	if err != nil || d.Form != apd.Finite {
		return 0, fmt.Errorf("invalid integer")
	}
	return d.Int64()
}

func gammaTime(value string) (time.Time, error) {
	value = strings.Replace(value, " ", "T", 1)
	if strings.HasSuffix(value, "+00") {
		value += ":00"
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04Z07:00", "2006-01-02T15:04:05.999999999", "2006-01-02T15:04", "2006-01-02", "20060102"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC().Truncate(time.Microsecond), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date-time %q", value)
}

func gammaSequence(raw json.RawMessage) ([]json.RawMessage, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []json.RawMessage{}, nil
	}
	if raw[0] == '"' {
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return nil, err
		}
		raw = json.RawMessage(text)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	if items == nil {
		return nil, fmt.Errorf("expected array")
	}
	return items, nil
}

func gammaStrings(raw json.RawMessage) ([]string, error) {
	items, err := gammaSequence(raw)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(items))
	for i, item := range items {
		if string(item) == "null" {
			return nil, fmt.Errorf("expected string")
		}
		if err = json.Unmarshal(item, &result[i]); err != nil {
			return nil, err
		}
	}
	return result, nil
}
