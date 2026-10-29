package chain

import (
	"math/big"
	"strings"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
)

// ParsePositionID accepts decimal or explicitly prefixed uint256 position IDs.
func ParsePositionID(id string) (*big.Int, error) {
	raw := strings.TrimSpace(id)
	base := 10
	if len(raw) > 2 {
		switch strings.ToLower(raw[:2]) {
		case "0x":
			base, raw = 16, raw[2:]
		case "0b":
			base, raw = 2, raw[2:]
		case "0o":
			base, raw = 8, raw[2:]
		}
	}
	value, ok := new(big.Int).SetString(raw, base)
	if !ok || raw == "" || strings.ContainsAny(raw, "_+- ") || value.Sign() < 0 || value.BitLen() > 256 {
		return nil, &sdkerrors.UserInputError{Message: "position ID must be a uint256"}
	}
	return value, nil
}
