package polymarket

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLiveCLOB performs public reads only. Live tests are deliberately opt-in.
func TestLiveCLOB(t *testing.T) {
	id := ClobAssetID(os.Getenv("POLYMARKET_LIVE_ASSET_ID"))
	if id == "" {
		t.Skip("set POLYMARKET_LIVE_ASSET_ID to enable public API reads")
	}
	client, createErr := NewPublicClient(ClientConfig{})
	if createErr != nil {
		t.Fatal(createErr)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	asset, assets := Asset(id), AssetsQuery{AssetIDs: []ClobAssetID{id}}
	tests := []struct {
		name string
		read func() error
	}{
		{name: "midpoint", read: func() error { _, err := client.GetMidpoint(ctx, asset); return err }},
		{name: "midpoints", read: func() error { _, err := client.GetMidpoints(ctx, assets); return err }},
		{name: "price", read: func() error { _, err := client.GetPrice(ctx, PriceRequest{AssetQuery: asset, Side: Buy}); return err }},
		{name: "prices", read: func() error {
			_, err := client.GetPrices(ctx, []PriceRequest{{AssetQuery: asset, Side: Buy}, {AssetQuery: asset, Side: Sell}})
			return err
		}},
		{name: "book", read: func() error {
			book, err := client.GetOrderBook(ctx, asset)
			if err == nil && book.AssetID != id {
				t.Errorf("book asset = %s, want %s", book.AssetID, id)
			}
			return err
		}},
		{name: "books", read: func() error {
			books, err := client.GetOrderBooks(ctx, assets)
			if err == nil && (len(books) != 1 || books[0].AssetID != id) {
				t.Errorf("books do not identify requested asset %s", id)
			}
			return err
		}},
		{name: "spread", read: func() error { _, err := client.GetSpread(ctx, asset); return err }},
		{name: "spreads", read: func() error { _, err := client.GetSpreads(ctx, assets); return err }},
		{name: "last_trade", read: func() error { _, err := client.GetLastTradePrice(ctx, asset); return err }},
		{name: "last_trades", read: func() error { _, err := client.GetLastTradePrices(ctx, assets); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(subtest *testing.T) {
			if err := test.read(); err != nil {
				subtest.Fatal(err)
			}
		})
	}
}
