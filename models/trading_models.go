// Authenticated models ported from the pinned SDK. This file is linted.

package models

import (
	"encoding/json"
	"time"

	"github.com/ericlagergren/decimal"
)

// APIKey identifies an account resource.
type APIKey string

// OrderID identifies an account resource.
type OrderID string

// CollateralReturnPlanResponse is an authenticated exchange response.
type CollateralReturnPlanResponse struct {
	PlanHash             HexString                        `json:"plan_hash"`
	ChainID              int64                            `json:"chain_id"`
	Wallet               EVMAddress                       `json:"wallet"`
	BlockNumber          int64                            `json:"block_number"`
	StartingPusd         *decimal.Big                     `json:"starting_pusd"`
	NetPusdOut           *decimal.Big                     `json:"net_pusd_out"`
	FinalPusd            *decimal.Big                     `json:"final_pusd"`
	Operations           []CollateralReturnOperation      `json:"operations"`
	OperationCount       int64                            `json:"operation_count"`
	Truncated            bool                             `json:"truncated"`
	EstimatedCost        float64                          `json:"estimated_cost"`
	RequiredPusdInput    *decimal.Big                     `json:"required_pusd_input"`
	RequiredPositions    []CollateralReturnPositionAmount `json:"required_positions"`
	PositionSummary      CollateralReturnPositionSummary  `json:"position_summary"`
	CandidatePositionIDs []PositionID                     `json:"candidate_position_ids"`
	RouterCall           CollateralReturnRouterCall       `json:"router_call"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *CollateralReturnPlanResponse) UnmarshalJSON(raw []byte) error {
	type plain CollateralReturnPlanResponse
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "plan_hash", required: true},
		{name: "chain_id", kind: "int", required: true},
		{name: "wallet", required: true},
		{name: "block_number", kind: "int", required: true},
		{name: "starting_pusd", kind: "string_decimal", required: true},
		{name: "net_pusd_out", kind: "string_decimal", required: true},
		{name: "final_pusd", kind: "string_decimal", required: true},
		{name: "operations", kind: "sequence", required: true},
		{name: "operation_count", kind: "int", required: true},
		{name: "truncated", kind: "bool", required: true},
		{name: "estimated_cost", kind: "float", required: true},
		{name: "required_pusd_input", kind: "string_decimal", required: true},
		{name: "required_positions", kind: "sequence", required: true},
		{name: "position_summary", fallback: json.RawMessage(`{}`)},
		{name: "candidate_position_ids", kind: "sequence", required: true},
		{name: "router_call", required: true},
	}); err != nil {
		return err
	}
	*m = CollateralReturnPlanResponse(result)
	return nil
}

// OpenOrder is an authenticated exchange response.
type OpenOrder struct {
	ID              string       `json:"id"`
	ConditionID     ConditionID  `json:"condition_id"`
	AssetID         ClobAssetID  `json:"asset_id"`
	Owner           string       `json:"owner"`
	MakerAddress    string       `json:"maker_address"`
	Side            OrderSide    `json:"side"`
	Price           *decimal.Big `json:"price"`
	OriginalSize    *decimal.Big `json:"original_size"`
	SizeMatched     *decimal.Big `json:"size_matched"`
	Outcome         string       `json:"outcome"`
	OrderType       string       `json:"order_type"`
	Status          string       `json:"status"`
	AssociateTrades []string     `json:"associate_trades"`
	CreatedAt       time.Time    `json:"created_at"`
	ExpiresAt       *time.Time   `json:"expires_at"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *OpenOrder) UnmarshalJSON(raw []byte) error {
	type plain OpenOrder
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "owner", required: true},
		{name: "maker_address", aliases: []string{"maker_address"}, required: true},
		{name: "side", required: true, allowed: []string{"BUY", "SELL"}},
		{name: "price", kind: "string_decimal", required: true},
		{name: "original_size", aliases: []string{"original_size"}, kind: "string_decimal", required: true},
		{name: "size_matched", aliases: []string{"size_matched"}, kind: "string_decimal", required: true},
		{name: "outcome", required: true},
		{name: "order_type", aliases: []string{"order_type"}, required: true},
		{name: "status", required: true},
		{name: "associate_trades", aliases: []string{"associate_trades"}, kind: "sequence", fallback: json.RawMessage("[]")},
		{name: "created_at", aliases: []string{"created_at"}, kind: "clob_epoch_iso", required: true},
		{name: "expires_at", aliases: []string{"expiration"}, kind: "clob_epoch_iso", nullable: true, steps: []string{"zero_time"}},
	}); err != nil {
		return err
	}
	*m = OpenOrder(result)
	return nil
}

