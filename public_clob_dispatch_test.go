package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
)

func callPublicClobParity(ctx context.Context, c *PublicClient, operation string, input json.RawMessage, cursor *string) (any, error) {
	switch operation {
	case "list_builder_trades":
		var p ListBuilderTradesParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListBuilderTrades(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_reward_markets":
		var p ListRewardMarketsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListRewardMarkets(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_current_rewards":
		var p ListCurrentRewardsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListCurrentRewards(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_market_rewards":
		var p ListMarketRewardsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListMarketRewards(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "get_current_rebates":
		var p GetCurrentRebatesParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetCurrentRebates(ctx, p)
	case "list_combo_markets":
		var p ListComboMarketsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListComboMarkets(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	default:
		return nil, fmt.Errorf("unknown operation %s", operation)
	}
}
