package streams

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/models"
	"github.com/coder/websocket"
)

type priceKey struct{ topic, symbol, provider string }
type priceListener struct {
	handle *Handle
	types  []string
}
type priceState struct {
	listeners []priceListener
	snapshot  models.StreamEvent
	accepted  bool
	ready     chan struct{}
}
type priceReply struct{ err error }

// Per-filter failures must not terminate unrelated subscriptions sharing a socket.
type priceRejections map[priceKey]error

func (r priceRejections) Error() string { return "price filters rejected" }
func (r priceRejections) Unwrap() []error {
	result := make([]error, 0, len(r))
	keys := make([]priceKey, 0, len(r))
	for key := range r {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(first, second priceKey) int {
		if order := strings.Compare(first.topic, second.topic); order != 0 {
			return order
		}
		if order := strings.Compare(first.symbol, second.symbol); order != 0 {
			return order
		}
		return strings.Compare(first.provider, second.provider)
	})
	for _, key := range keys {
		result = append(result, r[key])
	}
	return result
}

type pricePending struct {
	rid, operation string
	channels       []string
	keys           []priceKey
	rejected       priceRejections
	done           chan priceReply
	err            error
}
type priceSession struct {
	client      *Client
	provider    string
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	operation   sync.Mutex
	conn        *websocket.Conn
	keys        map[priceKey]*priceState
	pending     *pricePending
	rid         uint64
	lastSent    time.Time
	closed      bool
	lost        chan error
	workers     sync.WaitGroup
	lastMessage atomic.Int64
	keyTarget   int
	lastDrop    time.Time
}