// MakerOrder is an authenticated exchange response.
type MakerOrder struct {
	OrderID       string       `json:"order_id"`
	AssetID       ClobAssetID  `json:"asset_id"`
	MakerAddress  string       `json:"maker_address"`
	Owner         string       `json:"owner"`
	Side          OrderSide    `json:"side"`
	Price         *decimal.Big `json:"price"`
	MatchedAmount *decimal.Big `json:"matched_amount"`
	Outcome       string       `json:"outcome"`
	FeeRateBps    *decimal.Big `json:"fee_rate_bps"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *MakerOrder) UnmarshalJSON(raw []byte) error {
	type plain MakerOrder
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "order_id", aliases: []string{"order_id"}, required: true},
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "maker_address", aliases: []string{"maker_address"}, required: true},
		{name: "owner", required: true},
		{name: "side", required: true, allowed: []string{"BUY", "SELL"}},
		{name: "price", kind: "string_decimal", required: true},
		{name: "matched_amount", aliases: []string{"matched_amount"}, kind: "string_decimal", required: true},
		{name: "outcome", required: true},
		{name: "fee_rate_bps", aliases: []string{"fee_rate_bps"}, kind: "string_decimal", nullable: true, steps: []string{"empty_none"}},
	}); err != nil {
		return err
	}
	*m = MakerOrder(result)
	return nil
}

// ClobTrade is an authenticated exchange response.
type ClobTrade struct {
	ID              string       `json:"id"`
	ConditionID     ConditionID  `json:"condition_id"`
	AssetID         ClobAssetID  `json:"asset_id"`
	Owner           string       `json:"owner"`
	MakerAddress    string       `json:"maker_address"`
	TakerOrderID    string       `json:"taker_order_id"`
	Side            OrderSide    `json:"side"`
	TraderSide      string       `json:"trader_side"`
	Price           *decimal.Big `json:"price"`
	Size            *decimal.Big `json:"size"`
	Outcome         string       `json:"outcome"`
	Status          string       `json:"status"`
	FeeRateBps      *decimal.Big `json:"fee_rate_bps"`
	BucketIndex     int64        `json:"bucket_index"`
	TransactionHash *string      `json:"transaction_hash"`
	MakerOrders     []MakerOrder `json:"maker_orders"`
	MatchedAt       time.Time    `json:"matched_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *ClobTrade) UnmarshalJSON(raw []byte) error {
	type plain ClobTrade
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "owner", required: true},
		{name: "maker_address", aliases: []string{"maker_address"}, required: true},
		{name: "taker_order_id", aliases: []string{"taker_order_id"}, required: true},
		{name: "side", required: true, allowed: []string{"BUY", "SELL"}},
		{name: "trader_side", aliases: []string{"trader_side"}, required: true, allowed: []string{"TAKER", "MAKER"}},
		{name: "price", kind: "string_decimal", required: true},
		{name: "size", kind: "string_decimal", required: true},
		{name: "outcome", required: true},
		{name: "status", required: true, allowed: []string{"MATCHED", "MATCHED_NOT_BROADCASTED", "MINED", "CONFIRMED", "RETRYING", "FAILED"}, steps: []string{"trade_status"}},
		{name: "fee_rate_bps", aliases: []string{"fee_rate_bps"}, kind: "string_decimal", required: true},
		{name: "bucket_index", aliases: []string{"bucket_index"}, kind: "int", required: true},
		{name: "transaction_hash", aliases: []string{"transaction_hash"}, nullable: true, steps: []string{"empty_none"}},
		{name: "maker_orders", aliases: []string{"maker_orders"}, kind: "sequence", required: true},
		{name: "matched_at", aliases: []string{"match_time"}, kind: "clob_epoch_iso", required: true},
		{name: "updated_at", aliases: []string{"last_update"}, kind: "clob_epoch_iso", required: true},
	}); err != nil {
		return err
	}
	*m = ClobTrade(result)
	return nil
}

// BuilderAPIKeyInfo is an authenticated exchange response.
type BuilderAPIKeyInfo struct {
	Key       string     `json:"key"`
	CreatedAt *time.Time `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *BuilderAPIKeyInfo) UnmarshalJSON(raw []byte) error {
	type plain BuilderAPIKeyInfo
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "key", required: true},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "reward_epoch_ms", nullable: true},
		{name: "revoked_at", aliases: []string{"revokedAt"}, kind: "reward_epoch_ms", nullable: true},
	}); err != nil {
		return err
	}
	*m = BuilderAPIKeyInfo(result)
	return nil
}

// BuilderFeeRates is an authenticated exchange response.
type BuilderFeeRates struct {
	Maker *decimal.Big `json:"maker"`
	Taker *decimal.Big `json:"taker"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *BuilderFeeRates) UnmarshalJSON(raw []byte) error {
	normalized, err := normalizeBuilderFees(raw)
	if err != nil {
		return err
	}
	raw = normalized
	type plain BuilderFeeRates
	var result plain
	if err = decodeData(raw, &result, []dataField{
		{name: "maker", aliases: []string{"builder_maker_fee_rate_bps"}, kind: "data_decimal", required: true},
		{name: "taker", aliases: []string{"builder_taker_fee_rate_bps"}, kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = BuilderFeeRates(result)
	return nil
}

// CancelOrdersResponse is an authenticated exchange response.
type CancelOrdersResponse struct {
	Canceled    []OrderID         `json:"canceled"`
	NotCanceled map[string]string `json:"not_canceled"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *CancelOrdersResponse) UnmarshalJSON(raw []byte) error {
	type plain CancelOrdersResponse
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "canceled", kind: "sequence", required: true},
		{name: "not_canceled", aliases: []string{"not_canceled"}, fallback: json.RawMessage(`{}`)},
	}); err != nil {
		return err
	}
	*m = CancelOrdersResponse(result)
	return nil
}

// OrderHeartbeat is an authenticated exchange response.
type OrderHeartbeat struct {
	HeartbeatID string `json:"heartbeat_id"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *OrderHeartbeat) UnmarshalJSON(raw []byte) error {
	type plain OrderHeartbeat
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "heartbeat_id", required: true, kind: "nonempty_string"},
	}); err != nil {
		return err
	}
	*m = OrderHeartbeat(result)
	return nil
}

