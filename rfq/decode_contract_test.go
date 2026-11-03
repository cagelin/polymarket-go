package rfq

import (
	"encoding/json"
	"strings"
	"testing"
)

func rfqWireEvent(kind string) map[string]any {
	return map[string]any{
		"type": kind, "rfq_id": "request-1", "quote_id": "quote-1",
		"requestor_public_id": "maker-1", "requester_id": "taker-1",
		"condition_id": "0x03" + strings.Repeat("ab", 30), "leg_position_ids": []string{"123", "456"},
		"yes_position_id": "123", "no_position_id": "124", "direction": "BUY", "side": "YES",
		"requested_size":      map[string]any{"unit": "notional", "value_e6": "1234567"},
		"submission_deadline": int64(1700000000), "confirm_by": int64(1700000001),
		"signature_type": 3, "signer_address": "0x1111111111111111111111111111111111111111",
		"maker_address": "0x2222222222222222222222222222222222222222",
		"fill_size_e6":  "2345678", "price_e6": "123456", "size_e6": "3456789",
		"executed_at": int64(1700000002), "status": "CONFIRMED",
	}
}

func encodeRFQEvent(t *testing.T, fields map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("Marshal(RFQ fields) = %v, want nil", err)
	}
	return raw
}

func TestRFQConfirmationAndTradeAmounts(t *testing.T) {
	confirmation, err := parseEvent(encodeRFQEvent(t, rfqWireEvent("RFQ_CONFIRMATION_REQUEST")))
	if err != nil {
		t.Fatalf("parseEvent(confirmation) = %v, want nil", err)
	}
	if confirmation.Type != "confirmation_request" || confirmation.FillSize.String() != "2.345678" || confirmation.Price.String() != "0.123456" || confirmation.ConfirmBy != 1700000001 || confirmation.SignatureType != 3 {
		t.Errorf("parseEvent(confirmation) = %+v, want exact amounts, signature type and deadline", confirmation)
	}
	trade, err := parseEvent(encodeRFQEvent(t, rfqWireEvent("RFQ_TRADE")))
	if err != nil {
		t.Fatalf("parseEvent(trade) = %v, want nil", err)
	}
	if trade.Type != "trade" || trade.Size.String() != "3.456789" || trade.Price.String() != "0.123456" || trade.ExecutedAt != 1700000002 {
		t.Errorf("parseEvent(trade) = %+v, want exact price, size and execution time", trade)
	}
}

func TestRFQRejectsMalformedEvents(t *testing.T) {
	for _, sample := range []string{`null`, `[]`, `{}`, `{"type":1}`, `{"type":"FUTURE_EVENT"}`} {
		if _, err := parseEvent([]byte(sample)); err == nil {
			t.Errorf("parseEvent(%s) = nil error, want rejection", sample)
		}
	}
	cases := []struct {
		kind, field string
		value       any
	}{
		{"RFQ_REQUEST", "rfq_id", ""}, {"RFQ_REQUEST", "rfq_id", nil},
		{"RFQ_REQUEST", "direction", "HOLD"}, {"RFQ_REQUEST", "side", "NO"},
		{"RFQ_REQUEST", "condition_id", "0x01"}, {"RFQ_REQUEST", "leg_position_ids", nil},
		{"RFQ_REQUEST", "leg_position_ids", []int{1}}, {"RFQ_REQUEST", "submission_deadline", nil},
		{"RFQ_REQUEST", "requested_size", nil},
		{"RFQ_REQUEST", "requested_size", map[string]any{"value_e6": "1"}},
		{"RFQ_REQUEST", "requested_size", map[string]any{"unit": "coins", "value_e6": "1"}},
		{"RFQ_REQUEST", "requested_size", map[string]any{"unit": "shares", "value_e6": "-1"}},
		{"RFQ_CONFIRMATION_REQUEST", "confirm_by", nil},
		{"RFQ_CONFIRMATION_REQUEST", "signature_type", nil},
		{"RFQ_CONFIRMATION_REQUEST", "fill_size_e6", nil},
		{"RFQ_CONFIRMATION_REQUEST", "price_e6", "1.5"},
		{"RFQ_TRADE", "executed_at", nil}, {"RFQ_TRADE", "price_e6", ""},
		{"RFQ_TRADE", "size_e6", "NaN"}, {"RFQ_EXECUTION_UPDATE", "status", "UNKNOWN"},
	}
	for _, sample := range cases {
		t.Run(sample.kind+"/"+sample.field, func(subtest *testing.T) {
			fields := rfqWireEvent(sample.kind)
			fields[sample.field] = sample.value
			if _, err := parseEvent(encodeRFQEvent(subtest, fields)); err == nil {
				subtest.Errorf("parseEvent(%s, %s=%v) = nil error, want rejection", sample.kind, sample.field, sample.value)
			}
		})
	}
}
