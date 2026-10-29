package rfq

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/cagelin/polymarket-go/internal/endpoint"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
	"github.com/coder/websocket"
)

// OrderSigner signs RFQ orders using the caller's signing implementation.
type OrderSigner interface {
	Address() string
	Order(models.SignedOrder, string, models.SignerType) (string, error)
}

// Config binds one authenticated wallet to the quoter gateway.
type Config struct {
	URL         string
	HTTPClient  *http.Client
	Signer      OrderSigner
	Credentials models.APIKeyCreds
	Wallet      models.EVMAddress
	WalletType  models.WalletType
	QueueSize   int
	// HeartbeatInterval defaults to ten seconds; PongTimeout defaults to thirty seconds.
	HeartbeatInterval time.Duration
	PongTimeout       time.Duration
}

func (c Config) String() string { return "RFQConfig(<redacted>)" }

// GoString redacts authentication material.
func (c Config) GoString() string { return c.String() }

type acknowledgement struct {
	reference QuoteReference
	err       error
}
type ackKey struct{ kind, rfq, quote string }

// Session keeps one authenticated connection. Disconnection fails pending operations without replay.
type Session struct {
	mu       sync.Mutex
	sendMu   sync.Mutex
	config   Config
	conn     *websocket.Conn
	ctx      context.Context
	cancel   context.CancelFunc
	events   chan Event
	done     chan struct{}
	pending  map[ackKey][]chan acknowledgement
	closed   bool
	terminal error
}

// Open connects and waits for authentication before exposing the session.
func Open(ctx context.Context, config Config) (*Session, error) {
	signerValue := reflect.ValueOf(config.Signer)
	if !signerValue.IsValid() || signerValue.Kind() == reflect.Pointer && signerValue.IsNil() {
		return nil, &sdkerrors.UserInputError{Message: "RFQ signer is required"}
	}
	if config.URL == "" {
		config.URL = "wss://combos-rfq-gateway-quoter.polymarket.com/ws/rfq"
	}
	if _, err := endpoint.Parse(config.URL, "ws", "wss"); err != nil {
		return nil, err
	}
	if config.HeartbeatInterval < 0 || config.PongTimeout < 0 {
		return nil, &sdkerrors.UserInputError{Message: "invalid RFQ heartbeat configuration"}
	}
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = 10 * time.Second
	}
	if config.PongTimeout == 0 {
		config.PongTimeout = 30 * time.Second
	}
	if config.QueueSize == 0 {
		config.QueueSize = 1024
	}
	if config.QueueSize < 1 {
		return nil, &sdkerrors.UserInputError{Message: "RFQ queue size must be positive"}
	}
	// Dial owns the network body; a failed handshake may return buffered diagnostics.
	conn, response, err := websocket.Dial(ctx, config.URL, &websocket.DialOptions{HTTPClient: config.HTTPClient})
	if response != nil && response.Body != nil {
		_ = response.Body.Close() // The SDK does not consume diagnostic data.
	}
	if err != nil {
		return nil, &sdkerrors.TransportError{Cause: err}
	}
	lifetime, cancel := context.WithCancel(context.Background())
	s := &Session{config: config, conn: conn, ctx: lifetime, cancel: cancel, events: make(chan Event, config.QueueSize), done: make(chan struct{}), pending: make(map[ackKey][]chan acknowledgement)}
	conn.SetReadLimit(4 << 20)
	go s.read()
	auth := map[string]any{"type": "auth", "auth": map[string]string{"apiKey": config.Credentials.Key, "passphrase": config.Credentials.Passphrase, "secret": config.Credentials.Secret}, "identity": map[string]any{"signer_address": s.orderSigner(), "maker_address": config.Wallet, "signature_type": signing.SignatureType(config.WalletType)}}
	if _, err = s.request(ctx, ackKey{kind: "auth"}, auth); err != nil {
		_ = s.Close()
		return nil, err
	}
	//nolint:contextcheck // The authenticated session outlives the Open operation context.
	go s.heartbeat()
	return s, nil
}
func (s *Session) orderSigner() models.EVMAddress {
	if s.config.WalletType == models.WalletDeposit {
		return s.config.Wallet
	}
	return models.EVMAddress(s.config.Signer.Address())
}

