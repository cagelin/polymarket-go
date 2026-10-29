package streams

import (
	"fmt"
	"strconv"
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

// BenchmarkDispatchEvent compares indexed delivery with the previous linear scan.
// Selective cases match one subscriber; fanout cases match every subscriber.
func BenchmarkDispatchEvent(b *testing.B) {
	for _, subscriptionCount := range []int{1, 8, 64} {
		for _, distribution := range []string{"selective", "fanout"} {
			b.Run(fmt.Sprintf("%d/%s", subscriptionCount, distribution), func(bench *testing.B) {
				for _, implementation := range []string{"indexed", "linear_reference"} {
					bench.Run(implementation, func(sample *testing.B) {
						client := &Client{handles: make(map[*Handle][]selection)}
						for ordinal := range subscriptionCount {
							asset := strconv.Itoa(ordinal)
							if distribution == "fanout" {
								asset = "0"
							}
							listenerHandle := newHandle(64)
							client.order = append(client.order, listenerHandle)
							client.handles[listenerHandle] = []selection{{service: "market", topic: "market", assets: []string{asset}}}
						}
						client.rebuildDispatch()
						event := &models.MarketBookEvent{Topic: "market", Type: "book", Payload: models.MarketBookPayload{AssetID: "0"}}
						sample.ReportAllocs()
						for sample.Loop() {
							if implementation == "indexed" {
								client.dispatchEvent("market", event)
							} else {
								client.mu.Lock()
								for _, queuedHandle := range client.order {
									for _, chosen := range client.handles[queuedHandle] {
										if chosen.service == "market" && chosen.matches(event) {
											queuedHandle.push(event)
										}
									}
								}
								client.mu.Unlock()
							}
						}
					})
				}
			})
		}
	}
}
