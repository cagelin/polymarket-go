// Market models ported from polymarket-client 0.14.0.
// Regenerate with tests/parity/generate_market_types.py. These files are linted.

package models

import (
	"encoding/json"
	"time"

	"github.com/ericlagergren/decimal"
)

// Date is an ISO 8601 calendar date (YYYY-MM-DD).
type Date string

// ClobRewardID identifies a ClobReward resource.
type ClobRewardID string

// EVMAddress identifies a EVMAddress resource.
type EVMAddress string

// EventID identifies a Event resource.
type EventID string

// MarketID identifies a Market resource.
type MarketID string

// PositionID identifies a Position resource.
type PositionID string

// QuestionID identifies a Question resource.
type QuestionID string

// ResolutionRequestID identifies a ResolutionRequest resource.
type ResolutionRequestID string

// TagID identifies a Tag resource.
type TagID string

// TokenID identifies a Token resource.
type TokenID string

// ProtocolVersion describes known market states.
type ProtocolVersion string

// Supported ProtocolVersion values.
const (
	ProtocolV1 ProtocolVersion = "v1"
	ProtocolV2 ProtocolVersion = "v2"
)

// MarketResolutionStatus describes known market states.
type MarketResolutionStatus string

// Supported MarketResolutionStatus values.
const (
	ResolutionInactive MarketResolutionStatus = "inactive"
	ResolutionActive   MarketResolutionStatus = "active"
	ResolutionResolved MarketResolutionStatus = "resolved"
)

// UMAResolutionStatus describes known market states.
type UMAResolutionStatus string

// Supported UMAResolutionStatus values.
const (
	UMADisputed  UMAResolutionStatus = "disputed"
	UMAProposed  UMAResolutionStatus = "proposed"
	UMARequested UMAResolutionStatus = "requested"
	UMAResolved  UMAResolutionStatus = "resolved"
	UMASettled   UMAResolutionStatus = "settled"
)

// ClobReward represents reward configuration attached to a market condition.
type ClobReward struct {
	ID               ClobRewardID `json:"id"`
	ConditionID      ConditionID  `json:"condition_id"`
	AssetAddress     string       `json:"asset_address"`
	RewardsAmount    *decimal.Big `json:"rewards_amount"`
	RewardsDailyRate *decimal.Big `json:"rewards_daily_rate"`
	StartDate        Date         `json:"start_date"`
	EndDate          *Date        `json:"end_date"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *ClobReward) UnmarshalJSON(data []byte) error {
	type plain ClobReward
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "condition_id", aliases: []string{"conditionId"}, kind: "condition", required: true},
		{name: "asset_address", aliases: []string{"assetAddress"}, required: true},
		{name: "rewards_amount", aliases: []string{"rewardsAmount"}, kind: "decimal", required: true},
		{name: "rewards_daily_rate", aliases: []string{"rewardsDailyRate"}, kind: "decimal", required: true},
		{name: "start_date", aliases: []string{"startDate"}, kind: "date", required: true},
		{name: "end_date", aliases: []string{"endDate"}, kind: "date"},
	}); err != nil {
		return err
	}
	*m = ClobReward(result)
	return nil
}

// FeeSchedule represents fee schedule applied to market trading.
type FeeSchedule struct {
	Exponent   json.Number  `json:"exponent"`
	Rate       *decimal.Big `json:"rate"`
	TakerOnly  bool         `json:"taker_only"`
	RebateRate *decimal.Big `json:"rebate_rate"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *FeeSchedule) UnmarshalJSON(data []byte) error {
	type plain FeeSchedule
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "exponent", kind: "number", required: true},
		{name: "rate", kind: "decimal", required: true},
		{name: "taker_only", aliases: []string{"takerOnly"}, kind: "bool", required: true},
		{name: "rebate_rate", aliases: []string{"rebateRate"}, kind: "decimal", required: true},
	}); err != nil {
		return err
	}
	*m = FeeSchedule(result)
	return nil
}