// Closed reports whether the connection has ended.
func (s *Session) Closed() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.closed }

// Close cancels readers and fails pending acknowledgements.
func (s *Session) Close() error { s.finish(nil); return s.conn.CloseNow() }
func (s *Session) finish(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.terminal = err
	s.cancel()
	failure := err
	if failure == nil {
		failure = io.EOF
	}
	for key, queue := range s.pending {
		for _, pending := range queue {
			if pending != nil {
				pending <- acknowledgement{err: failure}
			}
		}
		delete(s.pending, key)
	}
	close(s.done)
}

// Next waits for one event. An explicit Close ends with io.EOF.
func (s *Session) Next(ctx context.Context) (Event, error) {
	select {
	case event := <-s.events:
		return event, nil
	default:
	}
	select {
	case event := <-s.events:
		return event, nil
	case <-ctx.Done():
		return Event{}, ctx.Err()
	case <-s.done:
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.terminal != nil {
			return Event{}, s.terminal
		}
		return Event{}, io.EOF
	}
}
func (s *Session) request(ctx context.Context, key ackKey, body any) (QuoteReference, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return QuoteReference{}, err
	}
	bounded, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	pending := make(chan acknowledgement, 1)
	s.sendMu.Lock()
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		s.sendMu.Unlock()
		return QuoteReference{}, io.EOF
	}
	s.pending[key] = append(s.pending[key], pending)
	s.mu.Unlock()
	err = s.conn.Write(bounded, websocket.MessageText, raw)
	s.sendMu.Unlock()
	if err != nil {
		s.finish(&sdkerrors.TransportError{Cause: err})
		return QuoteReference{}, err
	}
	select {
	case result := <-pending:
		return result.reference, result.err
	case <-bounded.Done():
		// Keep a tombstone so a late acknowledgement cannot satisfy a newer request.
		s.mu.Lock()
		for i, item := range s.pending[key] {
			if item == pending {
				s.pending[key][i] = nil
				break
			}
		}
		s.mu.Unlock()
		if ctx.Err() != nil {
			return QuoteReference{}, ctx.Err()
		}
		return QuoteReference{}, &sdkerrors.RFQTimeoutError{RFQID: key.rfq}
	}
}
func (s *Session) resolve(key ackKey, value acknowledgement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue := s.pending[key]
	if len(queue) == 0 {
		return
	}
	pending := queue[0]
	if len(queue) == 1 {
		delete(s.pending, key)
	} else {
		s.pending[key] = queue[1:]
	}
	if pending != nil {
		pending <- value
	}
}
func (s *Session) read() {
	defer func() { _ = s.conn.CloseNow() }()
	for {
		kind, raw, err := s.conn.Read(s.ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				s.finish(&sdkerrors.TransportError{Cause: err})
			}
			return
		}
		if kind != websocket.MessageText {
			continue
		}
		if string(raw) == "PING" {
			s.sendMu.Lock()
			writeErr := s.conn.Write(s.ctx, websocket.MessageText, []byte("PONG"))
			s.sendMu.Unlock()
			if writeErr != nil {
				s.finish(writeErr)
				return
			}
			continue
		}
		s.message(raw)
	}
}
func (s *Session) message(raw []byte) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return
	}
	var kind string
	if json.Unmarshal(fields["type"], &kind) != nil {
		return
	}
	if kind == "auth" {
		var success bool
		_ = json.Unmarshal(fields["success"], &success)
		if success {
			s.resolve(ackKey{kind: "auth"}, acknowledgement{})
		} else {
			s.resolve(ackKey{kind: "auth"}, acknowledgement{err: &sdkerrors.UserInputError{Message: "RFQ authentication failed"}})
		}
		return
	}
	if kind == "RFQ_ERROR" {
		s.reject(fields)
		return
	}
	if kind == "ACK_RFQ_QUOTE" || kind == "ACK_RFQ_QUOTE_CANCEL" || kind == "ACK_RFQ_CONFIRMATION_RESPONSE" {
		var reference QuoteReference
		if json.Unmarshal(raw, &reference) != nil || reference.RFQID == "" || reference.QuoteID == "" {
			return
		}
		key := ackKey{kind: kind, rfq: reference.RFQID}
		if kind != "ACK_RFQ_QUOTE" {
			key.quote = reference.QuoteID
		}
		s.resolve(key, acknowledgement{reference: reference})
		return
	}
	event, err := parseEvent(raw)
	if err != nil {
		return
	}
	select {
	case s.events <- event:
	default:
		s.finish(&sdkerrors.UnexpectedResponseError{Message: "RFQ event queue overflow"})
	}
}
func (s *Session) reject(fields map[string]json.RawMessage) {
	var failure Rejection
	for name, target := range map[string]*string{"request_type": &failure.RequestType, "rfq_id": &failure.RFQID, "quote_id": &failure.QuoteID, "code": &failure.Code, "error_id": &failure.ErrorID, "error": &failure.Message} {
		if value, ok := fields[name]; ok && string(value) != "null" {
			if json.Unmarshal(value, target) != nil {
				return
			}
		}
	}
	if failure.Message == "" {
		failure.Message = "RFQ request rejected"
	}
	kind := map[string]string{"RFQ_QUOTE": "ACK_RFQ_QUOTE", "RFQ_QUOTE_CANCEL": "ACK_RFQ_QUOTE_CANCEL", "RFQ_CONFIRMATION_RESPONSE": "ACK_RFQ_CONFIRMATION_RESPONSE"}[failure.RequestType]
	if kind == "" {
		return
	}
	key := ackKey{kind: kind, rfq: failure.RFQID}
	if kind != "ACK_RFQ_QUOTE" {
		key.quote = failure.QuoteID
	}
	if key.rfq == "" || (kind != "ACK_RFQ_QUOTE" && key.quote == "") {
		s.finish(&sdkerrors.UnexpectedResponseError{Message: "uncorrelated RFQ rejection"})
		return
	}
	s.resolve(key, acknowledgement{err: &failure})
}