func newPriceSession(c *Client, provider string) *priceSession {
	ctx, cancel := context.WithCancel(c.ctx)
	return &priceSession{client: c, provider: provider, ctx: ctx, cancel: cancel, keys: map[priceKey]*priceState{}, lost: make(chan error, 1), keyTarget: 64}
}
func (key priceKey) wire() map[string]any {
	filter := map[string]any{"symbol": key.symbol}
	if key.provider != "" {
		filter["provider"] = key.provider
	}
	if strings.HasSuffix(key.topic, ".twap") {
		filter["window_seconds"] = 60
	}
	return map[string]any{"channel": wireChannel(key.topic), "filter": filter}
}
func (s *priceSession) open(ctx context.Context) error {
	conn, err := s.client.dial(ctx, s.client.config.PriceURL)
	if err != nil {
		return err
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		_ = conn.CloseNow()
		return invalid("price subscription is closed")
	}
	s.conn = conn
	s.lastMessage.Store(time.Now().UnixNano())
	s.workers.Add(1)
	s.mu.Unlock()
	//nolint:contextcheck // Shared socket follows the session lifetime, not one subscription operation.
	go s.read(conn)
	credentials := s.client.config.Credentials
	err = s.request(ctx, "auth", nil, map[string]any{"op": "auth", "auth": map[string]string{"apiKey": credentials.Key, "secret": credentials.Secret, "passphrase": credentials.Passphrase}})
	if err != nil {
		_ = conn.CloseNow()
	}
	return err
}
func (s *priceSession) request(ctx context.Context, operation string, keys []priceKey, frame map[string]any) error {
	// The operation mutex is held by the caller through authentication and changes.
	s.mu.Lock()
	conn := s.conn
	if conn == nil || s.closed {
		s.mu.Unlock()
		return invalid("price connection is not open")
	}
	s.rid++
	rid := strconv.FormatUint(s.rid, 10)
	expected := map[string]string{"auth": "authed", "subscribe": "subscribed", "unsubscribe": "unsubscribed"}[operation]
	pending := &pricePending{rid: rid, operation: expected, done: make(chan priceReply, 1), keys: keys, rejected: priceRejections{}}
	for _, key := range keys {
		pending.channels = append(pending.channels, wireChannel(key.topic))
	}
	s.pending = pending
	delay := time.Until(s.lastSent.Add(110 * time.Millisecond))
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		if s.pending == pending {
			s.pending = nil
		}
		s.mu.Unlock()
	}()
	if delay > 0 {
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-s.ctx.Done():
			timer.Stop()
			return s.ctx.Err()
		case <-timer.C:
		}
	}
	frame["rid"] = rid
	s.mu.Lock()
	s.lastSent = time.Now()
	s.mu.Unlock()
	if err := sendFrames(ctx, conn, []any{frame}); err != nil {
		return err
	}
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	select {
	case reply := <-pending.done:
		return reply.err
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return s.ctx.Err()
	case <-timer.C:
		return &sdkerrors.TransportError{Cause: fmt.Errorf("price acknowledgement timed out")}
	}
}
func (s *priceSession) change(ctx context.Context, operation string, keys []priceKey) error {
	for offset := 0; offset < len(keys); {
		var batch []any
		end := offset
		for end < len(keys) {
			candidate := keys[end].wire()
			trial := append(slices.Clone(batch), candidate)
			raw, err := json.Marshal(map[string]any{"op": operation, "subscriptions": trial, "rid": "18446744073709551615"})
			if err != nil {
				return err
			}
			if len(batch) > 0 && len(raw) > 60000 {
				break
			}
			batch = trial
			end++
		}
		if err := s.request(ctx, operation, keys[offset:end], map[string]any{"op": operation, "subscriptions": batch}); err != nil {
			return err
		}
		offset = end
	}
	return nil
}
func (s *priceSession) read(conn *websocket.Conn) {
	defer s.workers.Done()
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				if now.Sub(time.Unix(0, s.lastMessage.Load())) >= 90*time.Second {
					_ = conn.CloseNow()
					return
				}
				if sendText(ctx, conn, []byte(`{"op":"ping"}`)) != nil {
					_ = conn.CloseNow()
					return
				}
			}
		}
	}()
	defer func() { cancel(); <-heartbeatDone }()
	for {
		kind, raw, err := conn.Read(ctx)
		if err != nil {
			s.mu.Lock()
			if s.conn == conn {
				s.conn = nil
				for _, state := range s.keys {
					if state.accepted {
						state.ready = make(chan struct{})
					}
					state.accepted = false
					state.snapshot = nil
				}
				if s.pending != nil {
					select {
					case s.pending.done <- priceReply{err: err}:
					default:
					}
				}
				select {
				case s.lost <- err:
				default:
				}
			}
			s.mu.Unlock()
			return
		}
		s.lastMessage.Store(time.Now().UnixNano())
		if kind != websocket.MessageText {
			continue
		}
		s.message(raw)
	}
}
func (s *priceSession) message(raw []byte) {
	var frame struct{ Op, Rid, Channel, Code string }
	if json.Unmarshal(raw, &frame) != nil {
		return
	}
	if slices.Contains([]string{"authed", "subscribed", "unsubscribed", "pong", "error"}, frame.Op) {
		s.mu.Lock()
		defer s.mu.Unlock()
		pending := s.pending
		if pending == nil || pending.rid != frame.Rid {
			return
		}
		switch frame.Op {
		case "error":
			if !slices.Contains([]string{"bad_op", "bad_channel", "bad_filter", "sub_limit", "rate_limited", "auth_required", "auth_unavailable", "auth_expired", "auth_attempts", "auth_invalid"}, frame.Code) {
				return
			}
			rejected := &sdkerrors.RequestRejectedError{Message: "price subscription rejected: " + frame.Code, Status: 200, Code: frame.Code}
			index := slices.Index(pending.channels, frame.Channel)
			if frame.Channel == "" || index < 0 {
				select {
				case pending.done <- priceReply{err: rejected}:
				default:
				}
				return
			}
			pending.channels[index] = ""
			pending.rejected[pending.keys[index]] = rejected
			pending.err = pending.rejected
		case pending.operation:
			if len(pending.channels) > 0 {
				index := slices.Index(pending.channels, frame.Channel)
				if frame.Channel == "" || index < 0 {
					return
				}
				pending.channels[index] = ""
			}
		default:
			return
		}
		for _, channel := range pending.channels {
			if channel != "" {
				return
			}
		}
		select {
		case pending.done <- priceReply{err: pending.err}:
		default:
		}
		return
	}
	event, err := models.ParseStreamEvent("price", raw)
	if err != nil {
		if errors.Is(err, models.ErrUnsupportedStreamEvent) {
			s.client.unsupported.Add(1)
			return
		}
		s.client.malformed.Add(1)
		return
	}
	details, err := priceDetails(event)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.keys[priceKey{topic: details.Topic, symbol: details.Payload.Symbol, provider: s.provider}]
	if state == nil {
		return
	}
	state.snapshot = snapshotFor(state.snapshot, event)
	if details.Dropped != nil && *details.Dropped > 0 && time.Since(s.lastDrop) >= 5*time.Second {
		target := s.keyTarget
		if time.Since(s.lastDrop) >= time.Minute {
			target = 64
		}
		s.keyTarget = max(1, target/2)
		s.lastDrop = time.Now()
	}
	for _, listener := range state.listeners {
		if len(listener.types) == 0 || slices.Contains(listener.types, event.StreamType()) {
			listener.handle.push(event)
		}
	}
}
func (s *priceSession) recover() {
	defer s.workers.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case lost := <-s.lost:
			s.mu.Lock()
			connected := s.conn != nil
			s.mu.Unlock()
			if connected {
				continue
			}
			code := websocket.CloseStatus(lost)
			if code == 4001 || code == 4008 {
				s.fail(lost)
				return
			}
			for attempt := 0; ; attempt++ {
				if !priceBackoff(s.ctx, int(code), attempt) {
					return
				}
				if s.ctx.Err() != nil {
					return
				}
				s.operation.Lock()
				s.mu.Lock()
				keys := make([]priceKey, 0, len(s.keys))
				for key := range s.keys {
					keys = append(keys, key)
				}
				s.mu.Unlock()
				if len(keys) == 0 {
					s.operation.Unlock()
					return
				}
				err := s.open(s.ctx)
				if err == nil {
					err = s.change(s.ctx, "subscribe", keys)
				}
				var rejectedFilters priceRejections
				if err == nil || errors.As(err, &rejectedFilters) {
					s.accept(keys, rejectedFilters)
					s.client.reconnects.Add(1)
					s.operation.Unlock()
					break
				}
				s.mu.Lock()
				conn := s.conn
				s.mu.Unlock()
				if conn != nil {
					_ = conn.CloseNow()
				}
				s.operation.Unlock()
				var rejected *sdkerrors.RequestRejectedError
				if errors.As(err, &rejected) && rejected.Code != "auth_unavailable" {
					s.fail(err)
					return
				}
			}
		}
	}
}
func (s *priceSession) fail(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.cancel()
	for _, state := range s.keys {
		for _, listener := range state.listeners {
			listener.handle.fail(fmt.Errorf("price stream: %w", err))
		}
	}
	if s.conn != nil {
		_ = s.conn.CloseNow()
	}
}
func (s *priceSession) close() {
	s.cancel()
	s.mu.Lock()
	s.closed = true
	if s.conn != nil {
		_ = s.conn.CloseNow()
	}
	s.mu.Unlock()
	s.workers.Wait()
}

