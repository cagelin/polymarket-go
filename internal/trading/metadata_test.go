package trading

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMetadataLoadsDoNotBlockUnrelatedKeys(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	client := fixtureClient(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/markets-by-token/123" {
			close(started)
			<-release
			w.WriteHeader(503)
			return
		}
		writeResponse(t, w, map[string]string{"maker": "0", "taker": "0.001"})
	})
	// Ensure cleanup releases the handler before fixtureClient closes its server.
	defer unblock()
	done := make(chan error, 1)
	go func() { _, err := client.market(t.Context(), "123", false); done <- err }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("market request did not start")
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if _, err := client.builderRate(ctx, "0x"+strings.Repeat("ab", 32)); err != nil {
		t.Errorf("builderRate while unrelated market is stalled = %v, want success", err)
	}
	unblock()
	if loadErr := <-done; loadErr == nil {
		t.Error("stalled market returned success, want rejection")
	}
}
func TestConcurrentColdMetadataIsFetchedOnce(t *testing.T) {
	var lookups, markets atomic.Int32
	condition := "0x" + strings.Repeat("ab", 32)
	client := fixtureClient(t, func(w http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/markets-by-token/") {
			lookups.Add(1)
			writeResponse(t, w, map[string]string{"condition_id": condition})
			return
		}
		markets.Add(1)
		writeResponse(t, w, map[string]any{"mts": "0.01", "nr": false, "t": []any{map[string]string{"t": "123"}}})
	})
	var workers sync.WaitGroup
	for range 16 {
		workers.Go(func() {
			if _, err := client.market(t.Context(), "123", false); err != nil {
				t.Error(err)
			}
		})
	}
	workers.Wait()
	if lookups.Load() != 1 || markets.Load() != 1 {
		t.Errorf("cold market requests = %d/%d, want 1/1", lookups.Load(), markets.Load())
	}
	client.metadata.keyMu.Lock()
	retained := len(client.metadata.keys)
	client.metadata.keyMu.Unlock()
	if retained != 0 {
		t.Errorf("metadata key locks retained = %d, want 0", retained)
	}
}
