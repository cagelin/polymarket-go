package streams

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cagelam/polymarket-go/models"
	"github.com/coder/websocket"
)

type testPeer struct {
	conn   *websocket.Conn
	frames chan []byte
	done   chan struct{}
}

func startServer(t *testing.T) (string, <-chan *testPeer) {
	t.Helper()
	peers := make(chan *testPeer, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.CloseNow() }()
		ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 10*time.Second)
		defer cancel()
		peer := &testPeer{conn: conn, frames: make(chan []byte, 30), done: make(chan struct{})}
		defer close(peer.done)
		select {
		case peers <- peer:
		case <-t.Context().Done():
			return
		}
		for {
			_, raw, readErr := conn.Read(ctx)
			if readErr != nil {
				return
			}
			select {
			case peer.frames <- raw:
			case <-t.Context().Done():
				return
			}
		}
	}))
	t.Cleanup(server.Close)
	return "ws" + strings.TrimPrefix(server.URL, "http"), peers
}
func receive[T any](t *testing.T, values <-chan T) T {
	t.Helper()
	select {
	case value := <-values:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for stream test event")
		var zero T
		return zero
	}
}
func writePeer(t *testing.T, peer *testPeer, raw string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := peer.conn.Write(ctx, websocket.MessageText, []byte(raw)); err != nil {
		t.Fatal(err)
	}
}
func nextEvent(t *testing.T, h *Handle) models.StreamEvent {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	event, err := h.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return event
}
func expectFrame(t *testing.T, peer *testPeer, want string) {
	t.Helper()
	raw := receive(t, peer.frames)
	var actual any
	if err := json.Unmarshal(raw, &actual); err != nil {
		t.Fatalf("frame %s: %v", raw, err)
	}
	equalJSON(t, actual, json.RawMessage(want))
}

func TestMarketMultiplexReconnectAndCancel(t *testing.T) {
	address, peers := startServer(t)
	client, err := New(Config{MarketURL: address})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	first, err := client.Subscribe(ctx, MarketSpec{AssetIDs: []string{"b", "a", "a"}})
	if err != nil {
		t.Fatal(err)
	}
	peer := receive(t, peers)
	expectFrame(t, peer, `{"type":"market","assets_ids":["a","b"],"custom_feature_enabled":false}`)
	writePeer(t, peer, `[{"bad":1},{"event_type":"book","market":"m","asset_id":"unselected","bids":[],"asks":[]},{"event_type":"book","market":"m","asset_id":"a","bids":[],"asks":[]}]`)
	if event, ok := nextEvent(t, first).(*models.MarketBookEvent); !ok || event.Payload.AssetID != "a" {
		t.Fatal("wrong market event")
	}
	second, err := client.Subscribe(ctx, MarketSpec{AssetIDs: []string{"a"}, CustomFeatureEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	expectFrame(t, peer, `{"operation":"subscribe","assets_ids":["a"],"custom_feature_enabled":true}`)
	writePeer(t, peer, `{"event_type":"best_bid_ask","market":"m","asset_id":"a","best_bid":"0.1"}`)
	if _, ok := nextEvent(t, second).(*models.MarketBestBidAskEvent); !ok {
		t.Fatal("wrong custom event")
	}
	if err = second.Close(); err != nil {
		t.Fatal(err)
	}
	expectFrame(t, peer, `{"operation":"subscribe","assets_ids":["a"],"custom_feature_enabled":false}`)
	_ = peer.conn.CloseNow()
	reconnected := receive(t, peers)
	expectFrame(t, reconnected, `{"type":"market","assets_ids":["a","b"],"custom_feature_enabled":false}`)
	writePeer(t, reconnected, `{"event_type":"book","market":"m","asset_id":"b","bids":[],"asks":[]}`)
	if event, ok := nextEvent(t, first).(*models.MarketBookEvent); !ok || event.Payload.AssetID != "b" {
		t.Fatal("custom event leaked or reconnect lost state")
	}
	if stats := client.StreamStats(); stats.Reconnects != 1 || stats.MalformedMessages != 1 {
		t.Fatalf("stats=%+v", stats)
	}
	cancel()
	receive(t, reconnected.done)
	if _, err = first.Next(t.Context()); !errors.Is(err, io.EOF) {
		t.Fatalf("closed handle error=%v", err)
	}
}

func TestHeartbeatStaleReconnect(t *testing.T) {
	address, peers := startServer(t)
	client, err := New(Config{MarketURL: address})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	client.groups["market"].heartbeatInterval = 10 * time.Millisecond
	client.groups["market"].staleAfter = 50 * time.Millisecond
	_, err = client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"a"}})
	if err != nil {
		t.Fatal(err)
	}
	first := receive(t, peers)
	receive(t, first.frames)
	if raw := string(receive(t, first.frames)); raw != "PING" {
		t.Fatalf("heartbeat=%q", raw)
	}
	// The peer deliberately omits PONG; the SDK must reconnect without user action.
	second := receive(t, peers)
	expectFrame(t, second, `{"type":"market","assets_ids":["a"],"custom_feature_enabled":false}`)
}

