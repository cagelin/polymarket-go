// Command positions reads one page of a wallet's public indexed positions without credentials.
package main

import (
	"context"
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
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: go run ./examples/positions <wallet-address>")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wallet, pageSize := os.Args[1], 10
	pager, err := client.ListPositions(polymarket.ListPositionsParams{User: &wallet, PageSize: &pageSize})
	if err != nil {
		return fmt.Errorf("select wallet positions: %w", err)
	}
	page, err := pager.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read wallet positions: %w", err)
	}
	for _, position := range page.Items {
		fmt.Printf("asset=%s shares=%s value=%s total_pnl=%s status=%s redeemable=%t\n",
			position.AssetID, position.CurrentSize, position.CurrentValue, position.TotalPnL, position.Status, position.Redeemable)
	}
	fmt.Printf("positions on this page=%d has_more=%t limit_reached=%t\n", len(page.Items), page.HasMore, page.LimitReached)
	if page.LimitReached {
		return fmt.Errorf("service pagination limit reached; completeness is unknown")
	}
	return nil
}
