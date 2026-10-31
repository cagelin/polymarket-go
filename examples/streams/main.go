// Command streams prints public crypto, equity, sports or comment events as JSON Lines.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
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
	feed := flag.String("feed", "binance", "public feed: Binance, Chainlink, TWAP, equity, sports or comments")
	flag.Parse()
	if *duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	spec, err := streamSpec(*feed, flag.Args())
	if err != nil {
		return err
	}
	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(signalContext, *duration)
	defer cancel()
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{StreamQueueSize: 256})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	handle, err := client.Subscribe(ctx, spec)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("subscribe to %s feed: %w", *feed, err)
	}
	defer func() {
		_ = handle.Close() // Close unsubscribes and always returns nil.
		log.Printf("typed queue dropped=%d stats=%+v", handle.Dropped(), client.StreamStats())
	}()
	encoder := json.NewEncoder(os.Stdout)
	for event, readErr := range handle.Events(ctx) {
		if ctx.Err() != nil {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read stream event: %w", readErr)
		}
		if err = encoder.Encode(event); err != nil {
			return fmt.Errorf("write stream event: %w", err)
		}
	}
	return nil
}

func streamSpec(feed string, arguments []string) (polymarket.Spec, error) {
	switch feed {
	case "binance":
		if len(arguments) == 0 {
			arguments = []string{"btcusdt"}
		}
		return polymarket.CryptoPricesSpec{Topic: "prices.crypto.binance", Symbols: arguments}, nil
	case "chainlink", "twap":
		if len(arguments) == 0 {
			arguments = []string{"btc/usd"}
		}
		if feed == "twap" {
			return polymarket.CryptoPricesChainlinkTwapSpec{WindowSeconds: 30, Symbols: arguments}, nil
		}
		return polymarket.CryptoPricesSpec{Topic: "prices.crypto.chainlink", Symbols: arguments}, nil
	case "equity":
		if len(arguments) != 1 {
			return nil, fmt.Errorf("equity feed requires one symbol")
		}
		return polymarket.EquityPricesSpec{Symbol: arguments[0]}, nil
	case "sports":
		if len(arguments) != 0 {
			return nil, fmt.Errorf("sports feed takes no positional arguments")
		}
		return polymarket.SportsSpec{}, nil
	case "comments":
		if len(arguments) != 1 {
			return nil, fmt.Errorf("comments feed requires one numeric event ID")
		}
		eventID, parseErr := strconv.ParseInt(arguments[0], 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("parse comment event ID: %w", parseErr)
		}
		if eventID <= 0 {
			return nil, fmt.Errorf("comment event ID must be positive")
		}
		parentType := "Event"
		return polymarket.CommentsSpec{ParentEntityID: &eventID, ParentEntityType: &parentType}, nil
	default:
		return nil, fmt.Errorf("unknown public feed %q", feed)
	}
}