// Market represents a Polymarket market.
type Market struct {
	ID             MarketID             `json:"id"`
	Version        *ProtocolVersion     `json:"version"`
	Slug           *string              `json:"slug"`
	ConditionID    *ConditionID         `json:"condition_id"`
	Question       *string              `json:"question"`
	GroupItemTitle *string              `json:"group_item_title"`
	Description    *string              `json:"description"`
	Category       *string              `json:"category"`
	Image          *string              `json:"image"`
	Icon           *string              `json:"icon"`
	State          MarketState          `json:"state"`
	Outcomes       MarketOutcomes       `json:"outcomes"`
	Metrics        MarketMetrics        `json:"metrics"`
	Prices         MarketPrices         `json:"prices"`
	Trading        MarketTrading        `json:"trading"`
	Resolution     MarketResolution     `json:"resolution"`
	Rewards        MarketRewards        `json:"rewards"`
	Sports         MarketSportsMetadata `json:"sports"`
	Events         []MarketEvent        `json:"events"`
	Tags           []MarketTag          `json:"tags"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *Market) UnmarshalJSON(data []byte) error {
	var err error
	data, err = normalizeMarket(data)
	if err != nil {
		return err
	}
	type plain Market
	var result plain
	if err = decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "version", kind: "version"},
		{name: "slug"},
		{name: "condition_id", aliases: []string{"conditionId", "condition"}, kind: "condition"},
		{name: "question"},
		{name: "group_item_title", aliases: []string{"groupItemTitle"}},
		{name: "description"},
		{name: "category"},
		{name: "image"},
		{name: "icon"},
		{name: "state", required: true},
		{name: "outcomes", required: true},
		{name: "metrics", required: true},
		{name: "prices", required: true},
		{name: "trading", required: true},
		{name: "resolution", required: true},
		{name: "rewards", required: true},
		{name: "sports", required: true},
		{name: "events", kind: "sequence", required: true},
		{name: "tags", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = Market(result)
	return nil
}

// MarketEvent represents event reference attached to a market.
type MarketEvent struct {
	ID    EventID `json:"id"`
	Slug  *string `json:"slug"`
	Title *string `json:"title"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketEvent) UnmarshalJSON(data []byte) error {
	type plain MarketEvent
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", kind: "id", required: true},
		{name: "slug"},
		{name: "title"},
	}); err != nil {
		return err
	}
	*m = MarketEvent(result)
	return nil
}

// MarketMetrics represents volume and liquidity metrics for a market.
type MarketMetrics struct {
	Volume        *decimal.Big `json:"volume"`
	VolumeNum     *decimal.Big `json:"volume_num"`
	Volume24hr    *decimal.Big `json:"volume_24hr"`
	Volume1wk     *decimal.Big `json:"volume_1wk"`
	Volume1mo     *decimal.Big `json:"volume_1mo"`
	Volume1yr     *decimal.Big `json:"volume_1yr"`
	VolumeAMM     *decimal.Big `json:"volume_amm"`
	VolumeClob    *decimal.Big `json:"volume_clob"`
	Liquidity     *decimal.Big `json:"liquidity"`
	LiquidityNum  *decimal.Big `json:"liquidity_num"`
	LiquidityClob *decimal.Big `json:"liquidity_clob"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketMetrics) UnmarshalJSON(data []byte) error {
	type plain MarketMetrics
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "volume", kind: "decimal"},
		{name: "volume_num", aliases: []string{"volumeNum"}, kind: "decimal"},
		{name: "volume_24hr", aliases: []string{"volume24hr"}, kind: "decimal"},
		{name: "volume_1wk", aliases: []string{"volume1wk"}, kind: "decimal"},
		{name: "volume_1mo", aliases: []string{"volume1mo"}, kind: "decimal"},
		{name: "volume_1yr", aliases: []string{"volume1yr"}, kind: "decimal"},
		{name: "volume_amm", aliases: []string{"volumeAmm"}, kind: "decimal"},
		{name: "volume_clob", aliases: []string{"volumeClob"}, kind: "decimal"},
		{name: "liquidity", kind: "decimal"},
		{name: "liquidity_num", aliases: []string{"liquidityNum"}, kind: "decimal"},
		{name: "liquidity_clob", aliases: []string{"liquidityClob"}, kind: "decimal"},
	}); err != nil {
		return err
	}
	*m = MarketMetrics(result)
	return nil
}

// MarketOutcome represents one tradable outcome in a binary market.
type MarketOutcome struct {
	Label      string       `json:"label"`
	TokenID    *TokenID     `json:"token_id"`
	PositionID *PositionID  `json:"position_id"`
	Price      *decimal.Big `json:"price"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketOutcome) UnmarshalJSON(data []byte) error {
	type plain MarketOutcome
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "label", required: true},
		{name: "token_id", aliases: []string{"tokenId"}},
		{name: "position_id", aliases: []string{"positionId"}},
		{name: "price", kind: "decimal"},
	}); err != nil {
		return err
	}
	*m = MarketOutcome(result)
	return nil
}

