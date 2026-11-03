package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/cagelin/polymarket-go"
)

func TestHistorySelectorsAndExactOutput(t *testing.T) {
	for _, scenario := range []struct {
		name  string
		args  []string
		query url.Values
	}{
		{name: "default", args: []string{"123"}, query: url.Values{"interval": {"1d"}, "limit": {"20"}}},
		{name: "interval", args: []string{"-interval", "1h", "123"}, query: url.Values{"interval": {"1h"}, "limit": {"20"}}},
		{name: "window", args: []string{"-start", "2023-11-14T22:13:20Z", "-end", "2023-11-14T22:14:20Z", "123"}, query: url.Values{"start": {"1700000000"}, "end": {"1700000060"}, "limit": {"20"}}},
		{name: "as-of", args: []string{"-as-of", "2023-11-14T22:13:20Z", "123"}, query: url.Values{"as_of": {"1700000000"}}},
	} {
		t.Run(scenario.name, func(subtest *testing.T) {
			var reads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				reads.Add(1)
				for key, values := range scenario.query {
					if request.URL.Query().Get(key) != values[0] {
						subtest.Errorf("runWith(%v) query %s=%q, want %q", scenario.args, key, request.URL.Query().Get(key), values[0])
					}
				}
				if _, err := fmt.Fprint(writer, `{"data":[{"timestamp":"1700000000.1234575","price":"0.1234567890123456789","resolution_seconds":60}],"pagination":{"has_more":false,"next_cursor":null}}`); err != nil {
					subtest.Error(err)
				}
			}))
			subtest.Cleanup(server.Close)
			var output bytes.Buffer
			err := runWith(polymarket.ClientConfig{DataURL: server.URL}, scenario.args, &output)
			if err != nil {
				subtest.Fatalf("runWith(%v) = %v, want nil", scenario.args, err)
			}
			var page struct {
				Items []struct{ Timestamp, Price string }
			}
			if err = json.Unmarshal(output.Bytes(), &page); err != nil {
				subtest.Fatalf("runWith(%v) JSON=%q error=%v", scenario.args, output.String(), err)
			}
			if len(page.Items) != 1 || page.Items[0].Timestamp != "2023-11-14T22:13:20.123458Z" || page.Items[0].Price != "0.1234567890123456789" || reads.Load() != 1 {
				subtest.Errorf("runWith(%v) page=%+v reads=%d, want one exact observation", scenario.args, page, reads.Load())
			}
		})
	}
}

func TestHistoryInvalidArgumentsDoNotRead(t *testing.T) {
	var reads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { reads.Add(1); writer.WriteHeader(500) }))
	t.Cleanup(server.Close)
	for _, args := range [][]string{{}, {"-end", "2023-11-14T22:13:20Z", "123"}, {"-start", "invalid", "123"}, {"-interval", "bad", "123"}, {"-interval", "1d", "-as-of", "2023-11-14T22:13:20Z", "123"}, {"-unknown"}} {
		var output bytes.Buffer
		if err := runWith(polymarket.ClientConfig{DataURL: server.URL}, args, &output); err == nil {
			t.Errorf("runWith(%v) = nil, want argument error", args)
		}
	}
	if reads.Load() != 0 {
		t.Errorf("invalid history arguments reads=%d, want 0", reads.Load())
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

func TestHistoryOutputFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(writer, `{"data":[],"pagination":{"has_more":false,"next_cursor":null}}`); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	failure := errors.New("output closed")
	err := runWith(polymarket.ClientConfig{DataURL: server.URL}, []string{"123"}, &failingOutput{cause: failure})
	if !errors.Is(err, failure) {
		t.Errorf("runWith(history, output failure)=%v, want preserved output error", err)
	}
}

func TestHelpDoesNotRead(t *testing.T) {
	var output bytes.Buffer
	if err := runWith(polymarket.ClientConfig{GammaURL: ":invalid", DataURL: ":invalid"}, []string{"-h"}, &output); err != nil || output.Len() == 0 {
		t.Errorf("runWith(-h)=%v output=%q, want help without constructing a client", err, output.String())
	}
}
