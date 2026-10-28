package streams

import (
	"regexp"
	"slices"
	"strings"
)

// CryptoPriceSpec selects authenticated USD crypto prices and history.
type CryptoPriceSpec struct {
	Symbols  []string
	Provider string
}

// CryptoTwapPriceSpec selects authenticated sixty-second crypto averages.
type CryptoTwapPriceSpec struct {
	Symbols  []string
	Provider string
}

// EquityPriceSpec selects an equity or FX symbol. Empty Types receives updates and history.
type EquityPriceSpec struct {
	Symbol   string
	Types    []string
	Provider string
}

// EquityTwapPriceSpec selects authenticated sixty-second equity or FX averages.
type EquityTwapPriceSpec struct {
	Symbol   string
	Provider string
}

var cryptoSymbol = regexp.MustCompile(`^[a-z0-9]+usd$`)
var equitySymbol = regexp.MustCompile(`^[a-zA-Z0-9._:/-]+$`)

func priceSelection(topic string, symbols []string, provider string, types []string) (selection, error) {
	if provider != "" && !slices.Contains([]string{"pyth", "chainlink"}, provider) {
		return selection{}, invalid("unknown price provider")
	}
	if len(symbols) == 0 {
		return selection{}, invalid("price symbols must be nonempty")
	}
	values := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		if strings.HasPrefix(topic, "prices.equity") {
			symbol = strings.ToLower(strings.TrimSpace(symbol))
			if len(symbol) < 1 || len(symbol) > 64 || !equitySymbol.MatchString(symbol) {
				return selection{}, invalid("invalid equity symbol")
			}
		} else if len(symbol) > 64 || !cryptoSymbol.MatchString(symbol) {
			return selection{}, invalid("crypto symbols must be canonical lowercase USD pairs")
		}
		if !slices.Contains(values, symbol) {
			values = append(values, symbol)
		}
	}
	for _, kind := range types {
		if kind != "update" && kind != "subscribe" {
			return selection{}, invalid("price event type must be update or subscribe")
		}
	}
	return selection{service: "price", topic: topic, symbols: values, provider: provider, types: slices.Clone(types)}, nil
}
func (p CryptoPriceSpec) normalize() (selection, error) {
	return priceSelection("prices.crypto", p.Symbols, p.Provider, nil)
}
func (p CryptoTwapPriceSpec) normalize() (selection, error) {
	return priceSelection("prices.crypto.twap", p.Symbols, p.Provider, nil)
}
func (p EquityPriceSpec) normalize() (selection, error) {
	return priceSelection("prices.equity", []string{p.Symbol}, p.Provider, p.Types)
}
func (p EquityTwapPriceSpec) normalize() (selection, error) {
	return priceSelection("prices.equity.twap", []string{p.Symbol}, p.Provider, nil)
}
