package trading

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/internal/transport"
	"github.com/cagelin/polymarket-go/models"
)

func fixtureClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	tr, err := transport.New(transport.Config{CLOBURL: server.URL, BuilderGatewayURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tr.Close() })
	signer, err := signing.New("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	//nolint:gosec // Public synthetic credentials for a local test server only.
	//noinspection SpellCheckingInspection
	return &Client{transport: tr, signer: signer, wallet: models.EVMAddress(signer.Address()), walletType: models.WalletEOA, signerType: models.SignerOwner, credentials: models.APIKeyCreds{Key: "key", Secret: "c2VjcmV0", Passphrase: "pass"}, builderKey: &models.BuilderAPIKey{Key: "builder", Secret: "c2VjcmV0", Passphrase: "pass"}, now: func() time.Time { return time.Unix(1720000000, 0) }, salt: func(int) (*big.Int, error) { return big.NewInt(12345), nil }}
}
func writeResponse(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Error(err)
	}
}
func TestPythonOrderPost(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/order-post-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Order     models.SignedOrder
			Body      json.RawMessage
			Signature string
		}
		Responses []struct {
			Response json.RawMessage
			Expected struct {
				Value json.RawMessage
				Error string
			}
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for i, tc := range fixture.Cases {
		t.Run(fmt.Sprintf("payload/%d", i), func(subtest *testing.T) {
			client := fixtureClient(subtest, func(w http.ResponseWriter, r *http.Request) {
				body, readErr := io.ReadAll(r.Body)
				if readErr != nil {
					subtest.Error(readErr)
				}
				equalJSON(subtest, json.RawMessage(body), tc.Body)
				if r.Header.Get("POLY_SIGNATURE") != tc.Signature {
					subtest.Error("wire body differs from the Python HMAC payload")
				}
				writeResponse(subtest, w, map[string]any{"errorMsg": "", "makingAmount": "1", "takingAmount": "2", "orderID": "order", "success": true, "status": "live"})
			})
			result, callErr := client.PostOrder(subtest.Context(), tc.Order)
			if callErr != nil {
				subtest.Fatal(callErr)
			}
			if _, ok := result.(*models.AcceptedOrder); !ok {
				subtest.Fatal("successful response not accepted")
			}
		})
	}
	for i, tc := range fixture.Responses {
		t.Run(fmt.Sprintf("response/%d", i), func(subtest *testing.T) {
			var orderResponse models.RawOrderResponse
			parseErr := json.Unmarshal(tc.Response, &orderResponse)
			if (parseErr != nil) != (tc.Expected.Error != "") {
				subtest.Fatalf("error %v; Python %s", parseErr, tc.Expected.Error)
			}
			if parseErr == nil {
				equalJSON(subtest, models.NormalizeOrderResponse(orderResponse), tc.Expected.Value)
			}
		})
	}
}
func TestCreateOrderRefreshAndConcurrentCache(t *testing.T) {
	var lookups, markets, posts atomic.Int32
	condition := "0x" + strings.Repeat("ab", 32)
	client := fixtureClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/markets-by-token/"):
			lookups.Add(1)
			if r.Header.Get("POLY_SIGNATURE") != "" {
				t.Error("metadata request carried account credentials")
			}
			writeResponse(t, w, map[string]string{"condition_id": condition})
		case strings.HasPrefix(r.URL.Path, "/clob-markets/"):
			count := markets.Add(1)
			tick := "0.1"
			if count > 1 {
				tick = "0.01"
			}
			writeResponse(t, w, map[string]any{"mts": tick, "nr": false, "t": []any{map[string]string{"t": "123"}}, "fd": nil})
		case r.URL.Path == "/order":
			posts.Add(1)
			writeResponse(t, w, map[string]any{"errorMsg": "", "makingAmount": "1", "takingAmount": "2", "orderID": "order", "success": true, "status": "live"})
		default:
			t.Errorf("unexpected request %s", r.URL)
			w.WriteHeader(500)
		}
	})
	p := CreateLimitOrderParams{AssetID: "123", Side: models.Buy, Price: dec("0.51"), Size: dec("12.345")}
	order, err := client.CreateLimitOrder(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}
	if markets.Load() != 2 || lookups.Load() != 1 || posts.Load() != 0 {
		t.Fatal("metadata refresh or create-only boundary failed")
	}
	if order.MakerAmount.String() != "6293400" || order.TakerAmount.String() != "12340000" {
		t.Fatalf("wrong amounts %v %v", order.MakerAmount, order.TakerAmount)
	}
	var workers sync.WaitGroup
	for range 12 {
		workers.Go(func() {
			if _, createErr := client.CreateLimitOrder(t.Context(), p); createErr != nil {
				t.Error(createErr)
			}
		})
	}
	workers.Wait()
	if markets.Load() != 2 || lookups.Load() != 1 {
		t.Fatal("cache missed for concurrent callers")
	}
	if _, err = client.PlaceLimitOrder(t.Context(), p); err != nil {
		t.Fatal(err)
	}
	if posts.Load() != 1 {
		t.Fatal("place did not submit exactly once")
	}
	p.Expiration = new(int64)
	*p.Expiration = 1720000179
	if _, err = client.CreateLimitOrder(t.Context(), p); err == nil {
		t.Fatal("unsafe expiry accepted")
	}
}
func TestHeartbeatMismatchAndNoRetry(t *testing.T) {
	var calls atomic.Int32
	client := fixtureClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		writeResponse(t, w, map[string]string{"error_msg": "Invalid Heartbeat ID", "heartbeat_id": "expected"})
	})
	_, err := client.SendOrderHeartbeat(t.Context(), "old")
	var mismatch *sdkerrors.OrderHeartbeatMismatchError
	var rejected *sdkerrors.RequestRejectedError
	if !errors.As(err, &mismatch) || mismatch.HeartbeatID != "expected" || !errors.As(err, &rejected) || rejected.Status != 400 || calls.Load() != 1 {
		t.Fatalf("heartbeat recovery incorrectly handled: %v", err)
	}
}
func tradeFixture(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/data-models-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name     string
			Response map[string]any
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, tc := range fixture.Cases {
		if tc.Name == "ClobTrade/full" {
			return tc.Response
		}
	}
	t.Fatal("missing trade fixture")
	return nil
}
func TestSettlementLifecycle(t *testing.T) {
	for _, outcome := range []string{"CONFIRMED", "FAILED", "MINED", "missing-hash"} {
		t.Run(outcome, func(subtest *testing.T) {
			trade := tradeFixture(subtest)
			trade["id"] = "trade"
			trade["status"] = outcome
			trade["transaction_hash"] = "hash"
			if outcome == "missing-hash" {
				trade["status"] = "CONFIRMED"
				trade["transaction_hash"] = nil
			}
			client := fixtureClient(subtest, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/data/trades" || r.URL.Query().Get("id") != "trade" {
					subtest.Error("incorrect settlement query")
				}
				writeResponse(subtest, w, map[string]any{"data": []any{trade}, "next_cursor": "LTE="})
			})
			timeout := 5 * time.Second
			if outcome == "MINED" {
				timeout = 10 * time.Millisecond
			}
			hashes, err := client.WaitForOrderFillSettlement(subtest.Context(), models.AcceptedOrder{TradeIDs: []string{"trade", "trade"}}, timeout)
			switch outcome {
			case "CONFIRMED":
				if err != nil || len(hashes) != 1 || hashes[0] != "hash" {
					subtest.Fatalf("settlement %v %v", hashes, err)
				}
			case "FAILED":
				if _, ok := errors.AsType[*sdkerrors.TransactionFailedError](err); !ok {
					subtest.Fatalf("failure %v", err)
				}
			case "MINED":
				if _, ok := errors.AsType[*sdkerrors.SettlementTimeoutError](err); !ok {
					subtest.Fatalf("MINED incorrectly treated as settled: %v", err)
				}
			case "missing-hash":
				if _, ok := errors.AsType[*sdkerrors.UnexpectedResponseError](err); !ok {
					subtest.Fatalf("missing confirmed hash accepted: %v", err)
				}
			}
		})
	}
}
func TestComboRetryKeepsIdenticalSignedOrder(t *testing.T) {
	var calls atomic.Int32
	var first []byte
	client := fixtureClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("POLY_API_KEY") == "" || r.Header.Get("POLY_BUILDER_API_KEY") == "" {
			t.Error("missing account or builder authentication")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if calls.Add(1) == 1 {
			first = body
			writeResponse(t, w, map[string]string{"rfq_id": "wrong", "status": "EXECUTING"})
			return
		}
		if string(first) != string(body) {
			t.Error("retry signed a new order")
		}
		writeResponse(t, w, map[string]string{"rfq_id": "rfq", "status": "EXECUTING"})
	})
	quote := models.ComboQuote{RFQID: "rfq", QuoteID: "quote", BuilderCode: signing.ZeroBytes32, Direction: models.Buy, PositionID: "123", BlendedPrice: dec("0.5"), MakerAmount: dec("1"), TakerAmount: dec("2"), TotalRequired: dec("1")}
	acceptance, err := client.AcceptComboQuote(t.Context(), quote)
	if err != nil || acceptance.Status != "executing" || acceptance.TakerOrderHash != nil || calls.Load() != 2 {
		t.Fatalf("retry outcome %+v %v", acceptance, err)
	}
}
func TestComboStatusNoBuilderHeadersAndTerminalFailure(t *testing.T) {
	client := fixtureClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("POLY_BUILDER_API_KEY") != "" {
			t.Error("status read included builder credentials")
		}
		writeResponse(t, w, map[string]any{"rfq_id": "rfq", "status": "FAILED", "error": map[string]string{"code": "future_code", "message": "failure"}})
	})
	value, err := client.WaitForComboFill(t.Context(), "rfq", time.Second, time.Millisecond)
	if err != nil || value.Status != "FAILED" || value.Error.Code != "future_code" {
		t.Fatalf("terminal failure %+v %v", value, err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err = waitContext(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("poll cancellation %v", err)
	}
}

func TestPythonSettlement(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/order-post-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Settlements []struct {
			States   []string
			Trades   []map[string]any
			TradeIDs []string `json:"trade_ids"`
			Existing []models.TransactionHash
			Expected struct {
				Value json.RawMessage
				Error string
			}
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Settlements) == 0 {
		t.Fatal("missing Python settlement cases")
	}
	for _, tc := range fixture.Settlements {
		t.Run(strings.Join(tc.States, "/"), func(subtest *testing.T) {
			client := fixtureClient(subtest, func(w http.ResponseWriter, r *http.Request) {
				items := make([]any, 0) // The protocol requires an empty array, not JSON null.
				for _, trade := range tc.Trades {
					if trade["id"] == r.URL.Query().Get("id") {
						items = append(items, trade)
					}
				}
				writeResponse(subtest, w, map[string]any{"data": items, "next_cursor": "LTE="})
			})
			timeout := 5 * time.Second
			if tc.Expected.Error == "TimeoutError" {
				timeout = 10 * time.Millisecond
			}
			value, callErr := client.WaitForOrderFillSettlement(subtest.Context(), models.AcceptedOrder{TradeIDs: tc.TradeIDs, TransactionsHashes: tc.Existing}, timeout)
			if tc.Expected.Error == "" {
				if callErr != nil {
					subtest.Fatal(callErr)
				}
				equalJSON(subtest, value, tc.Expected.Value)
			} else {
				kind := tc.Expected.Error
				if kind == "TimeoutError" {
					kind = "SettlementTimeoutError"
				}
				if callErr == nil || !strings.HasSuffix(fmt.Sprintf("%T", callErr), "."+kind) {
					subtest.Fatalf("error %v; Python %s", callErr, tc.Expected.Error)
				}
			}
		})
	}
}

func TestProtectedMarketOrderIncludesFeesAndUsesNegativeRiskExchange(t *testing.T) {
	var fees, books atomic.Int32
	builder := "0x" + strings.Repeat("01", 32)
	client := fixtureClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/markets-by-token/"):
			writeResponse(t, w, map[string]string{"condition_id": "0x" + strings.Repeat("ab", 32)})
		case strings.HasPrefix(r.URL.Path, "/clob-markets/"):
			writeResponse(t, w, map[string]any{"mts": "0.01", "nr": true, "t": []any{map[string]string{"t": "123"}}, "fd": map[string]string{"r": "0.025", "e": "2"}})
		case strings.HasPrefix(r.URL.Path, "/fees/builder-fees/"):
			fees.Add(1)
			writeResponse(t, w, map[string]string{"builder_maker_fee_rate_bps": "0", "builder_taker_fee_rate_bps": "10"})
		default:
			books.Add(1)
			t.Errorf("unexpected request %s", r.URL)
			w.WriteHeader(500)
		}
	})
	params := CreateMarketOrderParams{AssetID: "123", Side: models.Buy, Amount: dec("10.123456789"), MaxPrice: dec("0.5"), MaxSpend: dec("10"), BuilderCode: &builder}
	for range 2 {
		order, err := client.CreateMarketOrder(t.Context(), params)
		if err != nil {
			t.Fatal(err)
		}
		// Python fee adjustment is 9.958919457238889580480517864; order quantities floor to cents.
		if order.MakerAmount.String() != "9950000" || order.TakerAmount.String() != "19900000" {
			t.Fatalf("fee-adjusted order %v %v", order.MakerAmount, order.TakerAmount)
		}
		signature, signErr := client.signer.Order(*order, signing.Exchange("123", true), client.signerType)
		if signErr != nil || signature != order.Signature {
			t.Fatal("negative-risk exchange was not signed")
		}
	}
	if fees.Load() != 1 || books.Load() != 0 {
		t.Fatal("fee cache or protected-price path failed")
	}
}

