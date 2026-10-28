// Data models ported from polymarket-client 0.14.0. This file is linted.

package models

import (
	"encoding/json"
	"time"

	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/ericlagergren/decimal"
)

// ComboActivityID identifies a Data API resource.
type ComboActivityID string

// ComboConditionID identifies a Data API resource.
type ComboConditionID string

// HexString identifies a Data API resource.
type HexString string

// TransactionHash identifies a Data API resource.
type TransactionHash string

// Position represents A wallet position. Sizes are shares; prices, costs, values and PnL are USDC.
type Position struct {
	Wallet             EVMAddress     `json:"wallet"`
	AssetID            ClobAssetID    `json:"asset_id"`
	ConditionID        ConditionID    `json:"condition_id"`
	CurrentSize        *decimal.Big   `json:"current_size"`
	AvgPrice           *decimal.Big   `json:"avg_price"`
	EntryCostUSDC      *decimal.Big   `json:"entry_cost_usdc"`
	EntryFeesUSDC      *decimal.Big   `json:"entry_fees_usdc"`
	TotalCostUSDC      *decimal.Big   `json:"total_cost_usdc"`
	CurrentPrice       *decimal.Big   `json:"current_price"`
	CurrentValue       *decimal.Big   `json:"current_value"`
	TotalSize          *decimal.Big   `json:"total_size"`
	RealizedPnL        *decimal.Big   `json:"realized_pnl"`
	UnrealizedPnL      *decimal.Big   `json:"unrealized_pnl"`
	TotalPnL           *decimal.Big   `json:"total_pnl"`
	PercentPnL         *decimal.Big   `json:"percent_pnl"`
	PercentRealizedPnL *decimal.Big   `json:"percent_realized_pnl"`
	Status             PositionStatus `json:"status"`
	Redeemable         bool           `json:"redeemable"`
	Mergeable          bool           `json:"mergeable"`
	NegativeRisk       bool           `json:"negative_risk"`
	Archived           bool           `json:"archived"`
	Verified           bool           `json:"verified"`
	Title              *string        `json:"title"`
	Slug               *string        `json:"slug"`
	Icon               *string        `json:"icon"`
	EventSlug          *string        `json:"event_slug"`
	Outcome            *string        `json:"outcome"`
	OppositeOutcome    *string        `json:"opposite_outcome"`
	Name               *string        `json:"name"`
	ProfileImage       *string        `json:"profile_image"`
	EventID            *EventID       `json:"event_id"`
	OutcomeIndex       *int64         `json:"outcome_index"`
	OppositeAssetID    *ClobAssetID   `json:"opposite_asset_id"`
	EndDate            *Date          `json:"end_date"`
	LastEventAt        *time.Time     `json:"last_event_at"`
	FirstEntryAt       *time.Time     `json:"first_entry_at"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *Position) UnmarshalJSON(raw []byte) error {
	type plain Position
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "asset_id", aliases: []string{"token_id"}, required: true},
		{name: "condition_id", required: true, steps: []string{"validate_condition_id_response"}},
		{name: "current_size", kind: "data_decimal", required: true},
		{name: "avg_price", kind: "data_decimal", required: true},
		{name: "entry_cost_usdc", kind: "data_decimal", required: true},
		{name: "entry_fees_usdc", kind: "data_decimal", required: true},
		{name: "total_cost_usdc", kind: "data_decimal", required: true},
		{name: "current_price", kind: "data_decimal", required: true},
		{name: "current_value", kind: "data_decimal", required: true},
		{name: "total_size", kind: "data_decimal", required: true},
		{name: "realized_pnl", kind: "data_decimal", required: true},
		{name: "unrealized_pnl", kind: "data_decimal", required: true},
		{name: "total_pnl", kind: "data_decimal", required: true},
		{name: "percent_pnl", kind: "data_decimal", required: true},
		{name: "percent_realized_pnl", kind: "data_decimal", required: true},
		{name: "status", required: true, allowed: []string{"OPEN", "REDEEMABLE", "CLOSED"}},
		{name: "redeemable", kind: "bool", required: true},
		{name: "mergeable", kind: "bool", required: true},
		{name: "negative_risk", kind: "bool", required: true},
		{name: "archived", kind: "bool", required: true},
		{name: "verified", kind: "bool", required: true},
		{name: "title", nullable: true, steps: []string{"optional_text"}},
		{name: "slug", nullable: true, steps: []string{"optional_text"}},
		{name: "icon", nullable: true, steps: []string{"optional_text"}},
		{name: "event_slug", nullable: true, steps: []string{"optional_text"}},
		{name: "outcome", nullable: true, steps: []string{"optional_text"}},
		{name: "opposite_outcome", nullable: true, steps: []string{"optional_text"}},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", nullable: true, steps: []string{"optional_text"}},
		{name: "event_id", nullable: true, steps: []string{"optional_event_id"}},
		{name: "outcome_index", kind: "int", nullable: true, steps: []string{"optional_outcome_index"}},
		{name: "opposite_asset_id", aliases: []string{"opposite_token_id"}, nullable: true, steps: []string{"optional_text"}},
		{name: "end_date", kind: "date", nullable: true, steps: []string{"date_from_calendar_string"}},
		{name: "last_event_at", kind: "optional_epoch", nullable: true},
		{name: "first_entry_at", kind: "optional_epoch", nullable: true},
	}); err != nil {
		return err
	}
	*m = Position(result)
	return nil
}

// ComboPositionMarketEvent represents a Data API response.
type ComboPositionMarketEvent struct {
	EventID    *EventID `json:"event_id"`
	EventSlug  *string  `json:"event_slug"`
	EventTitle *string  `json:"event_title"`
	EventImage *string  `json:"event_image"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboPositionMarketEvent) UnmarshalJSON(raw []byte) error {
	type plain ComboPositionMarketEvent
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "event_id", nullable: true, steps: []string{"optional_event_id"}},
		{name: "event_slug", nullable: true, steps: []string{"optional_text"}},
		{name: "event_title", nullable: true, steps: []string{"optional_text"}},
		{name: "event_image", nullable: true, steps: []string{"optional_text"}},
	}); err != nil {
		return err
	}
	*m = ComboPositionMarketEvent(result)
	return nil
}

// ComboPositionMarket represents a Data API response.
type ComboPositionMarket struct {
	MarketID         *MarketID                 `json:"market_id"`
	Slug             *string                   `json:"slug"`
	Title            *string                   `json:"title"`
	Question         *string                   `json:"question"`
	GroupItemTitle   *string                   `json:"group_item_title"`
	SportsMarketType *string                   `json:"sports_market_type"`
	Line             *decimal.Big              `json:"line"`
	Outcomes         []string                  `json:"outcomes"`
	Outcome          *string                   `json:"outcome"`
	ImageURL         *string                   `json:"image_url"`
	IconURL          *string                   `json:"icon_url"`
	Category         *string                   `json:"category"`
	Subcategory      *string                   `json:"subcategory"`
	Tags             []string                  `json:"tags"`
	EndDate          *time.Time                `json:"end_date"`
	Event            *ComboPositionMarketEvent `json:"event"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboPositionMarket) UnmarshalJSON(raw []byte) error {
	type plain ComboPositionMarket
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "market_id", nullable: true},
		{name: "slug", nullable: true, steps: []string{"optional_text"}},
		{name: "title", nullable: true, steps: []string{"optional_text"}},
		{name: "question", nullable: true, steps: []string{"optional_text"}},
		{name: "group_item_title", nullable: true, steps: []string{"optional_text"}},
		{name: "sports_market_type", nullable: true, steps: []string{"optional_text"}},
		{name: "line", kind: "data_decimal", nullable: true},
		{name: "outcomes", kind: "sequence", nullable: true},
		{name: "outcome", nullable: true, steps: []string{"optional_text"}},
		{name: "image_url", nullable: true, steps: []string{"optional_text"}},
		{name: "icon_url", nullable: true, steps: []string{"optional_text"}},
		{name: "category", nullable: true, steps: []string{"optional_text"}},
		{name: "subcategory", nullable: true, steps: []string{"optional_text"}},
		{name: "tags", kind: "sequence", nullable: true},
		{name: "end_date", kind: "optional_epoch_or_iso", nullable: true},
		{name: "event", nullable: true},
	}); err != nil {
		return err
	}
	*m = ComboPositionMarket(result)
	return nil
}

// ComboPositionLeg represents A combo leg; leg_current_price is USDC per share.
type ComboPositionLeg struct {
	LegIndex        int64                `json:"leg_index"`
	LegPositionID   PositionID           `json:"leg_position_id"`
	LegConditionID  ConditionID          `json:"leg_condition_id"`
	LegOutcomeIndex int64                `json:"leg_outcome_index"`
	LegOutcomeLabel *string              `json:"leg_outcome_label"`
	LegStatus       ComboPositionStatus  `json:"leg_status"`
	LegResolvedAt   *time.Time           `json:"leg_resolved_at"`
	LegCurrentPrice *decimal.Big         `json:"leg_current_price"`
	Market          *ComboPositionMarket `json:"market"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboPositionLeg) UnmarshalJSON(raw []byte) error {
	type plain ComboPositionLeg
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "leg_index", kind: "int", required: true},
		{name: "leg_position_id", required: true},
		{name: "leg_condition_id", required: true, steps: []string{"validate_condition_id_response"}},
		{name: "leg_outcome_index", kind: "int", required: true},
		{name: "leg_outcome_label", nullable: true, steps: []string{"optional_text"}},
		{name: "leg_status", required: true, allowed: []string{"OPEN", "REDEEMABLE", "PARTIAL", "RESOLVED_PARTIAL", "RESOLVED_WIN", "RESOLVED_LOSS"}},
		{name: "leg_resolved_at", kind: "epoch_datetime", nullable: true},
		{name: "leg_current_price", kind: "data_decimal", nullable: true},
		{name: "market", nullable: true},
	}); err != nil {
		return err
	}
	*m = ComboPositionLeg(result)
	return nil
}

