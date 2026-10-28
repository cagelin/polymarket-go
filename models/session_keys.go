package models

import "time"

// SessionKeyScope supports known venues and future names without an SDK update.
type SessionKeyScope string

// Known session-key scopes; unknown names remain supported.
const (
	SessionScopeAll       SessionKeyScope = "ALL"
	SessionScopeCLOB      SessionKeyScope = "CLOB"
	SessionScopeCombosRFQ SessionKeyScope = "COMBOSRFQ"
)

// AuthorizedSessionKey describes an active externally managed signing key.
type AuthorizedSessionKey struct {
	Address    EVMAddress
	Scopes     []SessionKeyScope
	ValidUntil time.Time
}

// AuthorizeSessionKeyParams contains the public address only. Nil Scopes defaults to ALL.
type AuthorizeSessionKeyParams struct {
	Address        EVMAddress
	Scopes         []SessionKeyScope
	IdempotencyKey *string
}

// RevokeSessionKeyParams identifies an existing session key.
type RevokeSessionKeyParams struct {
	Address        EVMAddress
	IdempotencyKey *string
}

// AuthorizeSessionKeyResult requires both chain confirmation and registry visibility.
type AuthorizeSessionKeyResult struct {
	SessionKey  AuthorizedSessionKey
	Transaction TransactionOutcome
}
