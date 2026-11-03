package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cagelin/polymarket-go"
)

const eventFixture = `{"id":"7","title":"Daily temperature","slug":"daily-temperature","markets":[{"id":"8","question":"Above 20?","outcomes":["Yes","No"],"clobTokenIds":["123","124"]}]}`

func TestEventLookupAndTags(t *testing.T) {
	for _, selector := range [][]string{{"-id", "7"}, {"-slug", "daily-temperature"}, {"-url", "https://polymarket.com/event/daily-temperature"}} {
		t.Run(selector[0], func(subtest *testing.T) {
			var reads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				reads.Add(1)
				var body string
				switch request.URL.Path {
				case "/events/7", "/events/slug/daily-temperature":
					body = eventFixture
				case "/events/7/tags":
					body = `[{"id":"9","label":"Weather"},{"id":"10"}]`
				default:
					subtest.Errorf("event lookup path = %s, want event or tags", request.URL)
					writer.WriteHeader(500)
					return
				}
				if _, err := fmt.Fprint(writer, body); err != nil {
					subtest.Error(err)
				}
			}))
			subtest.Cleanup(server.Close)
			var output bytes.Buffer
			err := runWith(polymarket.ClientConfig{GammaURL: server.URL}, selector, &output)
			if err != nil {
				subtest.Fatalf("runWith(%v) = %v, want nil", selector, err)
			}
			for _, expected := range []string{"event=7 slug=daily-temperature title=Daily temperature markets=1", "market=8 question=Above 20?", "YES asset-id=123", "NO  asset-id=124", "event tags=2", "tag=9 label=Weather", "tag=10 label="} {
				if !strings.Contains(output.String(), expected) {
					subtest.Errorf("runWith(%v) output = %q, want %q", selector, output.String(), expected)
				}
			}
			if reads.Load() != 2 {
				subtest.Errorf("runWith(%v) requests = %d, want 2", selector, reads.Load())
			}
		})
	}
}

func TestEventListFiltersAndBounds(t *testing.T) {
	var reads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		reads.Add(1)
		q := request.URL.Query()
		if request.URL.Path != "/events/keyset" || q.Get("limit") != "2" || q.Get("closed") != "true" || q.Get("tag_slug") != "weather" || q.Get("series_id") != "4" {
			t.Errorf("runWith(list) query = %s, want selected list filters", request.URL)
		}
		if _, err := fmt.Fprintf(writer, `{"events":[%s,{"id":"11","markets":[{"id":"12","outcomes":["Yes","No"]}]}],"next_cursor":"next-page"}`, eventFixture); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	var output bytes.Buffer
	err := runWith(polymarket.ClientConfig{GammaURL: server.URL}, []string{"-limit", "2", "-closed", "-tag", "weather", "-series", "4"}, &output)
	if err != nil {
		t.Fatalf("runWith(list) = %v, want nil", err)
	}
	if reads.Load() != 1 || !strings.Contains(output.String(), "events on this page=2 has_more=true limit_reached=false") {
		t.Errorf("runWith(list) reads=%d output=%q, want one bounded page with continuation", reads.Load(), output.String())
	}
}

func TestEventInvalidArgumentsDoNotRead(t *testing.T) {
	var reads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { reads.Add(1); writer.WriteHeader(500) }))
	t.Cleanup(server.Close)
	for _, args := range [][]string{{"-id", "7", "-slug", "daily"}, {"-limit", "0"}, {"-series", "-1"}, {"extra"}, {"-unknown"}} {
		var output bytes.Buffer
		if err := runWith(polymarket.ClientConfig{GammaURL: server.URL}, args, &output); err == nil {
			t.Errorf("runWith(%v) = nil, want argument error", args)
		}
	}
	if reads.Load() != 0 {
		t.Errorf("invalid event arguments reads=%d, want 0", reads.Load())
	}
}

type failingOutput struct {
	remaining int
	cause     error
}

func (w *failingOutput) Write(data []byte) (int, error) {
	if w.remaining == 0 {
		return 0, w.cause
	}
	w.remaining--
	return len(data), nil
}

func TestEventOutputFailureStopsReads(t *testing.T) {
	for _, args := range [][]string{{"-id", "7"}, {}} {
		for failAfter := range 4 {
			var reads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				reads.Add(1)
				body := eventFixture
				if request.URL.Path == "/events/7/tags" {
					body = `[{"id":"9","label":"Weather"},{"id":"10"}]`
				}
				if request.URL.Path == "/events/keyset" {
					body = `{"events":[` + eventFixture + `,` + eventFixture + `,` + eventFixture + `],"next_cursor":null}`
				}
				if _, err := fmt.Fprint(writer, body); err != nil {
					t.Error(err)
				}
			}))
			failure := errors.New("output closed")
			err := runWith(polymarket.ClientConfig{GammaURL: server.URL}, args, &failingOutput{remaining: failAfter, cause: failure})
			server.Close()
			if !errors.Is(err, failure) {
				t.Errorf("runWith(%v, output fails after %d writes)=%v, want output cause", args, failAfter, err)
			}
			if failAfter == 0 && reads.Load() != 1 {
				t.Errorf("runWith(output failure) reads=%d, want 1 before tags or continuation", reads.Load())
			}
		}
	}
}

func TestHelpDoesNotRead(t *testing.T) {
	var output bytes.Buffer
	if err := runWith(polymarket.ClientConfig{GammaURL: ":invalid", DataURL: ":invalid"}, []string{"-h"}, &output); err != nil || output.Len() == 0 {
		t.Errorf("runWith(-h)=%v output=%q, want help without constructing a client", err, output.String())
	}
}
