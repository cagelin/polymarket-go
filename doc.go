// Package polymarket provides access to Polymarket APIs.
//
// This implementation covers public queries, authentication, trading, wallets,
// pagination, and real-time streams. Wallet submissions return confirmation handles.
// Decimal fields use decimal.Big and retain their decimal-string precision.
// Network operations accept a context, and clients support concurrent use.
package polymarket

const (
	// UpstreamVersion is the Python SDK version used for compatibility tests.
	UpstreamVersion = "0.14.0"
	// UpstreamCommit identifies the exact Python SDK source revision.
	UpstreamCommit = "7032ed06175b0bbea77b2d9e898bd9f252af2041"
)
