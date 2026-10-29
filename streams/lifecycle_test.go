package streams

import (
	"testing"
	"time"
)

func TestRepeatedConnectionCleanup(t *testing.T) {
	address, peers := startServer(t)
	started := time.Now()
	for cycle := range 64 {
		client, err := New(Config{MarketURL: address, QueueSize: 2})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = client.Close() }) // Close releases local connections and is idempotent.
		handle, err := client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"a"}})
		if err != nil {
			t.Fatal(err)
		}
		peer := receive(t, peers)
		expectFrame(t, peer, `{"type":"market","assets_ids":["a"],"custom_feature_enabled":false}`)
		writePeer(t, peer, `{"event_type":"book","market":"m","asset_id":"a","bids":[],"asks":[]}`)
		if event := nextEvent(t, handle); event == nil {
			t.Fatalf("cycle %d: no event received", cycle)
		}
		if err = client.Close(); err != nil {
			t.Fatal(err)
		}
		receive(t, peer.done)
		client.mu.Lock()
		retained := len(client.handles) + len(client.merged) + len(client.order)
		for _, connection := range client.groups {
			if connection.running || connection.conn != nil {
				retained++
			}
		}
		client.mu.Unlock()
		if retained != 0 {
			t.Fatalf("cycle %d: retained %d subscriptions/workers/sockets after Close", cycle, retained)
		}
	}
	t.Logf("64 connect/receive/close cycles completed in %s", time.Since(started))
}
