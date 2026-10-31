// Command raw-stream writes market frames and lifecycle observations as JSON Lines.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/cagelin/polymarket-go/streams"
)

// record keeps arbitrary frame bytes intact: encoding/json encodes Data as base64.
type record struct {
	Frame     *streams.RawFrame  `json:"frame,omitempty"`
	Lifecycle *streams.Lifecycle `json:"lifecycle,omitempty"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	duration := flag.Duration("duration", 30*time.Second, "subscription lifetime")
	flag.Parse()
	if flag.NArg() == 0 || *duration <= 0 {
		return fmt.Errorf("usage: go run ./examples/raw-stream [-duration 30s] <asset-id> [asset-id ...]")
	}
	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(signalContext, *duration)
	defer cancel()
	records := make(chan record, 64)
	var droppedFrames, droppedLifecycle atomic.Uint64
	client, err := streams.New(streams.Config{
		RawOnly:   true,
		ReadLimit: 1 << 20, // Bound each message to 1 MiB in this demonstration.
		RawHandler: func(frame streams.RawFrame) {
			// Retain the immutable buffer; never block the socket reader on output.
			select {
			case records <- record{Frame: &frame}:
			default:
				droppedFrames.Add(1)
			}
		},
		LifecycleHandler: func(observation streams.Lifecycle) {
			// Observations can arrive concurrently with frames and each other.
			select {
			case records <- record{Lifecycle: &observation}:
			default:
				droppedLifecycle.Add(1)
			}
		},
	})
	if err != nil {
		return fmt.Errorf("create raw stream client: %w", err)
	}
	defer func() {
		_ = client.Close() // Wait for callbacks outside the callbacks themselves.
		log.Printf("callback queue dropped: frames=%d lifecycle=%d; queued at exit=%d; stats=%+v",
			droppedFrames.Load(), droppedLifecycle.Load(), len(records), client.StreamStats())
	}()
	handle, err := client.Subscribe(ctx, streams.MarketSpec{AssetIDs: flag.Args()})
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("subscribe to raw market frames: %w", err)
	}
	defer func() { _ = handle.Close() }() // The handle owns the raw subscription lifetime.
	encoder := json.NewEncoder(os.Stdout)
	for {
		select {
		case <-ctx.Done():
			return nil // Queued records are deliberately not drained on exit.
		case nextRecord := <-records:
			if encodeErr := encoder.Encode(nextRecord); encodeErr != nil {
				return fmt.Errorf("write raw stream record: %w", encodeErr)
			}
		}
	}
}
