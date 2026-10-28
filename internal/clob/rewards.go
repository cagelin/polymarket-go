package clob

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"slices"
	"strings"
	"time"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/paging"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
)

// ListRewardMarketsParams filters public reward-bearing markets.
type ListRewardMarketsParams struct {
	Query         *string  `json:"query"`
	TagSlugs      []string `json:"tag_slugs"`
	MarketID      *int64   `json:"market_id"`
	EventIDs      []int64  `json:"event_ids"`
	EventTitle    *string  `json:"event_title"`
	OrderBy       *string  `json:"order_by"`
	SortDirection *string  `json:"sort_direction"`
	MinVolume24hr *float64 `json:"min_volume_24hr"`
	MaxVolume24hr *float64 `json:"max_volume_24hr"`
	MinSpread     *float64 `json:"min_spread"`
	MaxSpread     *float64 `json:"max_spread"`
	MinPrice      *float64 `json:"min_price"`
	MaxPrice      *float64 `json:"max_price"`
	PageSize      *int     `json:"page_size"`
}

// ListRewardMarkets lists markets and available reward configurations.
func (c *Client) ListRewardMarkets(p ListRewardMarketsParams) (*pagination.Paginator[models.RewardMarket], error) {
	size := 100
	if p.PageSize != nil {
		size = *p.PageSize
	}
	if size < 1 || size > 500 {
		return nil, &sdkerrors.UserInputError{Message: "page_size must be between 1 and 500"}
	}
	query := map[string]any{"page_size": size}
	paging.QueryOption(query, "q", p.Query)
	paging.QueryOption(query, "event_title", p.EventTitle)
	paging.QueryOption(query, "market_id", p.MarketID)
	paging.QueryOption(query, "order_by", p.OrderBy)
	paging.QueryOption(query, "position", p.SortDirection)
	if p.MarketID != nil && *p.MarketID <= 0 {
		return nil, &sdkerrors.UserInputError{Message: "market_id must be positive"}
	}
	if p.EventIDs != nil {
		values := make([]int64, len(p.EventIDs))
		copy(values, p.EventIDs)
		for _, value := range values {
			if value <= 0 {
				return nil, &sdkerrors.UserInputError{Message: "event_ids must be positive"}
			}
		}
		query["event_id"] = values
	}
	if p.TagSlugs != nil {
		values := append(make([]string, 0, len(p.TagSlugs)), p.TagSlugs...)
		for _, value := range values {
			if value == "" {
				return nil, &sdkerrors.UserInputError{Message: "tag_slugs must be nonempty"}
			}
		}
		query["tag_slug"] = values
	}
	if p.SortDirection != nil && *p.SortDirection != "ASC" && *p.SortDirection != "DESC" {
		return nil, &sdkerrors.UserInputError{Message: "sort_direction must be ASC or DESC"}
	}
	if p.OrderBy != nil && (*p.OrderBy == "" || !slices.Contains(strings.Fields("market_id created_at volume_24hr spread competitiveness max_spread min_size question one_day_price_change rate_per_day price end_date start_date reward_end_date"), *p.OrderBy)) {
		return nil, &sdkerrors.UserInputError{Message: "invalid reward market sort"}
	}
	for key, value := range map[string]*float64{"min_volume_24hr": p.MinVolume24hr, "max_volume_24hr": p.MaxVolume24hr, "min_spread": p.MinSpread, "max_spread": p.MaxSpread, "min_price": p.MinPrice, "max_price": p.MaxPrice} {
		if value != nil {
			if math.IsNaN(*value) || math.IsInf(*value, 0) || *value < 0 {
				return nil, &sdkerrors.UserInputError{Message: key + " must be finite and nonnegative"}
			}
			query[key] = *value
		}
	}
	return rawPaginator[models.RewardMarket](c, "/rewards/markets/multi", query, false), nil
}

// ListCurrentRewardsParams selects sponsored or native rewards.
type ListCurrentRewardsParams struct {
	Sponsored *bool `json:"sponsored"`
}

// ListCurrentRewards lists active reward configurations.
func (c *Client) ListCurrentRewards(p ListCurrentRewardsParams) (*pagination.Paginator[models.CurrentReward], error) {
	query := map[string]any{}
	paging.QueryOption(query, "sponsored", p.Sponsored)
	return rawPaginator[models.CurrentReward](c, "/rewards/markets/current", query, false), nil
}

// ListMarketRewardsParams selects rewards for one market condition.
type ListMarketRewardsParams struct {
	ConditionID models.ConditionID `json:"condition_id"`
	Sponsored   *bool              `json:"sponsored"`
}

