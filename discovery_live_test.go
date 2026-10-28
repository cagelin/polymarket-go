package polymarket

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLiveDiscovery only reads public market and history endpoints.
func TestLiveDiscovery(t *testing.T) {
	if os.Getenv("POLYMARKET_LIVE_DISCOVERY") != "1" {
		t.Skip("set POLYMARKET_LIVE_DISCOVERY=1 to enable public discovery reads")
	}
	client, createErr := NewPublicClient(ClientConfig{})
	if createErr != nil {
		t.Fatal(createErr)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	closed, size, order := false, 2, "volume"
	pager, err := client.ListMarkets(ListMarketsParams{Closed: &closed, Ascending: &closed, PageSize: &size, Order: &order})
	if err != nil {
		t.Fatal(err)
	}
	page, err := pager.FirstPage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) == 0 {
		t.Fatal("no public market returned")
	}
	market, err := client.GetMarket(ctx, MarketQuery{ID: page.Items[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	if market.ID != page.Items[0].ID {
		t.Fatal("market lookup returned a different ID")
	}
	if market.Slug != nil {
		bySlug, lookupErr := client.GetMarket(ctx, MarketQuery{Slug: *market.Slug})
		if lookupErr != nil {
			t.Fatal(lookupErr)
		}
		if bySlug.ID != market.ID {
			t.Fatal("slug lookup returned a different ID")
		}
	}
	if page.NextCursor != nil {
		next, pageErr := pager.FromCursor(page.NextCursor).FirstPage(ctx)
		if pageErr != nil {
			t.Fatal(pageErr)
		}
		if len(next.Items) > 0 && next.Items[0].ID == page.Items[0].ID {
			t.Fatal("cursor replayed the first page")
		}
	}
	if market.Outcomes.Yes.TokenID == nil {
		t.Fatal("market has no token ID for history lookup")
	}
	interval := HistoryDay
	history, err := client.ListPriceHistory(PriceHistoryParams{AssetID: ClobAssetID(*market.Outcomes.Yes.TokenID), Interval: &interval, PageSize: &size})
	if err != nil {
		t.Fatal(err)
	}
	points, err := history.FirstPage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(points.Items) == 0 {
		t.Fatal("history endpoint returned no observations")
	}
	if points.NextCursor != nil {
		continued, pageErr := history.FromCursor(points.NextCursor).FirstPage(ctx)
		if pageErr != nil {
			t.Fatal(pageErr)
		}
		if len(continued.Items) > 0 && continued.Items[0].Timestamp.Before(points.Items[len(points.Items)-1].Timestamp) {
			t.Fatal("history continuation moved backwards")
		}
	}
	t.Logf("verified market %s, market continuation and %d history observations", market.ID, len(points.Items))
}
