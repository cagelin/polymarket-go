package polymarket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func researchClient(t *testing.T, handler http.HandlerFunc, configure func(*ClientConfig)) *PublicClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	config := ClientConfig{CLOBURL: server.URL + "/clob", GammaURL: server.URL + "/gamma", DataURL: server.URL + "/data"}
	if configure != nil {
		configure(&config)
	}
	client, err := NewPublicClient(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	return client
}

func writeResearchResponse(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	if _, err := w.Write([]byte(body)); err != nil {
		t.Error(err)
	}
}

func TestCLOBHistoryAndRawResponse(t *testing.T) {
	body := " {\n\"history\":[{\"t\":1700000001,\"p\":0.12345678901234567890123456789}]}\n"
	var observed []RawHTTPResponse
	client := researchClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clob/prices-history" || r.URL.Query().Get("market") != "123" || r.URL.Query().Get("startTs") != "1700000000" || r.URL.Query().Get("endTs") != "1700000060" || r.URL.Query().Get("fidelity") != "1" {
			t.Errorf("CLOB history request = %s", r.URL)
		}
		w.Header().Set("X-Test", "original")
		writeResearchResponse(t, w, body)
	}, func(config *ClientConfig) {
		config.RawHTTPHandler = func(ctx context.Context, response RawHTTPResponse) error {
			if ctx != t.Context() {
				t.Error("callback lost caller context")
			}
			observed = append(observed, response)
			return nil
		}
	})
	start, end, fidelity := int64(1700000000), int64(1700000060), 1
	points, err := client.GetCLOBPriceHistory(t.Context(), PriceHistoryQuery{Market: "123", StartTs: &start, EndTs: &end, Fidelity: &fidelity})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0].Price.String() != "0.12345678901234567890123456789" || points[0].Timestamp.String() != "1700000001" {
		t.Fatalf("CLOB history = %+v", points)
	}
	if len(observed) != 1 || string(observed[0].Body) != body || !observed[0].Complete || observed[0].StatusCode != 200 || observed[0].Service != "clob" || observed[0].ReceivedAt.IsZero() {
		t.Fatalf("raw response = %+v", observed)
	}
	observed[0].Body[0] = '!'
	if points[0].Price.String() != "0.12345678901234567890123456789" {
		t.Error("raw mutation changed decoded model")
	}
}

func TestCLOBHistoryValidationAndShapes(t *testing.T) {
	var calls int
	response := `{"history":[]}`
	client := researchClient(t, func(w http.ResponseWriter, _ *http.Request) { calls++; writeResearchResponse(t, w, response) }, nil)
	start, end, tooFar, zero, fidelity := int64(1), int64(1+15*86400), int64(2+15*86400), int64(0), 0
	interval, invalid := HistoryMax, PriceHistoryInterval("bad")
	for _, params := range []PriceHistoryQuery{
		{}, {Market: "1"}, {Market: "1", StartTs: &start}, {Market: "1", EndTs: &end},
		{Market: "1", Interval: &interval, StartTs: &start}, {Market: "1", Interval: &invalid},
		{Market: "1", StartTs: &start, EndTs: &tooFar}, {Market: "1", StartTs: &end, EndTs: &start},
		{Market: "1", StartTs: &zero, EndTs: &end}, {Market: "1", Interval: &interval, Fidelity: &fidelity},
	} {
		_, err := client.GetCLOBPriceHistory(t.Context(), params)
		if _, inputFailure := errors.AsType[*UserInputError](err); !inputFailure {
			t.Errorf("GetCLOBPriceHistory(%+v) error = %v, want UserInputError", params, err)
		}
	}
	if calls != 0 {
		t.Errorf("invalid history queries made %d requests", calls)
	}
	if _, err := client.GetCLOBPriceHistory(t.Context(), PriceHistoryQuery{Market: "1", StartTs: &start, EndTs: &end}); err != nil {
		t.Fatalf("15-day boundary: %v", err)
	}
	for _, malformed := range []string{`null`, `{}`, `{"history":null}`, `{"history":[{"t":1.5,"p":0.2}]}`, `{"history":[{"t":1}]}`, `{"history":[{"t":1,"p":"NaN"}]}`} {
		response = malformed
		_, err := client.GetCLOBPriceHistory(t.Context(), PriceHistoryQuery{Market: "1", Interval: &interval})
		if _, shapeFailure := errors.AsType[*UnexpectedResponseError](err); !shapeFailure {
			t.Errorf("history %s error = %v, want UnexpectedResponseError", malformed, err)
		}
	}
}

