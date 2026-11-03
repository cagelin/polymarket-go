// Command pagination demonstrates bounded page iteration and query-bound cursor resumption.
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
	flags := flag.NewFlagSet("pagination", flag.ContinueOnError)
	flags.SetOutput(output)
	pageLimit := flags.Int("pages", 2, "maximum number of pages to read")
	cursor := flags.String("cursor", "", "resume cursor printed by a previous run of this same query")
	if parseErr := flags.Parse(args); parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			return nil
		}
		return parseErr
	}
	if flags.NArg() != 0 || *pageLimit < 1 {
		return fmt.Errorf("usage: go run ./examples/pagination [-pages positive-number] [-cursor cursor]")
	}
	client, err := polymarket.NewPublicClient(config)
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
		if _, outputErr := fmt.Fprintf(output, "page=%d items=%d has_more=%t limit_reached=%t\n", pageNumber, len(page.Items), page.HasMore, page.LimitReached); outputErr != nil {
			return fmt.Errorf("write pagination: %w", outputErr)
		}
		for _, market := range page.Items {
			if _, outputErr := fmt.Fprintf(output, "market=%s\n", market.ID); outputErr != nil {
				return fmt.Errorf("write pagination: %w", outputErr)
			}
		}
		if page.LimitReached {
			return fmt.Errorf("service pagination limit reached; completeness is unknown")
		}
		if !page.HasMore {
			if _, outputErr := fmt.Fprintln(output, "query complete"); outputErr != nil {
				return fmt.Errorf("write pagination: %w", outputErr)
			}
			break
		}
		if pageNumber == *pageLimit {
			if page.NextCursor == nil {
				return fmt.Errorf("continuing page has no cursor")
			}
			if _, outputErr := fmt.Fprintf(output, "stopped at local page limit; resume with -cursor %q\n", *page.NextCursor); outputErr != nil {
				return fmt.Errorf("write pagination: %w", outputErr)
			}
			break // Pages does not prefetch after the caller stops.
		}
	}
	return nil
}
