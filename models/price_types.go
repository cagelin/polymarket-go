package models

// EquityMarketStatus is the provider's reported market state.
type EquityMarketStatus string

// Known equity market states; a missing state is represented by nil.
const (
	EquityMarketOpen    EquityMarketStatus = "open"
	EquityMarketClosed  EquityMarketStatus = "closed"
	EquityMarketUnknown EquityMarketStatus = "unknown"
)
