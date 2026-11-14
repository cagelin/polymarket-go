package models

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cagelin/polymarket-go/internal/codec"
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
			valid := slices.Contains(rule.allowed, text)
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
	if len(value) == 0 {
		return nil, fmt.Errorf("empty JSON value")
	}
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
	case "nonempty_string":
		text, parseErr := codec.Required[string](fields, rule.name)
		if parseErr != nil || text == "" {
			return nil, fmt.Errorf("expected a nonempty string")
		}
		return value, nil
	case "notification_kind":
		return dataNotificationKind(value, rule)
	case "notification_id":
		return dataNotificationID(value)
	case "e6_decimal":
		return dataBaseUnitDecimal(value, rule)
	case "string_decimal":
		amount, parseErr := codec.DecimalField(fields, rule.name)
		if parseErr != nil {
			return nil, parseErr
		}
		return json.Marshal(amount.String())
	case rewardEpochMillis, rewardEpochMillisUnsigned, "clob_epoch_iso", "game_start_time", streamEpochMillis:
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
		return dataAssetReference(value, rule)
	case "condition_reference":
		return dataConditionReference(value, rule)
	case "event_reference":
		return dataEventReference(value)
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
		return dataBooleanLiteral(value, rule)
	case "payout_pair":
		return dataPayoutPair(value)
	default:
		return normalizeGammaField(value, gammaField{name: rule.name, kind: rule.kind, required: rule.required})
	}
	return value, nil
}

func dataStep(value json.RawMessage, step string) (json.RawMessage, error) {
	raw := string(value)
	if dataStepNull(raw, step) {
		return json.RawMessage("null"), nil
	}
	switch step {
	case "empty_zero":
		if raw == `""` {
			return json.RawMessage(`"0"`), nil
		}
		return value, nil
	case "trade_status":
		var status string
		if err := json.Unmarshal(value, &status); err != nil {
			return nil, err
		}
		return json.Marshal(strings.TrimPrefix(status, "TRADE_STATUS_"))

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
	case "price_window":
		if raw != "60" && raw != "60.0" {
			return nil, fmt.Errorf("expected TWAP window 60")
		}
	case "twap_window":
		if raw != "30" && raw != "60" && raw != "30.0" && raw != "60.0" {
			return nil, fmt.Errorf("expected TWAP window 30 or 60")
		}
	case "unset_price":
		return dataUnsetPrice(value)
	case "optional_event_id":
		return optionalDataEventID(value)
	case "required_event_id":
		return requiredDataEventID(value)
	case "validate_condition_id_response", "validate_optional_condition_id_response", "validate_combo_condition_id":
		return dataConditionID(value, step)
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

func dataNotificationKind(value json.RawMessage, rule dataField) (json.RawMessage, error) {
	text := string(value)
	if text == "true" {
		text = "1"
	}
	if text == "false" {
		text = "0"
	}
	number, parseErr := strconv.ParseFloat(text, 64)
	if parseErr != nil || number != float64(rule.minimum) {
		return nil, fmt.Errorf("invalid notification type")
	}
	return json.Marshal(rule.minimum)
}

func dataNotificationID(value json.RawMessage) (json.RawMessage, error) {
	var err error
	text := string(value)
	if value[0] == '"' {
		if err = json.Unmarshal(value, &text); err != nil {
			return nil, err
		}
	}
	text = codec.ASCIIDigits(text)
	digits := strings.TrimPrefix(text, "-")
	if digits == "" || strings.Trim(digits, "0123456789") != "" {
		return nil, fmt.Errorf("invalid notification ID")
	}
	integer, parseErr := strconv.ParseInt(text, 10, 64)
	if parseErr != nil {
		return nil, parseErr
	}
	return json.Marshal(integer)
}

func dataBaseUnitDecimal(value json.RawMessage, rule dataField) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{rule.name: value}
	var err error
	text, textErr := codec.Required[string](fields, rule.name)
	if textErr != nil {
		return nil, textErr
	}
	digits := codec.ASCIIDigits(text)
	if digits == "" || strings.Trim(digits, "0123456789") != "" {
		return nil, fmt.Errorf("invalid base-unit integer string")
	}
	fields[rule.name], err = json.Marshal(digits)
	if err != nil {
		return nil, err
	}
	amount, parseErr := codec.DecimalField(fields, rule.name)
	if parseErr != nil {
		return nil, parseErr
	}
	scaled := new(decimal.Big)
	if parseErr = codec.DecimalArithmetic("mul", scaled, amount, decimal.New(1, 6)); parseErr != nil {
		return nil, parseErr
	}
	return json.Marshal(scaled.String())
}

func dataEventReference(value json.RawMessage) (json.RawMessage, error) {
	var err error
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
}

func dataBooleanLiteral(value json.RawMessage, rule dataField) (json.RawMessage, error) {
	var err error
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
}

func dataPayoutPair(value json.RawMessage) (json.RawMessage, error) {
	var err error
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
		scale := min(amount.Scale(), 1000026, 28-result[i].Precision()+result[i].Scale())
		if result[i].Scale() < scale && scale-result[i].Scale() <= 28-result[i].Precision() {
			ctx.Quantize(result[i], scale)
			if err = result[i].Context.Err(); err != nil {
				return nil, err
			}
		}
	}
	return json.Marshal(result)
}

func optionalDataEventID(value json.RawMessage) (json.RawMessage, error) {
	raw := string(value)
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
}

func requiredDataEventID(value json.RawMessage) (json.RawMessage, error) {
	raw := string(value)
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
}

func dataConditionID(value json.RawMessage, step string) (json.RawMessage, error) {
	raw := string(value)
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

func dataUnsetPrice(value json.RawMessage) (json.RawMessage, error) {
	raw := string(value)
	if string(value) == `""` {
		return json.RawMessage("null"), nil
	}
	if raw == "69" || raw == "69.0" || raw == `"69"` {
		return json.RawMessage("null"), nil
	}
	return value, nil
}

func dataAssetReference(value json.RawMessage, rule dataField) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{rule.name: value}
	text, parseErr := codec.Required[string](fields, rule.name)
	if parseErr != nil || len(text) < 1 || len(text) > 78 || strings.Trim(text, "0123456789") != "" {
		return nil, fmt.Errorf("invalid asset ID")
	}
	return value, nil
}

func dataConditionReference(value json.RawMessage, rule dataField) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{rule.name: value}
	text, parseErr := codec.Required[string](fields, rule.name)
	if parseErr != nil || len(text) != 66 || !codec.ValidHex(text) {
		return nil, fmt.Errorf("invalid 32-byte condition ID")
	}
	return value, nil
}

func dataStepNull(raw, step string) bool {
	switch step {
	case "empty_none", "optional_text":
		return raw == `""`
	case "zero_time":
		return raw == "0" || raw == `"0"`
	case "optional_outcome_index":
		return raw == "999" || raw == "999.0"
	case "date_from_calendar_string":
		return raw == `""` || raw == `"1970-01-01"`
	case "zero_rank":
		return raw == "0" || raw == "0.0" || raw == "false"
	case "global_condition":
		return raw == `"GLOBAL"`
	default:
		return false
	}
}
