package streams

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// MarketBatchConfig applies the same policy to initial, additional and restored
// market subscriptions. Interval is the minimum time between batch sends.
// OnBatch reports completion or timeout. Return promptly; do not close the client
// from the callback. Disconnection abandons the current batch.
type MarketBatchConfig struct {
	Size        int
	Interval    time.Duration
	BookTimeout time.Duration
	OnBatch     func(MarketBatchResult)
}

// MarketBatchResult identifies the requested assets and those still missing a
// book at the deadline. Assets removed while waiting are excluded from Missing.
type MarketBatchResult struct {
	Generation uint64
	SentAt     time.Time
	Assets     []string
	Missing    []string
}

// Each socket owns its own pending subscriptions and book receipts.
type marketBatches struct {
	policy  *MarketBatchConfig
	mu      sync.Mutex
	changed chan struct{}
	pending []string
	sent    serverState
	batch   []string
	missing map[string]bool
	sentAt  time.Time
}

func (b *marketBatches) notify() {
	select {
	case b.changed <- struct{}{}:
	default:
	}
}

// send runs before the reader starts or with mu held by the worker.
func (b *marketBatches) send(ctx context.Context, conn *websocket.Conn, initial bool) error {
	size := min(b.policy.Size, len(b.pending))
	assets := b.pending[:size:size]
	target := serverState{assets: assets, custom: b.sent.custom}
	frames := initialFrames(serviceMarket, target)
	if !initial {
		frames = diffFrames(serviceMarket, serverState{}, target)
	}
	sentAt := time.Now()
	if err := sendFrames(ctx, conn, frames); err != nil {
		return err
	}
	b.sentAt = sentAt
	b.batch = assets
	b.pending = b.pending[size:]
	b.sent.assets = append(b.sent.assets, assets...)
	b.missing = make(map[string]bool, size)
	for _, asset := range assets {
		b.missing[asset] = true
	}
	return nil
}

// update removes sent assets immediately. New assets go through the same worker
// as the initial set, including when handles change during recovery.
func (b *marketBatches) update(ctx context.Context, conn *websocket.Conn, target serverState) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	sent := make(map[string]bool, len(b.sent.assets))
	for _, asset := range b.sent.assets {
		sent[asset] = true
	}
	var pending []string
	keep := serverState{custom: target.custom}
	for _, asset := range target.assets {
		if sent[asset] {
			keep.assets = append(keep.assets, asset)
		} else {
			pending = append(pending, asset)
		}
	}
	if err := sendFrames(ctx, conn, diffFrames(serviceMarket, b.sent, keep)); err != nil {
		_ = conn.CloseNow()
		return err
	}
	for asset := range b.missing {
		if _, ok := slices.BinarySearch(target.assets, asset); !ok {
			delete(b.missing, asset)
		}
	}
	b.sent, b.pending = keep, pending
	b.notify()
	return nil
}

type marketBookHeader struct {
	Kind  string `json:"event_type"`
	Asset string `json:"asset_id"`
}

func (b *marketBatches) observeBooks(raw []byte, received time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.missing) == 0 || received.Before(b.sentAt) || received.After(b.sentAt.Add(b.policy.BookTimeout)) {
		return
	}
	raw = bytes.TrimSpace(raw)
	var headers []marketBookHeader
	if len(raw) > 0 && raw[0] == '[' {
		if json.Unmarshal(raw, &headers) != nil {
			return
		}
	} else {
		var header marketBookHeader
		if json.Unmarshal(raw, &header) != nil {
			return
		}
		headers = []marketBookHeader{header}
	}
	for _, header := range headers {
		if header.Kind == "book" {
			delete(b.missing, header.Asset)
		}
	}
	if len(b.missing) == 0 {
		b.notify()
	}
}

// advance either finishes a batch, sends the next one, or returns the next
// deadline. A zero deadline waits for a subscription change.
func (b *marketBatches) advance(ctx context.Context, conn *websocket.Conn, generation uint64) (*MarketBatchResult, time.Time, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	if len(b.batch) > 0 {
		deadline := b.sentAt.Add(b.policy.BookTimeout)
		if len(b.missing) > 0 && now.Before(deadline) {
			return nil, deadline, nil
		}
		result := &MarketBatchResult{Generation: generation, SentAt: b.sentAt, Assets: b.batch}
		for _, asset := range b.batch {
			if b.missing[asset] {
				result.Missing = append(result.Missing, asset)
			}
		}
		b.batch, b.missing = nil, nil
		return result, time.Time{}, nil
	}
	if len(b.pending) == 0 {
		return nil, time.Time{}, nil
	}
	next := b.sentAt.Add(b.policy.Interval)
	if now.Before(next) {
		return nil, next, nil
	}
	err := b.send(ctx, conn, false)
	return nil, b.sentAt.Add(b.policy.BookTimeout), err
}

func (b *marketBatches) run(ctx context.Context, conn *websocket.Conn, generation uint64) {
	for ctx.Err() == nil {
		result, deadline, err := b.advance(ctx, conn, generation)
		if err != nil {
			_ = conn.CloseNow()
			return
		}
		if result != nil {
			b.policy.OnBatch(*result)
			continue
		}
		var wake <-chan time.Time
		if !deadline.IsZero() {
			wake = time.After(time.Until(deadline))
		}
		select {
		case <-ctx.Done():
			return
		case <-b.changed:
		case <-wake:
		}
	}
}
