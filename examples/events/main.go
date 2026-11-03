// Command events lists public events or fetches one event and its markets and tags.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
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
	return runWith(polymarket.ClientConfig{}, os.Args[1:], os.Stdout)
}

func runWith(config polymarket.ClientConfig, args []string, output io.Writer) error {
	flags := flag.NewFlagSet("events", flag.ContinueOnError)
	flags.SetOutput(output)
	eventID := flags.String("id", "", "fetch one event by ID")
	eventSlug := flags.String("slug", "", "fetch one event by slug")
	eventURL := flags.String("url", "", "fetch one event by its Polymarket event URL")
	pageSize := flags.Int("limit", 5, "number of events in the first list page (list mode only)")
	closed := flags.Bool("closed", false, "list closed events instead of open events (list mode only)")
	tagSlug := flags.String("tag", "", "filter event lists by tag slug (list mode only)")
	seriesID := flags.Int64("series", 0, "filter event lists by series ID (list mode only)")
	if parseErr := flags.Parse(args); parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			return nil
		}
		return parseErr
	}
	selectorCount := 0
	for _, selector := range []string{*eventID, *eventSlug, *eventURL} {
		if selector != "" {
			selectorCount++
		}
	}
	if flags.NArg() != 0 || selectorCount > 1 || *pageSize < 1 || *seriesID < 0 {
		return fmt.Errorf("usage: go run ./examples/events [-limit positive-number] [-closed] [-tag slug] [-series id] [-id event-id | -slug event-slug | -url event-url]")
	}
	client, err := polymarket.NewPublicClient(config)
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
		if outputErr := printEvent(output, event); outputErr != nil {
			return fmt.Errorf("write event: %w", outputErr)
		}
		tags, tagsErr := client.GetEventTags(ctx, event.ID)
		if tagsErr != nil {
			return fmt.Errorf("read event tags: %w", tagsErr)
		}
		if _, outputErr := fmt.Fprintf(output, "event tags=%d\n", len(tags)); outputErr != nil {
			return fmt.Errorf("write events: %w", outputErr)
		}
		for _, tag := range tags {
			label := ""
			if tag.Label != nil {
				label = *tag.Label
			}
			if _, outputErr := fmt.Fprintf(output, "  tag=%s label=%s\n", tag.ID, label); outputErr != nil {
				return fmt.Errorf("write events: %w", outputErr)
			}
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
		if outputErr := printEvent(output, &listedEvent); outputErr != nil {
			return fmt.Errorf("write event: %w", outputErr)
		}
	}
	if _, outputErr := fmt.Fprintf(output, "events on this page=%d has_more=%t limit_reached=%t\n", len(page.Items), page.HasMore, page.LimitReached); outputErr != nil {
		return fmt.Errorf("write events: %w", outputErr)
	}
	if page.LimitReached {
		return fmt.Errorf("service pagination limit reached; completeness is unknown")
	}
	return nil
}

func printEvent(output io.Writer, event *polymarket.Event) error {
	title, slug := "", ""
	if event.Title != nil {
		title = *event.Title
	}
	if event.Slug != nil {
		slug = *event.Slug
	}
	result := fmt.Appendf(nil, "event=%s slug=%s title=%s markets=%d\n", event.ID, slug, title, len(event.Markets))
	for _, market := range event.Markets {
		question := ""
		if market.Question != nil {
			question = *market.Question
		}
		result = fmt.Appendf(result, "  market=%s question=%s\n", market.ID, question)
		if market.Outcomes.Yes.TokenID != nil {
			result = fmt.Appendf(result, "    YES asset-id=%s\n", *market.Outcomes.Yes.TokenID)
		}
		if market.Outcomes.No.TokenID != nil {
			result = fmt.Appendf(result, "    NO  asset-id=%s\n", *market.Outcomes.No.TokenID)
		}
	}
	_, err := output.Write(result)
	return err
}
