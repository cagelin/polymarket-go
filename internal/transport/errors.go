package transport

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
)

func numericHeader(value string) *float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || math.IsInf(v, 0) || math.IsNaN(v) {
		return nil
	}
	return &v
}

func parseRateLimit(headers http.Header) *sdkerrors.RateLimitUpdate {
	u := &sdkerrors.RateLimitUpdate{
		Remaining: numericHeader(headers.Get("Poly-RateLimit-Remaining")),
		Reset:     numericHeader(headers.Get("Poly-RateLimit-Reset")),
		Tier:      strings.TrimSpace(headers.Get("Poly-RateLimit-Tier")),
		Warning:   strings.EqualFold(strings.TrimSpace(headers.Get("Poly-RateLimit-Warning")), "true"),
	}
	if u.Remaining == nil && u.Reset == nil && u.Tier == "" && !u.Warning {
		return nil
	}
	return u
}

func retryAfter(headers http.Header, body map[string]any) *float64 {
	header := strings.TrimSpace(headers.Get("Retry-After"))
	if seconds := numericHeader(header); seconds != nil && *seconds >= 0 {
		return seconds
	}
	if deadline, err := http.ParseTime(header); err == nil {
		seconds := math.Max(0, math.Ceil(time.Until(deadline).Seconds()))
		return &seconds
	}
	if seconds, ok := body["retry_after_seconds"].(float64); ok && seconds >= 0 && !math.IsInf(seconds, 0) && !math.IsNaN(seconds) {
		return &seconds
	}
	return nil
}

// RetryDelay applies a finite, representable Retry-After value to a minimum delay.
// A missing value preserves the minimum; invalid values return false.
func RetryDelay(minimum time.Duration, retryAfterSeconds *float64) (time.Duration, bool) {
	if retryAfterSeconds == nil {
		return minimum, true
	}
	seconds := *retryAfterSeconds
	if math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds >= float64(math.MaxInt64)/float64(time.Second) {
		return 0, false
	}
	return max(minimum, time.Duration(max(0, seconds)*float64(time.Second))), true
}

func responseError(res *http.Response, data []byte) error {
	contentType := strings.ToLower(res.Header.Get("Content-Type"))
	var body map[string]any
	if strings.Contains(contentType, "application/json") {
		_ = json.Unmarshal(data, &body)
	}
	after := retryAfter(res.Header, body)
	if res.StatusCode == http.StatusTooManyRequests {
		return &sdkerrors.RateLimitError{RetryAfter: after, RateLimit: parseRateLimit(res.Header)}
	}
	if res.StatusCode == 400 && body["error_msg"] == "Invalid Heartbeat ID" {
		if id, ok := body["heartbeat_id"].(string); ok {
			return &sdkerrors.OrderHeartbeatMismatchError{HeartbeatID: id, Rejection: &sdkerrors.RequestRejectedError{Message: "Invalid Heartbeat ID", Status: 400}}
		}
	}
	code, _ := body["code"].(string)
	message, _ := body["error"].(string)
	var restriction sdkerrors.TradingRestriction
	switch res.StatusCode {
	case 425:
		restriction = sdkerrors.TradingRestarting
	case 503:
		if code == "post_only_mode" {
			restriction = sdkerrors.TradingPostOnly
		} else if strings.Contains(message, "cancel-only") {
			restriction = sdkerrors.TradingCancelOnly
		}
	}
	if message == "" && strings.Contains(contentType, "text/plain") {
		message = strings.TrimSpace(string(data))
	}
	if message == "" {
		message = fmt.Sprintf("request failed with HTTP status %d", res.StatusCode)
	}
	return &sdkerrors.RequestRejectedError{Message: message, Status: res.StatusCode, Code: code, RetryAfter: after, Restriction: restriction}
}
