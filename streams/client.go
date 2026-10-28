package streams

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cagelam/polymarket-go/models"
	"github.com/coder/websocket"
)

// Config controls public stream endpoints. Zero values use production defaults.
type Config struct {
	MarketURL  string
	SportsURL  string
	RTDSURL    string
	HTTPClient *http.Client
	QueueSize  int // Defaults to 1024 per handle, including merged subscriptions.
}

// Client multiplexes subscriptions onto one socket per public service.
type Client struct {
	mu         sync.Mutex
	config     Config
	ctx        context.Context
	cancel     context.CancelFunc
	groups     map[string]*group
	handles    map[*Handle][]selection
	order      []*Handle
	closed     bool
	workers    sync.WaitGroup
	malformed  atomic.Uint64
	reconnects atomic.Uint64
}
type group struct {
	service, url                  string
	conn                          *websocket.Conn
	running                       bool
	heartbeatInterval, staleAfter time.Duration
}

// Stats reports aggregate stream diagnostics.
type Stats struct {
	MalformedMessages uint64
	Reconnects        uint64
}

// New creates a public stream client without opening sockets.
func New(config Config) (*Client, error) {
	if config.QueueSize == 0 {
		config.QueueSize = 1024
	}
	if config.QueueSize < 1 {
		return nil, invalid("stream queue size must be positive")
	}
	urls := map[string]string{"market": config.MarketURL, "sports": config.SportsURL, "rtds": config.RTDSURL}
	defaults := map[string]string{"market": "wss://ws-subscriptions-clob.polymarket.com/ws/market", "sports": "wss://sports-api.polymarket.com/ws", "rtds": "wss://ws-live-data.polymarket.com"}
	groups := make(map[string]*group, len(urls))
	for service, address := range urls {
		if address == "" {
			address = defaults[service]
		}
		parsed, err := url.Parse(address)
		if err != nil || (parsed.Scheme != "ws" && parsed.Scheme != "wss") || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
			return nil, invalid("stream URL must be absolute WS(S) without credentials or fragment")
		}
		interval, stale := 10*time.Second, 30*time.Second
		if service == "rtds" {
			interval, stale = 5*time.Second, 10*time.Minute
		}
		if service == "sports" {
			interval = time.Second
		}
		groups[service] = &group{service: service, url: address, heartbeatInterval: interval, staleAfter: stale}
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{config: config, ctx: ctx, cancel: cancel, groups: groups, handles: make(map[*Handle][]selection)}, nil
}

// Subscribe validates all selections, connects, and returns a merged event handle.
// The context controls the subscription lifetime. Public protocols do not acknowledge acceptance.
func (c *Client) Subscribe(ctx context.Context, specs ...Spec) (*Handle, error) {
	selections, err := normalizeSpecs(specs)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, invalid("client is closed")
	}
	before := map[string]serverState{}
	for _, s := range selections {
		if _, ok := before[s.service]; !ok {
			before[s.service] = deriveState(c.selections(s.service))
		}
	}
	h := newHandle(c.config.QueueSize)
	h.onClose = func() { c.remove(h) }
	c.handles[h] = selections
	c.order = append(c.order, h)
	for service, state := range before {
		g := c.groups[service]
		after := deriveState(c.selections(service))
		frames := diffFrames(service, state, after)
		if g.conn == nil {
			g.conn, err = c.dial(ctx, g.url)
			if err == nil {
				frames = initialFrames(service, after)
			}
		}
		if err == nil {
			err = sendFrames(ctx, g.conn, frames)
		}
		if err != nil {
			c.removeLocked(h)
			h.end()
			return nil, fmt.Errorf("subscribe %s: %w", service, err)
		}
		if !g.running {
			g.running = true
			c.workers.Add(1)
			//nolint:contextcheck // A shared socket follows the client lifetime, not one subscriber.
			go c.run(g)
		}
	}
	h.bindContext(ctx)
	return h, nil
}

// StreamStats returns counts of malformed messages and successful reconnections.
func (c *Client) StreamStats() Stats {
	return Stats{MalformedMessages: c.malformed.Load(), Reconnects: c.reconnects.Load()}
}