// OrderHeartbeatMismatchResponse is an authenticated exchange response.
type OrderHeartbeatMismatchResponse struct {
	ErrorMsg    string `json:"error_msg"`
	HeartbeatID string `json:"heartbeat_id"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *OrderHeartbeatMismatchResponse) UnmarshalJSON(raw []byte) error {
	type plain OrderHeartbeatMismatchResponse
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "error_msg", required: true, allowed: []string{"Invalid Heartbeat ID"}},
		{name: "heartbeat_id", required: true},
	}); err != nil {
		return err
	}
	*m = OrderHeartbeatMismatchResponse(result)
	return nil
}

// RawOrderResponse is an authenticated exchange response.
type RawOrderResponse struct {
	ErrorMsg           string       `json:"error_msg"`
	MakingAmount       *decimal.Big `json:"making_amount"`
	OrderID            string       `json:"order_id"`
	Status             string       `json:"status"`
	Success            bool         `json:"success"`
	TakingAmount       *decimal.Big `json:"taking_amount"`
	TradeIDs           []string     `json:"trade_ids"`
	TransactionsHashes []string     `json:"transactions_hashes"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *RawOrderResponse) UnmarshalJSON(raw []byte) error {
	type plain RawOrderResponse
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "error_msg", aliases: []string{"errorMsg"}, required: true},
		{name: "making_amount", aliases: []string{"makingAmount"}, kind: "string_decimal", required: true, steps: []string{"empty_zero"}},
		{name: "order_id", aliases: []string{"orderID"}, required: true},
		{name: "status", required: true},
		{name: "success", kind: "bool", required: true},
		{name: "taking_amount", aliases: []string{"takingAmount"}, kind: "string_decimal", required: true, steps: []string{"empty_zero"}},
		{name: "trade_ids", aliases: []string{"tradeIDs"}, kind: "sequence", fallback: json.RawMessage("[]")},
		{name: "transactions_hashes", aliases: []string{"transactionsHashes"}, kind: "sequence", fallback: json.RawMessage("[]")},
	}); err != nil {
		return err
	}
	*m = RawOrderResponse(result)
	return nil
}

// AcceptedOrder is an authenticated exchange response.
type AcceptedOrder struct {
	Ok                 bool              `json:"ok"`
	OrderID            OrderID           `json:"order_id"`
	Status             string            `json:"status"`
	MakingAmount       *decimal.Big      `json:"making_amount"`
	TakingAmount       *decimal.Big      `json:"taking_amount"`
	TradeIDs           []string          `json:"trade_ids"`
	TransactionsHashes []TransactionHash `json:"transactions_hashes"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *AcceptedOrder) UnmarshalJSON(raw []byte) error {
	type plain AcceptedOrder
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "ok", kind: "literal_bool_true", fallback: json.RawMessage("true")},
		{name: "order_id", required: true},
		{name: "status", required: true, allowed: []string{"live", "matched", "delayed"}},
		{name: "making_amount", kind: "data_decimal", required: true},
		{name: "taking_amount", kind: "data_decimal", required: true},
		{name: "trade_ids", kind: "sequence", required: true},
		{name: "transactions_hashes", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = AcceptedOrder(result)
	return nil
}

// RejectedOrder is an authenticated exchange response.
type RejectedOrder struct {
	Ok      bool   `json:"ok"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *RejectedOrder) UnmarshalJSON(raw []byte) error {
	type plain RejectedOrder
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "ok", kind: "literal_bool_false", fallback: json.RawMessage("false")},
		{name: "code", required: true, allowed: []string{"unmatched", "market_not_ready", "not_enough_balance", "invalid_nonce", "invalid_expiration", "post_only_would_cross", "post_only_mode", "fok_not_filled", "fak_not_filled", "unknown"}},
		{name: "message", required: true},
	}); err != nil {
		return err
	}
	*m = RejectedOrder(result)
	return nil
}

// UserEarning is an authenticated exchange response.
type UserEarning struct {
	AssetAddress string       `json:"asset_address"`
	AssetRate    *decimal.Big `json:"asset_rate"`
	ConditionID  ConditionID  `json:"condition_id"`
	Date         time.Time    `json:"date"`
	Earnings     *decimal.Big `json:"earnings"`
	MakerAddress string       `json:"maker_address"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *UserEarning) UnmarshalJSON(raw []byte) error {
	type plain UserEarning
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_address", aliases: []string{"asset_address"}, required: true},
		{name: "asset_rate", aliases: []string{"asset_rate"}, kind: "data_decimal", required: true},
		{name: "condition_id", aliases: []string{"condition_id"}, required: true, steps: []string{"condition_feed"}},
		{name: "date", kind: "reward_epoch_ms", required: true},
		{name: "earnings", kind: "data_decimal", required: true},
		{name: "maker_address", aliases: []string{"maker_address"}, required: true},
	}); err != nil {
		return err
	}
	*m = UserEarning(result)
	return nil
}

// TotalUserEarning is an authenticated exchange response.
type TotalUserEarning struct {
	AssetAddress      string       `json:"asset_address"`
	AssetRate         *decimal.Big `json:"asset_rate"`
	Date              time.Time    `json:"date"`
	Earnings          *decimal.Big `json:"earnings"`
	MakerAddress      string       `json:"maker_address"`
	NativeEarnings    *decimal.Big `json:"native_earnings"`
	SponsoredEarnings *decimal.Big `json:"sponsored_earnings"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *TotalUserEarning) UnmarshalJSON(raw []byte) error {
	type plain TotalUserEarning
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_address", aliases: []string{"asset_address"}, required: true},
		{name: "asset_rate", aliases: []string{"asset_rate"}, kind: "data_decimal", required: true},
		{name: "date", kind: "reward_epoch_ms", required: true},
		{name: "earnings", kind: "data_decimal", required: true},
		{name: "maker_address", aliases: []string{"maker_address"}, required: true},
		{name: "native_earnings", kind: "data_decimal", nullable: true},
		{name: "sponsored_earnings", kind: "data_decimal", nullable: true},
	}); err != nil {
		return err
	}
	*m = TotalUserEarning(result)
	return nil
}

