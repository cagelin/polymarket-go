package streams

import (
	"testing"

	"github.com/cagelin/polymarket-go/models"
)

// BenchmarkQueueDelivery measures the bounded queue without network latency.
func BenchmarkQueueDelivery(b *testing.B) {
	handle := newHandle(1024)
	event := &models.MarketBookEvent{Type: "book"}
	ctx := b.Context()
	b.ReportAllocs()
	for b.Loop() {
		handle.push(event)
		got, err := handle.Next(ctx)
		if err != nil || got != event {
			b.Fatalf("queue delivered %v, %v; want original event", got, err)
		}
	}
}

// BenchmarkQueueSlowConsumer measures eviction after the queue reaches capacity.
func BenchmarkQueueSlowConsumer(b *testing.B) {
	handle := newHandle(1024)
	event := &models.MarketBookEvent{Type: "book"}
	for range 1024 {
		handle.push(event)
	}
	b.ReportAllocs()
	var iterations uint64
	for b.Loop() {
		handle.push(event)
		iterations++
	}
	if handle.Dropped() != iterations {
		b.Fatalf("Dropped() = %d, want %d", handle.Dropped(), b.N)
	}
}
