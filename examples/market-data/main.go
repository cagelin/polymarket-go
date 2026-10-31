// Command market-data reads indexed holders and open interest for one condition.
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
	pageSize := flag.Int("limit", 10, "holders in the first page (1-1000)")
	flag.Parse()
	if flag.NArg() != 1 || *pageSize < 1 || *pageSize > 1000 {
		return fmt.Errorf("usage: go run ./examples/market-data [-limit 10] <condition-id>")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	conditions := []string{flag.Arg(0)}
	pager, err := client.ListMarketHolders(polymarket.ListMarketHoldersParams{ConditionIDs: conditions, PageSize: pageSize})
	if err != nil {
		return fmt.Errorf("select market holders: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	holders, err := pager.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read market holders: %w", err)
	}
	interest, err := client.GetOpenInterests(ctx, polymarket.GetOpenInterestsParams{ConditionIDs: conditions})
	if err != nil {
		return fmt.Errorf("read open interest: %w", err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(map[string]any{"holders": holders, "open_interest": interest}); err != nil {
		return fmt.Errorf("write market data: %w", err)
	}
	if holders.LimitReached {
		return fmt.Errorf("service pagination limit reached; completeness is unknown")
	}
	return nil
}
