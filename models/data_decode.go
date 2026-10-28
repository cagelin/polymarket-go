package models

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/ericlagergren/decimal"
)

type dataField struct {
	name               string
	aliases            []string
	kind               string
	required, nullable bool
	steps, allowed     []string
	fallback           json.RawMessage
	bounded            bool
	minimum, maximum   int64
}

func decodeData(raw []byte, target any, rules []dataField) error {
	fields, err := codec.Object(raw)
	if err != nil {
		return err
	}
	result := make(map[string]json.RawMessage, len(rules))
	for _, rule := range rules {
		value, present := fields[rule.name]
		for _, alias := range rule.aliases {
			if aliased, exists := fields[alias]; exists {
				value, present = aliased, true
				break
			}
		}
		if !present {
			if rule.required {
				return fmt.Errorf("%s is required", rule.name)
			}
			if rule.fallback != nil {
				result[rule.name] = rule.fallback
			}
			continue
		}
		if string(value) == "null" {
			if !rule.nullable {
				return fmt.Errorf("%s must not be null", rule.name)
			}
			result[rule.name] = value
			continue
		}
		normalized, normalizeErr := dataValue(value, rule)
		if normalizeErr != nil {
			return fmt.Errorf("%s: %w", rule.name, normalizeErr)
		}
		if string(normalized) == "null" {
			if !rule.nullable {
				return fmt.Errorf("%s must not be null", rule.name)
			}
			result[rule.name] = normalized
			continue
		}
		if len(rule.allowed) > 0 {
			var text string
			if err = json.Unmarshal(normalized, &text); err != nil {
				return err
			}
			valid := false
			for _, allowed := range rule.allowed {
				if text == allowed {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid %s", rule.name)
			}
		}
		if rule.bounded {
			var integer int64
			if err = json.Unmarshal(normalized, &integer); err != nil {
				return err
			}
			if integer < rule.minimum || integer > rule.maximum {
				return fmt.Errorf("%s is outside the allowed range", rule.name)
			}
		}
		result[rule.name] = normalized
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, target)
}

func dataValue(value json.RawMessage, rule dataField) (json.RawMessage, error) {
	var err error
	for _, step := range rule.steps {
		value, err = dataStep(value, step)
		if err != nil {
			return nil, err
		}
		if string(value) == "null" {
			return value, nil
		}
	}
	fields := map[string]json.RawMessage{rule.name: value}
	switch rule.kind {
	case "string_decimal":
		amount, parseErr := codec.DecimalField(fields, rule.name)
		if parseErr != nil {
			return nil, parseErr
		}
		return json.Marshal(amount.String())
	case "reward_epoch_ms", "reward_epoch_ms_unsigned", "clob_epoch_iso", "game_start_time", "stream_epoch_ms":
		return clobTime(value, rule.kind)
	case "data_decimal":
		amount, parseErr := codec.NumericDecimal(fields, rule.name, false)
		if parseErr != nil {
			return nil, parseErr
		}
		return json.Marshal(amount.String())
	case "epoch", "optional_epoch", "epoch_or_iso", "optional_epoch_or_iso":
		return dataTime(value, rule.kind)
	case "asset_reference":
		text, parseErr := codec.Required[string](fields, rule.name)
		if parseErr != nil || len(text) < 1 || len(text) > 78 || strings.Trim(text, "0123456789") != "" {
			return nil, fmt.Errorf("invalid asset ID")
		}
	case "condition_reference":
		text, parseErr := codec.Required[string](fields, rule.name)
		if parseErr != nil || len(text) != 66 || !codec.ValidHex(text) {
			return nil, fmt.Errorf("invalid 32-byte condition ID")
		}
	case "event_reference":
		text := string(value)
		if value[0] == '"' {
			if err = json.Unmarshal(value, &text); err != nil {
				return nil, err
			}
		}
		if text == "" || strings.Trim(strings.TrimPrefix(text, "-"), "0123456789") != "" {
			return nil, fmt.Errorf("invalid event ID")
		}
		number, parseErr := strconv.ParseInt(text, 10, 32)
		if parseErr != nil {
			return nil, parseErr
		}
		return json.Marshal(strconv.FormatInt(number, 10))
	case "strict_bool":
		var parsed bool
		if err = json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}
	case "strict_int":
		var parsed int64
		if err = json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}
	case "literal_bool_true", "literal_bool_false":
		// Python Literal[bool] accepts numerically equal 0/1, not string booleans.
		desired := rule.kind == "literal_bool_true"
		if string(value) == strconv.FormatBool(desired) {
			return value, nil
		}
		var number float64
		if err = json.Unmarshal(value, &number); err != nil || (desired && number != 1) || (!desired && number != 0) {
			return nil, fmt.Errorf("invalid boolean literal")
		}
		return json.Marshal(desired)
	case "payout_pair":
		var pair []json.RawMessage
		if err = json.Unmarshal(value, &pair); err != nil || len(pair) != 2 {
			return nil, fmt.Errorf("expected a payout pair")
		}
		result := make([]*decimal.Big, 2)
		ctx := codec.DecimalContext()
		for i, item := range pair {
			amount, parseErr := codec.NumericDecimal(map[string]json.RawMessage{"amount": item}, "amount", false)
			if parseErr != nil {
				return nil, parseErr
			}
			result[i] = new(decimal.Big).Copy(amount)
			// Division by 10^6 is exact before applying the arithmetic context.
			result[i].SetScale(amount.Scale() + 6)
			if err = codec.RoundDecimal(result[i]); err != nil {
				return nil, err
			}
			if result[i].Sign() != 0 {
				result[i].Context.Precision = decimal.UnlimitedPrecision
				result[i].Reduce()
			} else if amount.Sign() == 0 {
				result[i].SetScale(min(max(amount.Scale(), -999999), 1000026))
			}
			// Preserve Python's preferred quotient exponent when precision allows.
			scale := min(amount.Scale(), 1000026)
			if result[i].Scale() < scale && scale-result[i].Scale() <= 28-result[i].Precision() {
				ctx.Quantize(result[i], scale)
				if err = result[i].Context.Err(); err != nil {
					return nil, err
				}
			}
		}
		return json.Marshal(result)
	default:
		return normalizeGammaField(value, gammaField{name: rule.name, kind: rule.kind, required: rule.required})
	}
	return value, nil
}

