// Package streams provides public real-time subscriptions and bounded event queues.
package streams

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/models"
)

// Spec selects a supported public stream. Values are copied when subscribed.
type Spec interface{ normalize() (selection, error) }

// MarketSpec selects asset updates and optional market lifecycle events.
type MarketSpec struct {
	AssetIDs             []string
	CustomFeatureEnabled bool
}

// SportsSpec receives all public sports results.
type SportsSpec struct{}

// CommentsSpec filters comment events. Nil Types defaults to comment_created.
type CommentsSpec struct {
	Types            []string
	ParentEntityID   *int64
	ParentEntityType *string
}

// CryptoPricesSpec selects the public Binance or Chainlink feed, including all-symbol mode.
// These feeds remain necessary for unauthenticated access and their original quote currencies.
type CryptoPricesSpec struct {
	Topic   string
	Symbols []string
}

// CryptoPricesChainlinkTwapSpec selects public 30- or 60-second Chainlink averages.
type CryptoPricesChainlinkTwapSpec struct {
	WindowSeconds int
	Symbols       []string
}

// EquityPricesSpec selects public equity updates and optional history snapshots.
type EquityPricesSpec struct {
	Symbol string
	Types  []string
}

type selection struct {
	service, topic, symbol string
	provider               string
	assets, symbols, types []string
	custom                 bool
	all                    bool
	parentID               *int64
	parentType             *string
	window                 int
}

