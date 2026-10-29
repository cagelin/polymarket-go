package rfq

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/coder/websocket"
)

func TestHeartbeatTerminatesUnresponsivePeer(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		conn, err := websocket.Accept(w, request, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = conn.CloseNow() }()
		if _, _, err = conn.Read(request.Context()); err != nil {
			t.Error(err)
			return
		}
		if err = conn.Write(request.Context(), websocket.MessageText, []byte(`{"type":"auth","success":true}`)); err != nil {
			t.Error(err)
			return
		}
		// No further reads: the peer cannot respond to control-frame pings.
		<-release
	}))
	t.Cleanup(server.Close)
	defer close(release)
	signer, err := signing.New(strings.Repeat("0", 63) + "1")
	if err != nil {
		t.Fatal(err)
	}
	session, err := Open(t.Context(), Config{URL: "ws" + strings.TrimPrefix(server.URL, "http"), Signer: signer, HeartbeatInterval: 10 * time.Millisecond, PongTimeout: 20 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	_, err = session.Next(ctx)
	var failure *sdkerrors.TransportError
	if !errors.As(err, &failure) || !session.Closed() {
		t.Errorf("Next(unresponsive peer) = %v, closed=%t; want terminal TransportError", err, session.Closed())
	}
}
