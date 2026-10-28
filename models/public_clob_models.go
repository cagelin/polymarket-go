// Public CLOB models ported from the pinned Python SDK. This file is linted.

package models

import (
	"encoding/json"
	"time"

	"github.com/ericlagergren/decimal"
)

// CurrentRewardConfig represents a public CLOB or RFQ response.
type CurrentRewardConfig struct {
	ID           *int64       `json:"id"`
	AssetAddress string       `json:"asset_address"`
	StartDate    time.Time    `json:"start_date"`
	EndDate      *time.Time   `json:"end_date"`
	RatePerDay   *decimal.Big `json:"rate_per_day"`
	TotalRewards *decimal.Big `json:"total_rewards"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *CurrentRewardConfig) UnmarshalJSON(raw []byte) error {
	type plain CurrentRewardConfig
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "int", nullable: true},
		{name: "asset_address", aliases: []string{"asset_address"}, required: true},
		{name: "start_date", aliases: []string{"start_date"}, kind: "reward_epoch_ms", required: true},
		{name: "end_date", aliases: []string{"end_date"}, kind: "reward_epoch_ms", nullable: true},
		{name: "rate_per_day", aliases: []string{"rate_per_day"}, kind: "data_decimal", required: true},
		{name: "total_rewards", aliases: []string{"total_rewards"}, kind: "data_decimal", nullable: true},
	}); err != nil {
		return err
	}
	*m = CurrentRewardConfig(result)
	return nil
}

// CurrentReward represents a public CLOB or RFQ response.
type CurrentReward struct {
	ConditionID        ConditionID           `json:"condition_id"`
	RewardsMaxSpread   *float64              `json:"rewards_max_spread"`
	RewardsMinSize     *decimal.Big          `json:"rewards_min_size"`
	RewardsConfig      []CurrentRewardConfig `json:"rewards_config"`
	SponsoredDailyRate *decimal.Big          `json:"sponsored_daily_rate"`
	SponsorsCount      *int64                `json:"sponsors_count"`
	NativeDailyRate    *decimal.Big          `json:"native_daily_rate"`
	TotalDailyRate     *decimal.Big          `json:"total_daily_rate"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *CurrentReward) UnmarshalJSON(raw []byte) error {
	type plain CurrentReward
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", steps: []string{"condition_feed"}, aliases: []string{"condition_id"}, required: true},
		{name: "rewards_max_spread", aliases: []string{"rewards_max_spread"}, kind: "float", nullable: true},
		{name: "rewards_min_size", aliases: []string{"rewards_min_size"}, kind: "data_decimal", nullable: true},
		{name: "rewards_config", aliases: []string{"rewards_config"}, kind: "sequence", fallback: json.RawMessage("[]")},
		{name: "sponsored_daily_rate", aliases: []string{"sponsored_daily_rate"}, kind: "data_decimal", nullable: true},
		{name: "sponsors_count", aliases: []string{"sponsors_count"}, kind: "int", nullable: true},
		{name: "native_daily_rate", aliases: []string{"native_daily_rate"}, kind: "data_decimal", nullable: true},
		{name: "total_daily_rate", aliases: []string{"total_daily_rate"}, kind: "data_decimal", nullable: true},
	}); err != nil {
		return err
	}
	*m = CurrentReward(result)
	return nil
}

// MarketRewardConfig represents a public CLOB or RFQ response.
type MarketRewardConfig struct {
	AssetAddress string       `json:"asset_address"`
	StartDate    time.Time    `json:"start_date"`
	EndDate      *time.Time   `json:"end_date"`
	RatePerDay   *decimal.Big `json:"rate_per_day"`
	TotalRewards *decimal.Big `json:"total_rewards"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *MarketRewardConfig) UnmarshalJSON(raw []byte) error {
	type plain MarketRewardConfig
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_address", aliases: []string{"asset_address"}, required: true},
		{name: "start_date", aliases: []string{"start_date"}, kind: "reward_epoch_ms", required: true},
		{name: "end_date", aliases: []string{"end_date"}, kind: "reward_epoch_ms", nullable: true},
		{name: "rate_per_day", aliases: []string{"rate_per_day"}, kind: "data_decimal", required: true},
		{name: "total_rewards", aliases: []string{"total_rewards"}, kind: "data_decimal", nullable: true},
	}); err != nil {
		return err
	}
	*m = MarketRewardConfig(result)
	return nil
}

// MarketRewardToken represents a public CLOB or RFQ response.
type MarketRewardToken struct {
	AssetID ClobAssetID  `json:"asset_id"`
	Outcome string       `json:"outcome"`
	Price   *decimal.Big `json:"price"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *MarketRewardToken) UnmarshalJSON(raw []byte) error {
	type plain MarketRewardToken
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "outcome", required: true},
		{name: "price", kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = MarketRewardToken(result)
	return nil
}