func invalid(message string) error { return &sdkerrors.UserInputError{Message: message} }
func copyValues(values []string, required bool) ([]string, error) {
	if values == nil && !required {
		return nil, nil
	}
	if len(values) == 0 {
		return nil, invalid("subscription values must be nonempty")
	}
	if slices.Contains(values, "") {
		return nil, invalid("subscription values must not contain empty strings")
	}
	return slices.Clone(values), nil
}
func (p MarketSpec) normalize() (selection, error) {
	assets, err := copyValues(p.AssetIDs, true)
	return selection{service: "market", topic: "market", assets: assets, custom: p.CustomFeatureEnabled}, err
}
func (SportsSpec) normalize() (selection, error) {
	return selection{service: "sports", topic: "sports"}, nil
}
func (p CommentsSpec) normalize() (selection, error) {
	types := p.Types
	if types == nil {
		types = []string{"comment_created"}
	}
	types, err := copyValues(types, true)
	if err != nil {
		return selection{}, err
	}
	for _, kind := range types {
		if !slices.Contains([]string{"comment_created", "comment_removed", "reaction_created", "reaction_removed"}, kind) {
			return selection{}, invalid("invalid comment event type")
		}
	}
	var id *int64
	if p.ParentEntityID != nil {
		value := *p.ParentEntityID
		id = &value
	}
	var parent *string
	if p.ParentEntityType != nil {
		value := *p.ParentEntityType
		if value != "Event" && value != "Market" {
			return selection{}, invalid("comment parent must be Event or Market")
		}
		parent = &value
	}
	return selection{service: "rtds", topic: "comments", types: types, parentID: id, parentType: parent}, nil
}
func (p CryptoPricesSpec) normalize() (selection, error) {
	if p.Topic != "prices.crypto.binance" && p.Topic != "prices.crypto.chainlink" {
		return selection{}, invalid("invalid crypto price topic")
	}
	symbols, err := copyValues(p.Symbols, false)
	return selection{service: "rtds", topic: p.Topic, symbols: symbols}, err
}
func (p CryptoPricesChainlinkTwapSpec) normalize() (selection, error) {
	if p.WindowSeconds != 30 && p.WindowSeconds != 60 {
		return selection{}, invalid("window_seconds must be 30 or 60")
	}
	symbols, err := copyValues(p.Symbols, false)
	return selection{service: "rtds", topic: "prices.crypto.chainlink.twap", symbols: symbols, window: p.WindowSeconds}, err
}
func (p EquityPricesSpec) normalize() (selection, error) {
	if p.Symbol == "" {
		return selection{}, invalid("symbol must be nonempty")
	}
	types, err := copyValues(p.Types, false)
	if err != nil {
		return selection{}, err
	}
	for _, kind := range types {
		if kind != "update" && kind != "subscribe" {
			return selection{}, invalid("invalid equity event type")
		}
	}
	return selection{service: "rtds", topic: "prices.equity.pyth", symbol: strings.ToLower(p.Symbol), types: types}, nil
}
func (s selection) matches(event models.StreamEvent) bool {
	if event.StreamTopic() != s.topic {
		return false
	}
	contains := func(asset models.ClobAssetID) bool { return slices.Contains(s.assets, string(asset)) }
	switch value := event.(type) {
	case *models.UserOrderEvent:
		return s.userMatches(value.Payload.ConditionID)
	case *models.UserTradeEvent:
		return s.userMatches(value.Payload.ConditionID)
	case *models.MarketBookEvent:
		return contains(value.Payload.AssetID)
	case *models.MarketPriceChangeEvent:
		for _, change := range value.Payload.PriceChanges {
			if contains(change.AssetID) {
				return true
			}
		}
		return false
	case *models.MarketLastTradePriceEvent:
		return contains(value.Payload.AssetID)
	case *models.MarketTickSizeChangeEvent:
		return contains(value.Payload.AssetID)
	case *models.MarketBestBidAskEvent:
		return s.custom && contains(value.Payload.AssetID)
	case *models.NewMarketEvent:
		return s.custom
	case *models.MarketResolvedEvent:
		if !s.custom {
			return false
		}
		return slices.ContainsFunc(value.Payload.AssetIDs, contains)
	case *models.SportsResultEvent:
		return true
	case *models.CommentCreatedEvent:
		p := value.Payload
		parent := ""
		if p.ParentEntityID != nil {
			parent = *p.ParentEntityID
		}
		kind := ""
		if p.ParentEntityType != nil {
			kind = *p.ParentEntityType
		}
		return s.commentMatches(value.Type, parent, kind)
	case *models.CommentRemovedEvent:
		p := value.Payload
		parent := ""
		if p.ParentEntityID != nil {
			parent = strconv.FormatInt(*p.ParentEntityID, 10)
		}
		kind := ""
		if p.ParentEntityType != nil {
			kind = *p.ParentEntityType
		}
		return s.commentMatches(value.Type, parent, kind)
	case *models.ReactionCreatedEvent, *models.ReactionRemovedEvent:
		return s.parentID == nil && s.parentType == nil && slices.Contains(s.types, event.StreamType())
	case *models.CryptoPricesBinanceEvent:
		return s.symbols == nil || slices.Contains(s.symbols, value.Payload.Symbol)
	case *models.CryptoPricesChainlinkEvent:
		return s.symbols == nil || slices.Contains(s.symbols, value.Payload.Symbol)
	case *models.CryptoPricesChainlinkTwapEvent:
		return int64(s.window) == value.Payload.WindowSeconds && (s.symbols == nil || slices.Contains(s.symbols, value.Payload.Symbol))
	case *models.EquityPricesUpdateEvent:
		return s.equityMatches(value.Payload.Symbol, value.Type)
	case *models.EquityPricesSubscribeEvent:
		return s.equityMatches(value.Payload.Symbol, value.Type)
	default:
		return false
	}
}
func (s selection) commentMatches(kind, parent, parentType string) bool {
	return slices.Contains(s.types, kind) && (s.parentID == nil || strconv.FormatInt(*s.parentID, 10) == parent) && (s.parentType == nil || *s.parentType == parentType)
}
func (s selection) equityMatches(symbol, kind string) bool {
	return strings.ToLower(symbol) == s.symbol && (s.types == nil || slices.Contains(s.types, kind))
}
func normalizeSpecs(specs []Spec) ([]selection, error) {
	if len(specs) == 0 {
		return nil, invalid("provide at least one subscription")
	}
	result := make([]selection, 0, len(specs))
	for _, spec := range specs {
		if spec == nil || (reflect.ValueOf(spec).Kind() == reflect.Pointer && reflect.ValueOf(spec).IsNil()) {
			return nil, invalid("subscription must not be nil")
		}
		value, err := spec.normalize()
		if err != nil {
			return nil, fmt.Errorf("subscription: %w", err)
		}
		result = append(result, value)
	}
	return result, nil
}

// UserSpec selects account orders and trades. Empty Markets receives all account markets.
type UserSpec struct{ Markets []string }

func (p UserSpec) normalize() (selection, error) {
	values := append([]string(nil), p.Markets...)
	if slices.Contains(values, "") {
		return selection{}, invalid("market must be nonempty")
	}
	return selection{service: "user", topic: "user", assets: values, all: len(values) == 0}, nil
}
func (s selection) userMatches(condition models.ConditionID) bool {
	if s.all {
		return true
	}
	for _, market := range s.assets {
		if strings.EqualFold(market, string(condition)) {
			return true
		}
	}
	return false
}
