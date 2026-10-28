package streams

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/models"
)

func secureStreamConfig(address string) Config {
	return Config{UserURL: address, PriceURL: address, Credentials: &models.APIKeyCreds{Key: "key", Secret: "secret", Passphrase: "pass"}}
}

type subscribeResult struct {
	handle *Handle
	err    error
}

func beginSubscription(ctx context.Context, c *Client, specs ...Spec) <-chan subscribeResult {
	result := make(chan subscribeResult, 1)
	go func() { h, err := c.Subscribe(ctx, specs...); result <- subscribeResult{h, err} }()
	return result
}
func acknowledge(t *testing.T, peer *testPeer, want string) {
	t.Helper()
	raw := receive(t, peer.frames)
	var frame struct {
		Op, Rid       string
		Auth          map[string]string
		Subscriptions []struct {
			Channel string
			Filter  map[string]any
		}
	}
	if err := json.Unmarshal(raw, &frame); err != nil {
		t.Fatal(err)
	}
	if frame.Op != want || frame.Rid == "" {
		t.Fatalf("unexpected operation %s", raw)
	}
	if want == "auth" {
		if frame.Auth["apiKey"] != "key" || frame.Auth["secret"] != "secret" {
			t.Fatal("missing stream credentials")
		}
		writePeer(t, peer, `{"op":"authed","rid":"`+frame.Rid+`"}`)
		return
	}
	op := "subscribed"
	if want == "unsubscribe" {
		op = "unsubscribed"
	}
	for _, sub := range frame.Subscriptions {
		writePeer(t, peer, `{"op":"`+op+`","rid":"`+frame.Rid+`","channel":"`+sub.Channel+`"}`)
	}
}
func TestPriceAcceptanceSharingAndReconnect(t *testing.T) {
	address, peers := startServer(t)
	client, err := New(secureStreamConfig(address))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	waiting := beginSubscription(t.Context(), client, CryptoPriceSpec{Symbols: []string{"btcusd"}})
	peer := receive(t, peers)
	select {
	case <-waiting:
		t.Fatal("returned before authentication")
	default:
	}
	acknowledge(t, peer, "auth")
	select {
	case <-waiting:
		t.Fatal("returned before subscription acceptance")
	default:
	}
	acknowledge(t, peer, "subscribe")
	first := receive(t, waiting)
	if first.err != nil {
		t.Fatal(first.err)
	}
	defer func() { _ = first.handle.Close() }()
	writePeer(t, peer, `{"v":1,"channel":"price.crypto","seq":1,"ts":1720000000000,"payload":{"symbol":"BTCUSD","source":"pyth","timestamp":1720000000000,"value":0.5,"full_accuracy_value":"0.50000000000000000000001"}}`)
	event, ok := nextEvent(t, first.handle).(*models.CryptoPriceUpdateEvent)
	if !ok || event.Payload.Value.String() != "0.50000000000000000000001" {
		t.Fatal("exact price lost")
	}
	second, err := client.Subscribe(t.Context(), CryptoPriceSpec{Symbols: []string{"btcusd"}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = second.Close() }()
	if _, ok = nextEvent(t, second).(*models.CryptoPriceSnapshotEvent); !ok {
		t.Fatal("shared subscription did not receive history")
	}
	select {
	case extra := <-peers:
		t.Fatalf("duplicate key opened extra connection %v", extra)
	default:
	}
	_ = peer.conn.CloseNow()
	replacement := receive(t, peers)
	acknowledge(t, replacement, "auth")
	acknowledge(t, replacement, "subscribe")
	writePeer(t, replacement, `{"v":1,"channel":"price.crypto","seq":1,"ts":1720000000001,"payload":{"symbol":"btcusd","source":"chainlink","timestamp":1720000000001,"value":0.6,"full_accuracy_value":"0.6"}}`)
	if nextEvent(t, first.handle).StreamType() != "update" || nextEvent(t, second).StreamType() != "update" {
		t.Fatal("reconnect delivery failed")
	}
}
func TestPriceRejectionAndCancellation(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(map[bool]string{true: "cancelled", false: "rejected"}[cancelled], func(subtest *testing.T) {
			address, peers := startServer(subtest)
			client, err := New(secureStreamConfig(address))
			if err != nil {
				subtest.Fatal(err)
			}
			subtest.Cleanup(func() { _ = client.Close() })
			ctx, cancel := context.WithCancel(subtest.Context())
			defer cancel()
			waiting := beginSubscription(ctx, client, EquityPriceSpec{Symbol: " AAPL "})
			peer := receive(subtest, peers)
			acknowledge(subtest, peer, "auth")
			raw := receive(subtest, peer.frames)
			var frame struct{ Rid string }
			if err = json.Unmarshal(raw, &frame); err != nil {
				subtest.Fatal(err)
			}
			if cancelled {
				cancel()
			} else {
				writePeer(subtest, peer, `{"op":"error","rid":"`+frame.Rid+`","channel":"price.equity","code":"bad_filter"}`)
			}
			result := receive(subtest, waiting)
			if result.err == nil {
				subtest.Fatal("unaccepted subscription returned success")
			}
			if !cancelled {
				var rejected *sdkerrors.RequestRejectedError
				if !errors.As(result.err, &rejected) || rejected.Code != "bad_filter" {
					subtest.Fatalf("rejection %v", result.err)
				}
			}
		})
	}
}
func TestPriceProviderIsolationAndPublicRejection(t *testing.T) {
	public, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = public.Close() }()
	if _, err = public.Subscribe(t.Context(), CryptoPriceSpec{Symbols: []string{"btcusd"}}); err == nil {
		t.Fatal("public client accepted authenticated stream")
	}
	address, peers := startServer(t)
	client, err := New(secureStreamConfig(address))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	for _, provider := range []string{"pyth", "chainlink"} {
		waiting := beginSubscription(t.Context(), client, CryptoPriceSpec{Symbols: []string{"btcusd"}, Provider: provider})
		peer := receive(t, peers)
		acknowledge(t, peer, "auth")
		raw := receive(t, peer.frames)
		if !strings.Contains(string(raw), `"provider":"`+provider+`"`) {
			t.Fatalf("wrong provider %s", raw)
		}
		var frame struct{ Rid string }
		if err = json.Unmarshal(raw, &frame); err != nil {
			t.Fatal(err)
		}
		writePeer(t, peer, `{"op":"subscribed","rid":"`+frame.Rid+`","channel":"price.crypto"}`)
		result := receive(t, waiting)
		if result.err != nil {
			t.Fatal(result.err)
		}
	}
}
func TestUserAuthenticationAndAllMarketTransitions(t *testing.T) {
	address, peers := startServer(t)
	client, err := New(secureStreamConfig(address))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	first, err := client.Subscribe(t.Context(), UserSpec{Markets: []string{"b", "a"}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = first.Close() }()
	peer := receive(t, peers)
	expectFrame(t, peer, `{"type":"user","auth":{"apiKey":"key","secret":"secret","passphrase":"pass"},"markets":["a","b"]}`)
	all, err := client.Subscribe(t.Context(), UserSpec{})
	if err != nil {
		t.Fatal(err)
	}
	expectFrame(t, peer, `{"operation":"unsubscribe","markets":["a","b"]}`)
	if err = all.Close(); err != nil {
		t.Fatal(err)
	}
	expectFrame(t, peer, `{"operation":"subscribe","markets":["a","b"]}`)
	_ = peer.conn.CloseNow()
	replacement := receive(t, peers)
	expectFrame(t, replacement, `{"type":"user","auth":{"apiKey":"key","secret":"secret","passphrase":"pass"},"markets":["a","b"]}`)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	if _, err = first.Next(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("empty stream cancellation %v", err)
	}
}

func TestPriceRecoveryRejectsOnlyAffectedFilter(t *testing.T) {
	address, peers := startServer(t)
	client, err := New(secureStreamConfig(address))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	handles := map[string]*Handle{}
	var peer *testPeer
	for _, symbol := range []string{"btcusd", "ethusd"} {
		waiting := beginSubscription(t.Context(), client, CryptoPriceSpec{Symbols: []string{symbol}})
		if peer == nil {
			peer = receive(t, peers)
			acknowledge(t, peer, "auth")
		}
		acknowledge(t, peer, "subscribe")
		result := receive(t, waiting)
		if result.err != nil {
			t.Fatal(result.err)
		}
		handles[symbol] = result.handle
	}
	if peer == nil {
		t.Fatal("subscription did not establish a connection")
	}
	_ = peer.conn.CloseNow()
	replacement := receive(t, peers)
	acknowledge(t, replacement, "auth")
	raw := receive(t, replacement.frames)
	var request struct {
		Rid           string
		Subscriptions []struct {
			Channel string
			Filter  struct{ Symbol string }
		}
	}
	if err = json.Unmarshal(raw, &request); err != nil {
		t.Fatal(err)
	}
	if len(request.Subscriptions) != 2 {
		t.Fatalf("resubscriptions %s", raw)
	}
	for _, sub := range request.Subscriptions {
		if sub.Filter.Symbol == "btcusd" {
			writePeer(t, replacement, `{"op":"error","rid":"`+request.Rid+`","channel":"`+sub.Channel+`","code":"bad_filter"}`)
		} else {
			writePeer(t, replacement, `{"op":"subscribed","rid":"`+request.Rid+`","channel":"`+sub.Channel+`"}`)
		}
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	_, err = handles["btcusd"].Next(ctx)
	if _, ok := errors.AsType[*sdkerrors.RequestRejectedError](err); !ok {
		t.Fatalf("rejected filter: %v", err)
	}
	writePeer(t, replacement, `{"v":1,"channel":"price.crypto","seq":1,"ts":1720000000000,"payload":{"symbol":"ethusd","source":"pyth","timestamp":1720000000000,"value":0.5,"full_accuracy_value":"0.5"}}`)
	if nextEvent(t, handles["ethusd"]).StreamType() != "update" {
		t.Fatal("unrelated filter terminated")
	}
}

func TestPriceInitialAuthenticationUnavailableRetries(t *testing.T) {
	address, peers := startServer(t)
	client, err := New(secureStreamConfig(address))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	waiting := beginSubscription(t.Context(), client, CryptoPriceSpec{Symbols: []string{"btcusd"}})
	peer := receive(t, peers)
	var request struct{ Rid string }
	if err = json.Unmarshal(receive(t, peer.frames), &request); err != nil {
		t.Fatal(err)
	}
	writePeer(t, peer, `{"op":"error","rid":"`+request.Rid+`","code":"auth_unavailable"}`)
	replacement := receive(t, peers)
	acknowledge(t, replacement, "auth")
	acknowledge(t, replacement, "subscribe")
	result := receive(t, waiting)
	if result.err != nil {
		t.Fatal(result.err)
	}
}

func TestPriceRegistrationDuringRecoveryWaitsForAcceptance(t *testing.T) {
	address, peers := startServer(t)
	client, err := New(secureStreamConfig(address))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	waiting := beginSubscription(t.Context(), client, CryptoPriceSpec{Symbols: []string{"btcusd"}})
	peer := receive(t, peers)
	acknowledge(t, peer, "auth")
	acknowledge(t, peer, "subscribe")
	if result := receive(t, waiting); result.err != nil {
		t.Fatal(result.err)
	}
	_ = peer.conn.CloseNow()
	replacement := receive(t, peers)
	// Recovery holds the operation lock; this registration must get its own acknowledgement.
	added := beginSubscription(t.Context(), client, CryptoPriceSpec{Symbols: []string{"ethusd"}})
	acknowledge(t, replacement, "auth")
	acknowledge(t, replacement, "subscribe")
	select {
	case <-added:
		t.Fatal("new filter returned before its acknowledgement")
	default:
	}
	acknowledge(t, replacement, "subscribe")
	if result := receive(t, added); result.err != nil {
		t.Fatal(result.err)
	}
}

func TestPriceSessionFilterLimit(t *testing.T) {
	address, peers := startServer(t)
	client, err := New(secureStreamConfig(address))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	symbols := make([]string, 65)
	for i := range symbols {
		//noinspection SpellCheckingInspection
		symbols[i] = fmt.Sprintf("asset%dusd", i)
	}
	waiting := beginSubscription(t.Context(), client, CryptoPriceSpec{Symbols: symbols})
	total := 0
	for range 2 {
		peer := receive(t, peers)
		acknowledge(t, peer, "auth")
		var request struct {
			Rid           string
			Subscriptions []struct{ Channel string }
		}
		if err = json.Unmarshal(receive(t, peer.frames), &request); err != nil {
			t.Fatal(err)
		}
		count := len(request.Subscriptions)
		if count < 1 || count > 64 {
			t.Fatalf("session has %d filters", count)
		}
		total += count
		for _, sub := range request.Subscriptions {
			writePeer(t, peer, `{"op":"subscribed","rid":"`+request.Rid+`","channel":"`+sub.Channel+`"}`)
		}
	}
	if total != 65 {
		t.Fatalf("filter count %d", total)
	}
	if result := receive(t, waiting); result.err != nil {
		t.Fatal(result.err)
	}
}