// Close closes sockets, releases workers, and ends every handle.
func (c *Client) Close() error {
	c.cancel()
	c.mu.Lock()
	c.closed = true
	for _, g := range c.groups {
		if g.conn != nil {
			_ = g.conn.CloseNow()
			g.conn = nil
		}
	}
	for h := range c.handles {
		h.end()
	}
	clear(c.handles)
	c.order = nil
	c.mu.Unlock()
	c.workers.Wait()
	return nil
}
func (c *Client) selections(service string) []selection {
	var result []selection
	for _, h := range c.order {
		for _, s := range c.handles[h] {
			if s.service == service {
				result = append(result, s)
			}
		}
	}
	return result
}
func (c *Client) remove(h *Handle) { c.mu.Lock(); defer c.mu.Unlock(); c.removeLocked(h) }
func (c *Client) removeLocked(h *Handle) {
	selections, exists := c.handles[h]
	if !exists {
		return
	}
	before := map[string]serverState{}
	for _, s := range selections {
		before[s.service] = deriveState(c.selections(s.service))
	}
	delete(c.handles, h)
	for i, item := range c.order {
		if item == h {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	for service, state := range before {
		g := c.groups[service]
		if g.conn == nil {
			continue
		}
		remaining := c.selections(service)
		if len(remaining) == 0 {
			_ = g.conn.CloseNow()
			g.conn = nil
			continue
		}
		if err := sendFrames(c.ctx, g.conn, diffFrames(service, state, deriveState(remaining))); err != nil {
			_ = g.conn.CloseNow()
			g.conn = nil
		}
	}
}
func (c *Client) dial(ctx context.Context, address string) (*websocket.Conn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	//nolint:contextcheck // Either caller cancellation or whole-client closure must interrupt dialing.
	stop := context.AfterFunc(c.ctx, cancel)
	defer stop()
	conn, response, err := websocket.Dial(dialCtx, address, &websocket.DialOptions{HTTPClient: c.config.HTTPClient})
	if err != nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		return nil, err
	}
	conn.SetReadLimit(10 << 20)
	return conn, nil
}
func sendFrames(ctx context.Context, conn *websocket.Conn, frames []any) error {
	for _, frame := range frames {
		raw, err := json.Marshal(frame)
		if err != nil {
			return err
		}
		if err = sendText(ctx, conn, raw); err != nil {
			return err
		}
	}
	return nil
}
func sendText(ctx context.Context, conn *websocket.Conn, raw []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return conn.Write(writeCtx, websocket.MessageText, raw)
}
func (c *Client) run(g *group) {
	defer c.workers.Done()
	attempt := 0
	for {
		c.mu.Lock()
		if c.closed || len(c.selections(g.service)) == 0 {
			g.running = false
			c.mu.Unlock()
			return
		}
		conn := g.conn
		if conn == nil {
			var err error
			conn, err = c.dial(c.ctx, g.url)
			if err == nil {
				err = sendFrames(c.ctx, conn, initialFrames(g.service, deriveState(c.selections(g.service))))
			}
			if err != nil {
				if conn != nil {
					_ = conn.CloseNow()
				}
				c.mu.Unlock()
				if !c.backoff(attempt) {
					c.finishGroup(g)
					return
				}
				attempt++
				continue
			}
			g.conn = conn
			c.reconnects.Add(1)
		}
		c.mu.Unlock()
		attempt = 0
		c.readConnection(g, conn)
		_ = conn.CloseNow()
		c.mu.Lock()
		if g.conn == conn {
			g.conn = nil
		}
		c.mu.Unlock()
		if !c.backoff(attempt) {
			c.finishGroup(g)
			return
		}
		attempt++
	}
}
func (c *Client) finishGroup(g *group) { c.mu.Lock(); g.running = false; c.mu.Unlock() }
func (c *Client) backoff(attempt int) bool {
	capDelay := 250 * time.Millisecond
	for i := 0; i < attempt && capDelay < 30*time.Second; i++ {
		capDelay *= 2
	}
	if capDelay > 30*time.Second {
		capDelay = 30 * time.Second
	}
	draw, err := rand.Int(rand.Reader, big.NewInt(int64(capDelay)))
	delay := capDelay
	if err == nil {
		delay = time.Duration(draw.Int64())
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-c.ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
func (c *Client) readConnection(g *group, conn *websocket.Conn) {
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()
	var alive atomic.Int64
	alive.Store(time.Now().UnixNano())
	heartbeatDone := make(chan struct{})
	go func() { defer close(heartbeatDone); c.heartbeat(ctx, g, conn, &alive) }()
	defer func() { cancel(); <-heartbeatDone }()
	for {
		kind, raw, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if kind != websocket.MessageText {
			continue
		}
		text := string(raw)
		switch g.service {
		case "market":
			if text == "PONG" {
				alive.Store(time.Now().UnixNano())
				continue
			}
		case "sports":
			if text == "ping" {
				alive.Store(time.Now().UnixNano())
				if err = sendText(ctx, conn, []byte("pong")); err != nil {
					return
				}
				continue
			}
		case "rtds":
			alive.Store(time.Now().UnixNano())
		}
		c.deliver(g.service, raw)
	}
}
func (c *Client) heartbeat(ctx context.Context, g *group, conn *websocket.Conn, alive *atomic.Int64) {
	interval, stale, service := g.heartbeatInterval, g.staleAfter, g.service
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if now.Sub(time.Unix(0, alive.Load())) > stale {
				_ = conn.CloseNow()
				return
			}
			if service != "sports" {
				if err := sendText(ctx, conn, []byte("PING")); err != nil {
					_ = conn.CloseNow()
					return
				}
			}
		}
	}
}
func (c *Client) deliver(service string, raw []byte) {
	items := []json.RawMessage{raw}
	if service == "market" && len(raw) > 0 && raw[0] == '[' {
		if err := json.Unmarshal(raw, &items); err != nil {
			c.malformed.Add(1)
			return
		}
	}
	for _, item := range items {
		event, err := models.ParseStreamEvent(service, item)
		if err != nil {
			c.malformed.Add(1)
			continue
		}
		c.mu.Lock()
		for _, h := range c.order {
			for _, s := range c.handles[h] {
				if s.service == service && s.matches(event) {
					h.push(event)
				}
			}
		}
		c.mu.Unlock()
	}
}
