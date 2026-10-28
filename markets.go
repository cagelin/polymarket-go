package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// MarketQuery selects exactly one nonempty ID, slug or Polymarket URL.
type MarketQuery struct {
	ID         MarketID `json:"id"`
	Slug       string   `json:"slug"`
	URL        string   `json:"url"`
	IncludeTag *bool    `json:"include_tag"`
	Locale     *string  `json:"locale"`
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
	Ascending            *bool                   `json:"ascending"`
	Closed               *bool                   `json:"closed"`
	ClobTokenIDs         []string                `json:"clob_token_ids"`
	ConditionIDs         []string                `json:"condition_ids"`
	CYOM                 *bool                   `json:"cyom"`
	Decimalized          *bool                   `json:"decimalized"`
	EndDateMax           *string                 `json:"end_date_max"`
	EndDateMin           *string                 `json:"end_date_min"`
	GameID               *string                 `json:"game_id"`
	IDs                  []int64                 `json:"ids"`
	IncludeTag           *bool                   `json:"include_tag"`
	LiquidityNumMax      *float64                `json:"liquidity_num_max"`
	LiquidityNumMin      *float64                `json:"liquidity_num_min"`
	Locale               *string                 `json:"locale"`
	MarketMakerAddresses []string                `json:"market_maker_addresses"`
	Order                *string                 `json:"order"`
	OnchainEventIDs      []string                `json:"onchain_event_ids"`
	PositionIDs          []string                `json:"position_ids"`
	QuestionIDs          []string                `json:"question_ids"`
	RelatedTags          *bool                   `json:"related_tags"`
	RequestIDs           []string                `json:"request_ids"`
	ResolutionStatus     *MarketResolutionStatus `json:"resolution_status"`
	RFQEnabled           *bool                   `json:"rfq_enabled"`
	RewardsMinSize       *float64                `json:"rewards_min_size"`
	Slugs                []string                `json:"slug"`
	SportsMarketTypes    []string                `json:"sports_market_types"`
	StartDateMax         *string                 `json:"start_date_max"`
	StartDateMin         *string                 `json:"start_date_min"`
	TagID                *int64                  `json:"tag_id"`
	TagMatch             *TagMatch               `json:"tag_match"`
	UMAResolutionStatus  *string                 `json:"uma_resolution_status"`
	Versions             []ProtocolVersion       `json:"version"`
	VolumeNumMax         *float64                `json:"volume_num_max"`
	VolumeNumMin         *float64                `json:"volume_num_min"`
	PageSize             *int                    `json:"page_size"` // Nil defaults to 20.
}

// GetMarket fetches one binary market with typed outcomes, trading and resolution metadata.
func (c *PublicClient) GetMarket(ctx context.Context, query MarketQuery) (*Market, error) {
	path, err := marketPath(query)
	if err != nil {
		return nil, err
	}
	params := make(map[string]any)
	queryOption(params, "include_tag", query.IncludeTag)
	queryOption(params, "locale", query.Locale)
	data, err := c.requestAt(ctx, c.gammaURL, http.MethodGet, path, queryValues(params), nil)
	if err != nil {
		return nil, err
	}
	market, err := decodeResponse[Market](data)
	if err != nil {
		return nil, err
	}
	return &market, nil
}

// ListMarkets lazily lists binary markets. Nonbinary markets are omitted but
// their server continuation cursor is preserved. Query inputs are copied.
func (c *PublicClient) ListMarkets(params ListMarketsParams) (*Paginator[Market], error) {
	query, pageSize, err := marketsQuery(params)
	if err != nil {
		return nil, err
	}
	return keysetPaginator(c, "gamma", c.gammaURL, "/markets/keyset", "after_cursor", query, pageSize, parseMarketsPage), nil
}

func queryOption[T string | bool | int | int64 | float64](params map[string]any, key string, value *T) {
	if value != nil {
		params[key] = *value
	}
}

