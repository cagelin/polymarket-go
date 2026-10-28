package streams

import (
	"context"
	"errors"
	"io"
	"slices"
	"sync"
	"time"
)

type priceManager struct {
	mu       sync.Mutex
	sessions []*priceSession
	handles  map[*Handle][]*priceSession
	closed   bool
}

func (c *Client) subscribePrices(ctx context.Context, values []selection) (*Handle, error) {
	if c.config.Credentials == nil {
		return nil, invalid("price subscriptions require authentication")
	}
	ctx, stop := context.WithTimeout(ctx, 30*time.Second)
	defer stop()
	manager := &c.prices
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.closed {
		return nil, invalid("price client is closed")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if manager.handles == nil {
		manager.handles = map[*Handle][]*priceSession{}
	}
	h := newHandle(c.config.QueueSize)
	manager.handles[h] = nil
	h.onClose = func() { c.releasePrices(h) }
	pending := map[*priceSession][]priceKey{}
	created := map[*priceSession]bool{}
	for _, value := range values {
		for _, symbol := range value.symbols {
			key := priceKey{topic: value.topic, symbol: symbol, provider: value.provider}
			var chosen *priceSession
			for _, session := range manager.sessions {
				session.mu.Lock()
				_, exists := session.keys[key]
				available := !session.closed && session.provider == key.provider
				session.mu.Unlock()
				if available && exists {
					chosen = session
					break
				}
			}
			if chosen == nil {
				for _, session := range manager.sessions {
					session.mu.Lock()
					target := session.keyTarget
					if time.Since(session.lastDrop) >= time.Minute {
						target = 64
					}
					available := !session.closed && session.provider == key.provider && len(session.keys) < target
					session.mu.Unlock()
					if available {
						chosen = session
						break
					}
				}
			}
			if chosen == nil {
				chosen = newPriceSession(c, key.provider)
				manager.sessions = append(manager.sessions, chosen)
				created[chosen] = true
			}
			if !slices.Contains(manager.handles[h], chosen) {
				manager.handles[h] = append(manager.handles[h], chosen)
			}
			chosen.mu.Lock()
			state := chosen.keys[key]
			if state == nil {
				state = &priceState{ready: make(chan struct{})}
				chosen.keys[key] = state
				pending[chosen] = append(pending[chosen], key)
			}
			state.listeners = append(state.listeners, priceListener{handle: h, types: value.types})
			if state.snapshot != nil && (len(value.types) == 0 || slices.Contains(value.types, "subscribe")) {
				h.push(state.snapshot)
			}
			chosen.mu.Unlock()
		}
	}
	for session, keys := range pending {
		session.operation.Lock()
		var err error
		if created[session] {
			err = session.openInitial(ctx)
		}
		session.mu.Lock()
		connected := session.conn != nil
		session.mu.Unlock()
		if err == nil && connected {
			err = session.change(ctx, "subscribe", keys)
		}
		if err == nil && connected {
			session.accept(keys, nil)
			session.mu.Lock()
			if created[session] {
				session.workers.Add(1)
			}
			session.mu.Unlock()
		}
		session.operation.Unlock()
		if err != nil {
			// Roll back every registration for this merged handle, including successful batches.
			c.releasePricesLocked(h)
			h.end()
			return nil, err
		}
		if created[session] {
			go session.recover()
		}
	}
	for _, session := range manager.handles[h] {
		session.mu.Lock()
		var waits []<-chan struct{}
		for _, state := range session.keys {
			for _, listener := range state.listeners {
				if listener.handle == h && !state.accepted {
					waits = append(waits, state.ready)
					break
				}
			}
		}
		session.mu.Unlock()
		for _, ready := range waits {
			select {
			case <-ready:
			case <-ctx.Done():
				c.releasePricesLocked(h)
				h.end()
				return nil, ctx.Err()
			case <-h.done:
				h.mu.Lock()
				failure := h.failure
				h.mu.Unlock()
				c.releasePricesLocked(h)
				return nil, failure
			case <-session.ctx.Done():
				c.releasePricesLocked(h)
				h.end()
				return nil, invalid("price session ended before acceptance")
			}
		}
	}
	return h, nil
}
func (c *Client) releasePrices(h *Handle) {
	c.prices.mu.Lock()
	defer c.prices.mu.Unlock()
	c.releasePricesLocked(h)
}
func (c *Client) releasePricesLocked(h *Handle) {
	sessions := c.prices.handles[h]
	delete(c.prices.handles, h)
	for _, session := range sessions {
		session.operation.Lock()
		session.mu.Lock()
		var removed []priceKey
		for key, state := range session.keys {
			state.listeners = slices.DeleteFunc(state.listeners, func(value priceListener) bool { return value.handle == h })
			if len(state.listeners) == 0 {
				removed = append(removed, key)
				delete(session.keys, key)
			}
		}
		empty := len(session.keys) == 0
		connected := session.conn != nil
		closed := session.closed
		session.mu.Unlock()
		if empty {
			session.cancel()
			session.mu.Lock()
			session.closed = true
			if session.conn != nil {
				_ = session.conn.CloseNow()
			}
			session.mu.Unlock()
		} else if connected && !closed && len(removed) > 0 {
			if err := session.change(session.ctx, "unsubscribe", removed); err != nil {
				session.mu.Lock()
				if session.conn != nil {
					_ = session.conn.CloseNow()
				}
				session.mu.Unlock()
			}
		}
		session.operation.Unlock()
	}
}
func (c *Client) closePrices() {
	c.prices.mu.Lock()
	c.prices.closed = true
	sessions := slices.Clone(c.prices.sessions)
	for h := range c.prices.handles {
		h.end()
	}
	clear(c.prices.handles)
	c.prices.mu.Unlock()
	for _, session := range sessions {
		session.close()
	}
}

// subscribeMixed keeps one bounded public handle while retaining each protocol's acceptance rules.
func (c *Client) subscribeMixed(ctx context.Context, specs []Spec, values []selection) (*Handle, error) {
	var public []Spec
	var prices []selection
	for i, value := range values {
		if value.service == "price" {
			prices = append(prices, value)
		} else {
			public = append(public, specs[i])
		}
	}
	var children []*Handle
	if len(public) > 0 {
		child, err := c.Subscribe(ctx, public...)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	child, err := c.subscribePrices(ctx, prices)
	if err != nil {
		for _, handle := range children {
			_ = handle.Close()
		}
		return nil, err
	}
	children = append(children, child)
	merged := newHandle(c.config.QueueSize)
	merged.onClose = func() {
		for _, handle := range children {
			_ = handle.Close()
		}
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		_ = merged.Close()
		return nil, invalid("stream client is closed")
	}
	if c.merged == nil {
		c.merged = map[*Handle]struct{}{}
	}
	c.merged[merged] = struct{}{}
	c.workers.Add(len(children))
	c.mu.Unlock()
	merged.bindContext(ctx)
	for _, handle := range children {
		go func() {
			defer c.workers.Done()
			for {
				event, nextErr := handle.Next(c.ctx)
				if nextErr != nil {
					if !errors.Is(nextErr, io.EOF) && !errors.Is(nextErr, context.Canceled) {
						merged.fail(nextErr)
					}
					_ = merged.Close()
					c.mu.Lock()
					delete(c.merged, merged)
					c.mu.Unlock()
					return
				}
				merged.push(event)
			}
		}()
	}
	return merged, nil
}
