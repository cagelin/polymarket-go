package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
)

func callDataParity(ctx context.Context, c *PublicClient, operation string, input json.RawMessage, cursor *string) (any, error) {
	switch operation {
	case "list_trades":
		var p ListTradesParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListTrades(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_activity":
		var p ListActivityParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListActivity(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_combo_activity":
		var p ListComboActivityParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListComboActivity(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_positions":
		var p ListPositionsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListPositions(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_combo_positions":
		var p ListComboPositionsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListComboPositions(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "get_portfolio_value":
		var p GetPortfolioValueParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetPortfolioValue(ctx, p)
	case "get_user_stats":
		var p GetUserStatsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetUserStats(ctx, p)
	case "get_user_pnl":
		var p GetUserPnLParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetUserPnL(ctx, p)
	case "get_user_volume":
		var p GetUserVolumeParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetUserVolume(ctx, p)
	case "list_market_holders":
		var p ListMarketHoldersParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListMarketHolders(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "get_open_interests":
		var p GetOpenInterestsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetOpenInterests(ctx, p)
	case "get_event_live_volume":
		var p GetEventLiveVolumeParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetEventLiveVolume(ctx, p)
	case "get_token_references":
		var p GetTokenReferencesParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetTokenReferences(ctx, p)
	case "get_resolutions":
		var p GetResolutionsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetResolutions(ctx, p)
	case "list_trader_leaderboard":
		var p ListTraderLeaderboardParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListTraderLeaderboard(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "get_trader_leaderboard_standing":
		var p GetTraderLeaderboardStandingParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetTraderLeaderboardStanding(ctx, p)
	case "list_biggest_winners":
		var p ListBiggestWinnersParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListBiggestWinners(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_builder_leaderboard":
		var p ListBuilderLeaderboardParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListBuilderLeaderboard(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "get_builder_volumes":
		var p GetBuilderVolumesParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetBuilderVolumes(ctx, p)
	case "download_accounting_snapshot":
		var p DownloadAccountingSnapshotParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.DownloadAccountingSnapshot(ctx, p)
	default:
		return nil, fmt.Errorf("unknown operation %s", operation)
	}
}