// ListMarketRewards lists rewards for a market condition.
func (c *Client) ListMarketRewards(p ListMarketRewardsParams) (*pagination.Paginator[models.MarketReward], error) {
	condition, err := codec.Condition(string(p.ConditionID), "feed")
	if err != nil {
		return nil, &sdkerrors.UserInputError{Message: err.Error()}
	}
	query := map[string]any{}
	paging.QueryOption(query, "sponsored", p.Sponsored)
	return rawPaginator[models.MarketReward](c, "/rewards/markets/"+condition, query, false), nil
}

// GetCurrentRebatesParams selects a maker and calendar day.
type GetCurrentRebatesParams struct {
	Date         string            `json:"date"`
	MakerAddress models.EVMAddress `json:"maker_address"`
}

// GetCurrentRebates returns exact USDC maker rebates for a day.
func (c *Client) GetCurrentRebates(ctx context.Context, p GetCurrentRebatesParams) ([]models.RebatedFee, error) {
	parsed, err := time.Parse("2006-01-02", p.Date)
	if err != nil || parsed.Format("2006-01-02") != p.Date {
		return nil, &sdkerrors.UserInputError{Message: "date must be YYYY-MM-DD"}
	}
	address := string(p.MakerAddress)
	if len(address) != 42 || !codec.ValidHex(address) || strings.Trim(address[2:], "0") == "" {
		return nil, &sdkerrors.UserInputError{Message: "maker_address must be a nonzero EVM address"}
	}
	raw, err := c.transport.Request(ctx, c.transport.CLOBURL, "GET", "/rebates/current", paging.QueryValues(map[string]any{"date": p.Date, "maker_address": address}), nil)
	if err != nil {
		return nil, err
	}
	if string(raw) == "null" {
		return []models.RebatedFee{}, nil
	}
	values, err := codec.ParseList[models.RebatedFee](raw)
	if err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid rebates response", Cause: err}
	}
	return values, nil
}

func rawPaginator[T any](c *Client, path string, params map[string]any, requireCursor bool) *pagination.Paginator[T] {
	base := paging.QueryValues(params)
	return pagination.New(func(ctx context.Context, cursor *string) (pagination.Page[T], error) {
		query := make(url.Values, len(base)+1)
		for key, values := range base {
			query[key] = append([]string(nil), values...)
		}
		if cursor != nil {
			if *cursor == "" {
				return pagination.Page[T]{}, &sdkerrors.UserInputError{Message: "cursor must be nonempty"}
			}
			query.Set("next_cursor", *cursor)
		}
		raw, err := c.transport.Request(ctx, c.transport.CLOBURL, "GET", path, query, nil)
		if err != nil {
			return pagination.Page[T]{}, err
		}
		page, err := parseRawPage[T](raw, requireCursor)
		if err != nil {
			return page, &sdkerrors.UnexpectedResponseError{Message: "invalid CLOB page", Cause: err}
		}
		return page, nil
	}, nil)
}
func parseRawPage[T any](raw []byte, requireCursor bool) (pagination.Page[T], error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return pagination.Page[T]{}, err
	}
	items, err := codec.Required[[]T](fields, "data")
	if err != nil {
		return pagination.Page[T]{}, err
	}
	if requireCursor {
		if _, err = codec.Required[string](fields, "next_cursor"); err != nil {
			return pagination.Page[T]{}, err
		}
	}
	next, err := paging.ServerNextCursor(fields)
	if err != nil {
		return pagination.Page[T]{}, err
	}
	if next != nil && *next == "LTE=" {
		next = nil
	}
	return pagination.Page[T]{Items: items, HasMore: next != nil, NextCursor: next}, nil
}

// ListBuilderTradesParams filters trades attributed to a builder.
type ListBuilderTradesParams struct {
	BuilderCode string  `json:"builder_code"`
	Market      *string `json:"market"`
	AssetID     *string `json:"asset_id"`
	ID          *string `json:"id"`
	After       *string `json:"after"`
	Before      *string `json:"before"`
}

// ListBuilderTrades lists trades attributed to a builder code.
func (c *Client) ListBuilderTrades(p ListBuilderTradesParams) (*pagination.Paginator[models.BuilderTrade], error) {
	if len(p.BuilderCode) != 66 || !codec.ValidHex(p.BuilderCode) {
		return nil, &sdkerrors.UserInputError{Message: "builder_code must be a 32-byte hex string"}
	}
	query := map[string]any{"builder_code": p.BuilderCode}
	for key, value := range map[string]*string{"market": p.Market, "asset_id": p.AssetID, "id": p.ID, "after": p.After, "before": p.Before} {
		if value != nil {
			if *value == "" {
				return nil, &sdkerrors.UserInputError{Message: fmt.Sprintf("%s must be nonempty", key)}
			}
			query[key] = *value
		}
	}
	return rawPaginator[models.BuilderTrade](c, "/builder/trades", query, true), nil
}
