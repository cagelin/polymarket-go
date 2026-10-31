// Search demonstrates public event, market and tag discovery without credentials.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cagelin/polymarket-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	query := "election"
	if len(os.Args) > 1 {
		query = strings.Join(os.Args[1:], " ")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	size := 5
	pager, err := client.Search(polymarket.SearchParams{Query: query, PageSize: &size})
	if err != nil {
		return fmt.Errorf("select search query: %w", err)
	}
	page, err := pager.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read search page: %w", err)
	}
	for _, group := range page.Items {
		for _, event := range group.Events {
			title := ""
			if event.Title != nil {
				title = *event.Title
			}
			fmt.Printf("%s  %s (%d markets)\n", event.ID, title, len(event.Markets))
		}
		fmt.Printf("Tags: %d, profiles: %d, more pages: %t\n", len(group.Tags), len(group.Profiles), page.HasMore)
	}
	fmt.Printf("result groups=%d has_more=%t limit_reached=%t\n", len(page.Items), page.HasMore, page.LimitReached)
	return nil
}
