package polymarket

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func numericHeader(value string) *float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || math.IsInf(v, 0) || math.IsNaN(v) {
		return nil
	}
	return &v
}

func parseRateLimit(headers http.Header) *RateLimitUpdate {
	u := &RateLimitUpdate{
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

func responseError(res *http.Response, data []byte) error {
	contentType := strings.ToLower(res.Header.Get("Content-Type"))
	var body map[string]any
	if strings.Contains(contentType, "application/json") {
		_ = json.Unmarshal(data, &body)
	}
	after := retryAfter(res.Header, body)
	if res.StatusCode == http.StatusTooManyRequests {
		return &RateLimitError{RetryAfter: after, RateLimit: parseRateLimit(res.Header)}
	}
	code, _ := body["code"].(string)
	message, _ := body["error"].(string)
	var restriction TradingRestriction
	switch res.StatusCode {
	case 425:
		restriction = TradingRestarting
	case 503:
		if code == "post_only_mode" {
			restriction = TradingPostOnly
		} else if strings.Contains(message, "cancel-only") {
			restriction = TradingCancelOnly
		}
	}
	if message == "" && strings.Contains(contentType, "text/plain") {
		message = strings.TrimSpace(string(data))
	}
	if message == "" {
		message = fmt.Sprintf("request failed with HTTP status %d", res.StatusCode)
	}
	return &RequestRejectedError{Message: message, Status: res.StatusCode, Code: code, RetryAfter: after, Restriction: restriction}
}
