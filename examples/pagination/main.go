// Command pagination demonstrates bounded page iteration and query-bound cursor resumption.
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
	pageLimit := flag.Int("pages", 2, "maximum number of pages to read")
	cursor := flag.String("cursor", "", "resume cursor printed by a previous run of this same query")
	flag.Parse()
	if flag.NArg() != 0 || *pageLimit < 1 {
		return fmt.Errorf("usage: go run ./examples/pagination [-pages positive-number] [-cursor cursor]")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	closed, pageSize := false, 5
	pager, err := client.ListMarkets(polymarket.ListMarketsParams{Closed: &closed, PageSize: &pageSize})
	if err != nil {
		return fmt.Errorf("select markets: %w", err)
	}
	if *cursor != "" {
		pager = pager.FromCursor(cursor)
	}
	pageNumber := 0
	for page, pageErr := range pager.Pages(ctx) {
		if pageErr != nil {
			return fmt.Errorf("read market page: %w", pageErr)
		}
		pageNumber++
		fmt.Printf("page=%d items=%d has_more=%t limit_reached=%t\n", pageNumber, len(page.Items), page.HasMore, page.LimitReached)
		for _, market := range page.Items {
			fmt.Printf("market=%s\n", market.ID)
		}
		if page.LimitReached {
			return fmt.Errorf("service pagination limit reached; completeness is unknown")
		}
		if !page.HasMore {
			fmt.Println("query complete")
			break
		}
		if pageNumber == *pageLimit {
			if page.NextCursor == nil {
				return fmt.Errorf("continuing page has no cursor")
			}
			fmt.Printf("stopped at local page limit; resume with -cursor %q\n", *page.NextCursor)
			break // Pages does not prefetch after the caller stops.
		}
	}
	return nil
}
