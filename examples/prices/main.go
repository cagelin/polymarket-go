// Command prices reads batch quotes, midpoints, spreads and last trades.
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
	flag.Parse()
	if flag.NArg() == 0 {
		return fmt.Errorf("usage: go run ./examples/prices <asset-id> [asset-id ...]")
	}
	assets := make([]polymarket.ClobAssetID, 0, flag.NArg())
	requests := make([]polymarket.PriceRequest, 0, 2*flag.NArg())
	for _, argument := range flag.Args() {
		assetID := polymarket.ClobAssetID(argument)
		assets = append(assets, assetID)
		requests = append(requests,
			polymarket.PriceRequest{AssetQuery: polymarket.Asset(assetID), Side: polymarket.Buy},
			polymarket.PriceRequest{AssetQuery: polymarket.Asset(assetID), Side: polymarket.Sell})
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	selection := polymarket.AssetsQuery{AssetIDs: assets}
	prices, err := client.GetPrices(ctx, requests)
	if err != nil {
		return fmt.Errorf("read batch prices: %w", err)
	}
	midpoints, err := client.GetMidpoints(ctx, selection)
	if err != nil {
		return fmt.Errorf("read batch midpoints: %w", err)
	}
	spreads, err := client.GetSpreads(ctx, selection)
	if err != nil {
		return fmt.Errorf("read batch spreads: %w", err)
	}
	lastTrades, err := client.GetLastTradePrices(ctx, selection)
	if err != nil {
		return fmt.Errorf("read batch last trades: %w", err)
	}
	// These independent reads are not an atomic snapshot. Match results by asset ID.
	if err = json.NewEncoder(os.Stdout).Encode(map[string]any{
		"prices": prices, "midpoints": midpoints, "spreads": spreads, "last_trades": lastTrades,
	}); err != nil {
		return fmt.Errorf("write batch prices: %w", err)
	}
	return nil
}
