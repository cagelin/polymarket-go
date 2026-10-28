// Stream schemas ported from the pinned Python SDK. This file is linted.

package models

import (
	"encoding/json"
	"time"

	"github.com/ericlagergren/decimal"
)

// MarketBestBidAskEvent represents a public stream message or payload.
type MarketBestBidAskEvent struct {
	Topic   string                  `json:"topic"`
	Type    string                  `json:"type"`
	Payload MarketBestBidAskPayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketBestBidAskEvent) UnmarshalJSON(raw []byte) error {
	type plain MarketBestBidAskEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"market"}, fallback: json.RawMessage("\"market\"")},
		{name: "type", required: true, allowed: []string{"best_bid_ask"}},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = MarketBestBidAskEvent(result)
	return nil
}

// MarketBestBidAskPayload represents a public stream message or payload.
type MarketBestBidAskPayload struct {
	ConditionID ConditionID  `json:"condition_id"`
	AssetID     ClobAssetID  `json:"asset_id"`
	BestBid     *decimal.Big `json:"best_bid"`
	BestAsk     *decimal.Big `json:"best_ask"`
	Spread      *decimal.Big `json:"spread"`
	Timestamp   *time.Time   `json:"timestamp"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketBestBidAskPayload) UnmarshalJSON(raw []byte) error {
	type plain MarketBestBidAskPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "best_bid", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "best_ask", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "spread", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "timestamp", kind: "stream_epoch_ms", nullable: true},
	}); err != nil {
		return err
	}
	*m = MarketBestBidAskPayload(result)
	return nil
}

// MarketBookEvent represents a public stream message or payload.
type MarketBookEvent struct {
	Topic   string            `json:"topic"`
	Type    string            `json:"type"`
	Payload MarketBookPayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketBookEvent) UnmarshalJSON(raw []byte) error {
	type plain MarketBookEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"market"}, fallback: json.RawMessage("\"market\"")},
		{name: "type", required: true, allowed: []string{"book"}},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = MarketBookEvent(result)
	return nil
}

// MarketBookPayload represents a public stream message or payload.
type MarketBookPayload struct {
	ConditionID    ConditionID      `json:"condition_id"`
	AssetID        ClobAssetID      `json:"asset_id"`
	Bids           []OrderBookLevel `json:"bids"`
	Asks           []OrderBookLevel `json:"asks"`
	Hash           *string          `json:"hash"`
	Timestamp      *time.Time       `json:"timestamp"`
	MinOrderSize   *decimal.Big     `json:"min_order_size"`
	TickSize       *decimal.Big     `json:"tick_size"`
	NegRisk        *bool            `json:"neg_risk"`
	LastTradePrice *decimal.Big     `json:"last_trade_price"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketBookPayload) UnmarshalJSON(raw []byte) error {
	type plain MarketBookPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "bids", kind: "sequence", required: true},
		{name: "asks", kind: "sequence", required: true},
		{name: "hash", nullable: true},
		{name: "timestamp", kind: "stream_epoch_ms", nullable: true},
		{name: "min_order_size", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "tick_size", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "neg_risk", kind: "bool", nullable: true},
		{name: "last_trade_price", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
	}); err != nil {
		return err
	}
	*m = MarketBookPayload(result)
	return nil
}

// MarketEventMessage represents a public stream message or payload.
type MarketEventMessage struct {
	ID          string  `json:"id"`
	Ticker      *string `json:"ticker"`
	Slug        *string `json:"slug"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketEventMessage) UnmarshalJSON(raw []byte) error {
	type plain MarketEventMessage
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "ticker", nullable: true},
		{name: "slug", nullable: true},
		{name: "title", nullable: true},
		{name: "description", nullable: true},
	}); err != nil {
		return err
	}
	*m = MarketEventMessage(result)
	return nil
}

// MarketLastTradePriceEvent represents a public stream message or payload.
type MarketLastTradePriceEvent struct {
	Topic   string                      `json:"topic"`
	Type    string                      `json:"type"`
	Payload MarketLastTradePricePayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketLastTradePriceEvent) UnmarshalJSON(raw []byte) error {
	type plain MarketLastTradePriceEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"market"}, fallback: json.RawMessage("\"market\"")},
		{name: "type", required: true, allowed: []string{"last_trade_price"}},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = MarketLastTradePriceEvent(result)
	return nil
}

// MarketLastTradePricePayload represents a public stream message or payload.
type MarketLastTradePricePayload struct {
	ConditionID     ConditionID  `json:"condition_id"`
	AssetID         ClobAssetID  `json:"asset_id"`
	Price           *decimal.Big `json:"price"`
	Size            *decimal.Big `json:"size"`
	Side            OrderSide    `json:"side"`
	FeeRateBps      *decimal.Big `json:"fee_rate_bps"`
	TransactionHash *string      `json:"transaction_hash"`
	Timestamp       *time.Time   `json:"timestamp"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketLastTradePricePayload) UnmarshalJSON(raw []byte) error {
	type plain MarketLastTradePricePayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "price", kind: "data_decimal", required: true},
		{name: "size", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "side", steps: []string{"uppercase"}, required: true, allowed: []string{"BUY", "SELL"}},
		{name: "fee_rate_bps", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "transaction_hash", nullable: true},
		{name: "timestamp", kind: "stream_epoch_ms", nullable: true},
	}); err != nil {
		return err
	}
	*m = MarketLastTradePricePayload(result)
	return nil
}

// MarketPriceChangeEvent represents a public stream message or payload.
type MarketPriceChangeEvent struct {
	Topic   string                   `json:"topic"`
	Type    string                   `json:"type"`
	Payload MarketPriceChangePayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketPriceChangeEvent) UnmarshalJSON(raw []byte) error {
	type plain MarketPriceChangeEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"market"}, fallback: json.RawMessage("\"market\"")},
		{name: "type", required: true, allowed: []string{"price_change"}},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = MarketPriceChangeEvent(result)
	return nil
}

