package codec

import (
	"fmt"
	"strings"
)

// Condition validates and normalizes a condition according to the requested protocol grammar.
func Condition(value, grammar string) (string, error) {
	if !ValidHex(value) {
		return "", fmt.Errorf("expected a hex condition ID")
	}
	switch grammar {
	case "response":
		return value, nil
	case "feed":
		if len(value) == 64 || len(value) == 66 {
			return value, nil
		}
	case "market":
		value = strings.ToLower(value)
		if len(value) == 66 {
			return value, nil
		}
		if len(value) == 64 && (strings.HasPrefix(value, "0x01") || strings.HasPrefix(value, "0x02")) {
			return value + "00", nil
		}
	case "combo":
		value = strings.ToLower(value)
		if len(value) == 64 && strings.HasPrefix(value, "0x03") {
			return value, nil
		}
		if len(value) == 66 && strings.HasPrefix(value, "0x03") && (strings.HasSuffix(value, "00") || strings.HasSuffix(value, "01")) {
			return value[:64], nil
		}
	}
	return "", fmt.Errorf("condition ID does not match %s grammar", grammar)
}
