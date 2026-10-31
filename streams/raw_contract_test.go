package streams

import (
	"errors"
	"testing"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/models"
	"github.com/coder/websocket"
)

func TestRawConfigValidation(t *testing.T) {
	cases := []struct {
		name        string
		config      Config
		wantInvalid bool
		wantLimit   int64
	}{
		{name: "default", wantLimit: 10 << 20},
		{name: "explicit limit", config: Config{ReadLimit: 1024}, wantLimit: 1024},
		{name: "negative limit", config: Config{ReadLimit: -1}, wantInvalid: true},
		{name: "raw only without handler", config: Config{RawOnly: true}, wantInvalid: true},
		{name: "raw only", config: Config{RawOnly: true, RawHandler: func(RawFrame) {}}, wantLimit: 10 << 20},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(subtest *testing.T) {
			client, err := New(testCase.config)
			if testCase.wantInvalid {
				if _, invalidConfig := errors.AsType[*sdkerrors.UserInputError](err); !invalidConfig {
					subtest.Errorf("New(%+v) error=%v, want UserInputError", testCase.config, err)
				}
				return
			}
			if err != nil {
				subtest.Fatal(err)
			}
			subtest.Cleanup(func() { _ = client.Close() })
			if client.config.ReadLimit != testCase.wantLimit {
				subtest.Errorf("New(%+v).ReadLimit=%d, want %d", testCase.config, client.config.ReadLimit, testCase.wantLimit)
			}
		})
	}
}

func TestRawBinaryPongAndHeartbeat(t *testing.T) {
	address, peers := startServer(t)
	received := make(chan RawFrame, 4)
	states := make(chan Lifecycle, 32)
	client, err := New(Config{MarketURL: address, RawOnly: true,
		RawHandler:       func(frame RawFrame) { received <- frame },
		LifecycleHandler: func(observation Lifecycle) { states <- observation },
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	client.groups["market"].heartbeatInterval = 20 * time.Millisecond
	client.groups["market"].staleAfter = time.Second
	if _, err = client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"a"}}); err != nil {
		t.Fatal(err)
	}
	peer := receive(t, peers)
	receive(t, peer.frames)
	connected := receive(t, states)
	if connected.State != "connected" || connected.Generation == 0 || connected.At.IsZero() {
		t.Fatalf("initial lifecycle=%+v, want connected with nonzero generation and timestamp", connected)
	}
	if heartbeat := string(receive(t, peer.frames)); heartbeat != "PING" {
		t.Fatalf("heartbeat=%q, want PING", heartbeat)
	}
	writePeer(t, peer, "PONG")
	for observation := receive(t, states); observation.State != "pong"; observation = receive(t, states) {
		if observation.Generation != connected.Generation {
			t.Fatalf("heartbeat generation=%d, want %d", observation.Generation, connected.Generation)
		}
	}
	if err = peer.conn.Write(t.Context(), websocket.MessageBinary, []byte("PONG")); err != nil {
		t.Fatal(err)
	}
	binary := receive(t, received)
	if string(binary.Data) != "PONG" || binary.Sequence != 1 || binary.Generation != connected.Generation || binary.ReceivedAt.IsZero() || binary.Service != "market" {
		t.Errorf("binary message=%+v, want first complete binary PONG with connection metadata", binary)
	}
	writePeer(t, peer, "{broken")
	malformed := receive(t, received)
	if string(malformed.Data) != "{broken" || malformed.Sequence != 2 || client.StreamStats().MalformedMessages != 0 {
		t.Errorf("raw-only malformed=%+v stats=%+v, want second raw message without typed decoding", malformed, client.StreamStats())
	}
}

func TestCloseWaitsForRawCallback(t *testing.T) {
	address, peers := startServer(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})
	client, err := New(Config{MarketURL: address, RawOnly: true, RawHandler: func(RawFrame) {
		close(entered)
		<-release
		close(finished)
	}})
	if err != nil {
		t.Fatal(err)
	}
	released := false
	t.Cleanup(func() {
		if !released {
			close(release)
		}
		_ = client.Close()
	})
	if _, err = client.Subscribe(t.Context(), MarketSpec{AssetIDs: []string{"a"}}); err != nil {
		t.Fatal(err)
	}
	peer := receive(t, peers)
	receive(t, peer.frames)
	writePeer(t, peer, "{}")
	receive(t, entered)
	closed := make(chan error, 1)
	go func() { closed <- client.Close() }()
	// Closing the socket proves Close has started while the reader is in the callback.
	receive(t, peer.done)
	select {
	case closeErr := <-closed:
		t.Fatalf("Close() returned %v before the raw callback finished", closeErr)
	default:
	}
	close(release)
	released = true
	if closeErr := receive(t, closed); closeErr != nil {
		t.Fatal(closeErr)
	}
	receive(t, finished)
}

func TestRawOnlyPreservesOtherServices(t *testing.T) {
	address, peers := startServer(t)
	received := make(chan RawFrame, 1)
	client, err := New(Config{SportsURL: address, RawOnly: true, RawHandler: func(frame RawFrame) { received <- frame }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	handle, err := client.Subscribe(t.Context(), SportsSpec{})
	if err != nil {
		t.Fatal(err)
	}
	peer := receive(t, peers)
	writePeer(t, peer, `{"gameId":1,"leagueAbbreviation":"nba","status":"live","live":true,"ended":false,"score":"1-0"}`)
	if event, ok := nextEvent(t, handle).(*models.SportsResultEvent); !ok || event.Payload.Score != "1-0" {
		t.Errorf("raw-only sports event=%+v, want typed SportsResultEvent with score 1-0", event)
	}
	if count := len(received); count != 0 {
		t.Errorf("non-market raw callbacks=%d, want 0", count)
	}
}
