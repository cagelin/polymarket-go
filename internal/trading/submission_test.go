package trading

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cagelin/polymarket-go/models"
)

func TestPythonBatchOrderSubmission(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/order-post-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Order models.SignedOrder
			Body  json.RawMessage
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) < 2 {
		t.Fatal("missing batch input fixtures")
	}
	want := []json.RawMessage{fixture.Cases[0].Body, fixture.Cases[1].Body}
	var requests atomic.Int32
	client := fixtureClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != "POST" || r.URL.Path != "/orders" {
			t.Errorf("request = %s %s, want POST /orders", r.Method, r.URL.Path)
		}
		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			t.Error(readErr)
			return
		}
		expected, encodeErr := json.Marshal(want)
		if encodeErr != nil {
			t.Error(encodeErr)
			return
		}
		equalJSON(t, json.RawMessage(body), expected)
		writeResponse(t, w, []map[string]any{
			{"errorMsg": "", "makingAmount": "1", "takingAmount": "2", "orderID": "first", "success": true, "status": "live"},
			{"errorMsg": "", "makingAmount": "1", "takingAmount": "2", "orderID": "second", "success": true, "status": "live"},
		})
	})
	results, err := client.PostOrders(t.Context(), []models.SignedOrder{fixture.Cases[0].Order, fixture.Cases[1].Order})
	if err != nil || len(results) != 2 || requests.Load() != 1 {
		t.Fatalf("PostOrders = %v, %v; requests=%d, want two results from one request", results, err, requests.Load())
	}
	for _, result := range results {
		if _, ok := result.(*models.AcceptedOrder); !ok {
			t.Errorf("PostOrders result type %T, want AcceptedOrder", result)
		}
	}
}

func TestPlaceMarketOrderSubmitsOnce(t *testing.T) {
	var submits atomic.Int32
	client := fixtureClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/markets-by-token/"):
			writeResponse(t, w, map[string]string{"condition_id": "0x" + strings.Repeat("ab", 32)})
		case strings.HasPrefix(r.URL.Path, "/clob-markets/"):
			writeResponse(t, w, map[string]any{"mts": "0.01", "nr": false, "t": []any{map[string]string{"t": "123"}}})
		case r.URL.Path == "/order":
			submits.Add(1)
			var body struct {
				OrderType string
				Order     struct{ TokenID, MakerAmount, TakerAmount, Side, Signature string }
			}
			if decodeErr := json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
				t.Error(decodeErr)
				return
			}
			if body.OrderType != "FAK" || body.Order.TokenID != "123" || body.Order.Side != "BUY" || body.Order.MakerAmount != "1000000" || body.Order.TakerAmount != "2000000" || body.Order.Signature == "" {
				t.Errorf("PlaceMarketOrder submitted unexpected signed order: %+v", body)
			}
			writeResponse(t, w, map[string]any{"errorMsg": "", "makingAmount": "1", "takingAmount": "2", "orderID": "market", "success": true, "status": "matched"})
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			w.WriteHeader(500)
		}
	})
	_, err := client.PlaceMarketOrder(t.Context(), CreateMarketOrderParams{AssetID: "123", Side: models.Buy, Amount: dec("1"), MaxPrice: dec("0.5")})
	if err != nil || submits.Load() != 1 {
		t.Fatalf("PlaceMarketOrder error=%v submits=%d, want one successful submission", err, submits.Load())
	}
}
