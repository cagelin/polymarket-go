package data

import (
	"context"

	"github.com/cagelin/polymarket-go/models"
	"github.com/cagelin/polymarket-go/pagination"
)

// ListTradesParams selects filters for ListTrades. Nil optional fields are omitted.
type ListTradesParams struct {
	User         *string  `json:"user"`
	ConditionID  []string `json:"condition_id"`
	EventID      []int64  `json:"event_id"`
	Side         *string  `json:"side"`
	TakerOnly    *bool    `json:"taker_only"`
	FilterType   *string  `json:"filter_type"`
	FilterAmount *float64 `json:"filter_amount"`
	Start        *int64   `json:"start"`
	End          *int64   `json:"end"`
	FullHistory  bool     `json:"full_history"`
	PageSize     *int     `json:"page_size"`
}

// ListTrades reads public Data API information.
func (c *Client) ListTrades(p ListTradesParams) (*pagination.Paginator[models.Trade], error) {
	query := map[string]any{}
	put(query, "user", p.User)
	if p.ConditionID != nil {
		query["condition_id"] = append(make([]string, 0, len(p.ConditionID)), p.ConditionID...)
	}
	if p.EventID != nil {
		query["event_id"] = append(make([]int64, 0, len(p.EventID)), p.EventID...)
	}
	put(query, "side", p.Side)
	put(query, "taker_only", p.TakerOnly)
	put(query, "filter_type", p.FilterType)
	put(query, "filter_amount", p.FilterAmount)
	put(query, "start", p.Start)
	put(query, "end", p.End)
	query["full_history"] = p.FullHistory
	if err := normalizeQuery("list_trades", query); err != nil {
		return nil, err
	}
	return listData(c, "/v2/trades", query, p.PageSize, 1000, parseList[models.Trade])
}

// ListActivityParams selects filters for ListActivity. Nil optional fields are omitted.
type ListActivityParams struct {
	User          string   `json:"user"`
	ConditionID   []string `json:"condition_id"`
	EventID       []int64  `json:"event_id"`
	ActivityTypes []string `json:"activity_types"`
	Side          *string  `json:"side"`
	SortDirection *string  `json:"sort_direction"`
	Start         *int64   `json:"start"`
	End           *int64   `json:"end"`
	FullHistory   bool     `json:"full_history"`
	PageSize      *int     `json:"page_size"`
}

// ListActivity reads public Data API information.
func (c *Client) ListActivity(p ListActivityParams) (*pagination.Paginator[models.Activity], error) {
	query := map[string]any{}
	query["user"] = p.User
	if p.ConditionID != nil {
		query["condition_id"] = append(make([]string, 0, len(p.ConditionID)), p.ConditionID...)
	}
	if p.EventID != nil {
		query["event_id"] = append(make([]int64, 0, len(p.EventID)), p.EventID...)
	}
	if p.ActivityTypes != nil {
		query["activity_types"] = append(make([]string, 0, len(p.ActivityTypes)), p.ActivityTypes...)
	}
	put(query, "side", p.Side)
	put(query, "sort_direction", p.SortDirection)
	put(query, "start", p.Start)
	put(query, "end", p.End)
	query["full_history"] = p.FullHistory
	if err := normalizeQuery("list_activity", query); err != nil {
		return nil, err
	}
	return listData(c, "/v2/activity", query, p.PageSize, 1000, models.ParseActivities)
}

// ListComboActivityParams selects filters for ListComboActivity. Nil optional fields are omitted.
type ListComboActivityParams struct {
	User        string   `json:"user"`
	ConditionID []string `json:"condition_id"`
	PageSize    *int     `json:"page_size"`
}

// ListComboActivity reads public Data API information.
func (c *Client) ListComboActivity(p ListComboActivityParams) (*pagination.Paginator[models.ComboActivity], error) {
	query := map[string]any{}
	query["user"] = p.User
	if p.ConditionID != nil {
		query["condition_id"] = append(make([]string, 0, len(p.ConditionID)), p.ConditionID...)
	}
	if err := normalizeQuery("list_combo_activity", query); err != nil {
		return nil, err
	}
	return listData(c, "/v2/activity/combos", query, p.PageSize, 1000, models.ParseComboActivities)
}