// MarketPriceChangePayload represents a public stream message or payload.
type MarketPriceChangePayload struct {
	ConditionID  ConditionID   `json:"condition_id"`
	PriceChanges []PriceChange `json:"price_changes"`
	Timestamp    *time.Time    `json:"timestamp"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketPriceChangePayload) UnmarshalJSON(raw []byte) error {
	type plain MarketPriceChangePayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "price_changes", kind: "sequence", required: true},
		{name: "timestamp", kind: "stream_epoch_ms", nullable: true},
	}); err != nil {
		return err
	}
	*m = MarketPriceChangePayload(result)
	return nil
}

// MarketResolvedEvent represents a public stream message or payload.
type MarketResolvedEvent struct {
	Topic   string                `json:"topic"`
	Type    string                `json:"type"`
	Payload MarketResolvedPayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketResolvedEvent) UnmarshalJSON(raw []byte) error {
	type plain MarketResolvedEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"market"}, fallback: json.RawMessage("\"market\"")},
		{name: "type", required: true, allowed: []string{"market_resolved"}},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = MarketResolvedEvent(result)
	return nil
}

// MarketResolvedPayload represents a public stream message or payload.
type MarketResolvedPayload struct {
	ID             string              `json:"id"`
	ConditionID    ConditionID         `json:"condition_id"`
	AssetIDs       []ClobAssetID       `json:"asset_ids"`
	WinningAssetID *ClobAssetID        `json:"winning_asset_id"`
	WinningOutcome *string             `json:"winning_outcome"`
	EventMessage   *MarketEventMessage `json:"event_message"`
	Timestamp      *time.Time          `json:"timestamp"`
	Tags           []string            `json:"tags"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketResolvedPayload) UnmarshalJSON(raw []byte) error {
	type plain MarketResolvedPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "asset_ids", aliases: []string{"asset_ids", "assets_ids", "token_ids"}, kind: "sequence", nullable: true},
		{name: "winning_asset_id", aliases: []string{"winning_asset_id", "winning_token_id"}, nullable: true},
		{name: "winning_outcome", nullable: true},
		{name: "event_message", nullable: true},
		{name: "timestamp", kind: "stream_epoch_ms", nullable: true},
		{name: "tags", kind: "sequence", nullable: true},
	}); err != nil {
		return err
	}
	*m = MarketResolvedPayload(result)
	return nil
}

// MarketTickSizeChangeEvent represents a public stream message or payload.
type MarketTickSizeChangeEvent struct {
	Topic   string                      `json:"topic"`
	Type    string                      `json:"type"`
	Payload MarketTickSizeChangePayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketTickSizeChangeEvent) UnmarshalJSON(raw []byte) error {
	type plain MarketTickSizeChangeEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"market"}, fallback: json.RawMessage("\"market\"")},
		{name: "type", required: true, allowed: []string{"tick_size_change"}},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = MarketTickSizeChangeEvent(result)
	return nil
}

// MarketTickSizeChangePayload represents a public stream message or payload.
type MarketTickSizeChangePayload struct {
	ConditionID ConditionID  `json:"condition_id"`
	AssetID     ClobAssetID  `json:"asset_id"`
	OldTickSize *decimal.Big `json:"old_tick_size"`
	NewTickSize *decimal.Big `json:"new_tick_size"`
	Timestamp   *time.Time   `json:"timestamp"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *MarketTickSizeChangePayload) UnmarshalJSON(raw []byte) error {
	type plain MarketTickSizeChangePayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "old_tick_size", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "new_tick_size", kind: "data_decimal", required: true},
		{name: "timestamp", kind: "stream_epoch_ms", nullable: true},
	}); err != nil {
		return err
	}
	*m = MarketTickSizeChangePayload(result)
	return nil
}

// NewMarketEvent represents a public stream message or payload.
type NewMarketEvent struct {
	Topic   string           `json:"topic"`
	Type    string           `json:"type"`
	Payload NewMarketPayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *NewMarketEvent) UnmarshalJSON(raw []byte) error {
	type plain NewMarketEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"market"}, fallback: json.RawMessage("\"market\"")},
		{name: "type", required: true, allowed: []string{"new_market"}},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = NewMarketEvent(result)
	return nil
}

// NewMarketPayload represents a public stream message or payload.
type NewMarketPayload struct {
	ID                    string              `json:"id"`
	ConditionID           *ConditionID        `json:"condition_id"`
	Market                string              `json:"market"`
	Question              *string             `json:"question"`
	Slug                  *string             `json:"slug"`
	Description           *string             `json:"description"`
	AssetIDs              []ClobAssetID       `json:"asset_ids"`
	Outcomes              []string            `json:"outcomes"`
	EventMessage          *MarketEventMessage `json:"event_message"`
	Timestamp             *time.Time          `json:"timestamp"`
	Tags                  []string            `json:"tags"`
	Active                *bool               `json:"active"`
	ClobTokenIDs          []string            `json:"clob_token_ids"`
	SportsMarketType      *string             `json:"sports_market_type"`
	Line                  *decimal.Big        `json:"line"`
	GameStartTime         *time.Time          `json:"game_start_time"`
	OrderPriceMinTickSize *decimal.Big        `json:"order_price_min_tick_size"`
	GroupItemTitle        *string             `json:"group_item_title"`
	TakerBaseFee          *decimal.Big        `json:"taker_base_fee"`
	FeesEnabled           *bool               `json:"fees_enabled"`
	FeeSchedule           json.RawMessage     `json:"fee_schedule"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *NewMarketPayload) UnmarshalJSON(raw []byte) error {
	type plain NewMarketPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "condition_id", steps: []string{"condition_feed"}, nullable: true},
		{name: "market", required: true},
		{name: "question", nullable: true},
		{name: "slug", nullable: true},
		{name: "description", nullable: true},
		{name: "asset_ids", aliases: []string{"asset_ids", "assets_ids", "token_ids"}, kind: "sequence", nullable: true},
		{name: "outcomes", kind: "sequence", nullable: true},
		{name: "event_message", nullable: true},
		{name: "timestamp", kind: "stream_epoch_ms", nullable: true},
		{name: "tags", kind: "sequence", nullable: true},
		{name: "active", kind: "bool", nullable: true},
		{name: "clob_token_ids", kind: "sequence", nullable: true},
		{name: "sports_market_type", nullable: true},
		{name: "line", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "game_start_time", kind: "game_start_time", nullable: true},
		{name: "order_price_min_tick_size", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "group_item_title", nullable: true},
		{name: "taker_base_fee", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "fees_enabled", kind: "bool", nullable: true},
		{name: "fee_schedule", nullable: true},
	}); err != nil {
		return err
	}
	*m = NewMarketPayload(result)
	return nil
}