func dataStep(value json.RawMessage, step string) (json.RawMessage, error) {
	raw := string(value)
	switch step {
	case "condition_feed":
		if err := json.Unmarshal(value, &raw); err != nil {
			return nil, err
		}
		parsed, err := codec.Condition(raw, "feed")
		if err != nil {
			return nil, err
		}
		return json.Marshal(parsed)
	case "uppercase":
		if err := json.Unmarshal(value, &raw); err != nil {
			return nil, err
		}
		return json.Marshal(strings.ToUpper(raw))
	case "twap_window":
		if raw != "30" && raw != "60" && raw != "30.0" && raw != "60.0" {
			return nil, fmt.Errorf("expected TWAP window 30 or 60")
		}
	case "optional_text":
		if raw == `""` {
			return json.RawMessage("null"), nil
		}
	case "optional_outcome_index":
		if raw == "999" || raw == "999.0" {
			return json.RawMessage("null"), nil
		}
	case "date_from_calendar_string":
		if raw == `""` || raw == `"1970-01-01"` {
			return json.RawMessage("null"), nil
		}
	case "zero_rank":
		if raw == "0" || raw == "0.0" || raw == "false" {
			return json.RawMessage("null"), nil
		}
	case "unset_price":
		if raw == "69" || raw == "69.0" || raw == `"69"` {
			return json.RawMessage("null"), nil
		}
	case "global_condition":
		if raw == `"GLOBAL"` {
			return json.RawMessage("null"), nil
		}
	case "optional_event_id":
		if raw == `""` || raw == `"0"` || raw == "0" || raw == "0.0" || raw == "false" {
			return json.RawMessage("null"), nil
		}
		if value[0] == '"' {
			return value, nil
		}
		if raw == "true" {
			raw = "True"
		}
		return json.Marshal(raw)
	case "required_event_id":
		if value[0] == '"' {
			if err := json.Unmarshal(value, &raw); err != nil {
				return nil, err
			}
		}
		digits := codec.ASCIIDigits(raw)
		if digits == "" || strings.Trim(digits, "0123456789") != "" || strings.TrimLeft(digits, "0") == "" {
			return nil, fmt.Errorf("expected positive event ID")
		}
		return json.Marshal(raw)
	case "validate_condition_id_response", "validate_optional_condition_id_response", "validate_combo_condition_id":
		if err := json.Unmarshal(value, &raw); err != nil {
			return nil, err
		}
		grammar := "response"
		if step == "validate_combo_condition_id" {
			grammar = "combo"
		}
		normalized, err := codec.Condition(raw, grammar)
		if err != nil {
			return nil, err
		}
		return json.Marshal(normalized)
	}
	return value, nil
}

func dataTime(raw json.RawMessage, kind string) (json.RawMessage, error) {
	text := string(raw)
	optional := strings.HasPrefix(kind, "optional_")
	if optional && (text == `""` || (kind == "optional_epoch" && (text == "0" || text == "0.0" || text == `"0"` || text == "false"))) {
		return json.RawMessage("null"), nil
	}
	if raw[0] == '"' {
		if err := json.Unmarshal(raw, &text); err != nil {
			return nil, err
		}
		if strings.Contains(kind, "or_iso") && (text == "" || strings.Trim(codec.ASCIIDigits(text), "0123456789") != "") {
			parsed, err := codec.GammaTime(text)
			if err != nil {
				return nil, err
			}
			return json.Marshal(parsed)
		}
	}
	seconds, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsInf(seconds, 0) || math.IsNaN(seconds) || seconds < -62135596800 || seconds >= 253402300800 {
		return nil, fmt.Errorf("invalid epoch seconds")
	}
	whole, fraction := math.Modf(seconds)
	parsed := time.Unix(int64(whole), int64(math.RoundToEven(fraction*1e6))*1000).UTC()
	return json.Marshal(parsed)
}
