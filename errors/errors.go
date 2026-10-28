package errors

import (
	"fmt"
)

// UserInputError reports invalid arguments before a request is sent.
type UserInputError struct{ Message string }

func (e *UserInputError) Error() string { return e.Message }

// UnexpectedResponseError reports a response that violates the API contract.
type UnexpectedResponseError struct {
	Message string
	Cause   error
}

func (e *UnexpectedResponseError) Error() string { return e.Message }

func (e *UnexpectedResponseError) Unwrap() error { return e.Cause }

// TransportError wraps a network failure, including context cancellation.
type TransportError struct{ Cause error }

func (e *TransportError) Error() string { return fmt.Sprintf("polymarket request: %v", e.Cause) }

func (e *TransportError) Unwrap() error { return e.Cause }

// TradingRestriction describes a temporary exchange restriction.
type TradingRestriction string

// Trading restrictions reported by the exchange.
const (
	TradingRestarting TradingRestriction = "restarting"
	TradingPostOnly   TradingRestriction = "post_only"
	TradingCancelOnly TradingRestriction = "cancel_only"
)

// RequestRejectedError reports an unsuccessful HTTP response other than 429.
type RequestRejectedError struct {
	Message     string
	Status      int
	Code        string
	RetryAfter  *float64 // Seconds; nil means unavailable.
	Restriction TradingRestriction
}

func (e *RequestRejectedError) Error() string { return e.Message }

// RateLimitUpdate contains the exchange's optional rate-limit headers.
type RateLimitUpdate struct {
	Remaining *float64
	Reset     *float64
	Tier      string
	Warning   bool
}

// RateLimitError reports HTTP 429. The SDK does not retry automatically.
type RateLimitError struct {
	RetryAfter *float64
	RateLimit  *RateLimitUpdate
}

func (e *RateLimitError) Error() string { return "polymarket request was rate limited" }

// InsufficientLiquidityError reports an empty book or an unfillable FOK order.
type InsufficientLiquidityError struct{ Message string }

func (e *InsufficientLiquidityError) Error() string { return e.Message }

// PaginationLimitError means the server cannot establish whether more results exist.
type PaginationLimitError struct{ Message string }

func (e *PaginationLimitError) Error() string { return e.Message }