func marketsQuery(p ListMarketsParams) (map[string]any, int, error) {
	if p.TagMatch != nil && *p.TagMatch != TagMatchAny && *p.TagMatch != TagMatchAll {
		return nil, 0, &UserInputError{Message: "tag_match must be any or all"}
	}
	if p.ResolutionStatus != nil && *p.ResolutionStatus != ResolutionInactive && *p.ResolutionStatus != ResolutionActive && *p.ResolutionStatus != ResolutionResolved {
		return nil, 0, &UserInputError{Message: "invalid resolution_status"}
	}
	params := make(map[string]any)
	if p.Versions != nil {
		if len(p.Versions) < 1 || len(p.Versions) > 2 {
			return nil, 0, &UserInputError{Message: "version requires one or two values"}
		}
		versions := make([]string, len(p.Versions))
		for i, version := range p.Versions {
			if version != ProtocolV1 && version != ProtocolV2 {
				return nil, 0, &UserInputError{Message: "invalid protocol version"}
			}
			versions[i] = string(version)
		}
		params["version"] = versions
	}
	for key, values := range map[string][]string{
		"clob_token_ids": p.ClobTokenIDs, "condition_ids": p.ConditionIDs, "market_maker_address": p.MarketMakerAddresses,
		"onchain_event_ids": p.OnchainEventIDs, "position_ids": p.PositionIDs, "question_ids": p.QuestionIDs,
		"request_ids": p.RequestIDs, "slug": p.Slugs, "sports_market_types": p.SportsMarketTypes,
	} {
		if (key == "onchain_event_ids" || key == "request_ids") && len(values) > 100 {
			return nil, 0, &UserInputError{Message: key + " accepts at most 100 values"}
		}
		if len(values) > 0 {
			params[key] = append([]string(nil), values...)
		}
	}
	if len(p.IDs) > 0 {
		params["id"] = append([]int64(nil), p.IDs...)
	}
	queryOption(params, "ascending", p.Ascending)
	queryOption(params, "closed", p.Closed)
	queryOption(params, "cyom", p.CYOM)
	queryOption(params, "decimalized", p.Decimalized)
	queryOption(params, "end_date_max", p.EndDateMax)
	queryOption(params, "end_date_min", p.EndDateMin)
	queryOption(params, "game_id", p.GameID)
	queryOption(params, "include_tag", p.IncludeTag)
	queryOption(params, "liquidity_num_max", p.LiquidityNumMax)
	queryOption(params, "liquidity_num_min", p.LiquidityNumMin)
	queryOption(params, "locale", p.Locale)
	queryOption(params, "related_tags", p.RelatedTags)
	if p.ResolutionStatus != nil {
		params["resolution_status"] = string(*p.ResolutionStatus)
	}
	queryOption(params, "rfq_enabled", p.RFQEnabled)
	queryOption(params, "rewards_min_size", p.RewardsMinSize)
	queryOption(params, "start_date_max", p.StartDateMax)
	queryOption(params, "start_date_min", p.StartDateMin)
	queryOption(params, "tag_id", p.TagID)
	if p.TagMatch != nil {
		params["tag_match"] = string(*p.TagMatch)
	}
	queryOption(params, "uma_resolution_status", p.UMAResolutionStatus)
	queryOption(params, "volume_num_max", p.VolumeNumMax)
	queryOption(params, "volume_num_min", p.VolumeNumMin)
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
		return nil, 0, &UserInputError{Message: "page_size must be positive"}
	}
	return params, pageSize, nil
}

func marketPath(query MarketQuery) (string, error) {
	count := 0
	for _, value := range []string{string(query.ID), query.Slug, query.URL} {
		if value != "" {
			count++
		}
	}
	if count != 1 {
		return "", &UserInputError{Message: "provide exactly one of id, slug or url"}
	}
	if query.ID != "" {
		return "/markets/" + escapeSegment(string(query.ID)), nil
	}
	slug := query.Slug
	if query.URL != "" {
		u, err := url.Parse(query.URL)
		if err != nil || u.Scheme != "https" || u.User != nil || (u.Host != "polymarket.com" && u.Host != "www.polymarket.com") {
			return "", &UserInputError{Message: "expected a Polymarket HTTPS URL"}
		}
		segments := strings.FieldsFunc(u.EscapedPath(), func(r rune) bool { return r == '/' })
		switch {
		case len(segments) >= 2 && len(segments) <= 3 && segments[0] == "event":
			slug = segments[len(segments)-1]
		case len(segments) == 2 && segments[0] == "market":
			slug = segments[1]
		default:
			return "", &UserInputError{Message: "expected a Polymarket market URL"}
		}
	}
	return "/markets/slug/" + escapeSegment(slug), nil
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

func parseMarketsPage(data []byte) ([]Market, *string, error) {
	fields, err := object(data)
	if err != nil {
		return nil, nil, err
	}
	raw, err := required[[]json.RawMessage](fields, "markets")
	if err != nil {
		return nil, nil, err
	}
	items := make([]Market, 0, len(raw))
	for _, entry := range raw {
		marketFields, objectErr := object(entry)
		if objectErr != nil {
			return nil, nil, objectErr
		}
		outcomes, parseErr := gammaStrings(marketFields["outcomes"])
		if parseErr != nil {
			return nil, nil, parseErr
		}
		if len(outcomes) != 2 {
			continue
		}
		market, decodeErr := decodeResponse[Market](entry)
		if decodeErr != nil {
			return nil, nil, decodeErr
		}
		items = append(items, market)
	}
	next, err := serverNextCursor(fields)
	return items, next, err
}
