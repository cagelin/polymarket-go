package gamma

import (
	"context"
	"encoding/json"
	"net/http"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/paging"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
)

// EventQuery selects exactly one nonempty ID, slug or event URL.
type EventQuery struct {
	ID               models.EventID `json:"id"`
	Slug             string         `json:"slug"`
	URL              string         `json:"url"`
	IncludeBestLines *bool          `json:"include_best_lines"`
	IncludeChat      *bool          `json:"include_chat"`
	IncludeTemplate  *bool          `json:"include_template"`
	Locale           *string        `json:"locale"`
}

// Recurrence filters repeating events.
type Recurrence string

// Supported event recurrence filters.
const (
	RecurrenceDaily   Recurrence = "daily"
	RecurrenceWeekly  Recurrence = "weekly"
	RecurrenceMonthly Recurrence = "monthly"
)

// ListEventsParams contains the complete upstream event filter set.
// Closed defaults to false. Date/time filters use ISO strings. PageSize defaults
// to 20; nil pointers and empty slices otherwise omit filters.
type ListEventsParams struct {
	Ascending        *bool                    `json:"ascending"`
	Closed           bool                     `json:"closed"`
	CYOM             *bool                    `json:"cyom"`
	EndDateMax       *string                  `json:"end_date_max"`
	EndDateMin       *string                  `json:"end_date_min"`
	Ended            *bool                    `json:"ended"`
	EventDate        *string                  `json:"event_date"`
	EventWeek        *int64                   `json:"event_week"`
	ExcludeTagIDs    []int64                  `json:"exclude_tag_ids"`
	Featured         *bool                    `json:"featured"`
	FeaturedOrder    *bool                    `json:"featured_order"`
	GameIDs          []int64                  `json:"game_ids"`
	IDs              []int64                  `json:"ids"`
	IncludeBestLines *bool                    `json:"include_best_lines"`
	IncludeChat      *bool                    `json:"include_chat"`
	IncludeChildren  *bool                    `json:"include_children"`
	IncludeTemplate  *bool                    `json:"include_template"`
	LiquidityMax     *float64                 `json:"liquidity_max"`
	LiquidityMin     *float64                 `json:"liquidity_min"`
	Live             *bool                    `json:"live"`
	Locale           *string                  `json:"locale"`
	Order            *string                  `json:"order"`
	ParentEventID    *int64                   `json:"parent_event_id"`
	PartnerSlug      *string                  `json:"partner_slug"`
	Recurrence       *Recurrence              `json:"recurrence"`
	RelatedTags      *bool                    `json:"related_tags"`
	SeriesIDs        []int64                  `json:"series_ids"`
	Slugs            []string                 `json:"slug"`
	StartDateMax     *string                  `json:"start_date_max"`
	StartDateMin     *string                  `json:"start_date_min"`
	StartTimeMax     *string                  `json:"start_time_max"`
	StartTimeMin     *string                  `json:"start_time_min"`
	TagIDs           []int64                  `json:"tag_ids"`
	TagMatch         *TagMatch                `json:"tag_match"`
	TagSlug          *string                  `json:"tag_slug"`
	TitleSearch      *string                  `json:"title_search"`
	Versions         []models.ProtocolVersion `json:"version"`
	VolumeMax        *float64                 `json:"volume_max"`
	VolumeMin        *float64                 `json:"volume_min"`
	PageSize         *int                     `json:"page_size"`
}