// ComboPosition represents A combo position. current_size is shares; prices, costs and payouts are USDC.
type ComboPosition struct {
	ConditionID        ComboConditionID    `json:"condition_id"`
	PositionID         PositionID          `json:"position_id"`
	Wallet             EVMAddress          `json:"wallet"`
	OutcomeIndex       int64               `json:"outcome_index"`
	OutcomeLabel       string              `json:"outcome_label"`
	CurrentSize        *decimal.Big        `json:"current_size"`
	EntryAvgPriceUSDC  *decimal.Big        `json:"entry_avg_price_usdc"`
	EntryCostUSDC      *decimal.Big        `json:"entry_cost_usdc"`
	GrossEntryCostUSDC *decimal.Big        `json:"gross_entry_cost_usdc"`
	EntryFeesUSDC      *decimal.Big        `json:"entry_fees_usdc"`
	RealizedPayoutUSDC *decimal.Big        `json:"realized_payout_usdc"`
	Status             ComboPositionStatus `json:"status"`
	Redeemable         bool                `json:"redeemable"`
	FirstEntryAt       time.Time           `json:"first_entry_at"`
	ResolvedAt         *time.Time          `json:"resolved_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	LegsTotal          int64               `json:"legs_total"`
	LegsResolved       int64               `json:"legs_resolved"`
	LegsPending        int64               `json:"legs_pending"`
	Legs               []ComboPositionLeg  `json:"legs"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboPosition) UnmarshalJSON(raw []byte) error {
	type plain ComboPosition
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", aliases: []string{"combo_condition_id"}, required: true, steps: []string{"validate_combo_condition_id"}},
		{name: "position_id", aliases: []string{"combo_position_id"}, required: true},
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "outcome_index", kind: "int", required: true},
		{name: "outcome_label", required: true},
		{name: "current_size", kind: "data_decimal", required: true},
		{name: "entry_avg_price_usdc", kind: "data_decimal", required: true},
		{name: "entry_cost_usdc", kind: "data_decimal", required: true},
		{name: "gross_entry_cost_usdc", kind: "data_decimal", required: true},
		{name: "entry_fees_usdc", kind: "data_decimal", required: true},
		{name: "realized_payout_usdc", kind: "data_decimal", required: true},
		{name: "status", required: true, allowed: []string{"OPEN", "REDEEMABLE", "PARTIAL", "RESOLVED_PARTIAL", "RESOLVED_WIN", "RESOLVED_LOSS"}},
		{name: "redeemable", kind: "bool", required: true},
		{name: "first_entry_at", kind: "epoch_datetime", required: true},
		{name: "resolved_at", kind: "epoch_datetime", nullable: true},
		{name: "updated_at", kind: "epoch_datetime", required: true},
		{name: "legs_total", kind: "int", required: true},
		{name: "legs_resolved", kind: "int", required: true},
		{name: "legs_pending", kind: "int", required: true},
		{name: "legs", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = ComboPosition(result)
	return nil
}

// PortfolioValue represents Current portfolio value in USDC.
type PortfolioValue struct {
	Wallet EVMAddress   `json:"wallet"`
	Value  *decimal.Big `json:"value"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *PortfolioValue) UnmarshalJSON(raw []byte) error {
	type plain PortfolioValue
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "value", kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = PortfolioValue(result)
	return nil
}

// UserPnLPoint represents Cumulative wallet metrics. volume is shares; all monetary amounts are USDC.
type UserPnLPoint struct {
	Timestamp         time.Time    `json:"timestamp"`
	SourceBlock       int64        `json:"source_block"`
	TradeCount        int64        `json:"trade_count"`
	RealizedMarketPnL *decimal.Big `json:"realized_market_pnl"`
	RealizedLpPnL     *decimal.Big `json:"realized_lp_pnl"`
	RealizedComboPnL  *decimal.Big `json:"realized_combo_pnl"`
	RealizedPnL       *decimal.Big `json:"realized_pnl"`
	Volume            *decimal.Big `json:"volume"`
	VolumeUSDC        *decimal.Big `json:"volume_usdc"`
	UnrealizedPnL     *decimal.Big `json:"unrealized_pnl"`
	FeesRefunded      *decimal.Big `json:"fees_refunded"`
	MakerRebate       *decimal.Big `json:"maker_rebate"`
	TakerRebate       *decimal.Big `json:"taker_rebate"`
	RewardIncome      *decimal.Big `json:"reward_income"`
	YieldIncome       *decimal.Big `json:"yield_income"`
	ReferralIncome    *decimal.Big `json:"referral_income"`
	SponsoredIncome   *decimal.Big `json:"sponsored_income"`
	Deposits          *decimal.Big `json:"deposits"`
	Withdrawals       *decimal.Big `json:"withdrawals"`
	CashflowNet       *decimal.Big `json:"cashflow_net"`
	WalletIncome      *decimal.Big `json:"wallet_income"`
	PositionPnL       *decimal.Big `json:"position_pnl"`
	SettledPnL        *decimal.Big `json:"settled_pnl"`
	EconomicPnL       *decimal.Big `json:"economic_pnl"`
	TradePnL          *decimal.Big `json:"trade_pnl"`
	Fees              *decimal.Big `json:"fees"`
	FeesPaid          *decimal.Big `json:"fees_paid"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *UserPnLPoint) UnmarshalJSON(raw []byte) error {
	type plain UserPnLPoint
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "timestamp", kind: "epoch", required: true},
		{name: "source_block", kind: "int", required: true},
		{name: "trade_count", kind: "int", required: true},
		{name: "realized_market_pnl", kind: "data_decimal", required: true},
		{name: "realized_lp_pnl", kind: "data_decimal", required: true},
		{name: "realized_combo_pnl", kind: "data_decimal", required: true},
		{name: "realized_pnl", kind: "data_decimal", required: true},
		{name: "volume", kind: "data_decimal", required: true},
		{name: "volume_usdc", kind: "data_decimal", required: true},
		{name: "unrealized_pnl", kind: "data_decimal", nullable: true},
		{name: "fees_refunded", kind: "data_decimal", nullable: true},
		{name: "maker_rebate", kind: "data_decimal", nullable: true},
		{name: "taker_rebate", kind: "data_decimal", nullable: true},
		{name: "reward_income", kind: "data_decimal", nullable: true},
		{name: "yield_income", kind: "data_decimal", nullable: true},
		{name: "referral_income", kind: "data_decimal", nullable: true},
		{name: "sponsored_income", kind: "data_decimal", nullable: true},
		{name: "deposits", kind: "data_decimal", nullable: true},
		{name: "withdrawals", kind: "data_decimal", nullable: true},
		{name: "cashflow_net", kind: "data_decimal", nullable: true},
		{name: "wallet_income", kind: "data_decimal", nullable: true},
		{name: "position_pnl", kind: "data_decimal", nullable: true},
		{name: "settled_pnl", kind: "data_decimal", nullable: true},
		{name: "economic_pnl", kind: "data_decimal", nullable: true},
		{name: "trade_pnl", kind: "data_decimal", nullable: true},
		{name: "fees", kind: "data_decimal", nullable: true},
		{name: "fees_paid", kind: "data_decimal", nullable: true},
	}); err != nil {
		return err
	}
	*m = UserPnLPoint(result)
	return nil
}

