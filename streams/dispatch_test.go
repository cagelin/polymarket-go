package streams

import (
	"testing"

	"github.com/cagelin/polymarket-go/models"
)

func TestDispatchIndexPreservesOverlappingSelectionsAndClose(t *testing.T) {
	client := &Client{handles: make(map[*Handle][]selection)}
	handle := newHandle(8)
	client.handles[handle] = []selection{
		{service: "market", topic: "market", assets: []string{"123", "456"}},
		{service: "market", topic: "market", assets: []string{"123"}, custom: true},
	}
	client.order = []*Handle{handle}
	client.rebuildDispatch()
	// A single selection matching multiple asset changes must receive only one copy.
	event := &models.MarketPriceChangeEvent{Topic: "market", Type: "price_change"}
	event.Payload.PriceChanges = append(event.Payload.PriceChanges, models.PriceChange{AssetID: "123"}, models.PriceChange{AssetID: "456"})
	client.dispatchEvent("market", event)
	handle.mu.Lock()
	count := handle.count
	handle.mu.Unlock()
	if count != 2 {
		t.Errorf("dispatchEvent(overlapping filters) count = %d, want 2 selections", count)
	}
	handle.end()
	client.dispatchEvent("market", event)
	if handle.count != 0 {
		t.Errorf("dispatchEvent(closed handle) count = %d, want 0", handle.count)
	}
}
func TestUnsupportedEventHasSeparateCounter(t *testing.T) {
	client := &Client{}
	client.deliver("market", []byte(`{"event_type":"future_event"}`))
	client.deliver("market", []byte(`not json`))
	stats := client.StreamStats()
	if stats.UnsupportedMessages != 1 || stats.MalformedMessages != 1 {
		t.Errorf("StreamStats() = %+v, want one unsupported and one malformed", stats)
	}
}
