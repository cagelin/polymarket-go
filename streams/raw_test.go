package streams

import (
	"bytes"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRawOwnershipWithTypedArray(t *testing.T) {
	address, peers := startServer(t)
	received := make(chan RawFrame, 8)
	client, err := New(Config{MarketURL: address, RawHandler: func(frame RawFrame) { received <- frame }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	handle, err := client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"a"}})
	if err != nil {
		t.Fatal(err)
	}
	peer := receive(t, peers)
	receive(t, peer.frames)
	payload := `[{"event_type":"book","market":"m","asset_id":"a","bids":[],"asks":[]}]`
	writePeer(t, peer, payload)
	first := receive(t, received)
	nextEvent(t, handle)
	writePeer(t, peer, payload)
	receive(t, received)
	nextEvent(t, handle)
	if !bytes.Equal(first.Data, []byte(payload)) {
		t.Errorf("retained array bytes=%q, want %q", first.Data, payload)
	}
}

func TestRawOnlyReconnectUnknownMalformedAndClose(t *testing.T) {
	address, peers := startServer(t)
	received := make(chan RawFrame, 8)
	states := make(chan Lifecycle, 16)
	client, err := New(Config{MarketURL: address, RawOnly: true, RawHandler: func(frame RawFrame) { received <- frame }, LifecycleHandler: func(event Lifecycle) { states <- event }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	handle, err := client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"a"}})
	if err != nil {
		t.Fatal(err)
	}
	peer := receive(t, peers)
	receive(t, peer.frames)
	connected := receive(t, states)
	writePeer(t, peer, `{"event_type":"future"}`)
	first := receive(t, received)
	writePeer(t, peer, "\x00broken")
	second := receive(t, received)
	if err = peer.conn.CloseNow(); err != nil {
		t.Fatal(err)
	}
	reconnected := receive(t, peers)
	receive(t, reconnected.frames)
	writePeer(t, reconnected, `{"event_type":"future"}`)
	third := receive(t, received)
	if first.Sequence != 1 || second.Sequence != 2 || third.Sequence != 3 || third.Generation == first.Generation || connected.State != "connected" {
		t.Errorf("raw reconnect IDs=%+v,%+v,%+v, lifecycle=%+v; want monotonic sequences and new generation", first, second, third, connected)
	}
	if string(first.Data) != `{"event_type":"future"}` || string(second.Data) != "\x00broken" {
		t.Error("raw buffers changed across read/reconnect")
	}
	if client.StreamStats().MalformedMessages != 0 || handle.Dropped() != 0 {
		t.Error("raw-only mode unexpectedly ran typed parsing")
	}
	if err = client.Close(); err != nil {
		t.Fatal(err)
	}
	count := len(received)
	if count != 0 {
		t.Errorf("raw callbacks pending after Close=%d, want 0", count)
	}
}

func TestNonMarketArrayStillReportsMalformed(t *testing.T) {
	client, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	client.deliver("sports", []byte(`[{}]`))
	if client.StreamStats().MalformedMessages != 1 {
		t.Errorf("non-market invalid array stats=%+v, want one malformed", client.StreamStats())
	}
}

func TestReadLimitReportsLossAndBacksOff(t *testing.T) {
	address, peers := startServer(t)
	var rawCount, oversized atomic.Uint64
	limited := make(chan struct{}, 8)
	client, err := New(Config{MarketURL: address, RawOnly: true, ReadLimit: 1024,
		RawHandler: func(RawFrame) { rawCount.Add(1) },
		LifecycleHandler: func(event Lifecycle) {
			if event.State == "read_limit" {
				oversized.Add(1)
				limited <- struct{}{}
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if _, err = client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"a"}}); err != nil {
		t.Fatal(err)
	}
	deadline := time.NewTimer(1500 * time.Millisecond)
	defer deadline.Stop()
	connections := 0
	for {
		select {
		case peer := <-peers:
			connections++
			receive(t, peer.frames)
			writePeer(t, peer, strings.Repeat("x", 1025))
			// Observe the loss before allowing the timer to close the client.
			receive(t, limited)
		case <-deadline.C:
			if err = client.Close(); err != nil {
				t.Fatal(err)
			}
			if connections < 2 || connections > 4 {
				t.Errorf("rapid oversize reconnects=%d, want 2..4 with increasing minimum delays", connections)
			}
			if oversized.Load() != uint64(connections) || rawCount.Load() != 0 {
				t.Errorf("oversized=%d raw=%d connections=%d; want explicit loss without truncated raw", oversized.Load(), rawCount.Load(), connections)
			}
			return
		}
	}
}

func TestOnlyStableConnectionResetsBackoff(t *testing.T) {
	if got := resetReconnectAttempt(4, time.Second); got != 4 {
		t.Errorf("short connection attempt=%d, want 4", got)
	}
	if got := resetReconnectAttempt(4, 30*time.Second); got != 0 {
		t.Errorf("stable connection attempt=%d, want 0", got)
	}
}
