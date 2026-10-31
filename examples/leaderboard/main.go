// Command leaderboard reads a first page of public trader rankings.
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
	category := flag.String("category", "", "optional leaderboard category; empty uses the service default")
	window := flag.String("window", "week", "leaderboard window: day, week, month or all")
	sortBy := flag.String("sort", "PNL", "sort by PNL or VOLUME")
	pageSize := flag.Int("limit", 10, "traders in the first page (1-1000)")
	flag.Parse()
	if flag.NArg() != 0 || *pageSize < 1 || *pageSize > 1000 {
		return fmt.Errorf("usage: go run ./examples/leaderboard [-category category] [-window week] [-sort PNL|VOLUME] [-limit 10]")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	params := polymarket.ListTraderLeaderboardParams{Window: window, SortBy: sortBy, PageSize: pageSize}
	if *category != "" {
		params.Category = category
	}
	pager, err := client.ListTraderLeaderboard(params)
	if err != nil {
		return fmt.Errorf("select trader leaderboard: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	page, err := pager.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read trader leaderboard: %w", err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(page); err != nil {
		return fmt.Errorf("write trader leaderboard: %w", err)
	}
	if page.LimitReached {
		return fmt.Errorf("service pagination limit reached; completeness is unknown")
	}
	return nil
}
