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
func (e *MarketBestBidAskEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *MarketBestBidAskEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *MarketBookEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *MarketBookEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *MarketLastTradePriceEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *MarketLastTradePriceEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *MarketPriceChangeEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *MarketPriceChangeEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *MarketResolvedEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *MarketResolvedEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *MarketTickSizeChangeEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *MarketTickSizeChangeEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *NewMarketEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *NewMarketEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *SportsResultEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *SportsResultEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *CommentCreatedEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *CommentCreatedEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *CommentRemovedEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *CommentRemovedEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *CryptoPricesBinanceEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *CryptoPricesBinanceEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *CryptoPricesChainlinkEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *CryptoPricesChainlinkEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *CryptoPricesChainlinkTwapEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *CryptoPricesChainlinkTwapEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *EquityPricesSubscribeEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *EquityPricesSubscribeEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *EquityPricesUpdateEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *EquityPricesUpdateEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *ReactionCreatedEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *ReactionCreatedEvent) StreamType() string { return e.Type }

// StreamTopic identifies the source topic.
func (e *ReactionRemovedEvent) StreamTopic() string { return e.Topic }

// StreamType identifies the event variant.
func (e *ReactionRemovedEvent) StreamType() string { return e.Type }

// ParseStreamEvent decodes market, sports or RTDS wire events.
func ParseStreamEvent(service string, raw []byte) (StreamEvent, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	switch service {
	case "market":
		_, topic := fields["topic"]
		_, kind := fields["type"]
		_, payload := fields["payload"]
		if !topic || !kind || !payload {
			eventType := fields["event_type"]
			if len(eventType) == 0 || string(eventType) == "null" || string(eventType) == `""` {
				eventType = fields["type"]
			}
			delete(fields, "event_type")
			delete(fields, "type")
			delete(fields, "topic")
			body, encodeErr := json.Marshal(fields)
			if encodeErr != nil {
				return nil, encodeErr
			}
			fields = map[string]json.RawMessage{"topic": json.RawMessage(`"market"`), "type": eventType, "payload": body}
		}
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
