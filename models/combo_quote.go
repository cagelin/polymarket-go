package models

import (
	"fmt"

	"github.com/ericlagergren/decimal"
)

// RFQErrorDetail preserves known and future server error codes.
type RFQErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ComboQuote contains all fields needed to sign acceptance. Treat persisted quotes as signing-sensitive data.
type ComboQuote struct {
	RFQID         string       `json:"rfq_id"`
	QuoteID       string       `json:"quote_id"`
	BuilderCode   string       `json:"builder_code"`
	Direction     OrderSide    `json:"direction"`
	PositionID    string       `json:"position_id"`
	BlendedPrice  *decimal.Big `json:"blended_price"`
	MakerAmount   *decimal.Big `json:"maker_amount"`
	TakerAmount   *decimal.Big `json:"taker_amount"`
	TotalRequired *decimal.Big `json:"total_required"`
	NetReceive    *decimal.Big `json:"net_receive"`
	ExpiresAt     int64        `json:"expires_at"`
}

// UnmarshalJSON validates a persisted quote without accepting deprecated field aliases.
func (q *ComboQuote) UnmarshalJSON(raw []byte) error {
	type plain ComboQuote
	var value plain
	if err := decodeData(raw, &value, []dataField{
		{name: "rfq_id", required: true}, {name: "quote_id", required: true}, {name: "builder_code", required: true}, {name: "direction", required: true, allowed: []string{"BUY", "SELL"}}, {name: "position_id", required: true},
		{name: "blended_price", kind: "data_decimal", required: true}, {name: "maker_amount", kind: "data_decimal", required: true}, {name: "taker_amount", kind: "data_decimal", required: true}, {name: "total_required", kind: "data_decimal", required: true}, {name: "net_receive", kind: "data_decimal", nullable: true}, {name: "expires_at", kind: "int", required: true},
	}); err != nil {
		return err
	}
	if value.Direction == Sell && value.NetReceive == nil {
		return fmt.Errorf("SELL quote requires net receive")
	}
	*q = ComboQuote(value)
	return nil
}

// ComboQuoteResult returns a quote or a normal no-quote outcome.
type ComboQuoteResult struct {
	RFQID  string      `json:"rfq_id"`
	Quote  *ComboQuote `json:"quote"`
	Reason *string     `json:"reason"`
}

// ComboQuoteAcceptance distinguishes execution handoff from maker refusal or expiry.
type ComboQuoteAcceptance struct {
	RFQID          string          `json:"rfq_id"`
	Status         string          `json:"status"`
	TakerOrderHash *string         `json:"taker_order_hash"`
	Reason         *string         `json:"reason"`
	Error          *RFQErrorDetail `json:"error"`
}

// RFQStatusResult reports both RFQ lifecycle and chain execution state.
type RFQStatusResult struct {
	RFQID          string           `json:"rfq_id"`
	Status         string           `json:"status"`
	TakerOrderHash *string          `json:"taker_order_hash"`
	TxHash         *TransactionHash `json:"tx_hash"`
	Error          *RFQErrorDetail  `json:"error"`
}

// ComboFillResult reports a terminal combo outcome, including failed or canceled fills.
type ComboFillResult struct {
	RFQID  string           `json:"rfq_id"`
	Status string           `json:"status"`
	TxHash *TransactionHash `json:"tx_hash"`
	Error  *RFQErrorDetail  `json:"error"`
}
