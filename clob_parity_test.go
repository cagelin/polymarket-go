package polymarket

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/ericlagergren/decimal"
)

type parityInput struct {
	AssetID   ClobAssetID     `json:"asset_id"`
	AssetIDs  []ClobAssetID   `json:"asset_ids"`
	Side      OrderSide       `json:"side"`
	Requests  []parityInput   `json:"requests"`
	Amount    *string         `json:"amount"`
	Shares    *string         `json:"shares"`
	OrderType MarketOrderType `json:"order_type"`
}

func (p parityInput) asset() AssetQuery { return AssetQuery{AssetID: p.AssetID} }
func (p parityInput) assets() AssetsQuery {
	return AssetsQuery{AssetIDs: p.AssetIDs}
}
func (p parityInput) price() PriceRequest { return PriceRequest{AssetQuery: p.asset(), Side: p.Side} }

type parityCase struct {
	Name      string          `json:"name"`
	Operation string          `json:"operation"`
	Input     parityInput     `json:"input"`
	Response  json.RawMessage `json:"response"`
	Expected  struct {
		Value json.RawMessage `json:"value"`
		Error string          `json:"error"`
	} `json:"expected"`
	Request *struct {
		Method string            `json:"method"`
		Path   string            `json:"path"`
		Query  map[string]string `json:"query"`
		Body   json.RawMessage   `json:"body"`
	} `json:"request"`
}

func TestPythonCLOBParity(t *testing.T) {
	data, readErr := os.ReadFile("testdata/clob-parity.json")
	if readErr != nil {
		t.Fatal(readErr)
	}
	var fixture struct {
		Version string       `json:"upstream_version"`
		Commit  string       `json:"upstream_commit"`
		Cases   []parityCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != UpstreamVersion || fixture.Commit != UpstreamCommit {
		t.Fatal("fixture source differs from SDK metadata")
	}
	if len(fixture.Cases) == 0 {
		t.Fatal("no parity cases")
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(subtest *testing.T) {
			subtest.Parallel()
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if test.Request == nil {
					subtest.Error("unexpected HTTP request after invalid input")
					w.WriteHeader(500)
					return
				}
				if r.Method != test.Request.Method || r.URL.Path != test.Request.Path {
					subtest.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, test.Request.Method, test.Request.Path)
				}
				query := r.URL.Query()
				if len(query) != len(test.Request.Query) {
					subtest.Errorf("query = %v, want %v", query, test.Request.Query)
				}
				for k, v := range test.Request.Query {
					if query.Get(k) != v {
						subtest.Errorf("query %s = %q, want %q", k, query.Get(k), v)
					}
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					subtest.Error(err)
				}
				if test.Request.Method == http.MethodPost {
					if r.Header.Get("Content-Type") != "application/json" {
						subtest.Error("missing JSON content type")
					}
					compareJSON(subtest, body, test.Request.Body)
				} else if len(body) != 0 {
					subtest.Error("GET request contains a body")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(test.Response)
			}))
			defer server.Close()
			client, err := NewPublicClient(ClientConfig{CLOBURL: server.URL})
			if err != nil {
				subtest.Fatal(err)
			}
			defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
			value, err := callParity(subtest, client, test)
			if got := errorClass(err); got != test.Expected.Error {
				subtest.Fatalf("error = %s (%v), Python = %s", got, err, test.Expected.Error)
			}
			if test.Request == nil && calls.Load() != 0 {
				subtest.Fatal("invalid input performed network I/O")
			}
			if test.Request != nil && calls.Load() != 1 {
				subtest.Fatalf("requests = %d, want 1", calls.Load())
			}
			if err == nil {
				encoded, encodeErr := json.Marshal(value)
				if encodeErr != nil {
					subtest.Fatal(encodeErr)
				}
				compareJSON(subtest, encoded, test.Expected.Value)
			}
		})
	}
}

func errorClass(err error) string {
	if _, ok := errors.AsType[*PaginationLimitError](err); ok {
		return "PaginationLimitError"
	}
	if err == nil {
		return ""
	}
	var input *UserInputError
	var response *UnexpectedResponseError
	var liquidity *InsufficientLiquidityError
	var rate *RateLimitError
	var rejected *RequestRejectedError
	var transport *TransportError
	switch {
	case errors.As(err, &input):
		return "UserInputError"
	case errors.As(err, &response):
		return "UnexpectedResponseError"
	case errors.As(err, &liquidity):
		return "InsufficientLiquidityError"
	case errors.As(err, &rate):
		return "RateLimitError"
	case errors.As(err, &rejected):
		return "RequestRejectedError"
	case errors.As(err, &transport):
		return "TransportError"
	default:
		return "unknown: " + err.Error()
	}
}

func compareJSON(t *testing.T, got, want []byte) {
	t.Helper()
	var actual, expected any
	if err := json.Unmarshal(got, &actual); err != nil {
		t.Error(err)
		return
	}
	if err := json.Unmarshal(want, &expected); err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("JSON = %s\nPython = %s", got, want)
	}
}

func parityDecimal(t *testing.T, value *string) *decimal.Big {
	t.Helper()
	if value == nil {
		return nil
	}
	d, ok := new(decimal.Big).SetString(*value)
	if !ok {
		t.Fatalf("invalid decimal %q", *value)
	}
	return d
}

func callParity(t *testing.T, c *PublicClient, test parityCase) (any, error) {
	t.Helper()
	p := test.Input
	ctx := t.Context()
	switch test.Operation {
	case "midpoint":
		return c.GetMidpoint(ctx, p.asset())
	case "midpoints":
		return c.GetMidpoints(ctx, p.assets())
	case "price":
		return c.GetPrice(ctx, p.price())
	case "prices":
		requests := make([]PriceRequest, len(p.Requests))
		for i, r := range p.Requests {
			requests[i] = r.price()
		}
		return c.GetPrices(ctx, requests)
	case "order_book":
		return c.GetOrderBook(ctx, p.asset())
	case "order_books":
		return c.GetOrderBooks(ctx, p.assets())
	case "spread":
		return c.GetSpread(ctx, p.asset())
	case "spreads":
		return c.GetSpreads(ctx, p.assets())
	case "last_trade_price":
		return c.GetLastTradePrice(ctx, p.asset())
	case "last_trade_prices":
		return c.GetLastTradePrices(ctx, p.assets())
	case "estimate_market_price":
		return c.EstimateMarketPrice(ctx, EstimateMarketPriceParams{
			AssetQuery: p.asset(), Side: p.Side, Amount: parityDecimal(t, p.Amount),
			Shares: parityDecimal(t, p.Shares), OrderType: p.OrderType,
		})
	default:
		t.Fatalf("unknown operation %s", test.Operation)
		return nil, nil
	}
}

func TestCancellation(t *testing.T) {
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
	}))
	defer server.Close()
	client, createErr := NewPublicClient(ClientConfig{CLOBURL: server.URL})
	if createErr != nil {
		t.Fatal(createErr)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := client.GetMidpoint(ctx, Asset("123")); done <- err }()
	<-entered
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