// PriceChange represents a public stream message or payload.
type PriceChange struct {
	AssetID ClobAssetID  `json:"asset_id"`
	Price   *decimal.Big `json:"price"`
	Size    *decimal.Big `json:"size"`
	Side    OrderSide    `json:"side"`
	Hash    *string      `json:"hash"`
	BestBid *decimal.Big `json:"best_bid"`
	BestAsk *decimal.Big `json:"best_ask"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *PriceChange) UnmarshalJSON(raw []byte) error {
	type plain PriceChange
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "price", kind: "data_decimal", required: true},
		{name: "size", kind: "data_decimal", required: true},
		{name: "side", steps: []string{"uppercase"}, required: true, allowed: []string{"BUY", "SELL"}},
		{name: "hash", nullable: true},
		{name: "best_bid", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "best_ask", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
	}); err != nil {
		return err
	}
	*m = PriceChange(result)
	return nil
}

// SportsResultEvent represents a public stream message or payload.
type SportsResultEvent struct {
	Topic   string           `json:"topic"`
	Type    string           `json:"type"`
	Payload SportsGameResult `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *SportsResultEvent) UnmarshalJSON(raw []byte) error {
	type plain SportsResultEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"sports"}, fallback: json.RawMessage("\"sports\"")},
		{name: "type", allowed: []string{"sport_result"}, fallback: json.RawMessage("\"sport_result\"")},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = SportsResultEvent(result)
	return nil
}

// SportsGameResult represents a public stream message or payload.
type SportsGameResult struct {
	GameID             int64      `json:"game_id"`
	SportradarGameID   *string    `json:"sportradar_game_id"`
	Slug               *string    `json:"slug"`
	LeagueAbbreviation string     `json:"league_abbreviation"`
	HomeTeam           *string    `json:"home_team"`
	AwayTeam           *string    `json:"away_team"`
	Status             string     `json:"status"`
	Live               bool       `json:"live"`
	Ended              bool       `json:"ended"`
	Score              string     `json:"score"`
	Period             *string    `json:"period"`
	Elapsed            *string    `json:"elapsed"`
	FinishedAt         *time.Time `json:"finished_at"`
	Turn               *string    `json:"turn"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *SportsGameResult) UnmarshalJSON(raw []byte) error {
	normalized, err := normalizeStreamPayload(raw, "SportsGameResult")
	if err != nil {
		return err
	}
	raw = normalized
	type plain SportsGameResult
	var result plain
	if err = decodeData(raw, &result, []dataField{
		{name: "game_id", aliases: []string{"gameId"}, kind: "int", required: true},
		{name: "sportradar_game_id", aliases: []string{"sportradarGameId"}, nullable: true},
		{name: "slug", nullable: true},
		{name: "league_abbreviation", aliases: []string{"leagueAbbreviation"}, required: true},
		{name: "home_team", aliases: []string{"homeTeam"}, nullable: true},
		{name: "away_team", aliases: []string{"awayTeam"}, nullable: true},
		{name: "status", required: true},
		{name: "live", kind: "bool", required: true},
		{name: "ended", kind: "bool", required: true},
		{name: "score", required: true},
		{name: "period", nullable: true},
		{name: "elapsed", nullable: true},
		{name: "finished_at", aliases: []string{"finishedTimestamp"}, kind: "reward_epoch_ms_unsigned", nullable: true},
		{name: "turn", nullable: true},
	}); err != nil {
		return err
	}
	*m = SportsGameResult(result)
	return nil
}

// CommentCreatedEvent represents a public stream message or payload.
type CommentCreatedEvent struct {
	Topic     string     `json:"topic"`
	Type      string     `json:"type"`
	Timestamp *time.Time `json:"timestamp"`
	Payload   Comment    `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CommentCreatedEvent) UnmarshalJSON(raw []byte) error {
	type plain CommentCreatedEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"comments"}, fallback: json.RawMessage("\"comments\"")},
		{name: "type", required: true, allowed: []string{"comment_created"}},
		{name: "timestamp", kind: "reward_epoch_ms_unsigned", required: true, nullable: true},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = CommentCreatedEvent(result)
	return nil
}

// CommentRemovedEvent represents a public stream message or payload.
type CommentRemovedEvent struct {
	Topic     string                `json:"topic"`
	Type      string                `json:"type"`
	Timestamp *time.Time            `json:"timestamp"`
	Payload   CommentRemovedPayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CommentRemovedEvent) UnmarshalJSON(raw []byte) error {
	type plain CommentRemovedEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"comments"}, fallback: json.RawMessage("\"comments\"")},
		{name: "type", required: true, allowed: []string{"comment_removed"}},
		{name: "timestamp", kind: "reward_epoch_ms_unsigned", required: true, nullable: true},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = CommentRemovedEvent(result)
	return nil
}