// UserStats represents Wallet statistics; biggest_win is USDC.
type UserStats struct {
	Wallet            EVMAddress    `json:"wallet"`
	TradedMarketCount int64         `json:"traded_market_count"`
	BiggestWin        *decimal.Big  `json:"biggest_win"`
	Views             int64         `json:"views"`
	JoinDate          *time.Time    `json:"join_date"`
	AllTimePnL        *UserPnLPoint `json:"all_time_pnl"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *UserStats) UnmarshalJSON(raw []byte) error {
	type plain UserStats
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "traded_market_count", aliases: []string{"trades"}, kind: "int", required: true},
		{name: "biggest_win", kind: "data_decimal", required: true},
		{name: "views", kind: "int", required: true},
		{name: "join_date", kind: "optional_epoch", nullable: true},
		{name: "all_time_pnl", nullable: true},
	}); err != nil {
		return err
	}
	*m = UserStats(result)
	return nil
}

// UserPnLSeries represents A cumulative PnL time series for a wallet.
type UserPnLSeries struct {
	Wallet         EVMAddress      `json:"wallet"`
	Interval       UserPnLInterval `json:"interval"`
	Fidelity       UserPnLFidelity `json:"fidelity"`
	SourceFidelity UserPnLFidelity `json:"source_fidelity"`
	Points         []UserPnLPoint  `json:"points"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *UserPnLSeries) UnmarshalJSON(raw []byte) error {
	type plain UserPnLSeries
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "interval", required: true, allowed: []string{"max", "all", "1m", "1w", "1d", "12h", "6h"}},
		{name: "fidelity", required: true, allowed: []string{"1d", "18h", "12h", "3h", "1h"}},
		{name: "source_fidelity", required: true, allowed: []string{"1d", "18h", "12h", "3h", "1h"}},
		{name: "points", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = UserPnLSeries(result)
	return nil
}

// UserVolume represents Trading volume in shares and USDC, with a trade count.
type UserVolume struct {
	Volume     *decimal.Big `json:"volume"`
	VolumeUSDC *decimal.Big `json:"volume_usdc"`
	TradeCount int64        `json:"trade_count"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *UserVolume) UnmarshalJSON(raw []byte) error {
	type plain UserVolume
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "volume", kind: "data_decimal", required: true},
		{name: "volume_usdc", kind: "data_decimal", required: true},
		{name: "trade_count", kind: "int", required: true},
	}); err != nil {
		return err
	}
	*m = UserVolume(result)
	return nil
}

// ComboCompressActivity represents a Data API response.
type ComboCompressActivity struct {
	ID              ComboActivityID    `json:"id"`
	Wallet          EVMAddress         `json:"wallet"`
	ConditionID     ComboConditionID   `json:"condition_id"`
	PositionID      PositionID         `json:"position_id"`
	Amount          *decimal.Big       `json:"amount"`
	Timestamp       time.Time          `json:"timestamp"`
	TransactionHash TransactionHash    `json:"transaction_hash"`
	BlockNumber     int64              `json:"block_number"`
	Legs            []ComboPositionLeg `json:"legs"`
	Type            ComboActivityType  `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboCompressActivity) UnmarshalJSON(raw []byte) error {
	type plain ComboCompressActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "condition_id", aliases: []string{"combo_condition_id"}, required: true, steps: []string{"validate_combo_condition_id"}},
		{name: "position_id", aliases: []string{"combo_position_id"}, required: true},
		{name: "amount", aliases: []string{"amount_usdc"}, kind: "data_decimal", nullable: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "block_number", kind: "int", required: true},
		{name: "legs", kind: "sequence", required: true},
		{name: "type", required: true, allowed: []string{"COMPRESS"}},
	}); err != nil {
		return err
	}
	*m = ComboCompressActivity(result)
	return nil
}

// ComboConvertActivity represents a Data API response.
type ComboConvertActivity struct {
	ID              ComboActivityID    `json:"id"`
	Wallet          EVMAddress         `json:"wallet"`
	ConditionID     ComboConditionID   `json:"condition_id"`
	PositionID      PositionID         `json:"position_id"`
	Amount          *decimal.Big       `json:"amount"`
	Timestamp       time.Time          `json:"timestamp"`
	TransactionHash TransactionHash    `json:"transaction_hash"`
	BlockNumber     int64              `json:"block_number"`
	Legs            []ComboPositionLeg `json:"legs"`
	Type            ComboActivityType  `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboConvertActivity) UnmarshalJSON(raw []byte) error {
	type plain ComboConvertActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "condition_id", aliases: []string{"combo_condition_id"}, required: true, steps: []string{"validate_combo_condition_id"}},
		{name: "position_id", aliases: []string{"combo_position_id"}, required: true},
		{name: "amount", aliases: []string{"amount_usdc"}, kind: "data_decimal", nullable: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "block_number", kind: "int", required: true},
		{name: "legs", kind: "sequence", required: true},
		{name: "type", required: true, allowed: []string{"CONVERT"}},
	}); err != nil {
		return err
	}
	*m = ComboConvertActivity(result)
	return nil
}

// ComboMergeActivity represents a Data API response.
type ComboMergeActivity struct {
	ID              ComboActivityID    `json:"id"`
	Wallet          EVMAddress         `json:"wallet"`
	ConditionID     ComboConditionID   `json:"condition_id"`
	PositionID      PositionID         `json:"position_id"`
	Amount          *decimal.Big       `json:"amount"`
	Timestamp       time.Time          `json:"timestamp"`
	TransactionHash TransactionHash    `json:"transaction_hash"`
	BlockNumber     int64              `json:"block_number"`
	Legs            []ComboPositionLeg `json:"legs"`
	Type            ComboActivityType  `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboMergeActivity) UnmarshalJSON(raw []byte) error {
	type plain ComboMergeActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "condition_id", aliases: []string{"combo_condition_id"}, required: true, steps: []string{"validate_combo_condition_id"}},
		{name: "position_id", aliases: []string{"combo_position_id"}, required: true},
		{name: "amount", aliases: []string{"amount_usdc"}, kind: "data_decimal", nullable: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "block_number", kind: "int", required: true},
		{name: "legs", kind: "sequence", required: true},
		{name: "type", required: true, allowed: []string{"MERGE"}},
	}); err != nil {
		return err
	}
	*m = ComboMergeActivity(result)
	return nil
}