// ListPositionsParams selects filters for ListPositions. Nil optional fields are omitted.
type ListPositionsParams struct {
	User            *string  `json:"user"`
	ConditionID     []string `json:"condition_id"`
	Status          *string  `json:"status"`
	EventID         []int64  `json:"event_id"`
	Title           *string  `json:"title"`
	FilterType      *string  `json:"filter_type"`
	FilterAmount    *float64 `json:"filter_amount"`
	IncludeArchived *bool    `json:"include_archived"`
	SortBy          *string  `json:"sort_by"`
	SortDirection   *string  `json:"sort_direction"`
	Start           *int64   `json:"start"`
	End             *int64   `json:"end"`
	FullHistory     bool     `json:"full_history"`
	PageSize        *int     `json:"page_size"`
}

// ListPositions reads public Data API information.
func (c *Client) ListPositions(p ListPositionsParams) (*pagination.Paginator[models.Position], error) {
	query := map[string]any{}
	put(query, "user", p.User)
	if p.ConditionID != nil {
		query["condition_id"] = append(make([]string, 0, len(p.ConditionID)), p.ConditionID...)
	}
	put(query, "status", p.Status)
	if p.EventID != nil {
		query["event_id"] = append(make([]int64, 0, len(p.EventID)), p.EventID...)
	}
	put(query, "title", p.Title)
	put(query, "filter_type", p.FilterType)
	put(query, "filter_amount", p.FilterAmount)
	put(query, "include_archived", p.IncludeArchived)
	put(query, "sort_by", p.SortBy)
	put(query, "sort_direction", p.SortDirection)
	put(query, "start", p.Start)
	put(query, "end", p.End)
	query["full_history"] = p.FullHistory
	if err := normalizeQuery("list_positions", query); err != nil {
		return nil, err
	}
	return listData(c, "/v2/positions", query, p.PageSize, 1000, parseList[models.Position])
}

// ListComboPositionsParams selects filters for ListComboPositions. Nil optional fields are omitted.
type ListComboPositionsParams struct {
	User          string   `json:"user"`
	ConditionID   []string `json:"condition_id"`
	Status        []string `json:"status"`
	SortBy        *string  `json:"sort_by"`
	SortDirection *string  `json:"sort_direction"`
	UpdatedAfter  *int64   `json:"updated_after"`
	UpdatedBefore *int64   `json:"updated_before"`
	PageSize      *int     `json:"page_size"`
}

// ListComboPositions reads public Data API information.
func (c *Client) ListComboPositions(p ListComboPositionsParams) (*pagination.Paginator[models.ComboPosition], error) {
	query := map[string]any{}
	query["user"] = p.User
	if p.ConditionID != nil {
		query["condition_id"] = append(make([]string, 0, len(p.ConditionID)), p.ConditionID...)
	}
	if p.Status != nil {
		query["status"] = append(make([]string, 0, len(p.Status)), p.Status...)
	}
	put(query, "sort_by", p.SortBy)
	put(query, "sort_direction", p.SortDirection)
	put(query, "updated_after", p.UpdatedAfter)
	put(query, "updated_before", p.UpdatedBefore)
	if err := normalizeQuery("list_combo_positions", query); err != nil {
		return nil, err
	}
	return listData(c, "/v2/positions/combos", query, p.PageSize, 1000, parseList[models.ComboPosition])
}

// GetPortfolioValueParams selects filters for GetPortfolioValue. Nil optional fields are omitted.
type GetPortfolioValueParams struct {
	User         string   `json:"user"`
	ConditionIDs []string `json:"condition_ids"`
}