func TestPostBatchBoundsAndFailureDoesNotRetry(t *testing.T) {
	var calls atomic.Int32
	client := fixtureClient(t, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); w.WriteHeader(500) })
	for _, count := range []int{0, 16} {
		if _, err := client.PostOrders(t.Context(), make([]models.SignedOrder, count)); err == nil {
			t.Fatal("invalid batch accepted")
		}
	}
	if calls.Load() != 0 {
		t.Fatal("invalid batch reached server")
	}
	order := models.SignedOrder{Salt: big.NewInt(1), MakerAmount: big.NewInt(1), TakerAmount: big.NewInt(2)}
	if _, err := client.PostOrder(t.Context(), order); err == nil {
		t.Fatal("HTTP failure suppressed")
	}
	if calls.Load() != 1 {
		t.Fatal("order automatically retried")
	}
}

func TestBuilderRevocationUsesOnlyBuilderCredentials(t *testing.T) {
	client := fixtureClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/auth/builder-api-key" || r.Header.Get("POLY_BUILDER_API_KEY") != "builder" || r.Header.Get("POLY_BUILDER_SIGNATURE") == "" || r.Header.Get("POLY_API_KEY") != "" {
			t.Error("incorrect builder revocation authentication")
		}
		writeResponse(t, w, "OK")
	})
	if err := client.RevokeBuilderAPIKey(t.Context()); err != nil {
		t.Fatal(err)
	}
}