// UserRewardsEarning is an authenticated exchange response.
type UserRewardsEarning struct {
	ConditionID           ConditionID         `json:"condition_id"`
	EarningPercentage     float64             `json:"earning_percentage"`
	Earnings              []EarningBreakdown  `json:"earnings"`
	EventSlug             string              `json:"event_slug"`
	Image                 string              `json:"image"`
	MakerAddress          string              `json:"maker_address"`
	MarketCompetitiveness float64             `json:"market_competitiveness"`
	MarketSlug            string              `json:"market_slug"`
	Question              string              `json:"question"`
	RewardsConfig         []UserRewardsConfig `json:"rewards_config"`
	RewardsMaxSpread      float64             `json:"rewards_max_spread"`
	RewardsMinSize        *decimal.Big        `json:"rewards_min_size"`
	Tokens                []MarketRewardToken `json:"tokens"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *UserRewardsEarning) UnmarshalJSON(raw []byte) error {
	type plain UserRewardsEarning
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", aliases: []string{"condition_id"}, required: true, steps: []string{"condition_feed"}},
		{name: "earning_percentage", aliases: []string{"earning_percentage"}, kind: "float", required: true},
		{name: "earnings", kind: "sequence", required: true},
		{name: "event_slug", aliases: []string{"event_slug"}, required: true},
		{name: "image", required: true},
		{name: "maker_address", aliases: []string{"maker_address"}, required: true},
		{name: "market_competitiveness", aliases: []string{"market_competitiveness"}, kind: "float", required: true},
		{name: "market_slug", aliases: []string{"market_slug"}, required: true},
		{name: "question", required: true},
		{name: "rewards_config", aliases: []string{"rewards_config"}, kind: "sequence", required: true},
		{name: "rewards_max_spread", aliases: []string{"rewards_max_spread"}, kind: "float", required: true},
		{name: "rewards_min_size", aliases: []string{"rewards_min_size"}, kind: "data_decimal", required: true},
		{name: "tokens", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = UserRewardsEarning(result)
	return nil
}

// AutoRedeemedNotification is an authenticated exchange response.
type AutoRedeemedNotification struct {
	ID        int64                           `json:"id"`
	Owner     APIKey                          `json:"owner"`
	Timestamp time.Time                       `json:"timestamp"`
	Type      NotificationType                `json:"type"`
	Payload   AutoRedeemedNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *AutoRedeemedNotification) UnmarshalJSON(raw []byte) error {
	type plain AutoRedeemedNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 9},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = AutoRedeemedNotification(result)
	return nil
}

// AutoRedeemedNotificationPayload is an authenticated exchange response.
type AutoRedeemedNotificationPayload struct {
	ProxyWallet     EVMAddress      `json:"proxy_wallet"`
	Amount          *decimal.Big    `json:"amount"`
	ConditionID     ConditionID     `json:"condition_id"`
	Question        string          `json:"question"`
	Image           string          `json:"image"`
	MarketSlug      string          `json:"market_slug"`
	Position        *string         `json:"position"`
	MarketURL       *string         `json:"market_url"`
	PortfolioURL    *string         `json:"portfolio_url"`
	NegRisk         bool            `json:"neg_risk"`
	TransactionHash TransactionHash `json:"transaction_hash"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *AutoRedeemedNotificationPayload) UnmarshalJSON(raw []byte) error {
	type plain AutoRedeemedNotificationPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "proxy_wallet", aliases: []string{"proxyWallet"}, required: true},
		{name: "amount", kind: "data_decimal", required: true},
		{name: "condition_id", aliases: []string{"conditionId"}, required: true},
		{name: "question", required: true},
		{name: "image", required: true},
		{name: "market_slug", aliases: []string{"slug"}, required: true},
		{name: "position", nullable: true},
		{name: "market_url", aliases: []string{"marketUrl"}, nullable: true},
		{name: "portfolio_url", aliases: []string{"portfolioUrl"}, nullable: true},
		{name: "neg_risk", aliases: []string{"negRisk"}, kind: "bool", required: true},
		{name: "transaction_hash", aliases: []string{"txnHash"}, required: true},
	}); err != nil {
		return err
	}
	*m = AutoRedeemedNotificationPayload(result)
	return nil
}

// ChildCommentCreatedNotification is an authenticated exchange response.
type ChildCommentCreatedNotification struct {
	ID        int64                           `json:"id"`
	Owner     APIKey                          `json:"owner"`
	Timestamp time.Time                       `json:"timestamp"`
	Type      NotificationType                `json:"type"`
	Payload   ChildCommentNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *ChildCommentCreatedNotification) UnmarshalJSON(raw []byte) error {
	type plain ChildCommentCreatedNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 6},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = ChildCommentCreatedNotification(result)
	return nil
}