// CommentRemovedPayload represents a public stream message or payload.
type CommentRemovedPayload struct {
	ID               string          `json:"id"`
	Body             *string         `json:"body"`
	ParentEntityType *string         `json:"parent_entity_type"`
	ParentEntityID   *int64          `json:"parent_entity_id"`
	ParentCommentID  *string         `json:"parent_comment_id"`
	UserAddress      *string         `json:"user_address"`
	ReplyAddress     *string         `json:"reply_address"`
	CreatedAt        *time.Time      `json:"created_at"`
	UpdatedAt        *time.Time      `json:"updated_at"`
	Media            []CommentMedia  `json:"media"`
	Profile          *CommentProfile `json:"profile"`
	Reactions        []Reaction      `json:"reactions"`
	ReportCount      *int64          `json:"report_count"`
	ReactionCount    *int64          `json:"reaction_count"`
	TradeAsset       *string         `json:"trade_asset"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CommentRemovedPayload) UnmarshalJSON(raw []byte) error {
	type plain CommentRemovedPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "body", nullable: true},
		{name: "parent_entity_type", aliases: []string{"parentEntityType"}, nullable: true, allowed: []string{"Event", "Market"}},
		{name: "parent_entity_id", aliases: []string{"parentEntityID"}, kind: "int", nullable: true},
		{name: "parent_comment_id", aliases: []string{"parentCommentID"}, nullable: true},
		{name: "user_address", aliases: []string{"userAddress"}, nullable: true},
		{name: "reply_address", aliases: []string{"replyAddress"}, nullable: true},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime", nullable: true},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime", nullable: true},
		{name: "media", kind: "sequence", nullable: true},
		{name: "profile", nullable: true},
		{name: "reactions", kind: "sequence", nullable: true},
		{name: "report_count", aliases: []string{"reportCount"}, kind: "int", nullable: true},
		{name: "reaction_count", aliases: []string{"reactionCount"}, kind: "int", nullable: true},
		{name: "trade_asset", aliases: []string{"tradeAsset"}, nullable: true},
	}); err != nil {
		return err
	}
	*m = CommentRemovedPayload(result)
	return nil
}

// CryptoPricesBinanceEvent represents a public stream message or payload.
type CryptoPricesBinanceEvent struct {
	Topic     string             `json:"topic"`
	Type      string             `json:"type"`
	Timestamp *time.Time         `json:"timestamp"`
	Payload   PriceUpdatePayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CryptoPricesBinanceEvent) UnmarshalJSON(raw []byte) error {
	type plain CryptoPricesBinanceEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"prices.crypto.binance"}, fallback: json.RawMessage("\"prices.crypto.binance\"")},
		{name: "type", required: true, allowed: []string{"update"}},
		{name: "timestamp", kind: "reward_epoch_ms_unsigned", required: true, nullable: true},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = CryptoPricesBinanceEvent(result)
	return nil
}

// CryptoPricesChainlinkEvent represents a public stream message or payload.
type CryptoPricesChainlinkEvent struct {
	Topic     string             `json:"topic"`
	Type      string             `json:"type"`
	Timestamp *time.Time         `json:"timestamp"`
	Payload   PriceUpdatePayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CryptoPricesChainlinkEvent) UnmarshalJSON(raw []byte) error {
	type plain CryptoPricesChainlinkEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"prices.crypto.chainlink"}, fallback: json.RawMessage("\"prices.crypto.chainlink\"")},
		{name: "type", required: true, allowed: []string{"update"}},
		{name: "timestamp", kind: "reward_epoch_ms_unsigned", required: true, nullable: true},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = CryptoPricesChainlinkEvent(result)
	return nil
}

// CryptoPricesChainlinkTwapEvent represents a public stream message or payload.
type CryptoPricesChainlinkTwapEvent struct {
	Topic     string                           `json:"topic"`
	Type      string                           `json:"type"`
	Timestamp *time.Time                       `json:"timestamp"`
	Payload   CryptoPricesChainlinkTwapPayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CryptoPricesChainlinkTwapEvent) UnmarshalJSON(raw []byte) error {
	type plain CryptoPricesChainlinkTwapEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"prices.crypto.chainlink.twap"}, fallback: json.RawMessage("\"prices.crypto.chainlink.twap\"")},
		{name: "type", required: true, allowed: []string{"update"}},
		{name: "timestamp", kind: "reward_epoch_ms_unsigned", required: true, nullable: true},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = CryptoPricesChainlinkTwapEvent(result)
	return nil
}

// CryptoPricesChainlinkTwapPayload represents a public stream message or payload.
type CryptoPricesChainlinkTwapPayload struct {
	Symbol        string       `json:"symbol"`
	Timestamp     int64        `json:"timestamp"`
	Value         *decimal.Big `json:"value"`
	WindowSeconds int64        `json:"window_seconds"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CryptoPricesChainlinkTwapPayload) UnmarshalJSON(raw []byte) error {
	normalized, err := normalizeStreamPayload(raw, "CryptoPricesChainlinkTwapPayload")
	if err != nil {
		return err
	}
	raw = normalized
	type plain CryptoPricesChainlinkTwapPayload
	var result plain
	if err = decodeData(raw, &result, []dataField{
		{name: "symbol", required: true},
		{name: "timestamp", kind: "int", required: true},
		{name: "value", kind: "data_decimal", required: true},
		{name: "window_seconds", aliases: []string{"window_s"}, required: true, kind: "int", steps: []string{"twap_window"}},
	}); err != nil {
		return err
	}
	*m = CryptoPricesChainlinkTwapPayload(result)
	return nil
}

// EquityPriceSnapshotEntry represents a public stream message or payload.
type EquityPriceSnapshotEntry struct {
	Timestamp int64        `json:"timestamp"`
	Value     *decimal.Big `json:"value"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *EquityPriceSnapshotEntry) UnmarshalJSON(raw []byte) error {
	type plain EquityPriceSnapshotEntry
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "int", required: true},
		{name: "value", kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = EquityPriceSnapshotEntry(result)
	return nil
}

// EquityPriceSubscribePayload represents a public stream message or payload.
type EquityPriceSubscribePayload struct {
	Symbol string                     `json:"symbol"`
	Data   []EquityPriceSnapshotEntry `json:"data"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *EquityPriceSubscribePayload) UnmarshalJSON(raw []byte) error {
	type plain EquityPriceSubscribePayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "symbol", required: true},
		{name: "data", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = EquityPriceSubscribePayload(result)
	return nil
}

