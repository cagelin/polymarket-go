package models

import "math/big"

// SplitPositionParams selects exactly one market condition or combo leg set.
type SplitPositionParams struct {
	ConditionID ConditionID
	Legs        []string
	Amount      *big.Int
	Metadata    *string
}

// MergePositionsParams selects complementary outcomes. Nil Amount means maximum.
type MergePositionsParams struct {
	ConditionID ConditionID
	Legs        []string
	Amount      *big.Int
	Metadata    *string
}

// MergePositionRequest selects one condition, market or V2 position. Nil Amount means maximum.
type MergePositionRequest struct {
	ConditionID ConditionID
	MarketID    MarketID
	PositionID  PositionID
	Amount      *big.Int
}

// RedeemPositionsParams selects exactly one condition, market or V2 position.
type RedeemPositionsParams struct {
	ConditionID ConditionID
	MarketID    MarketID
	PositionID  PositionID
	Metadata    *string
}