// ChildCommentNotificationPayload is an authenticated exchange response.
type ChildCommentNotificationPayload struct {
	ID               CommentID       `json:"id"`
	Body             *string         `json:"body"`
	ParentEntityType *string         `json:"parent_entity_type"`
	ParentEntityID   *int64          `json:"parent_entity_id"`
	ParentCommentID  *CommentID      `json:"parent_comment_id"`
	UserAddress      *EVMAddress     `json:"user_address"`
	CreatedAt        *time.Time      `json:"created_at"`
	Profile          *CommentProfile `json:"profile"`
	EventSlug        *string         `json:"event_slug"`
	EventTitle       *string         `json:"event_title"`
	SeriesSlug       *string         `json:"series_slug"`
	SeriesTitle      *string         `json:"series_title"`
	Image            *string         `json:"image"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *ChildCommentNotificationPayload) UnmarshalJSON(raw []byte) error {
	type plain ChildCommentNotificationPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", required: true},
		{name: "body", nullable: true},
		{name: "parent_entity_type", aliases: []string{"parentEntityType"}, nullable: true, allowed: []string{"Event", "Series"}},
		{name: "parent_entity_id", aliases: []string{"parentEntityID"}, kind: "int", nullable: true},
		{name: "parent_comment_id", aliases: []string{"parentCommentID"}, nullable: true},
		{name: "user_address", aliases: []string{"userAddress"}, nullable: true},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "clob_epoch_iso", nullable: true},
		{name: "profile", nullable: true},
		{name: "event_slug", aliases: []string{"eventSlug"}, nullable: true},
		{name: "event_title", aliases: []string{"eventTitle"}, nullable: true},
		{name: "series_slug", aliases: []string{"seriesSlug"}, nullable: true},
		{name: "series_title", aliases: []string{"seriesTitle"}, nullable: true},
		{name: "image", nullable: true},
	}); err != nil {
		return err
	}
	*m = ChildCommentNotificationPayload(result)
	return nil
}

// ComboAutoRedeemedNotification is an authenticated exchange response.
type ComboAutoRedeemedNotification struct {
	ID        int64                                `json:"id"`
	Owner     APIKey                               `json:"owner"`
	Timestamp time.Time                            `json:"timestamp"`
	Type      NotificationType                     `json:"type"`
	Payload   ComboAutoRedeemedNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *ComboAutoRedeemedNotification) UnmarshalJSON(raw []byte) error {
	type plain ComboAutoRedeemedNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 10},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = ComboAutoRedeemedNotification(result)
	return nil
}

// ComboAutoRedeemedNotificationPayload is an authenticated exchange response.
type ComboAutoRedeemedNotificationPayload struct {
	ProxyWallet     EVMAddress       `json:"proxy_wallet"`
	Amount          *decimal.Big     `json:"amount"`
	PositionID      PositionID       `json:"position_id"`
	ConditionID     ComboConditionID `json:"condition_id"`
	OutcomeIndex    int64            `json:"outcome_index"`
	Legs            int64            `json:"legs"`
	PortfolioURL    *string          `json:"portfolio_url"`
	TransactionHash TransactionHash  `json:"transaction_hash"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *ComboAutoRedeemedNotificationPayload) UnmarshalJSON(raw []byte) error {
	type plain ComboAutoRedeemedNotificationPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "proxy_wallet", aliases: []string{"proxyWallet"}, required: true},
		{name: "amount", kind: "data_decimal", required: true},
		{name: "position_id", aliases: []string{"positionId"}, required: true},
		{name: "condition_id", aliases: []string{"conditionId"}, required: true, steps: []string{"condition_feed", "validate_combo_condition_id"}},
		{name: "outcome_index", aliases: []string{"outcomeIndex"}, kind: "int", required: true},
		{name: "legs", kind: "int", required: true},
		{name: "portfolio_url", aliases: []string{"portfolioUrl"}, nullable: true},
		{name: "transaction_hash", aliases: []string{"txnHash"}, required: true},
	}); err != nil {
		return err
	}
	*m = ComboAutoRedeemedNotificationPayload(result)
	return nil
}

// MarketNotificationPayload is an authenticated exchange response.
type MarketNotificationPayload struct {
	ConditionID              ConditionID                `json:"condition_id"`
	QuestionID               QuestionID                 `json:"question_id"`
	Question                 string                     `json:"question"`
	Description              string                     `json:"description"`
	MarketSlug               string                     `json:"market_slug"`
	Icon                     string                     `json:"icon"`
	Image                    string                     `json:"image"`
	Fpmm                     string                     `json:"fpmm"`
	Active                   bool                       `json:"active"`
	Closed                   bool                       `json:"closed"`
	Archived                 *bool                      `json:"archived"`
	AcceptingOrders          bool                       `json:"accepting_orders"`
	AcceptingOrdersTimestamp *time.Time                 `json:"accepting_orders_timestamp"`
	EnableOrderBook          *bool                      `json:"enable_order_book"`
	EndDate                  *time.Time                 `json:"end_date"`
	GameStartTime            *time.Time                 `json:"game_start_time"`
	SecondsDelay             int64                      `json:"seconds_delay"`
	MinimumOrderSize         *decimal.Big               `json:"minimum_order_size"`
	MinimumTickSize          *decimal.Big               `json:"minimum_tick_size"`
	MakerBaseFee             *int64                     `json:"maker_base_fee"`
	TakerBaseFee             *int64                     `json:"taker_base_fee"`
	NotificationsEnabled     *bool                      `json:"notifications_enabled"`
	NegRisk                  *bool                      `json:"neg_risk"`
	NegRiskMarketID          *string                    `json:"neg_risk_market_id"`
	NegRiskRequestID         *string                    `json:"neg_risk_request_id"`
	Is5050Outcome            *bool                      `json:"is_50_50_outcome"`
	Rewards                  *MarketNotificationRewards `json:"rewards"`
	Tokens                   []MarketNotificationToken  `json:"tokens"`
	Tags                     []string                   `json:"tags"`
	EventSlug                *string                    `json:"event_slug"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *MarketNotificationPayload) UnmarshalJSON(raw []byte) error {
	type plain MarketNotificationPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "condition_id", required: true},
		{name: "question_id", required: true},
		{name: "question", required: true},
		{name: "description", required: true},
		{name: "market_slug", required: true},
		{name: "icon", required: true},
		{name: "image", required: true},
		{name: "fpmm", required: true},
		{name: "active", kind: "bool", required: true},
		{name: "closed", kind: "bool", required: true},
		{name: "archived", kind: "bool", nullable: true},
		{name: "accepting_orders", kind: "bool", required: true},
		{name: "accepting_orders_timestamp", aliases: []string{"accepting_order_timestamp"}, kind: "clob_epoch_iso", nullable: true},
		{name: "enable_order_book", kind: "bool", nullable: true},
		{name: "end_date", aliases: []string{"end_date_iso"}, kind: "clob_epoch_iso", nullable: true},
		{name: "game_start_time", kind: "clob_epoch_iso", nullable: true},
		{name: "seconds_delay", kind: "int", required: true},
		{name: "minimum_order_size", kind: "data_decimal", required: true},
		{name: "minimum_tick_size", kind: "data_decimal", required: true},
		{name: "maker_base_fee", kind: "int", required: true, nullable: true},
		{name: "taker_base_fee", kind: "int", required: true, nullable: true},
		{name: "notifications_enabled", kind: "bool", nullable: true},
		{name: "neg_risk", kind: "bool", nullable: true},
		{name: "neg_risk_market_id", nullable: true},
		{name: "neg_risk_request_id", nullable: true},
		{name: "is_50_50_outcome", kind: "bool", nullable: true},
		{name: "rewards", nullable: true},
		{name: "tokens", kind: "sequence", required: true},
		{name: "tags", kind: "sequence", nullable: true},
		{name: "event_slug", aliases: []string{"eventSlug"}, nullable: true},
	}); err != nil {
		return err
	}
	*m = MarketNotificationPayload(result)
	return nil
}

// MarketNotificationRewards is an authenticated exchange response.
type MarketNotificationRewards struct {
	MinSize   *decimal.Big                    `json:"min_size"`
	MaxSpread float64                         `json:"max_spread"`
	Rates     []MarketNotificationRewardsRate `json:"rates"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *MarketNotificationRewards) UnmarshalJSON(raw []byte) error {
	type plain MarketNotificationRewards
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "min_size", kind: "data_decimal", required: true},
		{name: "max_spread", kind: "float", required: true},
		{name: "rates", kind: "sequence", nullable: true},
	}); err != nil {
		return err
	}
	*m = MarketNotificationRewards(result)
	return nil
}