// EquityPriceUpdatePayload represents a public stream message or payload.
type EquityPriceUpdatePayload struct {
	Symbol           string       `json:"symbol"`
	Value            *decimal.Big `json:"value"`
	Timestamp        int64        `json:"timestamp"`
	ReceivedAt       *int64       `json:"received_at"`
	IsCarriedForward *bool        `json:"is_carried_forward"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *EquityPriceUpdatePayload) UnmarshalJSON(raw []byte) error {
	normalized, err := normalizeStreamPayload(raw, "EquityPriceUpdatePayload")
	if err != nil {
		return err
	}
	raw = normalized
	type plain EquityPriceUpdatePayload
	var result plain
	if err = decodeData(raw, &result, []dataField{
		{name: "symbol", required: true},
		{name: "value", kind: "data_decimal", required: true},
		{name: "timestamp", kind: "int", required: true},
		{name: "received_at", kind: "int", nullable: true},
		{name: "is_carried_forward", kind: "bool", nullable: true},
	}); err != nil {
		return err
	}
	*m = EquityPriceUpdatePayload(result)
	return nil
}

// EquityPricesSubscribeEvent represents a public stream message or payload.
type EquityPricesSubscribeEvent struct {
	Topic     string                      `json:"topic"`
	Type      string                      `json:"type"`
	Timestamp *time.Time                  `json:"timestamp"`
	Payload   EquityPriceSubscribePayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *EquityPricesSubscribeEvent) UnmarshalJSON(raw []byte) error {
	type plain EquityPricesSubscribeEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"prices.equity.pyth"}, fallback: json.RawMessage("\"prices.equity.pyth\"")},
		{name: "type", required: true, allowed: []string{"subscribe"}},
		{name: "timestamp", kind: "reward_epoch_ms_unsigned", required: true, nullable: true},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = EquityPricesSubscribeEvent(result)
	return nil
}

// EquityPricesUpdateEvent represents a public stream message or payload.
type EquityPricesUpdateEvent struct {
	Topic     string                   `json:"topic"`
	Type      string                   `json:"type"`
	Timestamp *time.Time               `json:"timestamp"`
	Payload   EquityPriceUpdatePayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *EquityPricesUpdateEvent) UnmarshalJSON(raw []byte) error {
	type plain EquityPricesUpdateEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"prices.equity.pyth"}, fallback: json.RawMessage("\"prices.equity.pyth\"")},
		{name: "type", required: true, allowed: []string{"update"}},
		{name: "timestamp", kind: "reward_epoch_ms_unsigned", required: true, nullable: true},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = EquityPricesUpdateEvent(result)
	return nil
}

// PriceUpdatePayload represents a public stream message or payload.
type PriceUpdatePayload struct {
	Symbol    string       `json:"symbol"`
	Timestamp int64        `json:"timestamp"`
	Value     *decimal.Big `json:"value"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *PriceUpdatePayload) UnmarshalJSON(raw []byte) error {
	type plain PriceUpdatePayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "symbol", required: true},
		{name: "timestamp", kind: "int", required: true},
		{name: "value", kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = PriceUpdatePayload(result)
	return nil
}

// ReactionCreatedEvent represents a public stream message or payload.
type ReactionCreatedEvent struct {
	Topic     string     `json:"topic"`
	Type      string     `json:"type"`
	Timestamp *time.Time `json:"timestamp"`
	Payload   Reaction   `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *ReactionCreatedEvent) UnmarshalJSON(raw []byte) error {
	type plain ReactionCreatedEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"comments"}, fallback: json.RawMessage("\"comments\"")},
		{name: "type", required: true, allowed: []string{"reaction_created"}},
		{name: "timestamp", kind: "reward_epoch_ms_unsigned", required: true, nullable: true},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = ReactionCreatedEvent(result)
	return nil
}

// ReactionRemovedEvent represents a public stream message or payload.
type ReactionRemovedEvent struct {
	Topic     string     `json:"topic"`
	Type      string     `json:"type"`
	Timestamp *time.Time `json:"timestamp"`
	Payload   Reaction   `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *ReactionRemovedEvent) UnmarshalJSON(raw []byte) error {
	type plain ReactionRemovedEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"comments"}, fallback: json.RawMessage("\"comments\"")},
		{name: "type", required: true, allowed: []string{"reaction_removed"}},
		{name: "timestamp", kind: "reward_epoch_ms_unsigned", required: true, nullable: true},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = ReactionRemovedEvent(result)
	return nil
}

// UserOrderEvent represents a public stream message or payload.
type UserOrderEvent struct {
	Topic   string           `json:"topic"`
	Type    string           `json:"type"`
	Payload UserOrderPayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *UserOrderEvent) UnmarshalJSON(raw []byte) error {
	type plain UserOrderEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"user"}, fallback: json.RawMessage("\"user\"")},
		{name: "type", required: true, allowed: []string{"order"}},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = UserOrderEvent(result)
	return nil
}

// UserOrderPayload represents a public stream message or payload.
type UserOrderPayload struct {
	ID              string       `json:"id"`
	Owner           string       `json:"owner"`
	ConditionID     ConditionID  `json:"condition_id"`
	AssetID         ClobAssetID  `json:"asset_id"`
	Side            OrderSide    `json:"side"`
	OriginalSize    *decimal.Big `json:"original_size"`
	SizeMatched     *decimal.Big `json:"size_matched"`
	Price           *decimal.Big `json:"price"`
	OrderEventType  string       `json:"order_event_type"`
	Timestamp       *time.Time   `json:"timestamp"`
	CreatedAt       *time.Time   `json:"created_at"`
	ExpiresAt       *time.Time   `json:"expires_at"`
	OrderType       *string      `json:"order_type"`
	Status          *string      `json:"status"`
	MakerAddress    *string      `json:"maker_address"`
	OrderOwner      *string      `json:"order_owner"`
	AssociateTrades []string     `json:"associate_trades"`
	Outcome         *string      `json:"outcome"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *UserOrderPayload) UnmarshalJSON(raw []byte) error {
	type plain UserOrderPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "owner", required: true},
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "side", steps: []string{"uppercase"}, required: true, allowed: []string{"BUY", "SELL"}},
		{name: "original_size", kind: "data_decimal", required: true},
		{name: "size_matched", kind: "data_decimal", required: true},
		{name: "price", kind: "data_decimal", required: true},
		{name: "order_event_type", aliases: []string{"type"}, required: true, allowed: []string{"PLACEMENT", "UPDATE", "CANCELLATION"}},
		{name: "timestamp", kind: "clob_epoch_iso", nullable: true},
		{name: "created_at", kind: "clob_epoch_iso", nullable: true},
		{name: "expires_at", steps: []string{"zero_time"}, aliases: []string{"expiration"}, kind: "clob_epoch_iso", nullable: true},
		{name: "order_type", nullable: true, allowed: []string{"GTC", "FOK", "IOC", "GTD", "FAK"}},
		{name: "status", nullable: true},
		{name: "maker_address", nullable: true},
		{name: "order_owner", nullable: true},
		{name: "associate_trades", kind: "sequence", nullable: true},
		{name: "outcome", nullable: true},
	}); err != nil {
		return err
	}
	*m = UserOrderPayload(result)
	return nil
}

// UserTradeEvent represents a public stream message or payload.
type UserTradeEvent struct {
	Topic   string           `json:"topic"`
	Type    string           `json:"type"`
	Payload UserTradePayload `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *UserTradeEvent) UnmarshalJSON(raw []byte) error {
	type plain UserTradeEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "topic", allowed: []string{"user"}, fallback: json.RawMessage("\"user\"")},
		{name: "type", required: true, allowed: []string{"trade"}},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = UserTradeEvent(result)
	return nil
}

