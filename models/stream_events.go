package models

import (
	"encoding/json"
	"fmt"

	"github.com/cagelam/polymarket-go/internal/codec"
)

// StreamEvent is a typed event from a public subscription. Use a type switch for its payload.
type StreamEvent interface {
	StreamTopic() string
	StreamType() string
}

// StreamTopic identifies the source topic.
func (m *MarketBestBidAskEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *MarketBestBidAskEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *MarketBookEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *MarketBookEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *MarketLastTradePriceEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *MarketLastTradePriceEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *MarketPriceChangeEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *MarketPriceChangeEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *MarketResolvedEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *MarketResolvedEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *MarketTickSizeChangeEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *MarketTickSizeChangeEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *NewMarketEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *NewMarketEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *SportsResultEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *SportsResultEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *CommentCreatedEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *CommentCreatedEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *CommentRemovedEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *CommentRemovedEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *CryptoPricesBinanceEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *CryptoPricesBinanceEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *CryptoPricesChainlinkEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *CryptoPricesChainlinkEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *CryptoPricesChainlinkTwapEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *CryptoPricesChainlinkTwapEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *EquityPricesSubscribeEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *EquityPricesSubscribeEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *EquityPricesUpdateEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *EquityPricesUpdateEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *ReactionCreatedEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *ReactionCreatedEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *ReactionRemovedEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *ReactionRemovedEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *UserOrderEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *UserOrderEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *UserTradeEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *UserTradeEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *CryptoPriceSnapshotEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *CryptoPriceSnapshotEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *CryptoPriceUpdateEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *CryptoPriceUpdateEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *CryptoTwapPriceSnapshotEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *CryptoTwapPriceSnapshotEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *CryptoTwapPriceUpdateEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *CryptoTwapPriceUpdateEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *EquityPriceSnapshotEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *EquityPriceSnapshotEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *EquityPriceUpdateEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *EquityPriceUpdateEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *EquityTwapPriceSnapshotEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *EquityTwapPriceSnapshotEvent) StreamType() string { return m.Type }

// StreamTopic identifies the source topic.
func (m *EquityTwapPriceUpdateEvent) StreamTopic() string { return m.Topic }

// StreamType identifies the event variant.
func (m *EquityTwapPriceUpdateEvent) StreamType() string { return m.Type }

// ParseStreamEvent decodes market, sports or RTDS wire events.
func ParseStreamEvent(service string, raw []byte) (StreamEvent, error) {
	if service == "price" {
		return parsePriceEvent(raw)
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	switch service {
	case "market", "user":
		_, topic := fields["topic"]
		_, kind := fields["type"]
		_, payload := fields["payload"]
		if !topic || !kind || !payload {
			eventType := fields["event_type"]
			if service != "user" && (len(eventType) == 0 || string(eventType) == "null" || string(eventType) == `""`) {
				eventType = fields["type"]
			}
			delete(fields, "event_type")
			if service != "user" {
				delete(fields, "type")
			}
			delete(fields, "topic")
			body, encodeErr := json.Marshal(fields)
			if encodeErr != nil {
				return nil, encodeErr
			}
			fields = map[string]json.RawMessage{"topic": json.RawMessage(`"` + service + `"`), "type": eventType, "payload": body}
		}
	case "price_canonical":
	case "sports":
		fields = map[string]json.RawMessage{"topic": json.RawMessage(`"sports"`), "type": json.RawMessage(`"sport_result"`), "payload": raw}
	case "rtds":
		wire, parseErr := codec.Required[string](fields, "topic")
		if parseErr != nil {
			return nil, parseErr
		}
		api := map[string]string{"comments": "comments", "crypto_prices": "prices.crypto.binance", "crypto_prices_chainlink": "prices.crypto.chainlink", "crypto_prices_twap_thirty": "prices.crypto.chainlink.twap", "crypto_prices_twap_sixty": "prices.crypto.chainlink.twap", "equity_prices": "prices.equity.pyth"}[wire]
		if api == "" {
			return nil, fmt.Errorf("unknown RTDS topic %q", wire)
		}
		if wire == "crypto_prices_twap_thirty" || wire == "crypto_prices_twap_sixty" {
			payload, payloadErr := codec.Required[map[string]json.RawMessage](fields, "payload")
			if payloadErr != nil {
				return nil, payloadErr
			}
			expected := "30"
			if wire == "crypto_prices_twap_sixty" {
				expected = "60"
			}
			if string(payload["window_s"]) != expected {
				return nil, fmt.Errorf("TWAP window does not match topic")
			}
		}
		fields["topic"], err = json.Marshal(api)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown stream service %q", service)
	}
	topic, err := codec.Required[string](fields, "topic")
	if err != nil {
		return nil, err
	}
	kind, err := codec.Required[string](fields, "type")
	if err != nil {
		return nil, err
	}
	var event StreamEvent
	switch topic + ":" + kind {
	case "market:best_bid_ask":
		event = &MarketBestBidAskEvent{}
	case "market:book":
		event = &MarketBookEvent{}
	case "market:last_trade_price":
		event = &MarketLastTradePriceEvent{}
	case "market:price_change":
		event = &MarketPriceChangeEvent{}
	case "market:market_resolved":
		event = &MarketResolvedEvent{}
	case "market:tick_size_change":
		event = &MarketTickSizeChangeEvent{}
	case "market:new_market":
		event = &NewMarketEvent{}
	case "sports:sport_result":
		event = &SportsResultEvent{}
	case "comments:comment_created":
		event = &CommentCreatedEvent{}
	case "comments:comment_removed":
		event = &CommentRemovedEvent{}
	case "prices.crypto.binance:update":
		event = &CryptoPricesBinanceEvent{}
	case "prices.crypto.chainlink:update":
		event = &CryptoPricesChainlinkEvent{}
	case "prices.crypto.chainlink.twap:update":
		event = &CryptoPricesChainlinkTwapEvent{}
	case "prices.equity.pyth:subscribe":
		event = &EquityPricesSubscribeEvent{}
	case "prices.equity.pyth:update":
		event = &EquityPricesUpdateEvent{}
	case "comments:reaction_created":
		event = &ReactionCreatedEvent{}
	case "comments:reaction_removed":
		event = &ReactionRemovedEvent{}
	case "user:order":
		event = &UserOrderEvent{}
	case "user:trade":
		event = &UserTradeEvent{}
	case "prices.crypto:subscribe":
		event = &CryptoPriceSnapshotEvent{}
	case "prices.crypto:update":
		event = &CryptoPriceUpdateEvent{}
	case "prices.crypto.twap:subscribe":
		event = &CryptoTwapPriceSnapshotEvent{}
	case "prices.crypto.twap:update":
		event = &CryptoTwapPriceUpdateEvent{}
	case "prices.equity:subscribe":
		event = &EquityPriceSnapshotEvent{}
	case "prices.equity:update":
		event = &EquityPriceUpdateEvent{}
	case "prices.equity.twap:subscribe":
		event = &EquityTwapPriceSnapshotEvent{}
	case "prices.equity.twap:update":
		event = &EquityTwapPriceUpdateEvent{}
	default:
		return nil, fmt.Errorf("unknown stream event %s:%s", topic, kind)
	}
	raw, err = json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(raw, event); err != nil {
		return nil, err
	}
	return event, nil
}
