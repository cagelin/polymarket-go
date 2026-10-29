package streams

import (
	"strings"

	"github.com/cagelin/polymarket-go/models"
)

type dispatchKey struct{ service, topic, value string }
type subscriber struct {
	handle    *Handle
	selection selection
}
type dispatchIndex map[dispatchKey][]*subscriber

// rebuildDispatch publishes an immutable index. The caller holds c.mu.
func (c *Client) rebuildDispatch() {
	index := make(dispatchIndex)
	for _, handle := range c.order {
		for _, selected := range c.handles[handle] {
			listener := &subscriber{handle: handle, selection: selected}
			values := selected.assets
			switch selected.service {
			case "user":
				values = make([]string, len(selected.assets))
				for position, asset := range selected.assets {
					values[position] = strings.ToLower(asset)
				}
			case "rtds":
				values = selected.symbols
				if selected.symbol != "" {
					values = []string{selected.symbol}
				}
			}
			// Wildcard selections also receive events without an asset, such as new markets.
			if len(values) == 0 || selected.custom {
				values = append(append([]string(nil), values...), "")
			}
			for _, value := range values {
				key := dispatchKey{service: selected.service, topic: selected.topic, value: value}
				index[key] = append(index[key], listener)
			}
		}
	}
	c.dispatch.Store(&index)
}
func eventValues(event models.StreamEvent) []string {
	switch payload := event.(type) {
	case *models.MarketBookEvent:
		return []string{string(payload.Payload.AssetID)}
	case *models.MarketLastTradePriceEvent:
		return []string{string(payload.Payload.AssetID)}
	case *models.MarketTickSizeChangeEvent:
		return []string{string(payload.Payload.AssetID)}
	case *models.MarketBestBidAskEvent:
		return []string{string(payload.Payload.AssetID)}
	case *models.MarketPriceChangeEvent:
		result := make([]string, len(payload.Payload.PriceChanges))
		for position, change := range payload.Payload.PriceChanges {
			result[position] = string(change.AssetID)
		}
		return result
	case *models.MarketResolvedEvent:
		result := make([]string, len(payload.Payload.AssetIDs))
		for position, asset := range payload.Payload.AssetIDs {
			result[position] = string(asset)
		}
		return result
	case *models.UserOrderEvent:
		return []string{strings.ToLower(string(payload.Payload.ConditionID))}
	case *models.UserTradeEvent:
		return []string{strings.ToLower(string(payload.Payload.ConditionID))}
	case *models.CryptoPricesBinanceEvent:
		return []string{payload.Payload.Symbol}
	case *models.CryptoPricesChainlinkEvent:
		return []string{payload.Payload.Symbol}
	case *models.CryptoPricesChainlinkTwapEvent:
		return []string{payload.Payload.Symbol}
	case *models.EquityPricesUpdateEvent:
		return []string{strings.ToLower(payload.Payload.Symbol)}
	case *models.EquityPricesSubscribeEvent:
		return []string{strings.ToLower(payload.Payload.Symbol)}
	default:
		return nil
	}
}
func (c *Client) dispatchEvent(service string, event models.StreamEvent) {
	index := c.dispatch.Load()
	if index == nil {
		return
	}
	seen := make(map[*subscriber]bool)
	values := append(eventValues(event), "")
	for _, value := range values {
		key := dispatchKey{service: service, topic: event.StreamTopic(), value: value}
		for _, listener := range (*index)[key] {
			if seen[listener] {
				continue
			}
			seen[listener] = true
			if listener.selection.matches(event) {
				listener.handle.push(event)
			}
		}
	}
}
