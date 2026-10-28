package gamma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/paging"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
)

// MarketQuery selects exactly one nonempty ID, slug or Polymarket URL.
type MarketQuery struct {
	ID         models.MarketID `json:"id"`
	Slug       string          `json:"slug"`
	URL        string          `json:"url"`
	IncludeTag *bool           `json:"include_tag"`
	Locale     *string         `json:"locale"`
}

// TagMatch controls how multiple tag filters are combined.
type TagMatch string

// Supported tag matching modes.
const (
	TagMatchAny TagMatch = "any"
	TagMatchAll TagMatch = "all"
)

// ListMarketsParams contains the complete upstream market filter set.
// Nil pointers omit filters; date/time filters use ISO strings. Empty slices
// omit filters except Versions, where an explicitly empty slice is invalid.
type ListMarketsParams struct {
	Ascending            *bool                          `json:"ascending"`
	Closed               *bool                          `json:"closed"`
	ClobTokenIDs         []string                       `json:"clob_token_ids"`
	ConditionIDs         []string                       `json:"condition_ids"`
	CYOM                 *bool                          `json:"cyom"`
	Decimalized          *bool                          `json:"decimalized"`
	EndDateMax           *string                        `json:"end_date_max"`
	EndDateMin           *string                        `json:"end_date_min"`
	GameID               *string                        `json:"game_id"`
	IDs                  []int64                        `json:"ids"`
	IncludeTag           *bool                          `json:"include_tag"`
	LiquidityNumMax      *float64                       `json:"liquidity_num_max"`
	LiquidityNumMin      *float64                       `json:"liquidity_num_min"`
	Locale               *string                        `json:"locale"`
	MarketMakerAddresses []string                       `json:"market_maker_addresses"`
	Order                *string                        `json:"order"`
	OnchainEventIDs      []string                       `json:"onchain_event_ids"`
	PositionIDs          []string                       `json:"position_ids"`
	QuestionIDs          []string                       `json:"question_ids"`
	RelatedTags          *bool                          `json:"related_tags"`
	RequestIDs           []string                       `json:"request_ids"`
	ResolutionStatus     *models.MarketResolutionStatus `json:"resolution_status"`
	RFQEnabled           *bool                          `json:"rfq_enabled"`
	RewardsMinSize       *float64                       `json:"rewards_min_size"`
	Slugs                []string                       `json:"slug"`
	SportsMarketTypes    []string                       `json:"sports_market_types"`
	StartDateMax         *string                        `json:"start_date_max"`
	StartDateMin         *string                        `json:"start_date_min"`
	TagID                *int64                         `json:"tag_id"`
	TagMatch             *TagMatch                      `json:"tag_match"`
	UMAResolutionStatus  *string                        `json:"uma_resolution_status"`
	Versions             []models.ProtocolVersion       `json:"version"`
	VolumeNumMax         *float64                       `json:"volume_num_max"`
	VolumeNumMin         *float64                       `json:"volume_num_min"`
	PageSize             *int                           `json:"page_size"` // Nil defaults to 20.
}

// GetMarket fetches one binary market with typed outcomes, trading and resolution metadata.
func (c *Client) GetMarket(ctx context.Context, query MarketQuery) (*models.Market, error) {
	path, err := entityPath("market", string(query.ID), query.Slug, query.URL)
	if err != nil {
		return nil, err
	}
	params := make(map[string]any)
	paging.QueryOption(params, "include_tag", query.IncludeTag)
	paging.QueryOption(params, "locale", query.Locale)
	data, err := c.transport.Request(ctx, c.transport.GammaURL, http.MethodGet, path, paging.QueryValues(params), nil)
	if err != nil {
		return nil, err
	}
	market, err := codec.DecodeResponse[models.Market](data)
	if err != nil {
		return nil, err
	}
	return &market, nil
}

// ListMarkets lazily lists binary markets. Nonbinary markets are omitted but
// their server continuation cursor is preserved. Query inputs are copied.
func (c *Client) ListMarkets(params ListMarketsParams) (*pagination.Paginator[models.Market], error) {
	query, pageSize, err := marketsQuery(params)
	if err != nil {
		return nil, err
	}
	return paging.KeysetPaginator(c.transport, "gamma", c.transport.GammaURL, "/markets/keyset", "after_cursor", query, pageSize, parseMarketsPage), nil
}