// MarketOutcomes represents binary market outcomes.
type MarketOutcomes struct {
	Yes MarketOutcome `json:"yes"`
	No  MarketOutcome `json:"no"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketOutcomes) UnmarshalJSON(data []byte) error {
	type plain MarketOutcomes
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "yes", required: true},
		{name: "no", required: true},
	}); err != nil {
		return err
	}
	*m = MarketOutcomes(result)
	return nil
}

// MarketPrices represents current price and price-change data for a market.
type MarketPrices struct {
	BestBid             *decimal.Big `json:"best_bid"`
	BestAsk             *decimal.Big `json:"best_ask"`
	LastTradePrice      *decimal.Big `json:"last_trade_price"`
	Spread              *decimal.Big `json:"spread"`
	OneHourPriceChange  *decimal.Big `json:"one_hour_price_change"`
	OneDayPriceChange   *decimal.Big `json:"one_day_price_change"`
	OneWeekPriceChange  *decimal.Big `json:"one_week_price_change"`
	OneMonthPriceChange *decimal.Big `json:"one_month_price_change"`
	OneYearPriceChange  *decimal.Big `json:"one_year_price_change"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketPrices) UnmarshalJSON(data []byte) error {
	type plain MarketPrices
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "best_bid", aliases: []string{"bestBid"}, kind: "decimal"},
		{name: "best_ask", aliases: []string{"bestAsk"}, kind: "decimal"},
		{name: "last_trade_price", aliases: []string{"lastTradePrice"}, kind: "decimal"},
		{name: "spread", kind: "decimal"},
		{name: "one_hour_price_change", aliases: []string{"oneHourPriceChange"}, kind: "decimal"},
		{name: "one_day_price_change", aliases: []string{"oneDayPriceChange"}, kind: "decimal"},
		{name: "one_week_price_change", aliases: []string{"oneWeekPriceChange"}, kind: "decimal"},
		{name: "one_month_price_change", aliases: []string{"oneMonthPriceChange"}, kind: "decimal"},
		{name: "one_year_price_change", aliases: []string{"oneYearPriceChange"}, kind: "decimal"},
	}); err != nil {
		return err
	}
	*m = MarketPrices(result)
	return nil
}

// MarketResolution represents resolution metadata for a market.
type MarketResolution struct {
	OnchainEventID      *string                 `json:"onchain_event_id"`
	RequestID           *ResolutionRequestID    `json:"request_id"`
	ResolutionStatus    *MarketResolutionStatus `json:"resolution_status"`
	QuestionID          *QuestionID             `json:"question_id"`
	NegRiskRequestID    *ResolutionRequestID    `json:"neg_risk_request_id"`
	UMAResolutionStatus *UMAResolutionStatus    `json:"uma_resolution_status"`
	Source              *string                 `json:"source"`
	ResolvedBy          *EVMAddress             `json:"resolved_by"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketResolution) UnmarshalJSON(data []byte) error {
	type plain MarketResolution
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "onchain_event_id", aliases: []string{"onchainEventId"}, kind: "event_id"},
		{name: "request_id", aliases: []string{"requestId"}, kind: "request_id"},
		{name: "resolution_status", aliases: []string{"resolutionStatus"}, kind: "resolution"},
		{name: "question_id", aliases: []string{"questionId"}, kind: "empty_text"},
		{name: "neg_risk_request_id", aliases: []string{"negRiskRequestId"}, kind: "empty_text"},
		{name: "uma_resolution_status", aliases: []string{"umaResolutionStatus"}, kind: "uma"},
		{name: "source"},
		{name: "resolved_by", aliases: []string{"resolvedBy"}, kind: "empty_text"},
	}); err != nil {
		return err
	}
	*m = MarketResolution(result)
	return nil
}

// MarketRewards represents reward settings for a market.
type MarketRewards struct {
	ClobRewards           []ClobReward `json:"clob_rewards"`
	RewardsMinSize        *decimal.Big `json:"rewards_min_size"`
	RewardsMaxSpread      *float64     `json:"rewards_max_spread"`
	HoldingRewardsEnabled *bool        `json:"holding_rewards_enabled"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketRewards) UnmarshalJSON(data []byte) error {
	type plain MarketRewards
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "clob_rewards", aliases: []string{"clobRewards"}},
		{name: "rewards_min_size", aliases: []string{"rewardsMinSize"}, kind: "decimal"},
		{name: "rewards_max_spread", aliases: []string{"rewardsMaxSpread"}, kind: "float"},
		{name: "holding_rewards_enabled", aliases: []string{"holdingRewardsEnabled"}, kind: "bool"},
	}); err != nil {
		return err
	}
	*m = MarketRewards(result)
	return nil
}

