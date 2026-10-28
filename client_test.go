package polymarket

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cagelam/polymarket-go/internal/clob"
	"github.com/cagelam/polymarket-go/internal/codec"
)

func TestPythonHTTPParity(t *testing.T) {
	data, readErr := os.ReadFile("testdata/http-parity.json")
	if readErr != nil {
		t.Fatal(readErr)
	}
	type state struct {
		Remaining *float64
		Reset     *float64
		Tier      string
		Warning   bool
	}
	var cases []struct {
		Name     string
		Status   int
		Headers  map[string]string
		Body     json.RawMessage
		Expected struct {
			Error       string
			Code        string
			Restriction TradingRestriction
			RetryAfter  *float64 `json:"retry_after"`
			RateLimit   *state   `json:"rate_limit"`
		}
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				for key, value := range test.Headers {
					w.Header().Set(key, value)
				}
				w.WriteHeader(test.Status)
				_, _ = w.Write(test.Body)
			}))
			defer server.Close()
			var update *state
			client, err := NewPublicClient(ClientConfig{CLOBURL: server.URL, OnRateLimitUpdate: func(value RateLimitUpdate) error {
				update = &state{Remaining: value.Remaining, Reset: value.Reset, Tier: value.Tier, Warning: value.Warning}
				return errors.New("listener error must not replace response")
			}})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
			_, err = client.GetMidpoint(t.Context(), Asset("123"))
			if got := errorClass(err); got != test.Expected.Error {
				t.Fatalf("error = %s (%v), Python = %s", got, err, test.Expected.Error)
			}
			var after *float64
			var code string
			var restriction TradingRestriction
			if rate, ok := errors.AsType[*RateLimitError](err); ok {
				after = rate.RetryAfter
				if rate.RateLimit == nil && update != nil {
					t.Error("rate error lost rate-limit state")
				}
			} else if rejected, isRejected := errors.AsType[*RequestRejectedError](err); isRejected {
				after, code, restriction = rejected.RetryAfter, rejected.Code, rejected.Restriction
				if rejected.Status != test.Status {
					t.Errorf("HTTP status = %d, want %d", rejected.Status, test.Status)
				}
			}
			if !reflect.DeepEqual(after, test.Expected.RetryAfter) || code != test.Expected.Code || restriction != test.Expected.Restriction {
				t.Errorf("error metadata = (%v,%q,%q), Python = (%v,%q,%q)", after, code, restriction, test.Expected.RetryAfter, test.Expected.Code, test.Expected.Restriction)
			}
			if !reflect.DeepEqual(update, test.Expected.RateLimit) {
				t.Errorf("rate-limit update = %+v, Python = %+v", update, test.Expected.RateLimit)
			}
			if calls.Load() != 1 {
				t.Errorf("request count = %d, want 1 (no implicit retry)", calls.Load())
			}
		})
	}
}

func TestSharedClientOwnershipAndConcurrentUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, `{"mid":"0.5"}`) }))
	defer server.Close()
	shared := server.Client()
	shared.Timeout = 3 * time.Second
	client, createErr := NewPublicClient(ClientConfig{CLOBURL: server.URL, HTTPClient: shared})
	if createErr != nil {
		t.Fatal(createErr)
	}
	var workers sync.WaitGroup
	for range 20 {
		workers.Go(func() {
			if _, err := client.GetMidpoint(t.Context(), Asset("123")); err != nil {
				t.Error(err)
			}
		})
	}
	workers.Wait()
	for range 20 {
		workers.Go(func() {
			if err := client.Close(); err != nil {
				t.Error(err)
			}
		})
	}
	workers.Wait()
	if _, err := client.GetMidpoint(t.Context(), Asset("123")); errorClass(err) != "UserInputError" {
		t.Errorf("closed client error = %v", err)
	}
	res, err := shared.Get(server.URL)
	if err != nil {
		t.Fatalf("borrowed client no longer usable: %v", err)
	}
	if closeErr := res.Body.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if shared.Timeout != 3*time.Second {
		t.Error("SDK changed borrowed client timeout")
	}
}

func TestMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, `<html>upstream error</html>`) }))
	defer server.Close()
	client, createErr := NewPublicClient(ClientConfig{CLOBURL: server.URL})
	if createErr != nil {
		t.Fatal(createErr)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	if _, err := client.GetMidpoint(t.Context(), Asset("123")); errorClass(err) != "UnexpectedResponseError" {
		t.Errorf("malformed JSON error = %v", err)
	}
}

func TestClientConfig(t *testing.T) {
	for _, endpoint := range []string{"relative", "ftp://example.com", "https://user:secret@example.com", "https://example.com?x=1", "https://example.com/#fragment"} {
		if _, err := NewPublicClient(ClientConfig{CLOBURL: endpoint}); errorClass(err) != "UserInputError" {
			t.Errorf("NewPublicClient(%q) error = %v", endpoint, err)
		}
	}
}

func benchmarkBook(b *testing.B) *OrderBook {
	b.Helper()
	data, readErr := os.ReadFile("testdata/clob-parity.json")
	if readErr != nil {
		b.Fatal(readErr)
	}
	var fixture struct {
		Cases []parityCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		b.Fatal(err)
	}
	for _, test := range fixture.Cases {
		if test.Name == "order_book/valid" {
			book, err := codec.DecodeResponse[OrderBook](test.Response)
			if err != nil {
				b.Fatal(err)
			}
			return &book
		}
	}
	b.Fatal("benchmark book missing")
	return nil
}

func BenchmarkEstimateMarketPrice(b *testing.B) {
	book := benchmarkBook(b)
	amount := book.MinOrderSize
	b.ReportAllocs()
	for b.Loop() {
		if _, err := clob.EstimatePrice(book, Buy, amount, FOK); err != nil {
			b.Fatal(err)
		}
	}
}
