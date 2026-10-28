// Package polymarket provides access to Polymarket APIs.
//
// This initial implementation covers public CLOB prices and order books.
// Monetary values use apd.Decimal; no binary floating-point conversion is used.
// Network operations accept a context, and clients support concurrent use.
package polymarket

const (
	// UpstreamVersion is the Python SDK version used for compatibility tests.
	UpstreamVersion = "0.14.0"
	// UpstreamCommit identifies the exact Python SDK source revision.
	UpstreamCommit = "7032ed06175b0bbea77b2d9e898bd9f252af2041"
)
