// Command catalog discovers public tags, series, sports or teams.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/cagelin/polymarket-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	kind := flag.String("kind", "tags", "catalog: tags, series, sports or teams")
	pageSize := flag.Int("limit", 5, "first page size for tags, series or teams (1-50)")
	flag.Parse()
	if flag.NArg() != 0 || !slices.Contains([]string{"tags", "series", "sports", "teams"}, *kind) || *pageSize < 1 || *pageSize > 50 {
		return fmt.Errorf("usage: go run ./examples/catalog [-kind tags|series|sports|teams] [-limit 5]")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var result any
	switch *kind {
	case "tags":
		tags, tagsErr := client.ListTags(polymarket.ListTagsParams{PageSize: pageSize})
		if tagsErr != nil {
			return fmt.Errorf("select tags: %w", tagsErr)
		}
		result, err = tags.FirstPage(ctx)
	case "series":
		series, seriesErr := client.ListSeries(polymarket.ListSeriesParams{PageSize: pageSize})
		if seriesErr != nil {
			return fmt.Errorf("select series: %w", seriesErr)
		}
		result, err = series.FirstPage(ctx)
	case "teams":
		teams, teamsErr := client.ListTeams(polymarket.ListTeamsParams{PageSize: pageSize})
		if teamsErr != nil {
			return fmt.Errorf("select teams: %w", teamsErr)
		}
		result, err = teams.FirstPage(ctx)
	case "sports":
		result, err = client.GetSports(ctx)
	}
	if err != nil {
		return fmt.Errorf("read %s catalog: %w", *kind, err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(result); err != nil {
		return fmt.Errorf("write catalog: %w", err)
	}
	return nil
}
