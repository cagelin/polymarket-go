// Command market-stream consumes typed public order-book and best-bid/ask events.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/cagelin/polymarket-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	duration := flag.Duration("duration", 30*time.Second, "subscription lifetime")
	flag.Parse()
	if flag.NArg() == 0 || *duration <= 0 {
		return fmt.Errorf("usage: go run ./examples/market-stream [-duration 30s] <asset-id> [asset-id ...]")
	}
	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(signalContext, *duration)
	defer cancel()
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{StreamQueueSize: 256})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	handle, err := client.Subscribe(ctx, polymarket.MarketSpec{AssetIDs: flag.Args(), CustomFeatureEnabled: true})
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("subscribe to market events: %w", err)
	}
	defer func() {
		_ = handle.Close() // Close unsubscribes and always returns nil.
		log.Printf("typed queue dropped=%d stats=%+v", handle.Dropped(), client.StreamStats())
	}()
	for event, readErr := range handle.Events(ctx) {
		if ctx.Err() != nil {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read market event: %w", readErr)
		}
		switch message := event.(type) {
		case *polymarket.MarketBookEvent:
			fmt.Printf("book asset=%s bids=%d asks=%d\n", message.Payload.AssetID, len(message.Payload.Bids), len(message.Payload.Asks))
		case *polymarket.MarketBestBidAskEvent:
			fmt.Printf("quote asset=%s bid=%s ask=%s spread=%s\n", message.Payload.AssetID, message.Payload.BestBid, message.Payload.BestAsk, message.Payload.Spread)
		default:
			fmt.Printf("topic=%s type=%s\n", event.StreamTopic(), event.StreamType())
		}
	}
	return nil
}