// ComboRedeemActivity represents a Data API response.
type ComboRedeemActivity struct {
	ID              ComboActivityID    `json:"id"`
	Wallet          EVMAddress         `json:"wallet"`
	ConditionID     ComboConditionID   `json:"condition_id"`
	PositionID      PositionID         `json:"position_id"`
	Amount          *decimal.Big       `json:"amount"`
	Timestamp       time.Time          `json:"timestamp"`
	TransactionHash TransactionHash    `json:"transaction_hash"`
	BlockNumber     int64              `json:"block_number"`
	Legs            []ComboPositionLeg `json:"legs"`
	Type            ComboActivityType  `json:"type"`
	Payout          *decimal.Big       `json:"payout"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboRedeemActivity) UnmarshalJSON(raw []byte) error {
	type plain ComboRedeemActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "condition_id", aliases: []string{"combo_condition_id"}, required: true, steps: []string{"validate_combo_condition_id"}},
		{name: "position_id", aliases: []string{"combo_position_id"}, required: true},
		{name: "amount", aliases: []string{"amount_usdc"}, kind: "data_decimal", nullable: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "block_number", kind: "int", required: true},
		{name: "legs", kind: "sequence", required: true},
		{name: "type", required: true, allowed: []string{"REDEEM"}},
		{name: "payout", aliases: []string{"payout_usdc"}, kind: "data_decimal", nullable: true},
	}); err != nil {
		return err
	}
	*m = ComboRedeemActivity(result)
	return nil
}

// ComboSplitActivity represents a Data API response.
type ComboSplitActivity struct {
	ID              ComboActivityID    `json:"id"`
	Wallet          EVMAddress         `json:"wallet"`
	ConditionID     ComboConditionID   `json:"condition_id"`
	PositionID      PositionID         `json:"position_id"`
	Amount          *decimal.Big       `json:"amount"`
	Timestamp       time.Time          `json:"timestamp"`
	TransactionHash TransactionHash    `json:"transaction_hash"`
	BlockNumber     int64              `json:"block_number"`
	Legs            []ComboPositionLeg `json:"legs"`
	Type            ComboActivityType  `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboSplitActivity) UnmarshalJSON(raw []byte) error {
	type plain ComboSplitActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "condition_id", aliases: []string{"combo_condition_id"}, required: true, steps: []string{"validate_combo_condition_id"}},
		{name: "position_id", aliases: []string{"combo_position_id"}, required: true},
		{name: "amount", aliases: []string{"amount_usdc"}, kind: "data_decimal", nullable: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "block_number", kind: "int", required: true},
		{name: "legs", kind: "sequence", required: true},
		{name: "type", required: true, allowed: []string{"SPLIT"}},
	}); err != nil {
		return err
	}
	*m = ComboSplitActivity(result)
	return nil
}

// ComboTradeActivity represents a Data API response.
type ComboTradeActivity struct {
	Wallet                EVMAddress       `json:"wallet"`
	Timestamp             time.Time        `json:"timestamp"`
	TransactionHash       TransactionHash  `json:"transaction_hash"`
	Name                  *string          `json:"name"`
	Pseudonym             *string          `json:"pseudonym"`
	Bio                   *string          `json:"bio"`
	ProfileImage          *string          `json:"profile_image"`
	ProfileImageOptimized *string          `json:"profile_image_optimized"`
	Type                  ActivityType     `json:"type"`
	IsCombo               bool             `json:"is_combo"`
	ConditionID           ComboConditionID `json:"condition_id"`
	PositionID            PositionID       `json:"position_id"`
	Side                  OrderSide        `json:"side"`
	Shares                *decimal.Big     `json:"shares"`
	Amount                *decimal.Big     `json:"amount"`
	Price                 *decimal.Big     `json:"price"`
	Title                 *string          `json:"title"`
	Icon                  *string          `json:"icon"`
	Outcome               *string          `json:"outcome"`
	OutcomeIndex          *int64           `json:"outcome_index"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboTradeActivity) UnmarshalJSON(raw []byte) error {
	type plain ComboTradeActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "type", required: true, allowed: []string{"TRADE"}},
		{name: "is_combo", aliases: []string{"is_combo"}, kind: "literal_bool_true", required: true},
		{name: "condition_id", aliases: []string{"condition_id"}, required: true, steps: []string{"validate_combo_condition_id"}},
		{name: "position_id", aliases: []string{"token_id"}, required: true},
		{name: "side", required: true, allowed: []string{"BUY", "SELL"}},
		{name: "shares", aliases: []string{"size"}, kind: "data_decimal", required: true},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "price", kind: "data_decimal", required: true},
		{name: "title", nullable: true, steps: []string{"optional_text"}},
		{name: "icon", nullable: true, steps: []string{"optional_text"}},
		{name: "outcome", nullable: true, steps: []string{"optional_text"}},
		{name: "outcome_index", kind: "int", nullable: true, steps: []string{"optional_outcome_index"}},
	}); err != nil {
		return err
	}
	*m = ComboTradeActivity(result)
	return nil
}

// ComboUnwrapActivity represents a Data API response.
type ComboUnwrapActivity struct {
	ID              ComboActivityID    `json:"id"`
	Wallet          EVMAddress         `json:"wallet"`
	ConditionID     ComboConditionID   `json:"condition_id"`
	PositionID      PositionID         `json:"position_id"`
	Amount          *decimal.Big       `json:"amount"`
	Timestamp       time.Time          `json:"timestamp"`
	TransactionHash TransactionHash    `json:"transaction_hash"`
	BlockNumber     int64              `json:"block_number"`
	Legs            []ComboPositionLeg `json:"legs"`
	Type            ComboActivityType  `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboUnwrapActivity) UnmarshalJSON(raw []byte) error {
	type plain ComboUnwrapActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "condition_id", aliases: []string{"combo_condition_id"}, required: true, steps: []string{"validate_combo_condition_id"}},
		{name: "position_id", aliases: []string{"combo_position_id"}, required: true},
		{name: "amount", aliases: []string{"amount_usdc"}, kind: "data_decimal", nullable: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "block_number", kind: "int", required: true},
		{name: "legs", kind: "sequence", required: true},
		{name: "type", required: true, allowed: []string{"UNWRAP"}},
	}); err != nil {
		return err
	}
	*m = ComboUnwrapActivity(result)
	return nil
}

// ComboWrapActivity represents a Data API response.
type ComboWrapActivity struct {
	ID              ComboActivityID    `json:"id"`
	Wallet          EVMAddress         `json:"wallet"`
	ConditionID     ComboConditionID   `json:"condition_id"`
	PositionID      PositionID         `json:"position_id"`
	Amount          *decimal.Big       `json:"amount"`
	Timestamp       time.Time          `json:"timestamp"`
	TransactionHash TransactionHash    `json:"transaction_hash"`
	BlockNumber     int64              `json:"block_number"`
	Legs            []ComboPositionLeg `json:"legs"`
	Type            ComboActivityType  `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboWrapActivity) UnmarshalJSON(raw []byte) error {
	type plain ComboWrapActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "condition_id", aliases: []string{"combo_condition_id"}, required: true, steps: []string{"validate_combo_condition_id"}},
		{name: "position_id", aliases: []string{"combo_position_id"}, required: true},
		{name: "amount", aliases: []string{"amount_usdc"}, kind: "data_decimal", nullable: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "block_number", kind: "int", required: true},
		{name: "legs", kind: "sequence", required: true},
		{name: "type", required: true, allowed: []string{"WRAP"}},
	}); err != nil {
		return err
	}
	*m = ComboWrapActivity(result)
	return nil
}

// ConversionActivity represents a Data API response.
type ConversionActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	ConditionID           ConditionID     `json:"condition_id"`
	Amount                *decimal.Big    `json:"amount"`
	Title                 *string         `json:"title"`
	Slug                  *string         `json:"slug"`
	Icon                  *string         `json:"icon"`
	EventSlug             *string         `json:"event_slug"`
	Type                  ActivityType    `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ConversionActivity) UnmarshalJSON(raw []byte) error {
	type plain ConversionActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "condition_id", aliases: []string{"condition_id"}, required: true, steps: []string{"validate_condition_id_response"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "title", nullable: true, steps: []string{"optional_text"}},
		{name: "slug", nullable: true, steps: []string{"optional_text"}},
		{name: "icon", nullable: true, steps: []string{"optional_text"}},
		{name: "event_slug", nullable: true, steps: []string{"optional_text"}},
		{name: "type", required: true, allowed: []string{"CONVERSION"}},
	}); err != nil {
		return err
	}
	*m = ConversionActivity(result)
	return nil
}

