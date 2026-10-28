package polymarket

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestLivePublicMilestone(t *testing.T) {
	if os.Getenv("POLYMARKET_LIVE_PUBLIC") != "1" {
		t.Skip("set POLYMARKET_LIVE_PUBLIC=1 for public milestone reads")
	}
	c, err := NewPublicClient(ClientConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }() // Close only releases local resources and always returns nil.
	size := 2
	checks := []struct {
		name string
		read func(context.Context) error
	}{
		{"approvals", func(ctx context.Context) error {
			pager, queryErr := c.ListTraderLeaderboard(ListTraderLeaderboardParams{PageSize: &size})
			if queryErr != nil {
				return queryErr
			}
			page, readErr := pager.FirstPage(ctx)
			if readErr != nil {
				return readErr
			}
			if len(page.Items) == 0 {
				return fmt.Errorf("no public leaderboard wallet")
			}
			_, readErr = c.GetTradingApprovalsState(ctx, page.Items[0].Wallet)
			return readErr
		}},
		{"sports", func(ctx context.Context) error { _, readErr := c.GetSports(ctx); return readErr }},
		{"sports-types", func(ctx context.Context) error { _, readErr := c.GetSportsMarketTypes(ctx); return readErr }},
		{"series", func(ctx context.Context) error {
			p, queryErr := c.ListSeries(ListSeriesParams{PageSize: &size})
			if queryErr != nil {
				return queryErr
			}
			page, readErr := p.FirstPage(ctx)
			if readErr != nil {
				return readErr
			}
			if len(page.Items) > 0 {
				_, readErr = c.GetSeries(ctx, page.Items[0].ID, nil)
			}
			return readErr
		}},
		{"tags", func(ctx context.Context) error {
			p, queryErr := c.ListTags(ListTagsParams{PageSize: &size})
			if queryErr != nil {
				return queryErr
			}
			page, readErr := p.FirstPage(ctx)
			if readErr != nil {
				return readErr
			}
			if len(page.Items) > 0 {
				_, readErr = c.GetTag(ctx, TagQuery{ID: &page.Items[0].ID})
			}
			return readErr
		}},
		{"teams", func(ctx context.Context) error {
			p, queryErr := c.ListTeams(ListTeamsParams{PageSize: &size})
			if queryErr != nil {
				return queryErr
			}
			_, readErr := p.FirstPage(ctx)
			return readErr
		}},
		{"trades", func(ctx context.Context) error {
			p, queryErr := c.ListTrades(ListTradesParams{PageSize: &size})
			if queryErr != nil {
				return queryErr
			}
			_, readErr := p.FirstPage(ctx)
			return readErr
		}},
		{"leaderboard", func(ctx context.Context) error {
			p, queryErr := c.ListTraderLeaderboard(ListTraderLeaderboardParams{PageSize: &size})
			if queryErr != nil {
				return queryErr
			}
			_, readErr := p.FirstPage(ctx)
			return readErr
		}},
		{"open-interest", func(ctx context.Context) error {
			_, readErr := c.GetOpenInterests(ctx, GetOpenInterestsParams{})
			return readErr
		}},
		{"builder-volumes", func(ctx context.Context) error {
			_, readErr := c.GetBuilderVolumes(ctx, GetBuilderVolumesParams{})
			return readErr
		}},
		{"reward-markets", func(ctx context.Context) error {
			p, queryErr := c.ListRewardMarkets(ListRewardMarketsParams{PageSize: &size})
			if queryErr != nil {
				return queryErr
			}
			_, readErr := p.FirstPage(ctx)
			return readErr
		}},
		{"current-rewards", func(ctx context.Context) error {
			p, queryErr := c.ListCurrentRewards(ListCurrentRewardsParams{})
			if queryErr != nil {
				return queryErr
			}
			_, readErr := p.FirstPage(ctx)
			return readErr
		}},
		{"combo-markets", func(ctx context.Context) error {
			p, queryErr := c.ListComboMarkets(ListComboMarketsParams{PageSize: &size})
			if queryErr != nil {
				return queryErr
			}
			_, readErr := p.FirstPage(ctx)
			return readErr
		}},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			if readErr := check.read(ctx); readErr != nil {
				t.Fatal(readErr)
			}
		})
	}
}

func TestLivePublicStreams(t *testing.T) {
	if os.Getenv("POLYMARKET_LIVE_STREAMS") != "1" {
		t.Skip("set POLYMARKET_LIVE_STREAMS=1 for public stream reads")
	}
	c, err := NewPublicClient(ClientConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }() // Close only releases local resources and always returns nil.
	for _, check := range []struct {
		name string
		spec Spec
	}{{"comments", CommentsSpec{}}, {"twap-30", CryptoPricesChainlinkTwapSpec{WindowSeconds: 30}}, {"twap-60", CryptoPricesChainlinkTwapSpec{WindowSeconds: 60}}, {"equity", EquityPricesSpec{Symbol: "AAPL"}}, {"binance", CryptoPricesSpec{Topic: "prices.crypto.binance"}}, {"chainlink", CryptoPricesSpec{Topic: "prices.crypto.chainlink"}}, {"sports", SportsSpec{}}} {
		t.Run(check.name, func(t *testing.T) {
			t.Parallel()
			client, createErr := NewPublicClient(ClientConfig{})
			if createErr != nil {
				t.Fatal(createErr)
			}
			defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
			ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
			defer cancel()
			h, subscribeErr := client.Subscribe(ctx, check.spec)
			if subscribeErr != nil {
				t.Fatal(subscribeErr)
			}
			defer func() { _ = h.Close() }()
			event, readErr := h.Next(ctx)
			if readErr != nil {
				t.Fatal(readErr)
			}
			t.Logf("received typed %s/%s event", event.StreamTopic(), event.StreamType())
		})
	}
}

func TestLiveMarketStream(t *testing.T) {
	if os.Getenv("POLYMARKET_LIVE_STREAMS") != "1" {
		t.Skip("set POLYMARKET_LIVE_STREAMS=1 for public stream reads")
	}
	c, err := NewPublicClient(ClientConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }() // Close only releases local resources and always returns nil.
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	closed, size, order := false, 20, "volume24hr"
	pager, err := c.ListMarkets(ListMarketsParams{Closed: &closed, PageSize: &size, Order: &order, Ascending: &closed})
	if err != nil {
		t.Fatal(err)
	}
	page, err := pager.FirstPage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var assets []string
	for _, market := range page.Items {
		if market.Outcomes.Yes.TokenID != nil {
			assets = append(assets, string(*market.Outcomes.Yes.TokenID))
		}
	}
	if len(assets) == 0 {
		t.Fatal("no market assets available")
	}
	h, err := c.Subscribe(ctx, MarketSpec{AssetIDs: assets})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = h.Close() }()
	event, err := h.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("received typed %s/%s event", event.StreamTopic(), event.StreamType())
}
