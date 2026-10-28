package streams

import (
	"context"
	"errors"
	"io"
	"iter"
	"sync"

	"github.com/cagelam/polymarket-go/models"
)

// Handle receives typed events. Its bounded queue drops the oldest event when full.
// Close discards pending events and is safe to call repeatedly or concurrently.
type Handle struct {
	mu          sync.Mutex
	queue       []models.StreamEvent
	head, count int
	dropped     uint64
	ended       bool
	failure     error
	done        chan struct{}
	wake        chan struct{}
	closeOnce   sync.Once
	onClose     func()
	stopContext func() bool
}

func newHandle(size int) *Handle {
	return &Handle{queue: make([]models.StreamEvent, size), done: make(chan struct{}), wake: make(chan struct{}, 1)}
}

// Next waits for an event, context cancellation, or io.EOF after closure.
func (h *Handle) Next(ctx context.Context) (models.StreamEvent, error) {
	for {
		h.mu.Lock()
		if h.count > 0 {
			event := h.queue[h.head]
			h.queue[h.head] = nil
			h.head = (h.head + 1) % len(h.queue)
			h.count--
			if h.count > 0 {
				select {
				case h.wake <- struct{}{}:
				default:
				}
			}
			h.mu.Unlock()
			return event, nil
		}
		ended := h.ended
		failure := h.failure
		h.mu.Unlock()
		if ended {
			if failure != nil {
				return nil, failure
			}
			return nil, io.EOF
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-h.done:
		case <-h.wake:
		}
	}
}

// Events iterates until closure, cancellation, or the caller stops. It does not close the handle.
func (h *Handle) Events(ctx context.Context) iter.Seq2[models.StreamEvent, error] {
	return func(yield func(models.StreamEvent, error) bool) {
		for {
			event, err := h.Next(ctx)
			if errors.Is(err, io.EOF) {
				return
			}
			if !yield(event, err) || err != nil {
				return
			}
		}
	}
}

// Dropped reports events lost because the consumer did not keep up.
func (h *Handle) Dropped() uint64 { h.mu.Lock(); defer h.mu.Unlock(); return h.dropped }

// Close unsubscribes and wakes waiting consumers.
func (h *Handle) Close() error {
	h.closeOnce.Do(func() {
		if h.onClose != nil {
			h.onClose()
		}
		h.end()
	})
	return nil
}
func (h *Handle) end() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.ended {
		return
	}
	h.ended = true
	clear(h.queue)
	h.count = 0
	if h.stopContext != nil {
		h.stopContext()
	}
	close(h.done)
}
func (h *Handle) push(event models.StreamEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.ended {
		return
	}
	if h.count == len(h.queue) {
		h.queue[h.head] = nil
		h.head = (h.head + 1) % len(h.queue)
		h.count--
		h.dropped++
	}
	h.queue[(h.head+h.count)%len(h.queue)] = event
	h.count++
	select {
	case h.wake <- struct{}{}:
	default:
	}
}
func (h *Handle) bindContext(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.ended {
		h.stopContext = context.AfterFunc(ctx, func() { _ = h.Close() })
	}
}

func (h *Handle) fail(err error) { h.mu.Lock(); h.failure = err; h.mu.Unlock(); h.end() }