// CancelQuote waits for the server's cancellation acknowledgement.
func (s *Session) CancelQuote(ctx context.Context, quote QuoteReference) (QuoteReference, error) {
	return s.request(ctx, ackKey{kind: "ACK_RFQ_QUOTE_CANCEL", rfq: quote.RFQID, quote: quote.QuoteID}, map[string]any{"type": "RFQ_QUOTE_CANCEL", "rfq_id": quote.RFQID, "quote_id": quote.QuoteID, "signer_address": s.orderSigner(), "maker_address": s.config.Wallet})
}

// RespondToConfirmation sends CONFIRM or DECLINE and waits for acknowledgement.
func (s *Session) RespondToConfirmation(ctx context.Context, quote QuoteReference, decision string) (QuoteReference, error) {
	if decision != "CONFIRM" && decision != "DECLINE" {
		return QuoteReference{}, &sdkerrors.UserInputError{Message: "decision must be CONFIRM or DECLINE"}
	}
	return s.request(ctx, ackKey{kind: "ACK_RFQ_CONFIRMATION_RESPONSE", rfq: quote.RFQID, quote: quote.QuoteID}, map[string]string{"type": "RFQ_CONFIRMATION_RESPONSE", "rfq_id": quote.RFQID, "quote_id": quote.QuoteID, "decision": decision})
}

// heartbeat uses standard WebSocket control frames, independent of RFQ JSON messages.
func (s *Session) heartbeat() {
	ticker := time.NewTicker(s.config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			bounded, cancel := context.WithTimeout(s.ctx, s.config.PongTimeout)
			err := s.conn.Ping(bounded)
			cancel()
			if err != nil {
				s.finish(&sdkerrors.TransportError{Cause: err})
				_ = s.conn.CloseNow()
				return
			}
		}
	}
}