// MarketNotificationRewardsRate is an authenticated exchange response.
type MarketNotificationRewardsRate struct {
	AssetAddress string       `json:"asset_address"`
	DailyRate    *decimal.Big `json:"daily_rate"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *MarketNotificationRewardsRate) UnmarshalJSON(raw []byte) error {
	type plain MarketNotificationRewardsRate
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_address", required: true},
		{name: "daily_rate", aliases: []string{"rewards_daily_rate"}, kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = MarketNotificationRewardsRate(result)
	return nil
}

// MarketNotificationToken is an authenticated exchange response.
type MarketNotificationToken struct {
	AssetID ClobAssetID  `json:"asset_id"`
	Outcome string       `json:"outcome"`
	Price   *decimal.Big `json:"price"`
	Winner  bool         `json:"winner"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *MarketNotificationToken) UnmarshalJSON(raw []byte) error {
	type plain MarketNotificationToken
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "outcome", required: true},
		{name: "price", kind: "data_decimal", nullable: true},
		{name: "winner", kind: "bool", required: true},
	}); err != nil {
		return err
	}
	*m = MarketNotificationToken(result)
	return nil
}

// MarketRegisteredNotification is an authenticated exchange response.
type MarketRegisteredNotification struct {
	ID        int64                     `json:"id"`
	Owner     APIKey                    `json:"owner"`
	Timestamp time.Time                 `json:"timestamp"`
	Type      NotificationType          `json:"type"`
	Payload   MarketNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *MarketRegisteredNotification) UnmarshalJSON(raw []byte) error {
	type plain MarketRegisteredNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 3},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = MarketRegisteredNotification(result)
	return nil
}

// MarketResolvedNotification is an authenticated exchange response.
type MarketResolvedNotification struct {
	ID        int64                     `json:"id"`
	Owner     APIKey                    `json:"owner"`
	Timestamp time.Time                 `json:"timestamp"`
	Type      NotificationType          `json:"type"`
	Payload   MarketNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *MarketResolvedNotification) UnmarshalJSON(raw []byte) error {
	type plain MarketResolvedNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 4},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = MarketResolvedNotification(result)
	return nil
}

// OrderCancellationNotification is an authenticated exchange response.
type OrderCancellationNotification struct {
	ID        int64                    `json:"id"`
	Owner     APIKey                   `json:"owner"`
	Timestamp time.Time                `json:"timestamp"`
	Type      NotificationType         `json:"type"`
	Payload   OrderNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *OrderCancellationNotification) UnmarshalJSON(raw []byte) error {
	type plain OrderCancellationNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 1},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = OrderCancellationNotification(result)
	return nil
}

// OrderFillFailedNotification is an authenticated exchange response.
type OrderFillFailedNotification struct {
	ID        int64                    `json:"id"`
	Owner     APIKey                   `json:"owner"`
	Timestamp time.Time                `json:"timestamp"`
	Type      NotificationType         `json:"type"`
	Payload   OrderNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *OrderFillFailedNotification) UnmarshalJSON(raw []byte) error {
	type plain OrderFillFailedNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 8},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = OrderFillFailedNotification(result)
	return nil
}

// OrderFillNotification is an authenticated exchange response.
type OrderFillNotification struct {
	ID        int64                    `json:"id"`
	Owner     APIKey                   `json:"owner"`
	Timestamp time.Time                `json:"timestamp"`
	Type      NotificationType         `json:"type"`
	Payload   OrderNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *OrderFillNotification) UnmarshalJSON(raw []byte) error {
	type plain OrderFillNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 2},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = OrderFillNotification(result)
	return nil
}