// DepositActivity represents a Data API response.
type DepositActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	Amount                *decimal.Big    `json:"amount"`
	Type                  ActivityType    `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *DepositActivity) UnmarshalJSON(raw []byte) error {
	type plain DepositActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "type", required: true, allowed: []string{"DEPOSIT"}},
	}); err != nil {
		return err
	}
	*m = DepositActivity(result)
	return nil
}

// MakerRebateActivity represents a Data API response.
type MakerRebateActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	Amount                *decimal.Big    `json:"amount"`
	Type                  ActivityType    `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *MakerRebateActivity) UnmarshalJSON(raw []byte) error {
	type plain MakerRebateActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "type", required: true, allowed: []string{"MAKER_REBATE"}},
	}); err != nil {
		return err
	}
	*m = MakerRebateActivity(result)
	return nil
}

// MergeActivity represents a Data API response.
type MergeActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	ConditionID           ConditionID     `json:"condition_id"`
	Amount                *decimal.Big    `json:"amount"`
	Title                 *string         `json:"title"`
	Slug                  *string         `json:"slug"`
	Icon                  *string         `json:"icon"`
	EventSlug             *string         `json:"event_slug"`
	Type                  ActivityType    `json:"type"`
	IsCombo               bool            `json:"is_combo"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *MergeActivity) UnmarshalJSON(raw []byte) error {
	type plain MergeActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "condition_id", aliases: []string{"condition_id"}, required: true, steps: []string{"validate_condition_id_response"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "title", nullable: true, steps: []string{"optional_text"}},
		{name: "slug", nullable: true, steps: []string{"optional_text"}},
		{name: "icon", nullable: true, steps: []string{"optional_text"}},
		{name: "event_slug", nullable: true, steps: []string{"optional_text"}},
		{name: "type", required: true, allowed: []string{"MERGE"}},
		{name: "is_combo", kind: "bool", fallback: json.RawMessage("false")},
	}); err != nil {
		return err
	}
	*m = MergeActivity(result)
	return nil
}

// RedeemActivity represents Redeemed proceeds from a market or Combo token.
type RedeemActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	ConditionID           ConditionID     `json:"condition_id"`
	Amount                *decimal.Big    `json:"amount"`
	Title                 *string         `json:"title"`
	Slug                  *string         `json:"slug"`
	Icon                  *string         `json:"icon"`
	EventSlug             *string         `json:"event_slug"`
	Type                  ActivityType    `json:"type"`
	IsCombo               bool            `json:"is_combo"`
	Outcome               *string         `json:"outcome"`
	OutcomeIndex          *int64          `json:"outcome_index"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *RedeemActivity) UnmarshalJSON(raw []byte) error {
	type plain RedeemActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "condition_id", aliases: []string{"condition_id"}, required: true, steps: []string{"validate_condition_id_response"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "title", nullable: true, steps: []string{"optional_text"}},
		{name: "slug", nullable: true, steps: []string{"optional_text"}},
		{name: "icon", nullable: true, steps: []string{"optional_text"}},
		{name: "event_slug", nullable: true, steps: []string{"optional_text"}},
		{name: "type", required: true, allowed: []string{"REDEEM"}},
		{name: "is_combo", kind: "bool", fallback: json.RawMessage("false")},
		{name: "outcome", nullable: true, steps: []string{"optional_text"}},
		{name: "outcome_index", kind: "int", nullable: true, steps: []string{"optional_outcome_index"}},
	}); err != nil {
		return err
	}
	if result.IsCombo {
		if _, err := codec.Condition(string(result.ConditionID), "combo"); err != nil {
			return err
		}
	}
	*m = RedeemActivity(result)
	return nil
}

// ReferralRewardActivity represents a Data API response.
type ReferralRewardActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	Amount                *decimal.Big    `json:"amount"`
	Type                  ActivityType    `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ReferralRewardActivity) UnmarshalJSON(raw []byte) error {
	type plain ReferralRewardActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "type", required: true, allowed: []string{"REFERRAL_REWARD"}},
	}); err != nil {
		return err
	}
	*m = ReferralRewardActivity(result)
	return nil
}

// RewardActivity represents a Data API response.
type RewardActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	Amount                *decimal.Big    `json:"amount"`
	Type                  ActivityType    `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *RewardActivity) UnmarshalJSON(raw []byte) error {
	type plain RewardActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "type", required: true, allowed: []string{"REWARD"}},
	}); err != nil {
		return err
	}
	*m = RewardActivity(result)
	return nil
}

// SplitActivity represents a Data API response.
type SplitActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	ConditionID           ConditionID     `json:"condition_id"`
	Amount                *decimal.Big    `json:"amount"`
	Title                 *string         `json:"title"`
	Slug                  *string         `json:"slug"`
	Icon                  *string         `json:"icon"`
	EventSlug             *string         `json:"event_slug"`
	Type                  ActivityType    `json:"type"`
	IsCombo               bool            `json:"is_combo"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *SplitActivity) UnmarshalJSON(raw []byte) error {
	type plain SplitActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "condition_id", aliases: []string{"condition_id"}, required: true, steps: []string{"validate_condition_id_response"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "title", nullable: true, steps: []string{"optional_text"}},
		{name: "slug", nullable: true, steps: []string{"optional_text"}},
		{name: "icon", nullable: true, steps: []string{"optional_text"}},
		{name: "event_slug", nullable: true, steps: []string{"optional_text"}},
		{name: "type", required: true, allowed: []string{"SPLIT"}},
		{name: "is_combo", kind: "bool", fallback: json.RawMessage("false")},
	}); err != nil {
		return err
	}
	*m = SplitActivity(result)
	return nil
}

// TakerRebateActivity represents a Data API response.
type TakerRebateActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	Amount                *decimal.Big    `json:"amount"`
	Type                  ActivityType    `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *TakerRebateActivity) UnmarshalJSON(raw []byte) error {
	type plain TakerRebateActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "type", required: true, allowed: []string{"TAKER_REBATE"}},
	}); err != nil {
		return err
	}
	*m = TakerRebateActivity(result)
	return nil
}

// Trade represents An executed trade; size is shares and price is USDC per share.
type Trade struct {
	Wallet                EVMAddress      `json:"wallet"`
	AssetID               ClobAssetID     `json:"asset_id"`
	ConditionID           ConditionID     `json:"condition_id"`
	Side                  OrderSide       `json:"side"`
	Size                  *decimal.Big    `json:"size"`
	Price                 *decimal.Big    `json:"price"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Title                 *string         `json:"title"`
	Slug                  *string         `json:"slug"`
	Icon                  *string         `json:"icon"`
	EventSlug             *string         `json:"event_slug"`
	Outcome               *string         `json:"outcome"`
	OutcomeIndex          *int64          `json:"outcome_index"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *Trade) UnmarshalJSON(raw []byte) error {
	type plain Trade
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "asset_id", aliases: []string{"token_id"}, required: true},
		{name: "condition_id", required: true, steps: []string{"validate_condition_id_response"}},
		{name: "side", required: true, allowed: []string{"BUY", "SELL"}},
		{name: "size", kind: "data_decimal", required: true},
		{name: "price", kind: "data_decimal", required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", required: true},
		{name: "title", nullable: true, steps: []string{"optional_text"}},
		{name: "slug", nullable: true, steps: []string{"optional_text"}},
		{name: "icon", nullable: true, steps: []string{"optional_text"}},
		{name: "event_slug", nullable: true, steps: []string{"optional_text"}},
		{name: "outcome", nullable: true, steps: []string{"optional_text"}},
		{name: "outcome_index", kind: "int", nullable: true, steps: []string{"optional_outcome_index"}},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", nullable: true, steps: []string{"optional_text"}},
	}); err != nil {
		return err
	}
	*m = Trade(result)
	return nil
}

