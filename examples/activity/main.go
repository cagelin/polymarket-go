// Command activity reads tagged public wallet activity without credentials.
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
	pageSize := flag.Int("limit", 10, "activities in the first page (1-1000)")
	flag.Parse()
	if flag.NArg() != 1 || *pageSize < 1 || *pageSize > 1000 {
		return fmt.Errorf("usage: go run ./examples/activity [-limit 10] <wallet-address>")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	pager, err := client.ListActivity(polymarket.ListActivityParams{User: flag.Arg(0), PageSize: pageSize})
	if err != nil {
		return fmt.Errorf("select wallet activity: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	page, err := pager.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read wallet activity: %w", err)
	}
	for _, activity := range page.Items {
		// Inspect known variants and keep the remaining variants in the JSON output.
		switch entry := activity.(type) {
		case *polymarket.TradeActivity:
			log.Printf("trade asset=%s side=%s shares=%s", entry.AssetID, entry.Side, entry.Shares)
		case *polymarket.UnknownActivity:
			log.Printf("unknown activity type=%s; raw fields retained", entry.Type)
		}
	}
	if err = json.NewEncoder(os.Stdout).Encode(page); err != nil {
		return fmt.Errorf("write wallet activity: %w", err)
	}
	if page.LimitReached {
		return fmt.Errorf("service pagination limit reached; completeness is unknown")
	}
	return nil
}
