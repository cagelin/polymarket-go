package streams

import (
	"slices"
	"testing"
	"time"
)

func TestMarketBatchesReconnect(t *testing.T) {
	address, peers := startServer(t)
	results := make(chan MarketBatchResult, 8)
	policy := MarketBatchConfig{Size: 2, Interval: 60 * time.Millisecond, BookTimeout: 150 * time.Millisecond, OnBatch: func(result MarketBatchResult) { results <- result }}
	client, err := New(Config{MarketURL: address, MarketBatches: &policy, RawOnly: true, RawHandler: func(RawFrame) {}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	_, err = client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"d", "c", "b", "a"}, CustomFeatureEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	for generation := uint64(1); generation <= 2; generation++ {
		peer := receive(t, peers)
		expectFrame(t, peer, `{"type":"market","assets_ids":["a","b"],"custom_feature_enabled":true}`)
		writePeer(t, peer, `[{"event_type":"book","asset_id":"a"},{"event_type":"book","asset_id":"a"},{"event_type":"book","asset_id":"c"},{"event_type":"price_change","asset_id":"b"}]`)
		var missing []string
		if generation == 1 {
			missing = []string{"b"}
		} else {
			writePeer(t, peer, `{"event_type":"book","asset_id":"b"}`)
		}
		first := receive(t, results)
		if first.Generation != generation || !slices.Equal(first.Assets, []string{"a", "b"}) || !slices.Equal(first.Missing, missing) {
			t.Fatalf("first batch = %+v, want generation %d missing %v", first, generation, missing)
		}
		expectFrame(t, peer, `{"operation":"subscribe","assets_ids":["c","d"],"custom_feature_enabled":true}`)
		writePeer(t, peer, `[{"event_type":"book","asset_id":"c"},{"event_type":"book","asset_id":"d"}]`)
		second := receive(t, results)
		minimum := policy.Interval
		if generation == 1 {
			minimum = policy.BookTimeout
		}
		if second.Generation != generation || !slices.Equal(second.Assets, []string{"c", "d"}) || len(second.Missing) != 0 || second.SentAt.Sub(first.SentAt) < minimum {
			t.Fatalf("second batch = %+v, want generation %d complete after %v", second, generation, minimum)
		}
		if generation == 1 {
			_ = peer.conn.CloseNow()
		}
	}
}

func TestMarketBatchesChangeHandles(t *testing.T) {
	address, peers := startServer(t)
	results := make(chan MarketBatchResult, 8)
	policy := MarketBatchConfig{Size: 1, BookTimeout: time.Second, OnBatch: func(result MarketBatchResult) { results <- result }}
	client, err := New(Config{MarketURL: address, MarketBatches: &policy})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	first, err := client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"a", "b", "c"}})
	if err != nil {
		t.Fatal(err)
	}
	peer := receive(t, peers)
	expectFrame(t, peer, `{"type":"market","assets_ids":["a"],"custom_feature_enabled":false}`)
	_, err = client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"b", "d"}})
	if err != nil {
		t.Fatal(err)
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	expectFrame(t, peer, `{"operation":"unsubscribe","assets_ids":["a"]}`)
	for _, asset := range []string{"b", "d"} {
		expectFrame(t, peer, `{"operation":"subscribe","assets_ids":["`+asset+`"],"custom_feature_enabled":false}`)
		writePeer(t, peer, `{"event_type":"book","asset_id":"`+asset+`"}`)
	}
	for _, asset := range []string{"a", "b", "d"} {
		result := receive(t, results)
		if !slices.Equal(result.Assets, []string{asset}) || len(result.Missing) != 0 {
			t.Fatalf("batch = %+v, want %s resolved", result, asset)
		}
	}
	_ = peer.conn.CloseNow()
	restored := receive(t, peers)
	expectFrame(t, restored, `{"type":"market","assets_ids":["b"],"custom_feature_enabled":false}`)
	writePeer(t, restored, `{"event_type":"book","asset_id":"b"}`)
	expectFrame(t, restored, `{"operation":"subscribe","assets_ids":["d"],"custom_feature_enabled":false}`)
}

func TestMarketBatchesRestartIncompleteBatch(t *testing.T) {
	address, peers := startServer(t)
	results := make(chan MarketBatchResult, 2)
	raw := make(chan RawFrame, 4)
	client, err := New(Config{MarketURL: address, RawHandler: func(frame RawFrame) { raw <- frame }, MarketBatches: &MarketBatchConfig{Size: 2, BookTimeout: 300 * time.Millisecond, OnBatch: func(result MarketBatchResult) { results <- result }}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	_, err = client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"a", "b", "c"}})
	if err != nil {
		t.Fatal(err)
	}
	peer := receive(t, peers)
	firstFrame := `{"type":"market","assets_ids":["a","b"],"custom_feature_enabled":false}`
	expectFrame(t, peer, firstFrame)
	writePeer(t, peer, `{"event_type":"book","asset_id":"a"}`)
	receive(t, raw)
	_ = peer.conn.CloseNow()
	restored := receive(t, peers)
	expectFrame(t, restored, firstFrame)
	writePeer(t, restored, `{"event_type":"book","asset_id":"b"}`)
	result := receive(t, results)
	if result.Generation != 2 || !slices.Equal(result.Missing, []string{"a"}) {
		t.Fatalf("restored batch = %+v, want only a missing on generation 2", result)
	}
	expectFrame(t, restored, `{"operation":"subscribe","assets_ids":["c"],"custom_feature_enabled":false}`)
}