// TradeActivity represents a Data API response.
type TradeActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	Type                  ActivityType    `json:"type"`
	IsCombo               bool            `json:"is_combo"`
	ConditionID           ConditionID     `json:"condition_id"`
	AssetID               ClobAssetID     `json:"asset_id"`
	Side                  OrderSide       `json:"side"`
	Shares                *decimal.Big    `json:"shares"`
	Amount                *decimal.Big    `json:"amount"`
	Price                 *decimal.Big    `json:"price"`
	Outcome               *string         `json:"outcome"`
	OutcomeIndex          *int64          `json:"outcome_index"`
	Title                 *string         `json:"title"`
	Slug                  *string         `json:"slug"`
	Icon                  *string         `json:"icon"`
	EventSlug             *string         `json:"event_slug"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *TradeActivity) UnmarshalJSON(raw []byte) error {
	type plain TradeActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "type", required: true, allowed: []string{"TRADE"}},
		{name: "is_combo", aliases: []string{"is_combo"}, kind: "literal_bool_false", fallback: json.RawMessage("false")},
		{name: "condition_id", aliases: []string{"condition_id"}, required: true, steps: []string{"validate_condition_id_response"}},
		{name: "asset_id", aliases: []string{"token_id"}, required: true},
		{name: "side", required: true, allowed: []string{"BUY", "SELL"}},
		{name: "shares", aliases: []string{"size"}, kind: "data_decimal", required: true},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "price", kind: "data_decimal", required: true},
		{name: "outcome", nullable: true, steps: []string{"optional_text"}},
		{name: "outcome_index", kind: "int", nullable: true, steps: []string{"optional_outcome_index"}},
		{name: "title", nullable: true, steps: []string{"optional_text"}},
		{name: "slug", nullable: true, steps: []string{"optional_text"}},
		{name: "icon", nullable: true, steps: []string{"optional_text"}},
		{name: "event_slug", nullable: true, steps: []string{"optional_text"}},
	}); err != nil {
		return err
	}
	*m = TradeActivity(result)
	return nil
}

// UnknownActivity represents a Data API response.
type UnknownActivity struct {
	Type                  string                     `json:"type"`
	Wallet                *EVMAddress                `json:"wallet"`
	Timestamp             *time.Time                 `json:"timestamp"`
	TransactionHash       *TransactionHash           `json:"transaction_hash"`
	Name                  *string                    `json:"name"`
	Pseudonym             *string                    `json:"pseudonym"`
	Bio                   *string                    `json:"bio"`
	ProfileImage          *string                    `json:"profile_image"`
	ProfileImageOptimized *string                    `json:"profile_image_optimized"`
	Raw                   map[string]json.RawMessage `json:"raw"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *UnknownActivity) UnmarshalJSON(raw []byte) error {
	type plain UnknownActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "type", required: true},
		{name: "wallet", aliases: []string{"proxy_wallet"}, nullable: true},
		{name: "timestamp", kind: "optional_epoch", nullable: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, nullable: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "raw", fallback: json.RawMessage(`{}`)},
	}); err != nil {
		return err
	}
	*m = UnknownActivity(result)
	return nil
}

// WithdrawalActivity represents a Data API response.
type WithdrawalActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	Amount                *decimal.Big    `json:"amount"`
	Type                  ActivityType    `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *WithdrawalActivity) UnmarshalJSON(raw []byte) error {
	type plain WithdrawalActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "type", required: true, allowed: []string{"WITHDRAWAL"}},
	}); err != nil {
		return err
	}
	*m = WithdrawalActivity(result)
	return nil
}

// YieldActivity represents a Data API response.
type YieldActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	Amount                *decimal.Big    `json:"amount"`
	Type                  ActivityType    `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *YieldActivity) UnmarshalJSON(raw []byte) error {
	type plain YieldActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "type", required: true, allowed: []string{"YIELD"}},
	}); err != nil {
		return err
	}
	*m = YieldActivity(result)
	return nil
}

// MigrationActivity represents a Data API response.
type MigrationActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	Amount                *decimal.Big    `json:"amount"`
	Type                  ActivityType    `json:"type"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *MigrationActivity) UnmarshalJSON(raw []byte) error {
	type plain MigrationActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "type", required: true, allowed: []string{"MIGRATION"}},
	}); err != nil {
		return err
	}
	*m = MigrationActivity(result)
	return nil
}

// TipActivity represents a Data API response.
type TipActivity struct {
	Wallet                EVMAddress      `json:"wallet"`
	Timestamp             time.Time       `json:"timestamp"`
	TransactionHash       TransactionHash `json:"transaction_hash"`
	Name                  *string         `json:"name"`
	Pseudonym             *string         `json:"pseudonym"`
	Bio                   *string         `json:"bio"`
	ProfileImage          *string         `json:"profile_image"`
	ProfileImageOptimized *string         `json:"profile_image_optimized"`
	Amount                *decimal.Big    `json:"amount"`
	Type                  ActivityType    `json:"type"`
	Side                  *TipSide        `json:"side"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *TipActivity) UnmarshalJSON(raw []byte) error {
	type plain TipActivity
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "timestamp", kind: "epoch", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", aliases: []string{"profile_image"}, nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", aliases: []string{"profile_image_optimized"}, nullable: true, steps: []string{"optional_text"}},
		{name: "amount", aliases: []string{"usdc_size"}, kind: "data_decimal", required: true},
		{name: "type", required: true, allowed: []string{"TIP"}},
		{name: "side", nullable: true, steps: []string{"optional_text"}, allowed: []string{"IN", "OUT"}},
	}); err != nil {
		return err
	}
	*m = TipActivity(result)
	return nil
}

// MarketLiveVolume represents Market taker volume in shares.
type MarketLiveVolume struct {
	ConditionID *ConditionID `json:"condition_id"`
	TakerVolume *decimal.Big `json:"taker_volume"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *MarketLiveVolume) UnmarshalJSON(raw []byte) error {
	type plain MarketLiveVolume
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", required: true, nullable: true, steps: []string{"optional_text", "validate_optional_condition_id_response"}},
		{name: "taker_volume", kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = MarketLiveVolume(result)
	return nil
}

// LiveVolume represents Event taker volume in shares and its market breakdown.
type LiveVolume struct {
	TakerVolumeTotal *decimal.Big       `json:"taker_volume_total"`
	Markets          []MarketLiveVolume `json:"markets"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *LiveVolume) UnmarshalJSON(raw []byte) error {
	type plain LiveVolume
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "taker_volume_total", kind: "data_decimal", required: true},
		{name: "markets", aliases: []string{"conditions"}, kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = LiveVolume(result)
	return nil
}

// OpenInterest represents Open interest in USDC; condition_id=None denotes global interest.
type OpenInterest struct {
	ConditionID *ConditionID `json:"condition_id"`
	Value       *decimal.Big `json:"value"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *OpenInterest) UnmarshalJSON(raw []byte) error {
	type plain OpenInterest
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", required: true, nullable: true, steps: []string{"global_condition", "validate_optional_condition_id_response"}},
		{name: "value", kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = OpenInterest(result)
	return nil
}

// Holder represents A holder. amount is shares; prices, cost, value and PnL are USDC.
type Holder struct {
	Wallet                EVMAddress   `json:"wallet"`
	AssetID               ClobAssetID  `json:"asset_id"`
	Amount                *decimal.Big `json:"amount"`
	OutcomeIndex          *int64       `json:"outcome_index"`
	DisplayUsernamePublic bool         `json:"display_username_public"`
	Verified              bool         `json:"verified"`
	Name                  *string      `json:"name"`
	Pseudonym             *string      `json:"pseudonym"`
	Bio                   *string      `json:"bio"`
	ProfileImage          *string      `json:"profile_image"`
	ProfileImageOptimized *string      `json:"profile_image_optimized"`
	AvgPrice              *decimal.Big `json:"avg_price"`
	EntryCostUSDC         *decimal.Big `json:"entry_cost_usdc"`
	CurrentPrice          *decimal.Big `json:"current_price"`
	CurrentValue          *decimal.Big `json:"current_value"`
	RealizedPnL           *decimal.Big `json:"realized_pnl"`
	UnrealizedPnL         *decimal.Big `json:"unrealized_pnl"`
	TotalPnL              *decimal.Big `json:"total_pnl"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *Holder) UnmarshalJSON(raw []byte) error {
	type plain Holder
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"proxy_wallet"}, required: true},
		{name: "asset_id", aliases: []string{"token_id"}, required: true},
		{name: "amount", kind: "data_decimal", required: true},
		{name: "outcome_index", kind: "int", nullable: true, steps: []string{"optional_outcome_index"}},
		{name: "display_username_public", kind: "bool", required: true},
		{name: "verified", kind: "bool", required: true},
		{name: "name", nullable: true, steps: []string{"optional_text"}},
		{name: "pseudonym", nullable: true, steps: []string{"optional_text"}},
		{name: "bio", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image_optimized", nullable: true, steps: []string{"optional_text"}},
		{name: "avg_price", kind: "data_decimal", nullable: true},
		{name: "entry_cost_usdc", kind: "data_decimal", nullable: true},
		{name: "current_price", kind: "data_decimal", nullable: true},
		{name: "current_value", kind: "data_decimal", nullable: true},
		{name: "realized_pnl", kind: "data_decimal", nullable: true},
		{name: "unrealized_pnl", kind: "data_decimal", nullable: true},
		{name: "total_pnl", kind: "data_decimal", nullable: true},
	}); err != nil {
		return err
	}
	*m = Holder(result)
	return nil
}

