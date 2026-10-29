package trading

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
)

func TestRequesterCompetitionAcceptanceAndSettlement(t *testing.T) {
	for _, scenario := range []string{"filled", "no-quote", "declined"} {
		t.Run(scenario, func(subtest *testing.T) {
			var requests, accepts atomic.Int32
			client := fixtureClient(subtest, func(writer http.ResponseWriter, request *http.Request) {
				switch {
				case request.Method == "POST" && request.URL.Path == comboRequestsPath:
					requests.Add(1)
					if scenario == "no-quote" {
						writeResponse(subtest, writer, map[string]any{"rfq_id": "r", "status": "FAILED", "error": map[string]string{"code": "NO_QUOTES", "message": "no test counterparty"}})
						return
					}
					var body map[string]any
					if decodeErr := json.NewDecoder(request.Body).Decode(&body); decodeErr != nil {
						subtest.Error(decodeErr)
						return
					}
					body["rfq_id"] = "r"
					body["condition_id"] = "0x03" + strings.Repeat("ab", 30)
					body["yes_position_id"] = "123"
					body["no_position_id"] = "124"
					writeResponse(subtest, writer, map[string]any{"rfq_id": "r", "status": "AWAITING_REQUESTER_ACCEPTANCE", "expires_at": 1720000030000, "builder_code": signing.ZeroBytes32, "request": body, "quote": map[string]string{"quote_id": "q", "blended_price_e6": "500000", "maker_amount_e6": "50000", "taker_amount_e6": "100000", "total_required_e6": "50000"}})
				case request.Method == "POST" && request.URL.Path == comboRequestsPath+"/r/accept":
					accepts.Add(1)
					if scenario == "declined" {
						writeResponse(subtest, writer, map[string]any{"rfq_id": "r", "status": "FAILED", "error": map[string]string{"code": "MAKER_DECLINED", "message": "declined"}})
						return
					}
					writeResponse(subtest, writer, map[string]string{"rfq_id": "r", "status": "EXECUTING"})
				case request.Method == "GET" && request.URL.Path == comboRequestsPath+"/r":
					writeResponse(subtest, writer, map[string]string{"rfq_id": "r", "status": "CONFIRMED", "tx_hash": "0x" + strings.Repeat("a", 64)})
				default:
					subtest.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
					writer.WriteHeader(500)
				}
			})
			result, err := client.RequestComboQuote(subtest.Context(), RequestComboQuoteParams{LegPositionIDs: []string{"123", "456"}, Direction: models.Buy, Amount: dec("0.05")})
			if err != nil {
				subtest.Fatal(err)
			}
			if scenario == "no-quote" {
				if result.Quote != nil || result.Reason == nil || *result.Reason != "NO_QUOTES" || accepts.Load() != 0 {
					subtest.Fatal("no-quote was treated as acceptance")
				}
				return
			}
			if result.Quote == nil {
				subtest.Fatal("missing quote")
			}
			accepted, err := client.AcceptComboQuote(subtest.Context(), *result.Quote)
			if err != nil {
				subtest.Fatal(err)
			}
			if scenario == "declined" {
				if accepted.Status != "failed" || accepted.Reason == nil || *accepted.Reason != "MAKER_DECLINED" {
					subtest.Fatalf("lost refusal: %+v", accepted)
				}
				return
			}
			filled, err := client.WaitForComboFill(subtest.Context(), "r", time.Second, time.Millisecond)
			if err != nil || filled.Status != "FILLED" || filled.TxHash == nil || requests.Load() != 1 || accepts.Load() != 1 {
				subtest.Fatalf("fill=%+v err=%v requests=%d accepts=%d", filled, err, requests.Load(), accepts.Load())
			}
		})
	}
}
