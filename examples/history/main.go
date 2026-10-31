// Command history reads a bounded first page of an asset's price history.
package main

import (
	"context"
	"encoding/json"
	"flag"
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
	interval := flag.String("interval", "", "history interval: 1h, 6h, 1d, 1w, 1m, all or max (default 1d)")
	startText := flag.String("start", "", "window start in RFC3339; mutually exclusive with -interval and -as-of")
	endText := flag.String("end", "", "window end in RFC3339; requires -start, at most 15 days later")
	asOfText := flag.String("as-of", "", "historical price observation at RFC3339 time")
	flag.Parse()
	if flag.NArg() != 1 || (*endText != "" && *startText == "") {
		return fmt.Errorf("usage: go run ./examples/history [-interval 1d | -start time [-end time] | -as-of time] <asset-id>")
	}
	params := polymarket.PriceHistoryParams{AssetID: polymarket.ClobAssetID(flag.Arg(0))}
	if *interval != "" {
		historyInterval := polymarket.PriceHistoryInterval(*interval)
		params.Interval = &historyInterval
	}
	// Parse user-facing RFC3339 dates, then use Unix seconds in the SDK request.
	for _, bound := range []struct {
		name   string
		text   string
		target **int64
	}{
		{name: "start", text: *startText, target: &params.Start},
		{name: "end", text: *endText, target: &params.End},
		{name: "as-of", text: *asOfText, target: &params.AsOf},
	} {
		if bound.text == "" {
			continue
		}
		parsedTime, parseErr := time.Parse(time.RFC3339, bound.text)
		if parseErr != nil {
			return fmt.Errorf("parse %s time: %w", bound.name, parseErr)
		}
		seconds := parsedTime.Unix()
		*bound.target = &seconds
	}
	if params.Interval == nil && params.Start == nil && params.AsOf == nil {
		defaultInterval := polymarket.HistoryDay
		params.Interval = &defaultInterval
	}
	if params.AsOf == nil {
		pageSize := 20
		params.PageSize = &pageSize
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }()         // Close only releases local resources.
	pager, err := client.ListPriceHistory(params) // Validates selectors and time windows before a request.
	if err != nil {
		return fmt.Errorf("select price history: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	page, err := pager.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read price history: %w", err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(page); err != nil {
		return fmt.Errorf("write price history: %w", err)
	}
	if page.LimitReached {
		return fmt.Errorf("service pagination limit reached; completeness is unknown")
	}
	return nil
}
