package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cagelin/polymarket-go"
)

func TestPaginationStopsAndResumes(t *testing.T) {
	var reads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		reads.Add(1)
		if request.URL.Path != "/markets/keyset" {
			t.Errorf("runWith(pages) path=%s, want /markets/keyset", request.URL.Path)
		}
		body := `{"markets":[{"id":"1","outcomes":["Yes","No"]}],"next_cursor":"server-next"}`
		if request.URL.Query().Get("after_cursor") == "server-next" {
			body = `{"markets":[{"id":"2","outcomes":["Yes","No"]}],"next_cursor":null}`
		}
		if _, err := fmt.Fprint(writer, body); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	config := polymarket.ClientConfig{GammaURL: server.URL}
	var first bytes.Buffer
	if err := runWith(config, []string{"-pages", "1"}, &first); err != nil {
		t.Fatalf("runWith(pages=1)=%v, want nil", err)
	}
	if reads.Load() != 1 || !strings.Contains(first.String(), "market=1") {
		t.Fatalf("runWith(pages=1) reads=%d output=%q, want one page without prefetch", reads.Load(), first.String())
	}
	_, quoted, found := strings.Cut(first.String(), "resume with -cursor ")
	if !found {
		t.Fatalf("runWith(pages=1) output=%q, want resume cursor", first.String())
	}
	cursor, err := strconv.Unquote(strings.TrimSpace(quoted))
	if err != nil {
		t.Fatalf("Unquote(resume cursor)=%v, want nil", err)
	}
	var resumed bytes.Buffer
	if err = runWith(config, []string{"-cursor", cursor}, &resumed); err != nil {
		t.Fatalf("runWith(resume)=%v, want nil", err)
	}
	if reads.Load() != 2 || !strings.Contains(resumed.String(), "market=2") || strings.Contains(resumed.String(), "market=1") || !strings.Contains(resumed.String(), "query complete") {
		t.Errorf("runWith(resume) reads=%d output=%q, want second page only then completion", reads.Load(), resumed.String())
	}
}

func TestPaginationReportsReadFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(503) }))
	t.Cleanup(server.Close)
	var output bytes.Buffer
	err := runWith(polymarket.ClientConfig{GammaURL: server.URL}, nil, &output)
	rejection, ok := errors.AsType[*polymarket.RequestRejectedError](err)
	if !ok || rejection.Status != 503 || strings.Contains(output.String(), "query complete") {
		t.Errorf("runWith(HTTP 503)=%v output=%q, want wrapped rejection without completion", err, output.String())
	}
}

func TestPaginationRejectsArgumentsAndForeignCursor(t *testing.T) {
	var reads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { reads.Add(1); writer.WriteHeader(500) }))
	t.Cleanup(server.Close)
	for _, args := range [][]string{{"-pages", "0"}, {"extra"}, {"-unknown"}, {"-cursor", "invalid-cursor"}} {
		var output bytes.Buffer
		if err := runWith(polymarket.ClientConfig{GammaURL: server.URL}, args, &output); err == nil {
			t.Errorf("runWith(%v)=nil, want input error", args)
		}
	}
	if reads.Load() != 0 {
		t.Errorf("runWith(invalid arguments) reads=%d, want 0", reads.Load())
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

func TestPaginationOutputFailureStopsReads(t *testing.T) {
	for _, next := range []string{`null`, `"next-page"`} {
		for failAfter := range 3 {
			var reads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				reads.Add(1)
				if _, err := fmt.Fprintf(writer, `{"markets":[{"id":"1","outcomes":["Yes","No"]}],"next_cursor":%s}`, next); err != nil {
					t.Error(err)
				}
			}))
			failure := errors.New("output closed")
			err := runWith(polymarket.ClientConfig{GammaURL: server.URL}, []string{"-pages", "1"}, &failingOutput{remaining: failAfter, cause: failure})
			server.Close()
			if !errors.Is(err, failure) || reads.Load() != 1 {
				t.Errorf("runWith(next=%s, fail after %d writes)=%v reads=%d, want output error and no prefetch", next, failAfter, err, reads.Load())
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