func marketsQuery(p ListMarketsParams) (map[string]any, int, error) {
	if p.TagMatch != nil && *p.TagMatch != TagMatchAny && *p.TagMatch != TagMatchAll {
		return nil, 0, &sdkerrors.UserInputError{Message: "tag_match must be any or all"}
	}
	if p.ResolutionStatus != nil && *p.ResolutionStatus != models.ResolutionInactive && *p.ResolutionStatus != models.ResolutionActive && *p.ResolutionStatus != models.ResolutionResolved {
		return nil, 0, &sdkerrors.UserInputError{Message: "invalid resolution_status"}
	}
	params := make(map[string]any)
	if err := queryVersions(params, p.Versions); err != nil {
		return nil, 0, err
	}
	for key, values := range map[string][]string{
		"clob_token_ids": p.ClobTokenIDs, "condition_ids": p.ConditionIDs, "market_maker_address": p.MarketMakerAddresses,
		"onchain_event_ids": p.OnchainEventIDs, "position_ids": p.PositionIDs, "question_ids": p.QuestionIDs,
		"request_ids": p.RequestIDs, "slug": p.Slugs, "sports_market_types": p.SportsMarketTypes,
	} {
		if (key == "onchain_event_ids" || key == "request_ids") && len(values) > 100 {
			return nil, 0, &sdkerrors.UserInputError{Message: key + " accepts at most 100 values"}
		}
		if len(values) > 0 {
			params[key] = append([]string(nil), values...)
		}
	}
	if len(p.IDs) > 0 {
		params["id"] = append([]int64(nil), p.IDs...)
	}
	paging.QueryOption(params, "ascending", p.Ascending)
	paging.QueryOption(params, "closed", p.Closed)
	paging.QueryOption(params, "cyom", p.CYOM)
	paging.QueryOption(params, "decimalized", p.Decimalized)
	paging.QueryOption(params, "end_date_max", p.EndDateMax)
	paging.QueryOption(params, "end_date_min", p.EndDateMin)
	paging.QueryOption(params, "game_id", p.GameID)
	paging.QueryOption(params, "include_tag", p.IncludeTag)
	paging.QueryOption(params, "liquidity_num_max", p.LiquidityNumMax)
	paging.QueryOption(params, "liquidity_num_min", p.LiquidityNumMin)
	paging.QueryOption(params, "locale", p.Locale)
	paging.QueryOption(params, "related_tags", p.RelatedTags)
	if p.ResolutionStatus != nil {
		params["resolution_status"] = string(*p.ResolutionStatus)
	}
	paging.QueryOption(params, "rfq_enabled", p.RFQEnabled)
	paging.QueryOption(params, "rewards_min_size", p.RewardsMinSize)
	paging.QueryOption(params, "start_date_max", p.StartDateMax)
	paging.QueryOption(params, "start_date_min", p.StartDateMin)
	paging.QueryOption(params, "tag_id", p.TagID)
	if p.TagMatch != nil {
		params["tag_match"] = string(*p.TagMatch)
	}
	paging.QueryOption(params, "uma_resolution_status", p.UMAResolutionStatus)
	paging.QueryOption(params, "volume_num_max", p.VolumeNumMax)
	paging.QueryOption(params, "volume_num_min", p.VolumeNumMin)
	if p.Order != nil {
		tokens := strings.Split(*p.Order, ",")
		for i, token := range tokens {
			token = strings.TrimSpace(token)
			switch token {
			case "volume":
				token = "volumeNum"
			case "liquidity":
				token = "liquidityNum"
			}
			tokens[i] = token
		}
		params["order"] = strings.Join(tokens, ",")
	}
	pageSize := 20
	if p.PageSize != nil {
		pageSize = *p.PageSize
	}
	if pageSize < 1 {
		return nil, 0, &sdkerrors.UserInputError{Message: "page_size must be positive"}
	}
	return params, pageSize, nil
}

func entityPath(kind, id, slug, rawURL string) (string, error) {
	count := 0
	for _, value := range []string{id, slug, rawURL} {
		if value != "" {
			count++
		}
	}
	if count != 1 {
		return "", &sdkerrors.UserInputError{Message: "provide exactly one of id, slug or url"}
	}
	if id != "" {
		return "/" + kind + "s/" + escapeSegment(id), nil
	}
	if rawURL != "" {
		u, err := url.Parse(rawURL)
		if err != nil || u.Scheme != "https" || u.User != nil || (u.Host != "polymarket.com" && u.Host != "www.polymarket.com") {
			return "", &sdkerrors.UserInputError{Message: "expected a Polymarket HTTPS URL"}
		}
		segments := strings.FieldsFunc(u.EscapedPath(), func(r rune) bool { return r == '/' })
		switch {
		case kind == "market" && len(segments) >= 2 && len(segments) <= 3 && segments[0] == "event":
			slug = segments[len(segments)-1]
		case len(segments) == 2 && segments[0] == kind:
			slug = segments[1]
		default:
			return "", &sdkerrors.UserInputError{Message: "expected a Polymarket " + kind + " URL"}
		}
	}
	return "/" + kind + "s/slug/" + escapeSegment(slug), nil
}

func escapeSegment(value string) string {
	var escaped strings.Builder
	for _, b := range []byte(value) {
		if b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || strings.ContainsRune("-._~", rune(b)) {
			escaped.WriteByte(b)
		} else {
			fmt.Fprintf(&escaped, "%%%02X", b)
		}
	}
	return escaped.String()
}

func parseMarketsPage(data []byte) ([]models.Market, *string, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return nil, nil, err
	}
	raw, err := codec.Required[[]json.RawMessage](fields, "markets")
	if err != nil {
		return nil, nil, err
	}
	items := make([]models.Market, 0, len(raw))
	for _, entry := range raw {
		marketFields, objectErr := codec.Object(entry)
		if objectErr != nil {
			return nil, nil, objectErr
		}
		outcomes, parseErr := codec.GammaStrings(marketFields["outcomes"])
		if parseErr != nil {
			return nil, nil, parseErr
		}
		if len(outcomes) != 2 {
			continue
		}
		market, decodeErr := codec.DecodeResponse[models.Market](entry)
		if decodeErr != nil {
			return nil, nil, decodeErr
		}
		items = append(items, market)
	}
	next, err := paging.ServerNextCursor(fields)
	return items, next, err
}