// UserTradeMakerOrder represents a public stream message or payload.
type UserTradeMakerOrder struct {
	OrderID       string       `json:"order_id"`
	Owner         string       `json:"owner"`
	MakerAddress  *string      `json:"maker_address"`
	MatchedAmount *decimal.Big `json:"matched_amount"`
	Price         *decimal.Big `json:"price"`
	FeeRateBps    *decimal.Big `json:"fee_rate_bps"`
	AssetID       ClobAssetID  `json:"asset_id"`
	Side          OrderSide    `json:"side"`
	Outcome       *string      `json:"outcome"`
	OutcomeIndex  *int64       `json:"outcome_index"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *UserTradeMakerOrder) UnmarshalJSON(raw []byte) error {
	type plain UserTradeMakerOrder
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "order_id", required: true},
		{name: "owner", required: true},
		{name: "maker_address", nullable: true},
		{name: "matched_amount", kind: "data_decimal", required: true},
		{name: "price", kind: "data_decimal", required: true},
		{name: "fee_rate_bps", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "side", steps: []string{"uppercase"}, required: true, allowed: []string{"BUY", "SELL"}},
		{name: "outcome", nullable: true},
		{name: "outcome_index", kind: "int", nullable: true},
	}); err != nil {
		return err
	}
	*m = UserTradeMakerOrder(result)
	return nil
}

// UserTradePayload represents a public stream message or payload.
type UserTradePayload struct {
	ID              string                `json:"id"`
	TakerOrderID    string                `json:"taker_order_id"`
	ConditionID     ConditionID           `json:"condition_id"`
	AssetID         ClobAssetID           `json:"asset_id"`
	Side            OrderSide             `json:"side"`
	Size            *decimal.Big          `json:"size"`
	Price           *decimal.Big          `json:"price"`
	Status          *string               `json:"status"`
	Owner           string                `json:"owner"`
	Timestamp       *time.Time            `json:"timestamp"`
	FeeRateBps      *decimal.Big          `json:"fee_rate_bps"`
	MatchedAt       *time.Time            `json:"matched_at"`
	UpdatedAt       *time.Time            `json:"updated_at"`
	TradeOwner      *string               `json:"trade_owner"`
	MakerAddress    *string               `json:"maker_address"`
	TransactionHash *string               `json:"transaction_hash"`
	BucketIndex     *int64                `json:"bucket_index"`
	MakerOrders     []UserTradeMakerOrder `json:"maker_orders"`
	TraderSide      *string               `json:"trader_side"`
	Outcome         *string               `json:"outcome"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *UserTradePayload) UnmarshalJSON(raw []byte) error {
	type plain UserTradePayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "taker_order_id", required: true},
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "side", steps: []string{"uppercase"}, required: true, allowed: []string{"BUY", "SELL"}},
		{name: "size", kind: "data_decimal", required: true},
		{name: "price", kind: "data_decimal", required: true},
		{name: "status", steps: []string{"trade_status"}, required: true, allowed: []string{"MATCHED", "MATCHED_NOT_BROADCASTED", "MINED", "CONFIRMED", "RETRYING", "FAILED"}},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", nullable: true},
		{name: "fee_rate_bps", steps: []string{"optional_text"}, kind: "data_decimal", nullable: true},
		{name: "matched_at", aliases: []string{"match_time", "matchtime"}, kind: "clob_epoch_iso", nullable: true},
		{name: "updated_at", aliases: []string{"last_update"}, kind: "clob_epoch_iso", nullable: true},
		{name: "trade_owner", nullable: true},
		{name: "maker_address", nullable: true},
		{name: "transaction_hash", nullable: true},
		{name: "bucket_index", kind: "int", nullable: true},
		{name: "maker_orders", kind: "sequence", nullable: true},
		{name: "trader_side", steps: []string{"uppercase"}, nullable: true, allowed: []string{"TAKER", "MAKER"}},
		{name: "outcome", nullable: true},
	}); err != nil {
		return err
	}
	*m = UserTradePayload(result)
	return nil
}

// RealtimeEquityTwapSnapshot represents a public stream message or payload.
type RealtimeEquityTwapSnapshot struct {
	MarketStatus  *EquityMarketStatus  `json:"market_status"`
	Symbol        string               `json:"symbol"`
	Source        string               `json:"source"`
	Data          []RealtimePricePoint `json:"data"`
	WindowSeconds int64                `json:"window_seconds"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *RealtimeEquityTwapSnapshot) UnmarshalJSON(raw []byte) error {
	type plain RealtimeEquityTwapSnapshot
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "market_status", nullable: true, allowed: []string{"open", "closed", "unknown"}},
		{name: "symbol", required: true},
		{name: "source", required: true},
		{name: "data", kind: "sequence", required: true},
		{name: "window_seconds", kind: "int", steps: []string{"price_window"}, fallback: json.RawMessage("60")},
	}); err != nil {
		return err
	}
	*m = RealtimeEquityTwapSnapshot(result)
	return nil
}

// RealtimeEquityTwapUpdate represents a public stream message or payload.
type RealtimeEquityTwapUpdate struct {
	MarketStatus  *EquityMarketStatus `json:"market_status"`
	Timestamp     time.Time           `json:"timestamp"`
	Value         *decimal.Big        `json:"value"`
	Symbol        string              `json:"symbol"`
	Source        string              `json:"source"`
	WindowSeconds int64               `json:"window_seconds"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *RealtimeEquityTwapUpdate) UnmarshalJSON(raw []byte) error {
	type plain RealtimeEquityTwapUpdate
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "market_status", nullable: true, allowed: []string{"open", "closed", "unknown"}},
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "value", kind: "data_decimal", required: true},
		{name: "symbol", required: true},
		{name: "source", required: true},
		{name: "window_seconds", kind: "int", steps: []string{"price_window"}, fallback: json.RawMessage("60")},
	}); err != nil {
		return err
	}
	*m = RealtimeEquityTwapUpdate(result)
	return nil
}

