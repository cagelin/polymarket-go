// Command streams prints public Binance price updates for thirty seconds.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cagelin/polymarket-go"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err) // A stderr failure cannot recover this terminal error.
		os.Exit(1)
	}
}
func run() error {
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	handle, err := client.Subscribe(ctx, polymarket.CryptoPricesSpec{Topic: "prices.crypto.binance", Symbols: []string{"btcusdt"}})
	if err != nil {
		return err
	}
	defer func() { _ = handle.Close() }()
	encoder := json.NewEncoder(os.Stdout)
	for event, readErr := range handle.Events(ctx) {
		if ctx.Err() != nil {
			return nil
		}
		if readErr != nil {
			return readErr
		}
		if err = encoder.Encode(event); err != nil {
			return err
		}
	}
	return nil
}
