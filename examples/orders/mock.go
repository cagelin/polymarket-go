package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cagelin/polymarket-go"
	"github.com/ericlagergren/decimal"
)

// mockExchange implements http.RoundTripper with synthetic responses and no network transport.
// It illustrates SDK control flow, not exchange signature validation or matching rules.
type mockExchange struct {
	mu                sync.Mutex
	limitOpen         bool
	submissions       int
	rejectMarket      bool
	pendingSettlement bool
}

func newMockClient(ctx context.Context, exchange *mockExchange) (*polymarket.SecureClient, error) {
	clientConfig := polymarket.ClientConfig{
		CLOBURL:    "https://example.invalid",
		HTTPClient: &http.Client{Transport: exchange},
	}
	return polymarket.NewSecureClient(ctx, polymarket.SecureClientConfig{
		ClientConfig: clientConfig,
		// Scalar one is a public test key, not a usable account secret.
		PrivateKey: fmt.Sprintf("%064x", 1),
		Wallet:     "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf",
		Credentials: &polymarket.APIKeyCreds{
			Key: "example-key", Secret: base64.StdEncoding.EncodeToString([]byte("example")), Passphrase: "example-passphrase",
		},
	})
}

func (m *mockExchange) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Body != nil {
		defer func() { _ = request.Body.Close() }() // RoundTripper owns the request body.
	}
	if contextErr := request.Context().Err(); contextErr != nil {
		return nil, contextErr
	}
	if request.URL.Host != "example.invalid" {
		return nil, fmt.Errorf("mock exchange refuses external host %q", request.URL.Host)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	condition := polymarket.ConditionID("0x" + strings.Repeat("ab", 32))
	var payload any
	switch request.Method + " " + request.URL.Path {
	case "GET /auth/api-keys":
		payload = map[string]any{"apiKeys": []string{"example-key"}}
	case "GET /markets-by-token/123":
		payload = map[string]any{"condition_id": condition}
	case "GET /clob-markets/" + string(condition):
		payload = map[string]any{"mts": "0.01", "nr": false, "t": []map[string]string{{"t": "123"}}}
	case "POST /order":
		var submission struct {
			OrderType string `json:"orderType"`
			Order     struct {
				TokenID   string `json:"tokenId"`
				Signature string `json:"signature"`
			} `json:"order"`
		}
		if decodeErr := json.NewDecoder(request.Body).Decode(&submission); decodeErr != nil {
			return nil, fmt.Errorf("decode mock submission: %w", decodeErr)
		}
		if submission.Order.TokenID != "123" || submission.Order.Signature == "" {
			return nil, fmt.Errorf("mock submission requires the synthetic token and a signature")
		}
		m.submissions++
		switch submission.OrderType {
		case "GTC":
			m.limitOpen = true
			payload = map[string]any{"errorMsg": "", "makingAmount": "4", "takingAmount": "10", "orderID": "mock-limit", "success": true, "status": "live"}
		case "FOK":
			if m.rejectMarket {
				payload = map[string]any{"errorMsg": "invalid nonce", "makingAmount": "", "takingAmount": "", "orderID": "", "success": false, "status": ""}
			} else {
				payload = map[string]any{"errorMsg": "", "makingAmount": "5", "takingAmount": "10", "orderID": "mock-market", "success": true, "status": "matched", "tradeIDs": []string{"mock-fill"}}
			}
		default:
			return nil, fmt.Errorf("unsupported mock order type %q", submission.OrderType)
		}
	case "GET /data/orders":
		orders := make([]polymarket.OpenOrder, 0, 1)
		if m.limitOpen {
			orders = append(orders, polymarket.OpenOrder{
				ID: "mock-limit", ConditionID: condition, AssetID: "123", Owner: "example-key", MakerAddress: "mock-wallet",
				Side: polymarket.Buy, Price: decimal.New(4, 1), OriginalSize: decimal.New(10, 0), SizeMatched: decimal.New(0, 0),
				Outcome: "Yes", OrderType: "GTC", Status: "LIVE", AssociateTrades: []string{}, CreatedAt: time.Unix(1700000000, 0).UTC(),
			})
		}
		payload = map[string]any{"data": orders, "next_cursor": "LTE="}
	case "DELETE /order":
		var cancellation struct {
			OrderID string `json:"orderID"`
		}
		if decodeErr := json.NewDecoder(request.Body).Decode(&cancellation); decodeErr != nil {
			return nil, fmt.Errorf("decode mock cancellation: %w", decodeErr)
		}
		if cancellation.OrderID != "mock-limit" || !m.limitOpen {
			return nil, fmt.Errorf("mock order is not open")
		}
		m.limitOpen = false
		payload = map[string]any{"canceled": []string{cancellation.OrderID}, "not_canceled": map[string]string{}}
	case "GET /data/trades":
		if request.URL.Query().Get("id") != "mock-fill" {
			return nil, fmt.Errorf("mock settlement requires the placement trade ID")
		}
		status := "CONFIRMED"
		if m.pendingSettlement {
			status = "MATCHED"
		}
		hash := "0x" + strings.Repeat("cd", 32)
		payload = map[string]any{"data": []polymarket.ClobTrade{{
			ID: "mock-fill", ConditionID: condition, AssetID: "123", Owner: "example-key", MakerAddress: "mock-wallet",
			TakerOrderID: "mock-market", Side: polymarket.Buy, TraderSide: "TAKER", Price: decimal.New(5, 1), Size: decimal.New(10, 0),
			Outcome: "Yes", Status: status, FeeRateBps: decimal.New(0, 0), TransactionHash: &hash, MakerOrders: []polymarket.MakerOrder{},
			MatchedAt: time.Unix(1700000000, 0).UTC(), UpdatedAt: time.Unix(1700000001, 0).UTC(),
		}}, "next_cursor": "LTE="}
	default:
		return nil, fmt.Errorf("unsupported mock request %s %s", request.Method, request.URL.Path)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode mock response: %w", err)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(encoded)),
		Request:    request,
	}, nil
}