// RealtimeEquityPriceSnapshot represents a public stream message or payload.
type RealtimeEquityPriceSnapshot struct {
	MarketStatus *EquityMarketStatus  `json:"market_status"`
	Symbol       string               `json:"symbol"`
	Source       string               `json:"source"`
	Data         []RealtimePricePoint `json:"data"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *RealtimeEquityPriceSnapshot) UnmarshalJSON(raw []byte) error {
	type plain RealtimeEquityPriceSnapshot
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "market_status", nullable: true, allowed: []string{"open", "closed", "unknown"}},
		{name: "symbol", required: true},
		{name: "source", required: true},
		{name: "data", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = RealtimeEquityPriceSnapshot(result)
	return nil
}

// RealtimeEquityPriceUpdate represents a public stream message or payload.
type RealtimeEquityPriceUpdate struct {
	MarketStatus     *EquityMarketStatus `json:"market_status"`
	Timestamp        time.Time           `json:"timestamp"`
	Value            *decimal.Big        `json:"value"`
	Symbol           string              `json:"symbol"`
	Source           string              `json:"source"`
	ReceivedAt       *time.Time          `json:"received_at"`
	IsCarriedForward *bool               `json:"is_carried_forward"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *RealtimeEquityPriceUpdate) UnmarshalJSON(raw []byte) error {
	type plain RealtimeEquityPriceUpdate
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "market_status", nullable: true, allowed: []string{"open", "closed", "unknown"}},
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "value", kind: "data_decimal", required: true},
		{name: "symbol", required: true},
		{name: "source", required: true},
		{name: "received_at", kind: "epoch_datetime", nullable: true},
		{name: "is_carried_forward", kind: "bool", nullable: true},
	}); err != nil {
		return err
	}
	*m = RealtimeEquityPriceUpdate(result)
	return nil
}

// CryptoPriceSnapshotEvent represents a public stream message or payload.
type CryptoPriceSnapshotEvent struct {
	Timestamp time.Time             `json:"timestamp"`
	Seq       *int64                `json:"seq"`
	Dropped   *int64                `json:"dropped"`
	Topic     string                `json:"topic"`
	Type      string                `json:"type"`
	Payload   RealtimePriceSnapshot `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CryptoPriceSnapshotEvent) UnmarshalJSON(raw []byte) error {
	type plain CryptoPriceSnapshotEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "seq", kind: "int", nullable: true},
		{name: "dropped", kind: "int", nullable: true},
		{name: "topic", allowed: []string{"prices.crypto"}, fallback: json.RawMessage("\"prices.crypto\"")},
		{name: "type", allowed: []string{"subscribe"}, fallback: json.RawMessage("\"subscribe\"")},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = CryptoPriceSnapshotEvent(result)
	return nil
}

// CryptoPriceUpdateEvent represents a public stream message or payload.
type CryptoPriceUpdateEvent struct {
	Timestamp time.Time           `json:"timestamp"`
	Seq       *int64              `json:"seq"`
	Dropped   *int64              `json:"dropped"`
	Topic     string              `json:"topic"`
	Type      string              `json:"type"`
	Payload   RealtimePriceUpdate `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CryptoPriceUpdateEvent) UnmarshalJSON(raw []byte) error {
	type plain CryptoPriceUpdateEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "seq", kind: "int", nullable: true},
		{name: "dropped", kind: "int", nullable: true},
		{name: "topic", allowed: []string{"prices.crypto"}, fallback: json.RawMessage("\"prices.crypto\"")},
		{name: "type", allowed: []string{"update"}, fallback: json.RawMessage("\"update\"")},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = CryptoPriceUpdateEvent(result)
	return nil
}

// CryptoTwapPriceSnapshotEvent represents a public stream message or payload.
type CryptoTwapPriceSnapshotEvent struct {
	Timestamp time.Time            `json:"timestamp"`
	Seq       *int64               `json:"seq"`
	Dropped   *int64               `json:"dropped"`
	Topic     string               `json:"topic"`
	Type      string               `json:"type"`
	Payload   RealtimeTwapSnapshot `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CryptoTwapPriceSnapshotEvent) UnmarshalJSON(raw []byte) error {
	type plain CryptoTwapPriceSnapshotEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "seq", kind: "int", nullable: true},
		{name: "dropped", kind: "int", nullable: true},
		{name: "topic", allowed: []string{"prices.crypto.twap"}, fallback: json.RawMessage("\"prices.crypto.twap\"")},
		{name: "type", allowed: []string{"subscribe"}, fallback: json.RawMessage("\"subscribe\"")},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = CryptoTwapPriceSnapshotEvent(result)
	return nil
}

// CryptoTwapPriceUpdateEvent represents a public stream message or payload.
type CryptoTwapPriceUpdateEvent struct {
	Timestamp time.Time          `json:"timestamp"`
	Seq       *int64             `json:"seq"`
	Dropped   *int64             `json:"dropped"`
	Topic     string             `json:"topic"`
	Type      string             `json:"type"`
	Payload   RealtimeTwapUpdate `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *CryptoTwapPriceUpdateEvent) UnmarshalJSON(raw []byte) error {
	type plain CryptoTwapPriceUpdateEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "seq", kind: "int", nullable: true},
		{name: "dropped", kind: "int", nullable: true},
		{name: "topic", allowed: []string{"prices.crypto.twap"}, fallback: json.RawMessage("\"prices.crypto.twap\"")},
		{name: "type", allowed: []string{"update"}, fallback: json.RawMessage("\"update\"")},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = CryptoTwapPriceUpdateEvent(result)
	return nil
}

// EquityPriceSnapshotEvent represents a public stream message or payload.
type EquityPriceSnapshotEvent struct {
	Timestamp time.Time                   `json:"timestamp"`
	Seq       *int64                      `json:"seq"`
	Dropped   *int64                      `json:"dropped"`
	Topic     string                      `json:"topic"`
	Type      string                      `json:"type"`
	Payload   RealtimeEquityPriceSnapshot `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *EquityPriceSnapshotEvent) UnmarshalJSON(raw []byte) error {
	type plain EquityPriceSnapshotEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "seq", kind: "int", nullable: true},
		{name: "dropped", kind: "int", nullable: true},
		{name: "topic", allowed: []string{"prices.equity"}, fallback: json.RawMessage("\"prices.equity\"")},
		{name: "type", allowed: []string{"subscribe"}, fallback: json.RawMessage("\"subscribe\"")},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = EquityPriceSnapshotEvent(result)
	return nil
}

// EquityPriceUpdateEvent represents a public stream message or payload.
type EquityPriceUpdateEvent struct {
	Timestamp time.Time                 `json:"timestamp"`
	Seq       *int64                    `json:"seq"`
	Dropped   *int64                    `json:"dropped"`
	Topic     string                    `json:"topic"`
	Type      string                    `json:"type"`
	Payload   RealtimeEquityPriceUpdate `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *EquityPriceUpdateEvent) UnmarshalJSON(raw []byte) error {
	type plain EquityPriceUpdateEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "seq", kind: "int", nullable: true},
		{name: "dropped", kind: "int", nullable: true},
		{name: "topic", allowed: []string{"prices.equity"}, fallback: json.RawMessage("\"prices.equity\"")},
		{name: "type", allowed: []string{"update"}, fallback: json.RawMessage("\"update\"")},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = EquityPriceUpdateEvent(result)
	return nil
}

