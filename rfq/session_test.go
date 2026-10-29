package rfq

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
	"github.com/coder/websocket"
	"github.com/ericlagergren/decimal"
)

func TestSessionAuthenticationQuoteAndDisconnect(t *testing.T) {
	request := `{"type":"RFQ_REQUEST","rfq_id":"r1","requestor_public_id":"requestor","leg_position_ids":["123","456"],"condition_id":"0x03` + strings.Repeat("ab", 30) + `","yes_position_id":"123","no_position_id":"124","direction":"BUY","side":"YES","requested_size":{"unit":"shares","value_e6":"2000000"},"submission_deadline":1720000030}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = conn.CloseNow() }()
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_, raw, err := conn.Read(ctx)
		if err != nil {
			t.Error(err)
			return
		}
		var auth map[string]json.RawMessage
		if err = json.Unmarshal(raw, &auth); err != nil {
			t.Error(err)
			return
		}
		if string(auth["type"]) != `"auth"` || len(auth["identity"]) == 0 {
			t.Error("missing auth identity")
		}
		if err = conn.Write(ctx, websocket.MessageText, []byte(`{"type":"auth","success":true}`)); err != nil {
			t.Error(err)
			return
		}
		if err = conn.Write(ctx, websocket.MessageText, []byte(`{"type":"RFQ_REQUEST","rfq_id":99}`)); err != nil {
			t.Error(err)
			return
		}
		if err = conn.Write(ctx, websocket.MessageText, []byte(request)); err != nil {
			t.Error(err)
			return
		}
		_, raw, err = conn.Read(ctx)
		if err != nil {
			t.Error(err)
			return
		}
		var quote map[string]json.RawMessage
		if err = json.Unmarshal(raw, &quote); err != nil {
			t.Error(err)
			return
		}
		if string(quote["type"]) != `"RFQ_QUOTE"` || len(quote["signed_order"]) == 0 {
			t.Error("missing signed quote")
		}
		if err = conn.Write(ctx, websocket.MessageText, []byte(`{"type":"ACK_RFQ_QUOTE","rfq_id":"r1","quote_id":"q1"}`)); err != nil {
			t.Error(err)
			return
		}
		_, _, err = conn.Read(ctx)
		if err != nil {
			t.Error(err)
			return
		}
		if err = conn.Write(ctx, websocket.MessageText, []byte(`{"type":"RFQ_ERROR","request_type":"RFQ_QUOTE_CANCEL","rfq_id":"r1","quote_id":"q1","code":"FUTURE_CODE","error_id":"e1","error":"cannot cancel"}`)); err != nil {
			t.Error(err)
			return
		}
		_, _, _ = conn.Read(ctx)
	}))
	defer server.Close()
	signer, err := signing.New("0x0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	session, err := Open(t.Context(), Config{URL: "ws" + strings.TrimPrefix(server.URL, "http"), Signer: signer, Wallet: models.EVMAddress(signer.Address()), WalletType: models.WalletEOA, Credentials: models.APIKeyCreds{Key: "test", Secret: base64.StdEncoding.EncodeToString([]byte("secret")), Passphrase: "test"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	event, err := session.Next(t.Context())
	if err != nil || event.RFQID != "r1" {
		t.Fatalf("event %+v err %v", event, err)
	}
	quote, err := session.Quote(t.Context(), event, QuoteParams{Price: decimal.New(5, 1)})
	if err != nil || quote.QuoteID != "q1" {
		t.Fatalf("quote %+v err %v", quote, err)
	}
	_, err = session.CancelQuote(t.Context(), quote)
	var rejected *Rejection
	if !errors.As(err, &rejected) || rejected.Code != "FUTURE_CODE" || rejected.ErrorID != "e1" {
		t.Fatalf("lost rejection details: %v", err)
	}
	if err = session.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = session.Next(t.Context()); !errors.Is(err, io.EOF) {
		t.Fatalf("close: %v", err)
	}
}
func TestLateAcknowledgementCannotConfirmNewQuote(t *testing.T) {
	first := make(chan acknowledgement, 1)
	second := make(chan acknowledgement, 1)
	key := ackKey{kind: "ACK_RFQ_QUOTE", rfq: "r1"}
	session := Session{pending: map[ackKey][]chan acknowledgement{key: {first, second}}}
	session.mu.Lock()
	session.pending[key][0] = nil
	session.mu.Unlock()
	session.resolve(key, acknowledgement{reference: QuoteReference{QuoteID: "old"}})
	select {
	case <-second:
		t.Fatal("late acknowledgement confirmed newer quote")
	default:
	}
	session.resolve(key, acknowledgement{reference: QuoteReference{QuoteID: "new"}})
	if (<-second).reference.QuoteID != "new" {
		t.Fatal("lost next acknowledgement")
	}
}
