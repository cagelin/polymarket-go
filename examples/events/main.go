// Command events lists public events or fetches one event and its markets and tags.
package main

import (
	"context"
	"flag"
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
	eventID := flag.String("id", "", "fetch one event by ID")
	eventSlug := flag.String("slug", "", "fetch one event by slug")
	eventURL := flag.String("url", "", "fetch one event by its Polymarket event URL")
	pageSize := flag.Int("limit", 5, "number of events in the first list page (list mode only)")
	closed := flag.Bool("closed", false, "list closed events instead of open events (list mode only)")
	tagSlug := flag.String("tag", "", "filter event lists by tag slug (list mode only)")
	seriesID := flag.Int64("series", 0, "filter event lists by series ID (list mode only)")
	flag.Parse()
	selectorCount := 0
	for _, selector := range []string{*eventID, *eventSlug, *eventURL} {
		if selector != "" {
			selectorCount++
		}
	}
	if flag.NArg() != 0 || selectorCount > 1 || *pageSize < 1 || *seriesID < 0 {
		return fmt.Errorf("usage: go run ./examples/events [-limit positive-number] [-closed] [-tag slug] [-series id] [-id event-id | -slug event-slug | -url event-url]")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if selectorCount == 1 {
		event, lookupErr := client.GetEvent(ctx, polymarket.EventQuery{
			ID:   polymarket.EventID(*eventID),
			Slug: *eventSlug,
			URL:  *eventURL,
		})
		if lookupErr != nil {
			return fmt.Errorf("read event: %w", lookupErr)
		}
		printEvent(event)
		tags, tagsErr := client.GetEventTags(ctx, event.ID)
		if tagsErr != nil {
			return fmt.Errorf("read event tags: %w", tagsErr)
		}
		fmt.Printf("event tags=%d\n", len(tags))
		for _, tag := range tags {
			label := ""
			if tag.Label != nil {
				label = *tag.Label
			}
			fmt.Printf("  tag=%s label=%s\n", tag.ID, label)
		}
		return nil
	}
	ascending, order := false, "volume"
	params := polymarket.ListEventsParams{
		PageSize:  pageSize,
		Closed:    *closed,
		Ascending: &ascending,
		Order:     &order,
	}
	if *tagSlug != "" {
		params.TagSlug = tagSlug
	}
	if *seriesID > 0 {
		params.SeriesIDs = []int64{*seriesID}
	}
	pager, err := client.ListEvents(params)
	if err != nil {
		return fmt.Errorf("select events: %w", err)
	}
	page, err := pager.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read event page: %w", err)
	}
	for _, listedEvent := range page.Items {
		printEvent(&listedEvent)
	}
	fmt.Printf("events on this page=%d has_more=%t limit_reached=%t\n", len(page.Items), page.HasMore, page.LimitReached)
	if page.LimitReached {
		return fmt.Errorf("service pagination limit reached; completeness is unknown")
	}
	return nil
}

func printEvent(event *polymarket.Event) {
	title, slug := "", ""
	if event.Title != nil {
		title = *event.Title
	}
	if event.Slug != nil {
		slug = *event.Slug
	}
	fmt.Printf("event=%s slug=%s title=%s markets=%d\n", event.ID, slug, title, len(event.Markets))
	for _, market := range event.Markets {
		question := ""
		if market.Question != nil {
			question = *market.Question
		}
		fmt.Printf("  market=%s question=%s\n", market.ID, question)
		if market.Outcomes.Yes.TokenID != nil {
			fmt.Printf("    YES asset-id=%s\n", *market.Outcomes.Yes.TokenID)
		}
		if market.Outcomes.No.TokenID != nil {
			fmt.Printf("    NO  asset-id=%s\n", *market.Outcomes.No.TokenID)
		}
	}
}
