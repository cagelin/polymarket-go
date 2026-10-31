// Command portfolio reads a wallet's public profile, valuation, statistics and PnL.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cagelin/polymarket-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	interval := flag.String("interval", "1w", "PnL interval: max, all, 1m, 1w, 1d, 12h or 6h")
	fidelity := flag.String("fidelity", "1d", "PnL bucket size: 1d, 18h, 12h, 3h or 1h")
	flag.Parse()
	if flag.NArg() != 1 {
		return fmt.Errorf("usage: go run ./examples/portfolio [-interval 1w] [-fidelity 1d] <wallet-address>")
	}
	wallet := flag.Arg(0)
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pnl, err := client.GetUserPnL(ctx, polymarket.GetUserPnLParams{User: wallet, Interval: interval, Fidelity: fidelity})
	if err != nil {
		return fmt.Errorf("read wallet PnL: %w", err)
	}
	profile, err := client.GetPublicProfile(ctx, polymarket.EVMAddress(wallet))
	if err != nil {
		return fmt.Errorf("read public profile: %w", err)
	}
	value, err := client.GetPortfolioValue(ctx, polymarket.GetPortfolioValueParams{User: wallet})
	if err != nil {
		return fmt.Errorf("read portfolio value: %w", err)
	}
	stats, err := client.GetUserStats(ctx, polymarket.GetUserStatsParams{User: wallet})
	if err != nil {
		return fmt.Errorf("read wallet statistics: %w", err)
	}
	volume, err := client.GetUserVolume(ctx, polymarket.GetUserVolumeParams{User: wallet})
	if err != nil {
		return fmt.Errorf("read wallet volume: %w", err)
	}
	// Missing profiles/statistics remain null; independent reads need not share a timestamp.
	if err = json.NewEncoder(os.Stdout).Encode(map[string]any{
		"profile": profile, "value": value, "stats": stats, "pnl": pnl, "volume": volume,
	}); err != nil {
		return fmt.Errorf("write portfolio: %w", err)
	}
	return nil
}