func (s *priceSession) accept(keys []priceKey, rejected priceRejections) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return
	}
	for _, key := range keys {
		state := s.keys[key]
		if state == nil {
			continue
		}
		if err := rejected[key]; err != nil {
			for _, listener := range state.listeners {
				listener.handle.fail(err)
			}
			delete(s.keys, key)
		} else if !state.accepted {
			state.accepted = true
			close(state.ready)
		}
	}
}

func priceBackoff(ctx context.Context, code, attempt int) bool {
	maximum := min(1<<min(attempt, 5), 30)
	if code == 4003 {
		maximum = 10
	}
	//nolint:gosec // Retry jitter is not a security-sensitive random value.
	timer := time.NewTimer(time.Duration(rand.Float64() * float64(maximum) * float64(time.Second)))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (s *priceSession) openInitial(ctx context.Context) error {
	for attempt := 0; ; attempt++ {
		err := s.open(ctx)
		var rejected *sdkerrors.RequestRejectedError
		if !errors.As(err, &rejected) || rejected.Code != "auth_unavailable" {
			return err
		}
		if !priceBackoff(ctx, 1006, attempt) {
			return ctx.Err()
		}
	}
}

func wireChannel(topic string) string { return "price." + strings.TrimPrefix(topic, "prices.") }
