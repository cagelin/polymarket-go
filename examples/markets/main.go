// Command markets prints a page of open markets and the first available history point.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cagelin/polymarket-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	closed, ascending, size, order := false, false, 5, "volume"
	markets, err := client.ListMarkets(polymarket.ListMarketsParams{Closed: &closed, Ascending: &ascending, PageSize: &size, Order: &order})
	if err != nil {
		return fmt.Errorf("select open markets: %w", err)
	}
	page, err := markets.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read market page: %w", err)
	}
	if len(page.Items) == 0 {
		fmt.Println("no open markets returned")
		return nil
	}
	for _, market := range page.Items {
		question := ""
		if market.Question != nil {
			question = *market.Question
		}
		fmt.Printf("%s: %s\n", market.ID, question)
		if market.Outcomes.Yes.TokenID != nil {
			fmt.Printf("  YES asset-id=%s\n", *market.Outcomes.Yes.TokenID)
		}
		if market.Outcomes.No.TokenID != nil {
			fmt.Printf("  NO  asset-id=%s\n", *market.Outcomes.No.TokenID)
		}
	}
	fmt.Printf("has_more=%t limit_reached=%t\n", page.HasMore, page.LimitReached)
	if page.Items[0].Outcomes.Yes.TokenID == nil {
		fmt.Println("first market has no YES token; skipping history")
		return nil
	}
	interval := polymarket.HistoryDay
	history, err := client.ListPriceHistory(polymarket.PriceHistoryParams{AssetID: polymarket.ClobAssetID(*page.Items[0].Outcomes.Yes.TokenID), Interval: &interval})
	if err != nil {
		return fmt.Errorf("select price history: %w", err)
	}
	foundHistory := false
	for point, pointErr := range history.Items(ctx) {
		if pointErr != nil {
			return fmt.Errorf("read price history: %w", pointErr)
		}
		fmt.Printf("first history observation: %s price=%s\n", point.Timestamp.Format(time.RFC3339), point.Price)
		foundHistory = true
		break
	}
	if !foundHistory {
		fmt.Println("no history observations returned")
	}
	return nil
}
