package chain

import (
	"strings"
	"testing"
)

func TestParsePositionID(t *testing.T) {
	for _, text := range []string{"0", "255", " 0xFF ", "0b11111111", "0o377"} {
		number, err := ParsePositionID(text)
		if err != nil || number == nil || number.Sign() < 0 {
			t.Errorf("ParsePositionID(%q) = %v, %v; want uint256", text, number, err)
		}
	}
	for _, text := range []string{"", "0x", "-1", "+1", "1_0", "0b_1", "0x" + strings.Repeat("f", 65)} {
		if number, err := ParsePositionID(text); err == nil {
			t.Errorf("ParsePositionID(%q) = %v, want error", text, number)
		}
	}
}
