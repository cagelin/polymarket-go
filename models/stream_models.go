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