// GetPortfolioValue reads public Data API information.
func (c *Client) GetPortfolioValue(ctx context.Context, p GetPortfolioValueParams) (*models.PortfolioValue, error) {
	query := map[string]any{}
	query["user"] = p.User
	if p.ConditionIDs != nil {
		query["condition_ids"] = append(make([]string, 0, len(p.ConditionIDs)), p.ConditionIDs...)
	}
	if err := normalizeQuery("get_portfolio_value", query); err != nil {
		return nil, err
	}
	return getData(ctx, c, "/v2/value", query, parseObject[models.PortfolioValue](false))
}

// GetUserStatsParams selects filters for GetUserStats. Nil optional fields are omitted.
type GetUserStatsParams struct {
	User string `json:"user"`
}

// GetUserStats reads public Data API information.
func (c *Client) GetUserStats(ctx context.Context, p GetUserStatsParams) (*models.UserStats, error) {
	query := map[string]any{}
	query["user"] = p.User
	if err := normalizeQuery("get_user_stats", query); err != nil {
		return nil, err
	}
	return getData(ctx, c, "/v2/user-stats", query, parseObject[models.UserStats](true))
}

// GetUserPnLParams selects filters for GetUserPnL. Nil optional fields are omitted.
type GetUserPnLParams struct {
	User     string  `json:"user"`
	Interval *string `json:"interval"`
	Fidelity *string `json:"fidelity"`
}

// GetUserPnL reads public Data API information.
func (c *Client) GetUserPnL(ctx context.Context, p GetUserPnLParams) (*models.UserPnLSeries, error) {
	query := map[string]any{}
	query["user"] = p.User
	put(query, "interval", p.Interval)
	put(query, "fidelity", p.Fidelity)
	if err := normalizeQuery("get_user_pnl", query); err != nil {
		return nil, err
	}
	return getData(ctx, c, "/v2/user-pnl", query, parseObject[models.UserPnLSeries](false))
}

// GetUserVolumeParams selects filters for GetUserVolume. Nil optional fields are omitted.
type GetUserVolumeParams struct {
	User        string `json:"user"`
	Start       *int64 `json:"start"`
	End         *int64 `json:"end"`
	FullHistory bool   `json:"full_history"`
}

// GetUserVolume reads public Data API information.
func (c *Client) GetUserVolume(ctx context.Context, p GetUserVolumeParams) (*models.UserVolume, error) {
	query := map[string]any{}
	query["user"] = p.User
	put(query, "start", p.Start)
	put(query, "end", p.End)
	query["full_history"] = p.FullHistory
	if err := normalizeQuery("get_user_volume", query); err != nil {
		return nil, err
	}
	return getData(ctx, c, "/v2/user-volume", query, parseObject[models.UserVolume](false))
}

// ListMarketHoldersParams selects filters for ListMarketHolders. Nil optional fields are omitted.
type ListMarketHoldersParams struct {
	ConditionIDs []string `json:"condition_ids"`
	MinBalance   *float64 `json:"min_balance"`
	IncludePnL   *bool    `json:"include_pnl"`
	PageSize     *int     `json:"page_size"`
}

// ListMarketHolders reads public Data API information.
func (c *Client) ListMarketHolders(p ListMarketHoldersParams) (*pagination.Paginator[models.MetaHolder], error) {
	query := map[string]any{}
	if p.ConditionIDs != nil {
		query["condition_ids"] = append(make([]string, 0, len(p.ConditionIDs)), p.ConditionIDs...)
	}
	put(query, "min_balance", p.MinBalance)
	put(query, "include_pnl", p.IncludePnL)
	if err := normalizeQuery("list_market_holders", query); err != nil {
		return nil, err
	}
	maximum := 1000
	if p.IncludePnL != nil && *p.IncludePnL {
		maximum = 100
	}
	return listData(c, "/v2/holders", query, p.PageSize, maximum, parseList[models.MetaHolder])
}

// GetOpenInterestsParams selects filters for GetOpenInterests. Nil optional fields are omitted.
type GetOpenInterestsParams struct {
	ConditionIDs []string `json:"condition_ids"`
}

