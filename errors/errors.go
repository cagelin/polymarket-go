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

// SigningError reports invalid signing inputs or a cryptographic operation failure.
type SigningError struct{ Cause error }

func (e *SigningError) Error() string { return "polymarket signing failed" }
func (e *SigningError) Unwrap() error { return e.Cause }

// OrderHeartbeatMismatchError reports the heartbeat ID currently expected by the exchange.
type OrderHeartbeatMismatchError struct {
	HeartbeatID string
	Rejection   *RequestRejectedError
}

func (e *OrderHeartbeatMismatchError) Error() string { return e.Rejection.Error() }
func (e *OrderHeartbeatMismatchError) Unwrap() error { return e.Rejection }

// SettlementTimeoutError identifies fills that did not settle before the deadline.
type SettlementTimeoutError struct{ PendingTradeIDs []string }

func (e *SettlementTimeoutError) Error() string {
	return "order fills did not settle before the deadline"
}

// TransactionFailedError reports that all placement fills failed settlement.
type TransactionFailedError struct{ Message string }

func (e *TransactionFailedError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "all order fills failed settlement"
}

// RFQRequestRejectedError reports a builder gateway refusal and retains the original HTTP details.
type RFQRequestRejectedError struct{ Rejection *RequestRejectedError }

func (e *RFQRequestRejectedError) Error() string { return e.Rejection.Error() }
func (e *RFQRequestRejectedError) Unwrap() error { return e.Rejection }

// RFQTimeoutError reports a pending quote acceptance or fill; it does not imply failure.
type RFQTimeoutError struct{ RFQID string }

func (e *RFQTimeoutError) Error() string { return "RFQ outcome is still pending: " + e.RFQID }

// RPCError is a JSON-RPC server rejection returned with a successful HTTP response.
type RPCError struct {
	Method  string
	Code    int
	Message string
	Data    []byte
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("JSON-RPC %s failed (%d): %s", e.Method, e.Code, e.Message)
}

// TransactionTimeoutError retains the submitted identifiers for later reconciliation.
type TransactionTimeoutError struct {
	TransactionID   string
	TransactionHash string
}

func (e *TransactionTimeoutError) Error() string {
	return "transaction confirmation timed out; outcome remains unknown"
}