func TestSportsHeartbeatAndRTDSFiltering(t *testing.T) {
	sportsURL, sportsPeers := startServer(t)
	rtdsURL, rtdsPeers := startServer(t)
	client, err := New(Config{SportsURL: sportsURL, RTDSURL: rtdsURL})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	h, err := client.Subscribe(t.Context(), SportsSpec{}, CryptoPricesSpec{Topic: "prices.crypto.binance", Symbols: []string{"btcusdt"}})
	if err != nil {
		t.Fatal(err)
	}
	sports, rtds := receive(t, sportsPeers), receive(t, rtdsPeers)
	expectFrame(t, rtds, `{"action":"subscribe","subscriptions":[{"topic":"crypto_prices","type":"update"}]}`)
	writePeer(t, sports, "ping")
	if string(receive(t, sports.frames)) != "pong" {
		t.Fatal("missing sports pong")
	}
	writePeer(t, sports, `{"gameId":1,"leagueAbbreviation":"nba","status":"live","live":true,"ended":false,"score":"1-0"}`)
	if _, ok := nextEvent(t, h).(*models.SportsResultEvent); !ok {
		t.Fatal("wrong sports event")
	}
	writePeer(t, rtds, `{"topic":"crypto_prices","type":"update","timestamp":null,"payload":{"symbol":"ethusdt","timestamp":1,"value":"2"}}`)
	writePeer(t, rtds, `{"topic":"crypto_prices","type":"update","timestamp":null,"payload":{"symbol":"btcusdt","timestamp":1,"value":"1.1234567890123456789"}}`)
	event, ok := nextEvent(t, h).(*models.CryptoPricesBinanceEvent)
	if !ok || event.Payload.Symbol != "btcusdt" || event.Payload.Value.String() != "1.1234567890123456789" {
		t.Fatal("incorrect price/filter")
	}
	if err = h.Close(); err != nil {
		t.Fatal(err)
	}
	receive(t, sports.done)
	receive(t, rtds.done)
}

func TestQueueDropOldestAndConcurrentClose(t *testing.T) {
	h := newHandle(2)
	for _, kind := range []string{"one", "two", "three"} {
		h.push(&models.MarketBookEvent{Type: kind})
	}
	if h.Dropped() != 1 || nextEvent(t, h).StreamType() != "two" || nextEvent(t, h).StreamType() != "three" {
		t.Fatal("drop-oldest queue order")
	}
	var wait sync.WaitGroup
	for range 20 {
		wait.Go(func() { _ = h.Close() })
	}
	wait.Wait()
	if _, err := h.Next(t.Context()); !errors.Is(err, io.EOF) {
		t.Fatalf("closed queue=%v", err)
	}
}

func TestSubscribeRollbackAndClosedClient(t *testing.T) {
	client, err := New(Config{MarketURL: "ws://127.0.0.1:1"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	if _, err = client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"a"}}); err == nil {
		t.Fatal("expected connect error")
	}
	if len(client.handles) != 0 || len(client.order) != 0 {
		t.Fatal("failed subscription leaked")
	}
	var nilSpec *MarketSpec
	if _, err = client.Subscribe(t.Context(), nilSpec); err == nil {
		t.Fatal("nil spec accepted")
	}
	if err = client.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = client.Subscribe(t.Context(), SportsSpec{}); err == nil {
		t.Fatal("closed client accepted subscription")
	}
}

func TestCloseInterruptsSubscriptionDial(t *testing.T) {
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { close(entered); <-r.Context().Done() }))
	defer server.Close()
	client, err := New(Config{MarketURL: "ws" + strings.TrimPrefix(server.URL, "http")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	result := make(chan error, 1)
	go func() {
		_, subscribeErr := client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"asset"}})
		result <- subscribeErr
	}()
	receive(t, entered)
	closed := make(chan struct{})
	go func() { _ = client.Close(); close(closed) }()
	receive(t, closed)
	if subscribeErr := receive(t, result); subscribeErr == nil {
		t.Fatal("closed client completed a pending subscription")
	}
}

func TestConcurrentSubscribeAndClientClose(t *testing.T) {
	address, _ := startServer(t)
	client, err := New(Config{MarketURL: address})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			h, subscribeErr := client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"asset"}})
			if subscribeErr == nil {
				_ = h.Close()
			}
		})
	}
	workers.Go(func() { _ = client.Close() })
	workers.Wait()
	if len(client.handles) != 0 {
		t.Fatal("closed client retained handles")
	}
}