// GetOpenInterests reads public Data API information.
func (c *Client) GetOpenInterests(ctx context.Context, p GetOpenInterestsParams) ([]models.OpenInterest, error) {
	query := map[string]any{}
	if p.ConditionIDs != nil {
		query["condition_ids"] = append(make([]string, 0, len(p.ConditionIDs)), p.ConditionIDs...)
	}
	if err := normalizeQuery("get_open_interests", query); err != nil {
		return nil, err
	}
	return getData(ctx, c, "/v2/oi", query, parseList[models.OpenInterest])
}

// GetEventLiveVolumeParams selects filters for GetEventLiveVolume. Nil optional fields are omitted.
type GetEventLiveVolumeParams struct {
	EventIDs []int64 `json:"event_ids"`
}

// GetEventLiveVolume reads public Data API information.
func (c *Client) GetEventLiveVolume(ctx context.Context, p GetEventLiveVolumeParams) (*models.LiveVolume, error) {
	query := map[string]any{}
	if p.EventIDs != nil {
		query["event_ids"] = append(make([]int64, 0, len(p.EventIDs)), p.EventIDs...)
	}
	if err := normalizeQuery("get_event_live_volume", query); err != nil {
		return nil, err
	}
	return getData(ctx, c, "/v2/live-volume", query, parseObject[models.LiveVolume](false))
}

// GetTokenReferencesParams selects filters for GetTokenReferences. Nil optional fields are omitted.
type GetTokenReferencesParams struct {
	AssetIDs     []string `json:"asset_ids"`
	ConditionIDs []string `json:"condition_ids"`
}

// GetTokenReferences reads public Data API information.
func (c *Client) GetTokenReferences(ctx context.Context, p GetTokenReferencesParams) ([]models.TokenReference, error) {
	query := map[string]any{}
	if p.AssetIDs != nil {
		query["asset_ids"] = append(make([]string, 0, len(p.AssetIDs)), p.AssetIDs...)
	}
	if p.ConditionIDs != nil {
		query["condition_ids"] = append(make([]string, 0, len(p.ConditionIDs)), p.ConditionIDs...)
	}
	if err := normalizeQuery("get_token_references", query); err != nil {
		return nil, err
	}
	return getData(ctx, c, "/v2/tokens", query, parseList[models.TokenReference])
}

// GetResolutionsParams selects filters for GetResolutions. Nil optional fields are omitted.
type GetResolutionsParams struct {
	QuestionID   *string  `json:"question_id"`
	ConditionIDs []string `json:"condition_ids"`
	EventIDs     []int64  `json:"event_ids"`
}

// GetResolutions reads public Data API information.
func (c *Client) GetResolutions(ctx context.Context, p GetResolutionsParams) ([]models.Resolution, error) {
	query := map[string]any{}
	put(query, "question_id", p.QuestionID)
	if p.ConditionIDs != nil {
		query["condition_ids"] = append(make([]string, 0, len(p.ConditionIDs)), p.ConditionIDs...)
	}
	if p.EventIDs != nil {
		query["event_ids"] = append(make([]int64, 0, len(p.EventIDs)), p.EventIDs...)
	}
	if err := normalizeQuery("get_resolutions", query); err != nil {
		return nil, err
	}
	return getData(ctx, c, "/v2/resolutions", query, parseList[models.Resolution])
}

// ListTraderLeaderboardParams selects filters for ListTraderLeaderboard. Nil optional fields are omitted.
type ListTraderLeaderboardParams struct {
	Category *string `json:"category"`
	Window   *string `json:"window"`
	SortBy   *string `json:"sort_by"`
	PageSize *int    `json:"page_size"`
}

// ListTraderLeaderboard reads public Data API information.
func (c *Client) ListTraderLeaderboard(p ListTraderLeaderboardParams) (*pagination.Paginator[models.TraderLeaderboardEntry], error) {
	query := map[string]any{}
	put(query, "category", p.Category)
	put(query, "window", p.Window)
	put(query, "sort_by", p.SortBy)
	if err := normalizeQuery("list_trader_leaderboard", query); err != nil {
		return nil, err
	}
	return listData(c, "/v2/leaderboard", query, p.PageSize, 1000, parseList[models.TraderLeaderboardEntry])
}

