// Package rfq provides authenticated, long-lived quote sessions.
package rfq

import (
	"github.com/cagelam/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

// Event contains one quote request, confirmation request, execution update or trade.
// Type determines which fields are populated.
type Event struct {
	Type               string             `json:"type"`
	RFQID              string             `json:"rfq_id"`
	QuoteID            string             `json:"quote_id,omitempty"`
	RequestorPublicID  string             `json:"requestor_public_id,omitempty"`
	RequesterID        string             `json:"requester_id,omitempty"`
	LegPositionIDs     []string           `json:"leg_position_ids,omitempty"`
	ConditionID        models.ConditionID `json:"condition_id,omitempty"`
	YesPositionID      string             `json:"yes_position_id,omitempty"`
	NoPositionID       string             `json:"no_position_id,omitempty"`
	Direction          models.OrderSide   `json:"direction,omitempty"`
	Side               string             `json:"side,omitempty"`
	RequestedSize      RequestedSize      `json:"requested_size"`
	SubmissionDeadline int64              `json:"submission_deadline,omitempty"`
	SignerAddress      models.EVMAddress  `json:"signer_address,omitempty"`
	MakerAddress       models.EVMAddress  `json:"maker_address,omitempty"`
	SignatureType      int                `json:"signature_type,omitempty"`
	FillSize           *decimal.Big       `json:"fill_size,omitempty"`
	Price              *decimal.Big       `json:"price,omitempty"`
	Size               *decimal.Big       `json:"size,omitempty"`
	ConfirmBy          int64              `json:"confirm_by,omitempty"`
	ExecutedAt         int64              `json:"executed_at,omitempty"`
	Status             string             `json:"status,omitempty"`
	TransactionHash    *string            `json:"tx_hash,omitempty"`
}

// RequestedSize uses decimal shares or collateral notional.
type RequestedSize struct {
	Unit  string       `json:"unit"`
	Value *decimal.Big `json:"value"`
}

// QuoteReference identifies an acknowledged quote.
type QuoteReference struct {
	RFQID   string `json:"rfq_id"`
	QuoteID string `json:"quote_id"`
}

// QuoteParams supplies a decimal price and optional size. Source defaults to collateral.
type QuoteParams struct {
	Price  *decimal.Big
	Size   *decimal.Big
	Source string
}

// Rejection retains the server's evolving rejection code and correlation identifiers.
type Rejection struct{ Message, Code, ErrorID, RequestType, RFQID, QuoteID string }

func (e *Rejection) Error() string { return e.Message }