// MarketReward represents a public CLOB or RFQ response.
type MarketReward struct {
	ConditionID           ConditionID          `json:"condition_id"`
	Question              string               `json:"question"`
	MarketSlug            *string              `json:"market_slug"`
	EventSlug             *string              `json:"event_slug"`
	Image                 *string              `json:"image"`
	RewardsMaxSpread      *float64             `json:"rewards_max_spread"`
	RewardsMinSize        *decimal.Big         `json:"rewards_min_size"`
	MarketCompetitiveness *float64             `json:"market_competitiveness"`
	Tokens                []MarketRewardToken  `json:"tokens"`
	RewardsConfig         []MarketRewardConfig `json:"rewards_config"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *MarketReward) UnmarshalJSON(raw []byte) error {
	type plain MarketReward
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", steps: []string{"condition_feed"}, aliases: []string{"condition_id"}, required: true},
		{name: "question", required: true},
		{name: "market_slug", aliases: []string{"market_slug"}, nullable: true},
		{name: "event_slug", aliases: []string{"event_slug"}, nullable: true},
		{name: "image", nullable: true},
		{name: "rewards_max_spread", aliases: []string{"rewards_max_spread"}, kind: "float", nullable: true},
		{name: "rewards_min_size", aliases: []string{"rewards_min_size"}, kind: "data_decimal", nullable: true},
		{name: "market_competitiveness", aliases: []string{"market_competitiveness"}, kind: "float", nullable: true},
		{name: "tokens", kind: "sequence", required: true},
		{name: "rewards_config", aliases: []string{"rewards_config"}, kind: "sequence", fallback: json.RawMessage("[]")},
	}); err != nil {
		return err
	}
	*m = MarketReward(result)
	return nil
}

// RewardMarket represents a public CLOB or RFQ response.
type RewardMarket struct {
	ConditionID           ConditionID           `json:"condition_id"`
	MarketID              MarketID              `json:"market_id"`
	MarketSlug            string                `json:"market_slug"`
	Question              string                `json:"question"`
	Image                 string                `json:"image"`
	MarketCompetitiveness float64               `json:"market_competitiveness"`
	RewardsConfig         []CurrentRewardConfig `json:"rewards_config"`
	RewardsMaxSpread      float64               `json:"rewards_max_spread"`
	RewardsMinSize        *decimal.Big          `json:"rewards_min_size"`
	Spread                *decimal.Big          `json:"spread"`
	Tokens                []MarketRewardToken   `json:"tokens"`
	GroupItemTitle        string                `json:"group_item_title"`
	Volume24hr            *decimal.Big          `json:"volume_24hr"`
	EventID               EventID               `json:"event_id"`
	EventSlug             string                `json:"event_slug"`
	CreatedAt             time.Time             `json:"created_at"`
	OneDayPriceChange     *decimal.Big          `json:"one_day_price_change"`
	EndDate               *time.Time            `json:"end_date"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *RewardMarket) UnmarshalJSON(raw []byte) error {
	type plain RewardMarket
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", steps: []string{"condition_feed"}, required: true},
		{name: "market_id", required: true},
		{name: "market_slug", required: true},
		{name: "question", required: true},
		{name: "image", required: true},
		{name: "market_competitiveness", kind: "float", required: true},
		{name: "rewards_config", kind: "sequence", required: true},
		{name: "rewards_max_spread", kind: "float", required: true},
		{name: "rewards_min_size", kind: "data_decimal", required: true},
		{name: "spread", kind: "data_decimal", required: true},
		{name: "tokens", kind: "sequence", required: true},
		{name: "group_item_title", required: true},
		{name: "volume_24hr", kind: "data_decimal", required: true},
		{name: "event_id", required: true},
		{name: "event_slug", required: true},
		{name: "created_at", kind: "reward_epoch_ms", required: true},
		{name: "one_day_price_change", kind: "data_decimal", required: true},
		{name: "end_date", kind: "reward_epoch_ms", nullable: true},
	}); err != nil {
		return err
	}
	*m = RewardMarket(result)
	return nil
}