// GetTraderLeaderboardStandingParams selects filters for GetTraderLeaderboardStanding. Nil optional fields are omitted.
type GetTraderLeaderboardStandingParams struct {
	User     string  `json:"user"`
	Category *string `json:"category"`
	Window   *string `json:"window"`
}

// GetTraderLeaderboardStanding reads public Data API information.
func (c *Client) GetTraderLeaderboardStanding(ctx context.Context, p GetTraderLeaderboardStandingParams) (*models.TraderLeaderboardStanding, error) {
	query := map[string]any{}
	query["user"] = p.User
	put(query, "category", p.Category)
	put(query, "window", p.Window)
	if err := normalizeQuery("get_trader_leaderboard_standing", query); err != nil {
		return nil, err
	}
	return getData(ctx, c, "/v2/leaderboard", query, parseObject[models.TraderLeaderboardStanding](true))
}

// ListBiggestWinnersParams selects filters for ListBiggestWinners. Nil optional fields are omitted.
type ListBiggestWinnersParams struct {
	Category *string `json:"category"`
	Window   *string `json:"window"`
	PageSize *int    `json:"page_size"`
}

// ListBiggestWinners reads public Data API information.
func (c *Client) ListBiggestWinners(p ListBiggestWinnersParams) (*pagination.Paginator[models.BiggestWinner], error) {
	query := map[string]any{}
	put(query, "category", p.Category)
	put(query, "window", p.Window)
	if err := normalizeQuery("list_biggest_winners", query); err != nil {
		return nil, err
	}
	return listData(c, "/v2/biggest-winners", query, p.PageSize, 1000, models.ParseBiggestWinners)
}

// ListBuilderLeaderboardParams selects filters for ListBuilderLeaderboard. Nil optional fields are omitted.
type ListBuilderLeaderboardParams struct {
	Window   *string `json:"window"`
	PageSize *int    `json:"page_size"`
}

// ListBuilderLeaderboard reads public Data API information.
func (c *Client) ListBuilderLeaderboard(p ListBuilderLeaderboardParams) (*pagination.Paginator[models.BuilderStanding], error) {
	query := map[string]any{}
	put(query, "window", p.Window)
	if err := normalizeQuery("list_builder_leaderboard", query); err != nil {
		return nil, err
	}
	return listData(c, "/v2/builders/leaderboard", query, p.PageSize, 1000, parseList[models.BuilderStanding])
}

// GetBuilderVolumesParams selects filters for GetBuilderVolumes. Nil optional fields are omitted.
type GetBuilderVolumesParams struct {
	Interval    *string `json:"interval"`
	BucketLimit *int    `json:"bucket_limit"`
}

// GetBuilderVolumes reads public Data API information.
func (c *Client) GetBuilderVolumes(ctx context.Context, p GetBuilderVolumesParams) ([]models.BuilderVolumePoint, error) {
	query := map[string]any{}
	put(query, "interval", p.Interval)
	put(query, "bucket_limit", p.BucketLimit)
	if err := normalizeQuery("get_builder_volumes", query); err != nil {
		return nil, err
	}
	return getData(ctx, c, "/v2/builders/volume", query, parseList[models.BuilderVolumePoint])
}

// DownloadAccountingSnapshotParams selects filters for DownloadAccountingSnapshot. Nil optional fields are omitted.
type DownloadAccountingSnapshotParams struct {
	User string `json:"user"`
}

// DownloadAccountingSnapshot reads public Data API information.
func (c *Client) DownloadAccountingSnapshot(ctx context.Context, p DownloadAccountingSnapshotParams) ([]byte, error) {
	query := map[string]any{}
	query["user"] = p.User
	if err := normalizeQuery("download_accounting_snapshot", query); err != nil {
		return nil, err
	}
	values, err := queryValues(query)
	if err != nil {
		return nil, err
	}
	return c.transport.Request(ctx, c.transport.DataURL, "GET", "/v1/accounting/snapshot", values, nil)
}