// GetEvent fetches an event and its binary child markets.
func (c *Client) GetEvent(ctx context.Context, query EventQuery) (*models.Event, error) {
	path, err := entityPath("event", string(query.ID), query.Slug, query.URL)
	if err != nil {
		return nil, err
	}
	params := make(map[string]any)
	paging.QueryOption(params, "include_best_lines", query.IncludeBestLines)
	paging.QueryOption(params, "include_chat", query.IncludeChat)
	paging.QueryOption(params, "include_template", query.IncludeTemplate)
	paging.QueryOption(params, "locale", query.Locale)
	data, err := c.transport.Request(ctx, c.transport.GammaURL, http.MethodGet, path, paging.QueryValues(params), nil)
	if err != nil {
		return nil, err
	}
	event, err := codec.DecodeResponse[models.Event](data)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// ListEvents lazily lists events with query-bound continuation cursors.
func (c *Client) ListEvents(p ListEventsParams) (*pagination.Paginator[models.Event], error) {
	if err := validateRecurrence(p.Recurrence); err != nil {
		return nil, err
	}
	if p.TagMatch != nil && *p.TagMatch != TagMatchAny && *p.TagMatch != TagMatchAll {
		return nil, &sdkerrors.UserInputError{Message: "tag_match must be any or all"}
	}
	params := map[string]any{"closed": p.Closed}
	if err := queryVersions(params, p.Versions); err != nil {
		return nil, err
	}
	paging.QueryOption(params, "ascending", p.Ascending)
	paging.QueryOption(params, "cyom", p.CYOM)
	paging.QueryOption(params, "end_date_max", p.EndDateMax)
	paging.QueryOption(params, "end_date_min", p.EndDateMin)
	paging.QueryOption(params, "ended", p.Ended)
	paging.QueryOption(params, "event_date", p.EventDate)
	paging.QueryOption(params, "event_week", p.EventWeek)
	if len(p.ExcludeTagIDs) > 0 {
		params["exclude_tag_id"] = append([]int64(nil), p.ExcludeTagIDs...)
	}
	paging.QueryOption(params, "featured", p.Featured)
	paging.QueryOption(params, "featured_order", p.FeaturedOrder)
	if len(p.GameIDs) > 0 {
		params["game_id"] = append([]int64(nil), p.GameIDs...)
	}
	if len(p.IDs) > 0 {
		params["id"] = append([]int64(nil), p.IDs...)
	}
	paging.QueryOption(params, "include_best_lines", p.IncludeBestLines)
	paging.QueryOption(params, "include_chat", p.IncludeChat)
	paging.QueryOption(params, "include_children", p.IncludeChildren)
	paging.QueryOption(params, "include_template", p.IncludeTemplate)
	paging.QueryOption(params, "liquidity_max", p.LiquidityMax)
	paging.QueryOption(params, "liquidity_min", p.LiquidityMin)
	paging.QueryOption(params, "live", p.Live)
	paging.QueryOption(params, "locale", p.Locale)
	paging.QueryOption(params, "order", p.Order)
	paging.QueryOption(params, "parent_event_id", p.ParentEventID)
	paging.QueryOption(params, "partner_slug", p.PartnerSlug)
	if p.Recurrence != nil {
		params["recurrence"] = string(*p.Recurrence)
	}
	paging.QueryOption(params, "related_tags", p.RelatedTags)
	if len(p.SeriesIDs) > 0 {
		params["series_id"] = append([]int64(nil), p.SeriesIDs...)
	}
	if len(p.Slugs) > 0 {
		params["slug"] = append([]string(nil), p.Slugs...)
	}
	paging.QueryOption(params, "start_date_max", p.StartDateMax)
	paging.QueryOption(params, "start_date_min", p.StartDateMin)
	paging.QueryOption(params, "start_time_max", p.StartTimeMax)
	paging.QueryOption(params, "start_time_min", p.StartTimeMin)
	if len(p.TagIDs) > 0 {
		params["tag_id"] = append([]int64(nil), p.TagIDs...)
	}
	if p.TagMatch != nil {
		params["tag_match"] = string(*p.TagMatch)
	}
	paging.QueryOption(params, "tag_slug", p.TagSlug)
	paging.QueryOption(params, "title_search", p.TitleSearch)
	paging.QueryOption(params, "volume_max", p.VolumeMax)
	paging.QueryOption(params, "volume_min", p.VolumeMin)
	size := 20
	if p.PageSize != nil {
		size = *p.PageSize
	}
	if size < 1 {
		return nil, &sdkerrors.UserInputError{Message: "page_size must be positive"}
	}
	return paging.KeysetPaginator(c.transport, "gamma", c.transport.GammaURL, "/events/keyset", "after_cursor", params, size, parseEventsPage), nil
}

func validateRecurrence(value *Recurrence) error {
	if value != nil && *value != RecurrenceDaily && *value != RecurrenceWeekly && *value != RecurrenceMonthly {
		return &sdkerrors.UserInputError{Message: "recurrence must be daily, weekly or monthly"}
	}
	return nil
}

func queryVersions(params map[string]any, values []models.ProtocolVersion) error {
	if values == nil {
		return nil
	}
	if len(values) < 1 || len(values) > 2 {
		return &sdkerrors.UserInputError{Message: "version requires one or two values"}
	}
	versions := make([]string, len(values))
	for i, version := range values {
		if version != models.ProtocolV1 && version != models.ProtocolV2 {
			return &sdkerrors.UserInputError{Message: "invalid protocol version"}
		}
		versions[i] = string(version)
	}
	params["version"] = versions
	return nil
}

func parseEventsPage(data []byte) ([]models.Event, *string, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return nil, nil, err
	}
	items, err := codec.Required[[]models.Event](fields, "events")
	if err != nil {
		return nil, nil, err
	}
	next, err := paging.ServerNextCursor(fields)
	return items, next, err
}

// GetEventTags fetches an event's complete tag references.
func (c *Client) GetEventTags(ctx context.Context, id models.EventID) ([]models.TagReference, error) {
	return c.entityTags(ctx, "events", string(id))
}

// GetMarketTags fetches a market's complete tag references.
func (c *Client) GetMarketTags(ctx context.Context, id models.MarketID) ([]models.TagReference, error) {
	return c.entityTags(ctx, "markets", string(id))
}

func (c *Client) entityTags(ctx context.Context, kind, id string) ([]models.TagReference, error) {
	if id == "" {
		return nil, &sdkerrors.UserInputError{Message: "id must be nonempty"}
	}
	data, err := c.transport.Request(ctx, c.transport.GammaURL, http.MethodGet, "/"+kind+"/"+escapeSegment(id)+"/tags", nil, nil)
	if err != nil {
		return nil, err
	}
	var items []models.TagReference
	if err = json.Unmarshal(data, &items); err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid tags response", Cause: err}
	}
	if items == nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "expected tags array"}
	}
	return items, nil
}
