package transport

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
)

func TestResponseBodyLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("12345")) }))
	t.Cleanup(server.Close)
	for _, limit := range []int64{4, 5, 6} {
		t.Run(fmt.Sprint(limit), func(subtest *testing.T) {
			client, err := New(Config{MaxResponseBodyBytes: limit})
			if err != nil {
				subtest.Fatal(err)
			}
			subtest.Cleanup(func() { _ = client.Close() })
			raw, err := client.Request(subtest.Context(), server.URL, "GET", "", nil, nil)
			var oversized *sdkerrors.UnexpectedResponseError
			if limit < 5 {
				if !errors.As(err, &oversized) || raw != nil {
					subtest.Errorf("Request(limit=%d) = %q, %v; want size error and no body", limit, raw, err)
				}
			} else if err != nil || string(raw) != "12345" {
				subtest.Errorf("Request(limit=%d) = %q, %v; want full body", limit, raw, err)
			}
		})
	}
}
func TestHeldRequestsReuseConnections(t *testing.T) {
	var connections atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) }))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.Start()
	t.Cleanup(server.Close)
	client, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	for range 3 {
		if _, err = client.LongRequest(t.Context(), server.URL, "GET", "", nil, nil, time.Second); err != nil {
			t.Fatal(err)
		}
	}
	if got := connections.Load(); got != 1 {
		t.Errorf("LongRequest repeated connections = %d, want 1", got)
	}
}