// RebatedFee represents a public CLOB or RFQ response.
type RebatedFee struct {
	Date            Date         `json:"date"`
	ConditionID     ConditionID  `json:"condition_id"`
	AssetAddress    string       `json:"asset_address"`
	MakerAddress    string       `json:"maker_address"`
	RebatedFeesUSDC *decimal.Big `json:"rebated_fees_usdc"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *RebatedFee) UnmarshalJSON(raw []byte) error {
	type plain RebatedFee
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "date", kind: "date", required: true},
		{name: "condition_id", steps: []string{"condition_feed"}, required: true},
		{name: "asset_address", required: true},
		{name: "maker_address", required: true},
		{name: "rebated_fees_usdc", kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = RebatedFee(result)
	return nil
}

// BuilderTrade represents a public CLOB or RFQ response.
type BuilderTrade struct {
	ID              string       `json:"id"`
	TradeType       string       `json:"trade_type"`
	TakerOrderHash  string       `json:"taker_order_hash"`
	Builder         string       `json:"builder"`
	ConditionID     ConditionID  `json:"condition_id"`
	AssetID         ClobAssetID  `json:"asset_id"`
	Side            OrderSide    `json:"side"`
	Size            *decimal.Big `json:"size"`
	SizeUSDC        *decimal.Big `json:"size_usdc"`
	Price           *decimal.Big `json:"price"`
	Status          string       `json:"status"`
	Outcome         string       `json:"outcome"`
	OutcomeIndex    int64        `json:"outcome_index"`
	Owner           string       `json:"owner"`
	Maker           string       `json:"maker"`
	TransactionHash string       `json:"transaction_hash"`
	MatchedAt       time.Time    `json:"matched_at"`
	BucketIndex     int64        `json:"bucket_index"`
	Fee             *decimal.Big `json:"fee"`
	FeeUSDC         *decimal.Big `json:"fee_usdc"`
	ErrorMsg        *string      `json:"error_msg"`
	CreatedAt       *time.Time   `json:"created_at"`
	UpdatedAt       *time.Time   `json:"updated_at"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *BuilderTrade) UnmarshalJSON(raw []byte) error {
	type plain BuilderTrade
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "trade_type", aliases: []string{"tradeType"}, required: true},
		{name: "taker_order_hash", aliases: []string{"takerOrderHash"}, required: true},
		{name: "builder", required: true},
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "asset_id", aliases: []string{"asset_id", "assetId", "token_id"}, required: true},
		{name: "side", required: true, allowed: []string{"BUY", "SELL"}},
		{name: "size", kind: "string_decimal", required: true},
		{name: "size_usdc", aliases: []string{"sizeUsdc"}, kind: "string_decimal", required: true},
		{name: "price", kind: "string_decimal", required: true},
		{name: "status", required: true},
		{name: "outcome", required: true},
		{name: "outcome_index", aliases: []string{"outcomeIndex"}, kind: "int", required: true},
		{name: "owner", required: true},
		{name: "maker", required: true},
		{name: "transaction_hash", aliases: []string{"transactionHash"}, required: true},
		{name: "matched_at", aliases: []string{"matchTime"}, kind: "clob_epoch_iso", required: true},
		{name: "bucket_index", aliases: []string{"bucketIndex"}, kind: "int", required: true},
		{name: "fee", kind: "string_decimal", required: true},
		{name: "fee_usdc", aliases: []string{"feeUsdc"}, kind: "string_decimal", required: true},
		{name: "error_msg", aliases: []string{"err_msg"}, nullable: true},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "clob_epoch_iso", nullable: true},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "clob_epoch_iso", nullable: true},
	}); err != nil {
		return err
	}
	*m = BuilderTrade(result)
	return nil
}

// ComboMarketOutcome represents a public CLOB or RFQ response.
type ComboMarketOutcome struct {
	Label      string       `json:"label"`
	PositionID PositionID   `json:"position_id"`
	Price      *decimal.Big `json:"price"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *ComboMarketOutcome) UnmarshalJSON(raw []byte) error {
	type plain ComboMarketOutcome
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "label", required: true},
		{name: "position_id", required: true},
		{name: "price", kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = ComboMarketOutcome(result)
	return nil
}

// ComboMarketOutcomes represents a public CLOB or RFQ response.
type ComboMarketOutcomes struct {
	Yes ComboMarketOutcome `json:"yes"`
	No  ComboMarketOutcome `json:"no"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *ComboMarketOutcomes) UnmarshalJSON(raw []byte) error {
	type plain ComboMarketOutcomes
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "yes", required: true},
		{name: "no", required: true},
	}); err != nil {
		return err
	}
	*m = ComboMarketOutcomes(result)
	return nil
}

// ComboMarket represents a public CLOB or RFQ response.
type ComboMarket struct {
	ID          MarketID            `json:"id"`
	ConditionID ConditionID         `json:"condition_id"`
	Slug        string              `json:"slug"`
	Title       string              `json:"title"`
	Outcomes    ComboMarketOutcomes `json:"outcomes"`
	Image       string              `json:"image"`
	Volume      float64             `json:"volume"`
	Tags        []string            `json:"tags"`
}

// UnmarshalJSON validates response values and wire aliases.
func (m *ComboMarket) UnmarshalJSON(raw []byte) error {
	normalized, err := normalizeComboMarket(raw)
	if err != nil {
		return err
	}
	raw = normalized
	type plain ComboMarket
	var result plain
	if err = decodeData(raw, &result, []dataField{
		{name: "id", kind: "id", required: true},
		{name: "condition_id", steps: []string{"condition_feed"}, required: true},
		{name: "slug", required: true},
		{name: "title", required: true},
		{name: "outcomes", required: true},
		{name: "image", required: true},
		{name: "volume", kind: "float", required: true},
		{name: "tags", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = ComboMarket(result)
	return nil
}
