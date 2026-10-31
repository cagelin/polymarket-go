// Command market fetches one market's metadata and tags without credentials.
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
	marketID := flag.String("id", "", "Gamma market ID (not an asset or condition ID)")
	marketSlug := flag.String("slug", "", "market slug")
	marketURL := flag.String("url", "", "Polymarket market page URL")
	flag.Parse()
	selectorCount := 0
	for _, selector := range []string{*marketID, *marketSlug, *marketURL} {
		if selector != "" {
			selectorCount++
		}
	}
	if flag.NArg() != 0 || selectorCount != 1 {
		return fmt.Errorf("provide exactly one of -id, -slug or -url")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	market, err := client.GetMarket(ctx, polymarket.MarketQuery{ID: polymarket.MarketID(*marketID), Slug: *marketSlug, URL: *marketURL})
	if err != nil {
		return fmt.Errorf("read market metadata: %w", err)
	}
	tags, err := client.GetMarketTags(ctx, market.ID)
	if err != nil {
		return fmt.Errorf("read market tags: %w", err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(map[string]any{"market": market, "tags": tags}); err != nil {
		return fmt.Errorf("write market metadata: %w", err)
	}
	return nil
}