// EquityTwapPriceSnapshotEvent represents a public stream message or payload.
type EquityTwapPriceSnapshotEvent struct {
	Timestamp time.Time                  `json:"timestamp"`
	Seq       *int64                     `json:"seq"`
	Dropped   *int64                     `json:"dropped"`
	Topic     string                     `json:"topic"`
	Type      string                     `json:"type"`
	Payload   RealtimeEquityTwapSnapshot `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *EquityTwapPriceSnapshotEvent) UnmarshalJSON(raw []byte) error {
	type plain EquityTwapPriceSnapshotEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "seq", kind: "int", nullable: true},
		{name: "dropped", kind: "int", nullable: true},
		{name: "topic", allowed: []string{"prices.equity.twap"}, fallback: json.RawMessage("\"prices.equity.twap\"")},
		{name: "type", allowed: []string{"subscribe"}, fallback: json.RawMessage("\"subscribe\"")},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = EquityTwapPriceSnapshotEvent(result)
	return nil
}

// EquityTwapPriceUpdateEvent represents a public stream message or payload.
type EquityTwapPriceUpdateEvent struct {
	Timestamp time.Time                `json:"timestamp"`
	Seq       *int64                   `json:"seq"`
	Dropped   *int64                   `json:"dropped"`
	Topic     string                   `json:"topic"`
	Type      string                   `json:"type"`
	Payload   RealtimeEquityTwapUpdate `json:"payload"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *EquityTwapPriceUpdateEvent) UnmarshalJSON(raw []byte) error {
	type plain EquityTwapPriceUpdateEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "seq", kind: "int", nullable: true},
		{name: "dropped", kind: "int", nullable: true},
		{name: "topic", allowed: []string{"prices.equity.twap"}, fallback: json.RawMessage("\"prices.equity.twap\"")},
		{name: "type", allowed: []string{"update"}, fallback: json.RawMessage("\"update\"")},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = EquityTwapPriceUpdateEvent(result)
	return nil
}

// RealtimePricePoint represents a public stream message or payload.
type RealtimePricePoint struct {
	Timestamp time.Time    `json:"timestamp"`
	Value     *decimal.Big `json:"value"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *RealtimePricePoint) UnmarshalJSON(raw []byte) error {
	type plain RealtimePricePoint
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "value", kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = RealtimePricePoint(result)
	return nil
}

// RealtimePriceSnapshot represents a public stream message or payload.
type RealtimePriceSnapshot struct {
	Symbol string               `json:"symbol"`
	Source string               `json:"source"`
	Data   []RealtimePricePoint `json:"data"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *RealtimePriceSnapshot) UnmarshalJSON(raw []byte) error {
	type plain RealtimePriceSnapshot
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "symbol", required: true},
		{name: "source", required: true},
		{name: "data", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = RealtimePriceSnapshot(result)
	return nil
}

// RealtimePriceUpdate represents a public stream message or payload.
type RealtimePriceUpdate struct {
	Timestamp        time.Time    `json:"timestamp"`
	Value            *decimal.Big `json:"value"`
	Symbol           string       `json:"symbol"`
	Source           string       `json:"source"`
	ReceivedAt       *time.Time   `json:"received_at"`
	IsCarriedForward *bool        `json:"is_carried_forward"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *RealtimePriceUpdate) UnmarshalJSON(raw []byte) error {
	type plain RealtimePriceUpdate
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "value", kind: "data_decimal", required: true},
		{name: "symbol", required: true},
		{name: "source", required: true},
		{name: "received_at", kind: "epoch_datetime", nullable: true},
		{name: "is_carried_forward", kind: "bool", nullable: true},
	}); err != nil {
		return err
	}
	*m = RealtimePriceUpdate(result)
	return nil
}

// RealtimeTwapSnapshot represents a public stream message or payload.
type RealtimeTwapSnapshot struct {
	Symbol        string               `json:"symbol"`
	Source        string               `json:"source"`
	Data          []RealtimePricePoint `json:"data"`
	WindowSeconds int64                `json:"window_seconds"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *RealtimeTwapSnapshot) UnmarshalJSON(raw []byte) error {
	type plain RealtimeTwapSnapshot
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "symbol", required: true},
		{name: "source", required: true},
		{name: "data", kind: "sequence", required: true},
		{name: "window_seconds", kind: "int", steps: []string{"price_window"}, fallback: json.RawMessage("60")},
	}); err != nil {
		return err
	}
	*m = RealtimeTwapSnapshot(result)
	return nil
}

// RealtimeTwapUpdate represents a public stream message or payload.
type RealtimeTwapUpdate struct {
	Timestamp     time.Time    `json:"timestamp"`
	Value         *decimal.Big `json:"value"`
	Symbol        string       `json:"symbol"`
	Source        string       `json:"source"`
	WindowSeconds int64        `json:"window_seconds"`
}

// UnmarshalJSON validates stream fields and preserves exact amounts.
func (m *RealtimeTwapUpdate) UnmarshalJSON(raw []byte) error {
	type plain RealtimeTwapUpdate
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch_datetime", required: true},
		{name: "value", kind: "data_decimal", required: true},
		{name: "symbol", required: true},
		{name: "source", required: true},
		{name: "window_seconds", kind: "int", steps: []string{"price_window"}, fallback: json.RawMessage("60")},
	}); err != nil {
		return err
	}
	*m = RealtimeTwapUpdate(result)
	return nil
}
