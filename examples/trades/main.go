// Command trades reads a first page of public trades with optional wallet/market filters.
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
	wallet := flag.String("wallet", "", "optional public wallet address")
	condition := flag.String("condition", "", "optional market condition ID (not an asset ID)")
	side := flag.String("side", "", "optional BUY or SELL filter")
	pageSize := flag.Int("limit", 10, "trades in the first page (1-1000)")
	flag.Parse()
	if flag.NArg() != 0 || *pageSize < 1 || *pageSize > 1000 {
		return fmt.Errorf("usage: go run ./examples/trades [-wallet address] [-condition condition-id] [-side BUY|SELL] [-limit 10]")
	}
	params := polymarket.ListTradesParams{PageSize: pageSize}
	if *wallet != "" {
		params.User = wallet
	}
	if *condition != "" {
		params.ConditionID = []string{*condition}
	}
	if *side != "" {
		params.Side = side
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	pager, err := client.ListTrades(params)
	if err != nil {
		return fmt.Errorf("select public trades: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	page, err := pager.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read public trades: %w", err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(page); err != nil {
		return fmt.Errorf("write public trades: %w", err)
	}
	if page.LimitReached {
		return fmt.Errorf("service pagination limit reached; completeness is unknown")
	}
	return nil
}
