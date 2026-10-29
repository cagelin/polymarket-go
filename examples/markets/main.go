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
		return err
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	closed, size, order := false, 5, "volume"
	markets, err := client.ListMarkets(polymarket.ListMarketsParams{Closed: &closed, Ascending: &closed, PageSize: &size, Order: &order})
	if err != nil {
		return err
	}
	page, err := markets.FirstPage(ctx)
	if err != nil {
		return err
	}
	for _, market := range page.Items {
		question := ""
		if market.Question != nil {
			question = *market.Question
		}
		fmt.Printf("%s: %s\n", market.ID, question)
	}
	if len(page.Items) == 0 || page.Items[0].Outcomes.Yes.TokenID == nil {
		return nil
	}
	interval := polymarket.HistoryDay
	history, err := client.ListPriceHistory(polymarket.PriceHistoryParams{AssetID: polymarket.ClobAssetID(*page.Items[0].Outcomes.Yes.TokenID), Interval: &interval})
	if err != nil {
		return err
	}
	for point, pointErr := range history.Items(ctx) {
		if pointErr != nil {
			return pointErr
		}
		fmt.Printf("first history observation: %s price=%s\n", point.Timestamp.Format(time.RFC3339), point.Price)
		break
	}
	return nil
}