// OrderNotificationPayload is an authenticated exchange response.
type OrderNotificationPayload struct {
	AssetID         ClobAssetID      `json:"asset_id"`
	ConditionID     ConditionID      `json:"condition_id"`
	OrderID         OrderID          `json:"order_id"`
	Side            OrderSide        `json:"side"`
	OrderType       *string          `json:"order_type"`
	Price           *decimal.Big     `json:"price"`
	OriginalSize    *decimal.Big     `json:"original_size"`
	MatchedSize     *decimal.Big     `json:"matched_size"`
	RemainingSize   *decimal.Big     `json:"remaining_size"`
	Outcome         string           `json:"outcome"`
	OutcomeIndex    int64            `json:"outcome_index"`
	TransactionHash *TransactionHash `json:"transaction_hash"`
	TradeID         *string          `json:"trade_id"`
	Question        *string          `json:"question"`
	MarketSlug      *string          `json:"market_slug"`
	Icon            *string          `json:"icon"`
	Image           *string          `json:"image"`
	EventSlug       *string          `json:"event_slug"`
	SeriesSlug      *string          `json:"series_slug"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *OrderNotificationPayload) UnmarshalJSON(raw []byte) error {
	type plain OrderNotificationPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_id", aliases: []string{"asset_id", "token_id"}, required: true},
		{name: "condition_id", aliases: []string{"market"}, required: true},
		{name: "order_id", required: true},
		{name: "side", required: true, allowed: []string{"BUY", "SELL"}},
		{name: "order_type", aliases: []string{"type"}, nullable: true, allowed: []string{"GTC", "GTD", "FAK", "FOK"}, steps: []string{"empty_none"}},
		{name: "price", kind: "string_decimal", required: true},
		{name: "original_size", kind: "string_decimal", required: true},
		{name: "matched_size", kind: "string_decimal", required: true},
		{name: "remaining_size", kind: "string_decimal", required: true},
		{name: "outcome", required: true},
		{name: "outcome_index", kind: "int", required: true},
		{name: "transaction_hash", nullable: true, steps: []string{"empty_none"}},
		{name: "trade_id", nullable: true, steps: []string{"empty_none"}},
		{name: "question", nullable: true},
		{name: "market_slug", nullable: true},
		{name: "icon", nullable: true},
		{name: "image", nullable: true},
		{name: "event_slug", aliases: []string{"eventSlug"}, nullable: true},
		{name: "series_slug", aliases: []string{"seriesSlug"}, nullable: true},
	}); err != nil {
		return err
	}
	*m = OrderNotificationPayload(result)
	return nil
}

// RewardPayoutNotification is an authenticated exchange response.
type RewardPayoutNotification struct {
	ID        int64                           `json:"id"`
	Owner     APIKey                          `json:"owner"`
	Timestamp time.Time                       `json:"timestamp"`
	Type      NotificationType                `json:"type"`
	Payload   RewardPayoutNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *RewardPayoutNotification) UnmarshalJSON(raw []byte) error {
	type plain RewardPayoutNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 5},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = RewardPayoutNotification(result)
	return nil
}

// RewardPayoutNotificationPayload is an authenticated exchange response.
type RewardPayoutNotificationPayload struct {
	ProxyWallet     EVMAddress      `json:"proxy_wallet"`
	Reward          *decimal.Big    `json:"reward"`
	TransactionHash TransactionHash `json:"transaction_hash"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *RewardPayoutNotificationPayload) UnmarshalJSON(raw []byte) error {
	type plain RewardPayoutNotificationPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "proxy_wallet", aliases: []string{"proxyWallet"}, required: true},
		{name: "reward", kind: "data_decimal", required: true},
		{name: "transaction_hash", aliases: []string{"txnHash"}, required: true},
	}); err != nil {
		return err
	}
	*m = RewardPayoutNotificationPayload(result)
	return nil
}

// TakerTierUpgradedNotification is an authenticated exchange response.
type TakerTierUpgradedNotification struct {
	ID        int64                                `json:"id"`
	Owner     APIKey                               `json:"owner"`
	Timestamp time.Time                            `json:"timestamp"`
	Type      NotificationType                     `json:"type"`
	Payload   TakerTierUpgradedNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *TakerTierUpgradedNotification) UnmarshalJSON(raw []byte) error {
	type plain TakerTierUpgradedNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 11},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = TakerTierUpgradedNotification(result)
	return nil
}

// TakerTierUpgradedNotificationPayload is an authenticated exchange response.
type TakerTierUpgradedNotificationPayload struct {
	PreviousTier int64 `json:"previous_tier"`
	Tier         int64 `json:"tier"`
	RebateBps    int64 `json:"rebate_bps"`
	SnapshotDate Date  `json:"snapshot_date"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *TakerTierUpgradedNotificationPayload) UnmarshalJSON(raw []byte) error {
	type plain TakerTierUpgradedNotificationPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "previous_tier", kind: "strict_int", required: true},
		{name: "tier", kind: "strict_int", required: true},
		{name: "rebate_bps", kind: "strict_int", required: true},
		{name: "snapshot_date", kind: "date", required: true},
	}); err != nil {
		return err
	}
	*m = TakerTierUpgradedNotificationPayload(result)
	return nil
}

// YieldPayoutNotification is an authenticated exchange response.
type YieldPayoutNotification struct {
	ID        int64                          `json:"id"`
	Owner     APIKey                         `json:"owner"`
	Timestamp time.Time                      `json:"timestamp"`
	Type      NotificationType               `json:"type"`
	Payload   YieldPayoutNotificationPayload `json:"payload"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *YieldPayoutNotification) UnmarshalJSON(raw []byte) error {
	type plain YieldPayoutNotification
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "id", kind: "notification_id", required: true},
		{name: "owner", required: true},
		{name: "timestamp", kind: "clob_epoch_iso", required: true},
		{name: "type", required: true, kind: "notification_kind", minimum: 7},
		{name: "payload", required: true},
	}); err != nil {
		return err
	}
	*m = YieldPayoutNotification(result)
	return nil
}

