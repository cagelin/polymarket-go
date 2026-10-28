package models

import "math/big"

// ERC20TradingApproval is a missing token allowance, expressed in base units.
type ERC20TradingApproval struct {
	TokenAddress EVMAddress `json:"token_address"`
	Spender      EVMAddress `json:"spender"`
	Amount       *big.Int   `json:"amount"`
}

// ERC1155TradingApproval is a missing operator approval.
type ERC1155TradingApproval struct {
	TokenAddress EVMAddress `json:"token_address"`
	Operator     EVMAddress `json:"operator"`
}

// MissingTradingApprovals lists approvals still required by the pinned environment.
type MissingTradingApprovals struct {
	ERC20   []ERC20TradingApproval   `json:"erc20"`
	ERC1155 []ERC1155TradingApproval `json:"erc1155"`
}

// TradingApprovalsState reflects indexed state; recent on-chain changes may lag.
type TradingApprovalsState struct {
	Missing         MissingTradingApprovals `json:"missing"`
	IsFullyApproved bool                    `json:"is_fully_approved"`
}
