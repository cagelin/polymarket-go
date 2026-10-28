package models

import "math/big"

// TransactionCall is one EVM call. A nil Value means zero native tokens.
type TransactionCall struct {
	To    EVMAddress `json:"to"`
	Data  string     `json:"data"`
	Value *big.Int   `json:"value,omitempty"`
}

// TransactionOutcome identifies a confirmed on-chain transaction.
type TransactionOutcome struct {
	TransactionHash string  `json:"transaction_hash"`
	TransactionID   *string `json:"transaction_id"`
}

// RelayerAPIKey authorizes sponsored wallet operations.
type RelayerAPIKey struct {
	Key     string     `json:"key"`
	Address EVMAddress `json:"address"`
}

func (k RelayerAPIKey) String() string { return "RelayerAPIKey(<redacted>)" }

// GoString omits credential values from diagnostic formatting.
func (k RelayerAPIKey) GoString() string { return k.String() }
