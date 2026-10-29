package codec

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/ericlagergren/decimal"
)

// DecodeResponse wraps response decoding failures in the SDK error type.
func DecodeResponse[T any](data []byte) (T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return result, &sdkerrors.UnexpectedResponseError{Message: "response did not match expected shape", Cause: err}
	}
	return result, nil
}

// ParseList decodes a required JSON array.
func ParseList[T any](data []byte) ([]T, error) {
	result, err := DecodeResponse[[]T](data)
	if err == nil && result == nil {
		err = &sdkerrors.UnexpectedResponseError{Message: "expected response array"}
	}
	return result, err
}

// Object decodes a required JSON object.
func Object(data []byte) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if fields == nil {
		return nil, fmt.Errorf("expected Object")
	}
	return fields, nil
}

// Required decodes a required nonnull field.
func Required[T any](fields map[string]json.RawMessage, key string) (T, error) {
	var result T
	data, ok := fields[key]
	if !ok || bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return result, fmt.Errorf("%s is Required", key)
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("%s: %w", key, err)
	}
	return result, nil
}

// DecimalField decodes a finite decimal string.
func DecimalField(fields map[string]json.RawMessage, key string) (*decimal.Big, error) {
	value, err := Required[string](fields, key)
	if err != nil {
		return nil, err
	}
	d, ok := new(decimal.Big).SetString(ASCIIDigits(strings.ReplaceAll(strings.TrimSpace(value), "_", "")))
	if !ok || !d.IsFinite() || d.Precision()-d.Scale()-1 > decimal.MaxScale || d.Scale() > 1999999999999999997 {
		return nil, fmt.Errorf("%s must be finite", key)
	}
	return d, nil
}

// ASCIIDigits maps Unicode decimal digits accepted by Python to ASCII.
func ASCIIDigits(value string) string {
	return strings.Map(func(r rune) rune {
		if r < 128 {
			return r
		}
		for _, span := range unicode.Nd.R16 {
			if uint32(r) >= uint32(span.Lo) && uint32(r) <= uint32(span.Hi) && (uint32(r)-uint32(span.Lo))%uint32(span.Stride) == 0 {
				return '0' + rune((uint32(r)-uint32(span.Lo))/uint32(span.Stride)%10)
			}
		}
		for _, span := range unicode.Nd.R32 {
			if uint32(r) >= span.Lo && uint32(r) <= span.Hi && (uint32(r)-span.Lo)%span.Stride == 0 {
				return '0' + rune((uint32(r)-span.Lo)/span.Stride%10)
			}
		}
		return r
	}, value)
}

// BoolField decodes Python-compatible boolean values.
func BoolField(fields map[string]json.RawMessage, key string) (bool, error) {
	value, err := Required[any](fields, key)
	if err != nil {
		return false, err
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case float64:
		if v == 0 || v == 1 {
			return v == 1, nil
		}
	case string:
		switch strings.ToLower(v) {
		case "0", "off", "f", "false", "n", "no":
			return false, nil
		case "1", "on", "t", "true", "y", "yes":
			return true, nil
		}
	}
	return false, fmt.Errorf("%s must be boolean", key)
}

// PythonFloat formats a float as Python does for query fingerprints.
func PythonFloat(value float64) string {
	if math.IsNaN(value) {
		return "NaN"
	}
	if math.IsInf(value, 1) {
		return "Infinity"
	}
	if math.IsInf(value, -1) {
		return "-Infinity"
	}
	format := byte('f')
	if magnitude := math.Abs(value); magnitude != 0 && (magnitude < 1e-4 || magnitude >= 1e16) {
		format = 'e'
	}
	s := strconv.FormatFloat(value, format, -1, 64)
	if !strings.ContainsAny(s, ".e") {
		s += ".0"
	}
	return s
}

// ValidHex checks a hexadecimal identifier.
func ValidHex(value string) bool {
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

// NumericDecimal decodes a string or numeric decimal field.
func NumericDecimal(fields map[string]json.RawMessage, key string, optional bool) (*decimal.Big, error) {
	raw, ok := fields[key]
	if optional && (!ok || string(raw) == "null" || string(raw) == `""`) {
		return nil, nil
	}
	if !ok || string(raw) == "null" {
		return nil, fmt.Errorf("%s is Required", key)
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
			text = PythonFloat(value)
		}
	}
	encoded, err := json.Marshal(text)
	if err != nil {
		return nil, err
	}
	return DecimalField(map[string]json.RawMessage{key: encoded}, key)
}

// IntegerField decodes a Python-compatible integer field.
func IntegerField(fields map[string]json.RawMessage, key string) (int64, error) {
	raw, err := Required[json.RawMessage](fields, key)
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
	value = ASCIIDigits(strings.ReplaceAll(strings.TrimSpace(value), "_", ""))
	if result, parseErr := strconv.ParseInt(value, 10, 64); parseErr == nil {
		return result, nil
	}
	d, ok := new(decimal.Big).SetString(value)
	if !ok || !d.IsFinite() || !d.IsInt() {
		return 0, fmt.Errorf("invalid integer")
	}
	result, ok := d.Int64()
	if !ok {
		return 0, fmt.Errorf("integer outside int64 range")
	}
	return result, nil
}

// GammaTime parses Gamma timestamp strings.
func GammaTime(value string) (time.Time, error) {
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

// GammaSequence decodes arrays including JSON-encoded relationship arrays.
func GammaSequence(raw json.RawMessage) ([]json.RawMessage, error) {
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

// GammaStrings validates a sequence of strings.
func GammaStrings(raw json.RawMessage) ([]string, error) {
	items, err := GammaSequence(raw)
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

// GammaEpochTime follows Pydantic datetime coercion for image metadata fields.
func GammaEpochTime(value json.RawMessage) (json.RawMessage, error) {
	text := string(value)
	if value[0] == '"' {
		if err := json.Unmarshal(value, &text); err != nil {
			return nil, err
		}
		if parsed, err := GammaTime(text); err == nil {
			return json.Marshal(parsed)
		}
	}
	seconds, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return nil, fmt.Errorf("invalid datetime")
	}
	if math.Abs(seconds) > 2e10 {
		seconds /= 1000
	}
	if seconds < -62135596800 || seconds >= 253402300800 {
		return nil, fmt.Errorf("datetime outside supported range")
	}
	whole, fraction := math.Modf(seconds)
	parsed := time.Unix(int64(whole), int64(math.Round(fraction*1e6))*1000).UTC()
	return json.Marshal(parsed)
}

// DecodeGammaSequence rejects JSON strings, which are accepted only by explicit
// relationship normalizers, and validates each element through its model.
func DecodeGammaSequence(value json.RawMessage) (json.RawMessage, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(value, &items); err != nil {
		return nil, err
	}
	if items == nil {
		return nil, fmt.Errorf("expected an array")
	}
	return value, nil
}