// YieldPayoutNotificationPayload is an authenticated exchange response.
type YieldPayoutNotificationPayload struct {
	ProxyWallet     EVMAddress      `json:"proxy_wallet"`
	Amount          *decimal.Big    `json:"amount"`
	TransactionHash TransactionHash `json:"transaction_hash"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *YieldPayoutNotificationPayload) UnmarshalJSON(raw []byte) error {
	type plain YieldPayoutNotificationPayload
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "proxy_wallet", aliases: []string{"proxyWallet"}, required: true},
		{name: "amount", kind: "data_decimal", required: true},
		{name: "transaction_hash", aliases: []string{"txnHash"}, required: true},
	}); err != nil {
		return err
	}
	*m = YieldPayoutNotificationPayload(result)
	return nil
}

// CollateralReturnOperation is an authenticated exchange response.
type CollateralReturnOperation struct {
	Kind           string            `json:"kind"`
	ConditionID    *ComboConditionID `json:"condition_id"`
	EventID        *EventID          `json:"event_id"`
	PositionID     *PositionID       `json:"position_id"`
	ConditionIndex int64             `json:"condition_index"`
	Amount         *decimal.Big      `json:"amount"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *CollateralReturnOperation) UnmarshalJSON(raw []byte) error {
	type plain CollateralReturnOperation
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "kind", required: true},
		{name: "condition_id", nullable: true, steps: []string{"validate_combo_condition_id", "condition_feed"}},
		{name: "event_id", nullable: true},
		{name: "position_id", nullable: true},
		{name: "condition_index", kind: "int", fallback: json.RawMessage("0")},
		{name: "amount", kind: "e6_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = CollateralReturnOperation(result)
	return nil
}

// CollateralReturnPositionAmount is an authenticated exchange response.
type CollateralReturnPositionAmount struct {
	PositionID PositionID   `json:"position_id"`
	Amount     *decimal.Big `json:"amount"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *CollateralReturnPositionAmount) UnmarshalJSON(raw []byte) error {
	type plain CollateralReturnPositionAmount
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "position_id", required: true},
		{name: "amount", kind: "e6_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = CollateralReturnPositionAmount(result)
	return nil
}

// CollateralReturnPositionSummary is an authenticated exchange response.
type CollateralReturnPositionSummary struct {
	Consumed []CollateralReturnPositionAmount `json:"consumed"`
	Created  []CollateralReturnPositionAmount `json:"created"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *CollateralReturnPositionSummary) UnmarshalJSON(raw []byte) error {
	type plain CollateralReturnPositionSummary
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "consumed", kind: "sequence", fallback: json.RawMessage("[]")},
		{name: "created", kind: "sequence", fallback: json.RawMessage("[]")},
	}); err != nil {
		return err
	}
	*m = CollateralReturnPositionSummary(result)
	return nil
}

// CollateralReturnRouterCall is an authenticated exchange response.
type CollateralReturnRouterCall struct {
	To   EVMAddress `json:"to"`
	Data HexString  `json:"data"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *CollateralReturnRouterCall) UnmarshalJSON(raw []byte) error {
	type plain CollateralReturnRouterCall
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "to", required: true},
		{name: "data", required: true},
	}); err != nil {
		return err
	}
	*m = CollateralReturnRouterCall(result)
	return nil
}

// EarningBreakdown is an authenticated exchange response.
type EarningBreakdown struct {
	AssetAddress string       `json:"asset_address"`
	AssetRate    *decimal.Big `json:"asset_rate"`
	Earnings     *decimal.Big `json:"earnings"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *EarningBreakdown) UnmarshalJSON(raw []byte) error {
	type plain EarningBreakdown
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_address", aliases: []string{"asset_address"}, required: true},
		{name: "asset_rate", aliases: []string{"asset_rate"}, kind: "data_decimal", required: true},
		{name: "earnings", kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = EarningBreakdown(result)
	return nil
}

// UserRewardsConfig is an authenticated exchange response.
type UserRewardsConfig struct {
	AssetAddress string       `json:"asset_address"`
	EndDate      time.Time    `json:"end_date"`
	RatePerDay   *decimal.Big `json:"rate_per_day"`
	StartDate    time.Time    `json:"start_date"`
	TotalRewards *decimal.Big `json:"total_rewards"`
}

// UnmarshalJSON validates canonical fields and exchange wire aliases.
func (m *UserRewardsConfig) UnmarshalJSON(raw []byte) error {
	type plain UserRewardsConfig
	var result plain
	if err := decodeData(raw, &result, []dataField{
		{name: "asset_address", aliases: []string{"asset_address"}, required: true},
		{name: "end_date", aliases: []string{"end_date"}, kind: "reward_epoch_ms", required: true},
		{name: "rate_per_day", aliases: []string{"rate_per_day"}, kind: "data_decimal", required: true},
		{name: "start_date", aliases: []string{"start_date"}, kind: "reward_epoch_ms", required: true},
		{name: "total_rewards", aliases: []string{"total_rewards"}, kind: "data_decimal", required: true},
	}); err != nil {
		return err
	}
	*m = UserRewardsConfig(result)
	return nil
}