func TestCountEventsAndExactFilters(t *testing.T) {
	var requests []string
	client := researchClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.String())
		if r.URL.Path == "/gamma/events/pagination" {
			if r.URL.Query().Get("closed") != "true" || r.URL.Query().Get("tag_id") != "7" || r.URL.Query().Get("title_search") != "temperature" || r.URL.Query().Get("limit") != "1" || r.URL.Query().Get("offset") != "0" {
				t.Errorf("CountEvents query = %s", r.URL)
			}
			writeResearchResponse(t, w, `{"totalResults":9007199254740993}`)
		} else {
			if r.URL.Query().Get("filter_amount") != "0.12345678901234567890123456789" {
				t.Errorf("filter amount lost precision: %s", r.URL)
			}
			writeResearchResponse(t, w, `{"data":[],"pagination":{"has_more":false,"next_cursor":null}}`)
		}
	}, nil)
	title := "temperature"
	total, err := client.CountEvents(t.Context(), ListEventsParams{Closed: true, TagIDs: []int64{7}, TitleSearch: &title})
	if err != nil || total != 9007199254740993 {
		t.Fatalf("CountEvents = %d, %v", total, err)
	}
	amount := json.Number("0.12345678901234567890123456789")
	pager, err := client.ListTrades(ListTradesParams{FilterAmount: &amount})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pager.FirstPage(t.Context()); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []json.Number{"-1", "NaN", "Inf", "null", "1/2", ""} {
		if _, err = client.ListTrades(ListTradesParams{FilterAmount: &invalid}); err == nil {
			t.Errorf("ListTrades accepted %q", invalid)
		}
	}
	if len(requests) != 2 {
		t.Errorf("requests = %v, want count plus one trade page", requests)
	}
}

func TestRawResponseErrorsAndLimits(t *testing.T) {
	for _, scenario := range []struct {
		name, body string
		status     int
		limit      int64
		complete   bool
	}{
		{"http-error", " {\"error\":\"missing\"}\n", 404, 0, true},
		{"parse-error", "not json", 200, 0, true},
		{"limit", "123456", 200, 5, false},
	} {
		t.Run(scenario.name, func(subtest *testing.T) {
			var observed []RawHTTPResponse
			client := researchClient(subtest, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(scenario.status)
				writeResearchResponse(subtest, w, scenario.body)
			}, func(config *ClientConfig) {
				config.MaxResponseBodyBytes = scenario.limit
				config.RawHTTPHandler = func(_ context.Context, response RawHTTPResponse) error {
					observed = append(observed, response)
					return nil
				}
			})
			interval := HistoryMax
			_, err := client.GetCLOBPriceHistory(subtest.Context(), PriceHistoryQuery{Market: "1", Interval: &interval})
			if err == nil || len(observed) != 1 {
				subtest.Fatalf("call error=%v raw=%+v", err, observed)
			}
			want := scenario.body
			if scenario.limit > 0 {
				want = want[:scenario.limit]
			}
			if string(observed[0].Body) != want || observed[0].Complete != scenario.complete {
				subtest.Errorf("raw response = %+v", observed[0])
			}
		})
	}
}

func TestRawResponseRetryAndHandlerFailure(t *testing.T) {
	var statuses []int
	calls := 0
	failure := errors.New("record failed")
	failHandler := false
	client := researchClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			writeResearchResponse(t, w, `{"error":"rate"}`)
			return
		}
		writeResearchResponse(t, w, `{"data":[],"pagination":{"has_more":false,"next_cursor":null}}`)
	}, func(config *ClientConfig) {
		config.RawHTTPHandler = func(_ context.Context, response RawHTTPResponse) error {
			statuses = append(statuses, response.StatusCode)
			if failHandler {
				return failure
			}
			return nil
		}
	})
	pager, err := client.ListTrades(ListTradesParams{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pager.FirstPage(t.Context()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(statuses, []int{429, 200}) {
		t.Errorf("raw attempts = %v, want [429 200]", statuses)
	}
	failHandler = true
	if _, err = pager.FirstPage(t.Context()); !errors.Is(err, failure) {
		t.Errorf("handler failure = %v, want preserved cause", err)
	}
	if calls != 3 {
		t.Errorf("handler failure retried: %d calls", calls)
	}
}

func TestRawResponseConcurrentCalls(t *testing.T) {
	var mutex sync.Mutex
	records := 0
	client := researchClient(t, func(w http.ResponseWriter, _ *http.Request) { writeResearchResponse(t, w, `{"totalResults":1}`) }, func(config *ClientConfig) {
		config.RawHTTPHandler = func(_ context.Context, response RawHTTPResponse) error {
			mutex.Lock()
			defer mutex.Unlock()
			records++
			if !strings.Contains(response.URL, "/events/pagination") {
				t.Errorf("unexpected raw URL: %s", response.URL)
			}
			return nil
		}
	})
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			if _, err := client.CountEvents(t.Context(), ListEventsParams{}); err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()
	if records != 8 {
		t.Errorf("concurrent records = %d, want 8", records)
	}
}
