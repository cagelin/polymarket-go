package polymarket

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLiveEvents reads public events, search results and tags without credentials.
func TestLiveEvents(t *testing.T) {
	if os.Getenv("POLYMARKET_LIVE_EVENTS") != "1" {
		t.Skip("set POLYMARKET_LIVE_EVENTS=1 to enable public event reads")
	}
	client, err := NewPublicClient(ClientConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	size, descending, order := 2, false, "volume"
	pager, err := client.ListEvents(ListEventsParams{PageSize: &size, Ascending: &descending, Order: &order})
	if err != nil {
		t.Fatal(err)
	}
	page, err := pager.FirstPage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) == 0 || page.NextCursor == nil {
		t.Fatal("ListEvents returned no events or continuation")
	}
	next, err := pager.FromCursor(page.NextCursor).FirstPage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Items) == 0 || next.Items[0].ID == page.Items[0].ID {
		t.Fatal("event continuation was empty or repeated the first event")
	}
	event, err := client.GetEvent(ctx, EventQuery{ID: page.Items[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != page.Items[0].ID || event.Slug == nil {
		t.Fatal("event lookup returned the wrong event or no slug")
	}
	linked, err := client.GetEvent(ctx, EventQuery{URL: "https://polymarket.com/event/" + *event.Slug})
	if err != nil {
		t.Fatal(err)
	}
	if linked.ID != event.ID {
		t.Fatal("event URL lookup returned a different event")
	}
	tags, err := client.GetEventTags(ctx, event.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(event.Markets) == 0 {
		t.Fatal("event has no binary market for tag lookup")
	}
	marketTags, err := client.GetMarketTags(ctx, event.Markets[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	search, err := client.Search(SearchParams{Query: "election", PageSize: &size})
	if err != nil {
		t.Fatal(err)
	}
	results, err := search.FirstPage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(results.Items) != 1 || len(results.Items[0].Events) == 0 || results.NextCursor == nil {
		t.Fatal("search returned no event results or continuation")
	}
	continued, err := search.FromCursor(results.NextCursor).FirstPage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(continued.Items) != 1 || len(continued.Items[0].Events) == 0 {
		t.Fatal("search continuation returned no events")
	}
	if continued.Items[0].Events[0].ID == results.Items[0].Events[0].ID {
		t.Fatal("search continuation repeated the first event")
	}
	t.Logf("verified event %s, event/search continuation, %d event tags and %d market tags", event.ID, len(tags), len(marketTags))
}