// MarketSportsMetadata represents sports-specific metadata for a market.
type MarketSportsMetadata struct {
	SportsMarketType *string    `json:"sports_market_type"`
	Line             *float64   `json:"line"`
	GameID           *string    `json:"game_id"`
	GameStartTime    *time.Time `json:"game_start_time"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketSportsMetadata) UnmarshalJSON(data []byte) error {
	type plain MarketSportsMetadata
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "sports_market_type", aliases: []string{"sportsMarketType"}},
		{name: "line", kind: "float"},
		{name: "game_id", aliases: []string{"gameId"}},
		{name: "game_start_time", aliases: []string{"gameStartTime"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = MarketSportsMetadata(result)
	return nil
}

// MarketState represents operational state and timing for a market.
type MarketState struct {
	Active          *bool      `json:"active"`
	Closed          *bool      `json:"closed"`
	Archived        *bool      `json:"archived"`
	ComboStatus     *string    `json:"combo_status"`
	AcceptingOrders *bool      `json:"accepting_orders"`
	EnableOrderBook *bool      `json:"enable_order_book"`
	NegRisk         *bool      `json:"neg_risk"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	ClosedTime      *time.Time `json:"closed_time"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketState) UnmarshalJSON(data []byte) error {
	type plain MarketState
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "active", kind: "bool"},
		{name: "closed", kind: "bool"},
		{name: "archived", kind: "bool"},
		{name: "combo_status", aliases: []string{"comboStatus"}},
		{name: "accepting_orders", aliases: []string{"acceptingOrders"}, kind: "bool"},
		{name: "enable_order_book", aliases: []string{"enableOrderBook"}, kind: "bool"},
		{name: "neg_risk", aliases: []string{"negRisk"}, kind: "bool"},
		{name: "start_date", aliases: []string{"startDate"}, kind: "datetime"},
		{name: "end_date", aliases: []string{"endDate"}, kind: "datetime"},
		{name: "closed_time", aliases: []string{"closedTime"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = MarketState(result)
	return nil
}

// MarketTag represents tag reference attached to a market.
type MarketTag struct {
	ID    TagID   `json:"id"`
	Slug  *string `json:"slug"`
	Label *string `json:"label"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketTag) UnmarshalJSON(data []byte) error {
	type plain MarketTag
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "slug"},
		{name: "label"},
	}); err != nil {
		return err
	}
	*m = MarketTag(result)
	return nil
}

// MarketTrading represents trading configuration and constraints for a market.
type MarketTrading struct {
	MinimumOrderSize *decimal.Big `json:"minimum_order_size"`
	MinimumTickSize  *decimal.Big `json:"minimum_tick_size"`
	SecondsDelay     *int64       `json:"seconds_delay"`
	FeesEnabled      *bool        `json:"fees_enabled"`
	FeeType          *string      `json:"fee_type"`
	FeeSchedule      *FeeSchedule `json:"fee_schedule"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *MarketTrading) UnmarshalJSON(data []byte) error {
	type plain MarketTrading
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "minimum_order_size", aliases: []string{"minimumOrderSize"}, kind: "decimal"},
		{name: "minimum_tick_size", aliases: []string{"minimumTickSize"}, kind: "decimal"},
		{name: "seconds_delay", aliases: []string{"secondsDelay"}, kind: "int"},
		{name: "fees_enabled", aliases: []string{"feesEnabled"}, kind: "bool"},
		{name: "fee_type", aliases: []string{"feeType"}},
		{name: "fee_schedule", aliases: []string{"feeSchedule"}},
	}); err != nil {
		return err
	}
	*m = MarketTrading(result)
	return nil
}
