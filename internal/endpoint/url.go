// Package endpoint validates service endpoints before credentials are used.
package endpoint

import (
	"net/url"
	"slices"
	"strings"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
)

// Parse accepts an absolute endpoint without credentials, query, or fragment.
func Parse(raw string, schemes ...string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || !slices.Contains(schemes, parsed.Scheme) || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || strings.Contains(raw, "#") || parsed.Opaque != "" {
		return "", &sdkerrors.UserInputError{Message: "service URL must be absolute without credentials, query or fragment"}
	}
	return strings.TrimRight(raw, "/"), nil
}
