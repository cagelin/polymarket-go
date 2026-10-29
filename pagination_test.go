package polymarket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cagelin/polymarket-go/pagination"
)

func TestPythonPaginationParity(t *testing.T) {
	data, readErr := os.ReadFile("testdata/pagination-parity.json")
	if readErr != nil {
		t.Fatal(readErr)
	}
	var tests []struct {
		Name     string
		Pages    []Page[string]
		Initial  *string
		Expected []Page[string]
		Error    string
		Calls    []*string
	}
	if err := json.Unmarshal(data, &tests); err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		t.Run(test.Name, func(subtest *testing.T) {
			calls := make([]*string, 0)
			pager := pagination.New(func(_ context.Context, cursor *string) (Page[string], error) {
				calls = append(calls, cursor)
				if len(calls) > len(test.Pages) {
					return Page[string]{}, errors.New("too many page requests")
				}
				return test.Pages[len(calls)-1], nil
			}, test.Initial)
			var pages []Page[string]
			var resultErr error
			for page, err := range pager.Pages(subtest.Context()) {
				if err != nil {
					resultErr = err
					break
				}
				pages = append(pages, page)
			}
			if !reflect.DeepEqual(pages, test.Expected) {
				subtest.Errorf("pages = %+v, Python = %+v", pages, test.Expected)
			}
			if !reflect.DeepEqual(calls, test.Calls) {
				subtest.Errorf("cursors = %+v, Python = %+v", calls, test.Calls)
			}
			if got := errorClass(resultErr); got != test.Error {
				subtest.Errorf("error = %s, Python = %s", got, test.Error)
			}
		})
	}
}

func TestPaginatorEmptyAndEarlyStop(t *testing.T) {
	calls := 0
	pager := pagination.New(func(_ context.Context, _ *string) (Page[int], error) {
		calls++
		next := "next"
		return Page[int]{Items: []int{1, 2}, HasMore: true, NextCursor: &next}, nil
	}, nil)
	empty := pager.FromCursor(nil)
	page, err := empty.FirstPage(t.Context())
	if err != nil || len(page.Items) != 0 || page.HasMore {
		t.Fatalf("empty page = %+v, %v", page, err)
	}
	for yieldedPage, pageErr := range empty.Pages(t.Context()) {
		t.Errorf("empty paginator yielded %+v, %v", yieldedPage, pageErr)
	}
	if calls != 0 {
		t.Fatal("empty paginator fetched data")
	}
	for item, itemErr := range pager.Items(t.Context()) {
		if itemErr != nil || item != 1 {
			t.Errorf("first item = %d, %v", item, itemErr)
		}
		break
	}
	if calls != 1 {
		t.Errorf("requests after early break = %d, want 1", calls)
	}
}

func TestPaginatorCopiesInputsAndSupportsConcurrentReads(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Query().Get("slug") != "original" || r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("closed") != "false" {
			t.Errorf("query changed after construction: %v", r.URL.Query())
		}
		_, _ = w.Write([]byte(`{"markets":[],"next_cursor":null}`))
	}))
	defer server.Close()
	client, createErr := NewPublicClient(ClientConfig{GammaURL: server.URL})
	if createErr != nil {
		t.Fatal(createErr)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	closed, size := false, 2
	slugs := []string{"original"}
	pager, err := client.ListMarkets(ListMarketsParams{Slugs: slugs, Closed: &closed, PageSize: &size})
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 0 {
		t.Fatal("paginator is not lazy")
	}
	closed = true
	size = 3
	slugs[0] = "changed"
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			if _, fetchErr := pager.FirstPage(t.Context()); fetchErr != nil {
				t.Error(fetchErr)
			}
		})
	}
	group.Wait()
	if calls.Load() != 8 {
		t.Errorf("requests = %d, want 8", calls.Load())
	}
	var seen string
	custom := pagination.New(func(_ context.Context, cursor *string) (Page[int], error) { seen = *cursor; return Page[int]{}, nil }, nil)
	cursor := "original"
	resumed := custom.FromCursor(&cursor)
	cursor = "changed"
	if _, err = resumed.FirstPage(t.Context()); err != nil {
		t.Fatal(err)
	}
	if seen != "original" {
		t.Errorf("cursor was not copied: %s", seen)
	}
}

func TestHistoryRateLimitRetry(t *testing.T) {
	for _, test := range []struct {
		name      string
		status    int
		delay     string
		failures  int
		wantCalls int
		wantError string
	}{
		{name: "eventual_success", status: 429, delay: "0", failures: 2, wantCalls: 3},
		{name: "exhausted", status: 429, delay: "0", failures: 3, wantCalls: 3, wantError: "RateLimitError"},
		{name: "long_delay", status: 429, delay: "6", failures: 1, wantCalls: 1, wantError: "RateLimitError"},
		{name: "server_error", status: 503, delay: "0", failures: 1, wantCalls: 1, wantError: "RequestRejectedError"},
	} {
		t.Run(test.name, func(subtest *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if int(calls.Add(1)) <= test.failures {
					w.Header().Set("Retry-After", test.delay)
					w.WriteHeader(test.status)
					return
				}
				_, _ = w.Write([]byte(`{"data":[],"pagination":{"has_more":false,"next_cursor":null}}`))
			}))
			defer server.Close()
			client, createErr := NewPublicClient(ClientConfig{DataURL: server.URL})
			if createErr != nil {
				subtest.Fatal(createErr)
			}
			defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
			interval := HistoryHour
			pager, err := client.ListPriceHistory(PriceHistoryParams{AssetID: "123", Interval: &interval})
			if err != nil {
				subtest.Fatal(err)
			}
			_, err = pager.FirstPage(subtest.Context())
			if errorClass(err) != test.wantError || int(calls.Load()) != test.wantCalls {
				subtest.Errorf("error = %v, requests = %d; want %s, %d", err, calls.Load(), test.wantError, test.wantCalls)
			}
		})
	}
}

func TestHistoryRetryCancellation(t *testing.T) {
	entered := make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			close(entered)
		}
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(429)
	}))
	defer server.Close()
	client, createErr := NewPublicClient(ClientConfig{DataURL: server.URL})
	if createErr != nil {
		t.Fatal(createErr)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	interval := HistoryHour
	pager, err := client.ListPriceHistory(PriceHistoryParams{AssetID: "123", Interval: &interval})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, fetchErr := pager.FirstPage(ctx); done <- fetchErr }()
	<-entered
	cancel()
	select {
	case resultErr := <-done:
		if !errors.Is(resultErr, context.Canceled) {
			t.Errorf("error = %v, want cancellation", resultErr)
		}
	case <-time.After(time.Second):
		t.Fatal("retry wait ignored cancellation")
	}
	if calls.Load() != 1 {
		t.Errorf("retried after cancellation: %d", calls.Load())
	}
}
