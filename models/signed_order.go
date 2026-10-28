package models

import "math/big"

// OrderType controls how long an exchange order can remain open.
type OrderType string

// Resting order lifetimes; market orders use FAK or FOK.
const (
	GTC OrderType = "GTC"
	GTD OrderType = "GTD"
)

// SignedOrder carries a locally signed order. Creating one never submits a trade.
// Amounts are integer base units; AssetID is serialized as tokenId only on the wire.
type SignedOrder struct {
	Builder       string      `json:"builder"`
	Expiration    int64       `json:"expiration"`
	Maker         EVMAddress  `json:"maker"`
	MakerAmount   *big.Int    `json:"maker_amount"`
	Metadata      string      `json:"metadata"`
	OrderType     OrderType   `json:"order_type"`
	Salt          *big.Int    `json:"salt"`
	Side          OrderSide   `json:"side"`
	Signature     string      `json:"signature"`
	SignatureType int         `json:"signature_type"`
	Signer        EVMAddress  `json:"signer"`
	TakerAmount   *big.Int    `json:"taker_amount"`
	Timestamp     int64       `json:"timestamp"`
	AssetID       ClobAssetID `json:"asset_id"`
	PostOnly      bool        `json:"post_only"`
}