// MetaHolder represents Holders grouped by outcome asset.
type MetaHolder struct {
	AssetID ClobAssetID `json:"asset_id"`
	Holders []Holder    `json:"holders"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *MetaHolder) UnmarshalJSON(raw []byte) error {
	type plain MetaHolder
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_id", aliases: []string{"token_id"}, required: true},
		{name: "holders", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = MetaHolder(result)
	return nil
}

// TraderLeaderboardEntry represents A ranked trader; PnL is USDC and volume is shares.
type TraderLeaderboardEntry struct {
	Rank         int64        `json:"rank"`
	Wallet       EVMAddress   `json:"wallet"`
	PnL          *decimal.Big `json:"pnl"`
	Volume       *decimal.Big `json:"volume"`
	UserName     *string      `json:"user_name"`
	ProfileImage *string      `json:"profile_image"`
	XUsername    *string      `json:"x_username"`
	Verified     bool         `json:"verified"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *TraderLeaderboardEntry) UnmarshalJSON(raw []byte) error {
	type plain TraderLeaderboardEntry
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "rank", kind: "int", required: true},
		{name: "wallet", aliases: []string{"user_id"}, required: true},
		{name: "pnl", kind: "data_decimal", required: true},
		{name: "volume", kind: "data_decimal", required: true},
		{name: "user_name", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", nullable: true, steps: []string{"optional_text"}},
		{name: "x_username", nullable: true, steps: []string{"optional_text"}},
		{name: "verified", kind: "bool", required: true},
	}); err != nil {
		return err
	}
	*m = TraderLeaderboardEntry(result)
	return nil
}

// TraderLeaderboardStanding represents Trader rankings; PnL is USDC and volume is shares. Unranked values are None.
type TraderLeaderboardStanding struct {
	Wallet       EVMAddress   `json:"wallet"`
	PnL          *decimal.Big `json:"pnl"`
	Volume       *decimal.Big `json:"volume"`
	UserName     *string      `json:"user_name"`
	ProfileImage *string      `json:"profile_image"`
	XUsername    *string      `json:"x_username"`
	Verified     bool         `json:"verified"`
	PnLRank      *int64       `json:"pnl_rank"`
	VolumeRank   *int64       `json:"volume_rank"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *TraderLeaderboardStanding) UnmarshalJSON(raw []byte) error {
	type plain TraderLeaderboardStanding
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "wallet", aliases: []string{"user_id"}, required: true},
		{name: "pnl", kind: "data_decimal", required: true},
		{name: "volume", kind: "data_decimal", required: true},
		{name: "user_name", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", nullable: true, steps: []string{"optional_text"}},
		{name: "x_username", nullable: true, steps: []string{"optional_text"}},
		{name: "verified", kind: "bool", required: true},
		{name: "pnl_rank", aliases: []string{"rank_pnl"}, kind: "int", nullable: true, steps: []string{"zero_rank"}},
		{name: "volume_rank", aliases: []string{"rank_volume"}, kind: "int", nullable: true, steps: []string{"zero_rank"}},
	}); err != nil {
		return err
	}
	*m = TraderLeaderboardStanding(result)
	return nil
}

// MarketBiggestWinner represents A winning market position; PnL and values are USDC.
type MarketBiggestWinner struct {
	Rank         int64             `json:"rank"`
	Wallet       EVMAddress        `json:"wallet"`
	PnL          *decimal.Big      `json:"pnl"`
	InitialValue *decimal.Big      `json:"initial_value"`
	FinalValue   *decimal.Big      `json:"final_value"`
	ResolvedAt   time.Time         `json:"resolved_at"`
	UserName     *string           `json:"user_name"`
	ProfileImage *string           `json:"profile_image"`
	EventTitle   *string           `json:"event_title"`
	Kind         BiggestWinnerKind `json:"kind"`
	ConditionID  ConditionID       `json:"condition_id"`
	AssetID      ClobAssetID       `json:"asset_id"`
	EventID      EventID           `json:"event_id"`
	EventSlug    *string           `json:"event_slug"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *MarketBiggestWinner) UnmarshalJSON(raw []byte) error {
	type plain MarketBiggestWinner
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "rank", aliases: []string{"win_rank"}, kind: "int", required: true},
		{name: "wallet", aliases: []string{"user_id"}, required: true},
		{name: "pnl", kind: "data_decimal", required: true},
		{name: "initial_value", kind: "data_decimal", required: true},
		{name: "final_value", kind: "data_decimal", required: true},
		{name: "resolved_at", kind: "epoch_or_iso", required: true},
		{name: "user_name", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", nullable: true, steps: []string{"optional_text"}},
		{name: "event_title", nullable: true, steps: []string{"optional_text"}},
		{name: "kind", required: true, allowed: []string{"market"}},
		{name: "condition_id", required: true, steps: []string{"validate_condition_id_response"}},
		{name: "asset_id", aliases: []string{"position_id"}, required: true},
		{name: "event_id", required: true, steps: []string{"required_event_id"}},
		{name: "event_slug", nullable: true, steps: []string{"optional_text"}},
	}); err != nil {
		return err
	}
	*m = MarketBiggestWinner(result)
	return nil
}

// ComboBiggestWinner represents A winning combo position; PnL and values are USDC.
type ComboBiggestWinner struct {
	Rank         int64             `json:"rank"`
	Wallet       EVMAddress        `json:"wallet"`
	PnL          *decimal.Big      `json:"pnl"`
	InitialValue *decimal.Big      `json:"initial_value"`
	FinalValue   *decimal.Big      `json:"final_value"`
	ResolvedAt   time.Time         `json:"resolved_at"`
	UserName     *string           `json:"user_name"`
	ProfileImage *string           `json:"profile_image"`
	EventTitle   *string           `json:"event_title"`
	Kind         BiggestWinnerKind `json:"kind"`
	ConditionID  ComboConditionID  `json:"condition_id"`
	PositionID   PositionID        `json:"position_id"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *ComboBiggestWinner) UnmarshalJSON(raw []byte) error {
	type plain ComboBiggestWinner
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "rank", aliases: []string{"win_rank"}, kind: "int", required: true},
		{name: "wallet", aliases: []string{"user_id"}, required: true},
		{name: "pnl", kind: "data_decimal", required: true},
		{name: "initial_value", kind: "data_decimal", required: true},
		{name: "final_value", kind: "data_decimal", required: true},
		{name: "resolved_at", kind: "epoch_or_iso", required: true},
		{name: "user_name", nullable: true, steps: []string{"optional_text"}},
		{name: "profile_image", nullable: true, steps: []string{"optional_text"}},
		{name: "event_title", nullable: true, steps: []string{"optional_text"}},
		{name: "kind", required: true, allowed: []string{"combo"}},
		{name: "condition_id", required: true, steps: []string{"validate_combo_condition_id"}},
		{name: "position_id", required: true},
	}); err != nil {
		return err
	}
	*m = ComboBiggestWinner(result)
	return nil
}

// BuilderStanding represents Builder rankings; volume is shares.
type BuilderStanding struct {
	Rank         int64        `json:"rank"`
	BuilderName  string       `json:"builder_name"`
	BuilderCode  HexString    `json:"builder_code"`
	ProfileImage *string      `json:"profile_image"`
	Verified     bool         `json:"verified"`
	Volume       *decimal.Big `json:"volume"`
	ActiveUsers  int64        `json:"active_users"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *BuilderStanding) UnmarshalJSON(raw []byte) error {
	type plain BuilderStanding
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "rank", kind: "int", required: true},
		{name: "builder_name", required: true},
		{name: "builder_code", required: true},
		{name: "profile_image", nullable: true, steps: []string{"optional_text"}},
		{name: "verified", kind: "bool", required: true},
		{name: "volume", kind: "data_decimal", required: true},
		{name: "active_users", kind: "int", required: true},
	}); err != nil {
		return err
	}
	*m = BuilderStanding(result)
	return nil
}

// BuilderVolumePoint represents Builder volume in shares for a calendar bucket.
type BuilderVolumePoint struct {
	Rank         int64        `json:"rank"`
	BuilderName  string       `json:"builder_name"`
	BuilderCode  HexString    `json:"builder_code"`
	ProfileImage *string      `json:"profile_image"`
	Verified     bool         `json:"verified"`
	Volume       *decimal.Big `json:"volume"`
	ActiveUsers  int64        `json:"active_users"`
	BucketDate   Date         `json:"bucket_date"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *BuilderVolumePoint) UnmarshalJSON(raw []byte) error {
	type plain BuilderVolumePoint
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "rank", kind: "int", required: true},
		{name: "builder_name", required: true},
		{name: "builder_code", required: true},
		{name: "profile_image", nullable: true, steps: []string{"optional_text"}},
		{name: "verified", kind: "bool", required: true},
		{name: "volume", kind: "data_decimal", required: true},
		{name: "active_users", kind: "int", required: true},
		{name: "bucket_date", aliases: []string{"date"}, kind: "date", required: true},
	}); err != nil {
		return err
	}
	*m = BuilderVolumePoint(result)
	return nil
}

// TokenReference represents Outcome asset reference facts and settlement metadata.
type TokenReference struct {
	AssetID               ClobAssetID  `json:"asset_id"`
	ConditionID           ConditionID  `json:"condition_id"`
	StructuralConditionID *ConditionID `json:"structural_condition_id"`
	Module                TokenModule  `json:"module"`
	OutcomeIndex          int64        `json:"outcome_index"`
	ClobIndex             *int64       `json:"clob_index"`
	Outcome               *string      `json:"outcome"`
	OppositeAssetID       *ClobAssetID `json:"opposite_asset_id"`
	Resolved              bool         `json:"resolved"`
	FinalPrice            *decimal.Big `json:"final_price"`
	Title                 *string      `json:"title"`
	MarketSlug            *string      `json:"market_slug"`
	Closed                *bool        `json:"closed"`
	NegRisk               *bool        `json:"neg_risk"`
	NegRiskMarketID       *ConditionID `json:"neg_risk_market_id"`
	QuestionIndex         *int64       `json:"question_index"`
	EventID               *EventID     `json:"event_id"`
	EventSlug             *string      `json:"event_slug"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *TokenReference) UnmarshalJSON(raw []byte) error {
	type plain TokenReference
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_id", aliases: []string{"token_id"}, kind: "asset_reference", required: true},
		{name: "condition_id", kind: "condition_reference", required: true},
		{name: "structural_condition_id", kind: "condition_reference", required: true, nullable: true},
		{name: "module", required: true, allowed: []string{"v1_ctf", "binary", "neg_risk", "combo"}},
		{name: "outcome_index", kind: "strict_int", required: true, minimum: -32768, bounded: true, maximum: 32767},
		{name: "clob_index", kind: "strict_int", required: true, nullable: true, minimum: -32768, bounded: true, maximum: 32767},
		{name: "outcome", required: true, nullable: true},
		{name: "opposite_asset_id", aliases: []string{"opposite_token_id"}, kind: "asset_reference", required: true, nullable: true},
		{name: "resolved", kind: "strict_bool", required: true},
		{name: "final_price", kind: "data_decimal", required: true, nullable: true},
		{name: "title", required: true, nullable: true},
		{name: "market_slug", required: true, nullable: true},
		{name: "closed", kind: "strict_bool", required: true, nullable: true},
		{name: "neg_risk", kind: "strict_bool", required: true, nullable: true},
		{name: "neg_risk_market_id", kind: "condition_reference", required: true, nullable: true},
		{name: "question_index", kind: "strict_int", required: true, nullable: true, minimum: -2147483648, bounded: true, maximum: 2147483647},
		{name: "event_id", kind: "event_reference", required: true, nullable: true},
		{name: "event_slug", required: true, nullable: true},
	}); err != nil {
		return err
	}
	*m = TokenReference(result)
	return nil
}

// Resolution represents Resolution lifecycle and per-outcome payouts in USDC per share.
type Resolution struct {
	QuestionID             *QuestionID                    `json:"question_id"`
	ConditionID            *ConditionID                   `json:"condition_id"`
	Status                 ResolutionStatus               `json:"status"`
	ExtendedReview         bool                           `json:"extended_review"`
	ExpectedSettlementTime *time.Time                     `json:"expected_settlement_time"`
	SettlementTimeBasis    *ResolutionSettlementTimeBasis `json:"settlement_time_basis"`
	WasDisputed            bool                           `json:"was_disputed"`
	QuestionRulesUpdated   bool                           `json:"question_rules_updated"`
	ProposedPrice          *decimal.Big                   `json:"proposed_price"`
	ReproposedPrice        *decimal.Big                   `json:"reproposed_price"`
	Price                  *decimal.Big                   `json:"price"`
	TransactionHash        *TransactionHash               `json:"transaction_hash"`
	LogIndex               *int64                         `json:"log_index"`
	LastUpdatedAt          time.Time                      `json:"last_updated_at"`
	MarketType             *ResolutionMarketType          `json:"market_type"`
	Payouts                []*decimal.Big                 `json:"payouts"`
	ResolutionSource       *ResolutionSource              `json:"resolution_source"`
	Reporter               *ResolutionReporter            `json:"reporter"`
	WasArbitrated          *bool                          `json:"was_arbitrated"`
	ResolvedBlock          *int64                         `json:"resolved_block"`
	ResolvedAt             *time.Time                     `json:"resolved_at"`
}

// UnmarshalJSON validates Data API field types and wire aliases.
func (m *Resolution) UnmarshalJSON(raw []byte) error {
	type plain Resolution
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "question_id", nullable: true},
		{name: "condition_id", nullable: true, steps: []string{"validate_optional_condition_id_response"}},
		{name: "status", required: true, allowed: []string{"initialized", "posed", "proposed", "challenged", "reproposed", "disputed", "active", "arbitration", "resolved"}},
		{name: "extended_review", kind: "bool", required: true},
		{name: "expected_settlement_time", kind: "epoch_datetime", nullable: true},
		{name: "settlement_time_basis", nullable: true, allowed: []string{"managed_proposal_expiration", "proposal_expiration", "liveness", "dvm_round_estimate"}},
		{name: "was_disputed", kind: "bool", required: true},
		{name: "question_rules_updated", aliases: []string{"new_version_q"}, kind: "bool", required: true},
		{name: "proposed_price", kind: "data_decimal", nullable: true, steps: []string{"unset_price"}},
		{name: "reproposed_price", kind: "data_decimal", nullable: true, steps: []string{"unset_price"}},
		{name: "price", kind: "data_decimal", nullable: true, steps: []string{"unset_price"}},
		{name: "transaction_hash", nullable: true, steps: []string{"optional_text"}},
		{name: "log_index", kind: "int", nullable: true, steps: []string{"optional_text"}},
		{name: "last_updated_at", aliases: []string{"last_update_timestamp"}, kind: "epoch_or_iso", required: true},
		{name: "market_type", nullable: true, allowed: []string{"BINARY", "INCREMENTAL_NEGRISK", "ATOMIC_NEGRISK"}},
		{name: "payouts", kind: "payout_pair", nullable: true},
		{name: "resolution_source", nullable: true, allowed: []string{"reported", "derived"}},
		{name: "reporter", nullable: true, allowed: []string{"UMA_OO", "CHAINLINK", "EOA"}},
		{name: "was_arbitrated", kind: "bool", nullable: true},
		{name: "resolved_block", kind: "int", nullable: true},
		{name: "resolved_at", kind: "epoch_datetime", nullable: true},
	}); err != nil {
		return err
	}
	*m = Resolution(result)
	return nil
}
