package gamma

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/paging"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
)

// SearchSort selects the event ordering within search results.
type SearchSort string

// Supported search orderings.
const (
	SearchByVolume      SearchSort = "volume"
	SearchByVolume24hr  SearchSort = "volume_24hr"
	SearchByLiquidity   SearchSort = "liquidity"
	SearchByCompetitive SearchSort = "competitive"
	SearchByClosedTime  SearchSort = "closed_time"
	SearchByStartDate   SearchSort = "start_date"
	SearchByEndDate     SearchSort = "end_date"
)

// SearchParams controls public search. PageSize defaults to 10 per result type.
type SearchParams struct {
	Query             string      `json:"q"`
	Ascending         *bool       `json:"ascending"`
	Cache             *bool       `json:"cache"`
	EventsStatus      *string     `json:"events_status"`
	EventsTag         []string    `json:"events_tag"`
	ExcludeTagIDs     []int64     `json:"exclude_tag_ids"`
	KeepClosedMarkets *int64      `json:"keep_closed_markets"`
	Optimized         *bool       `json:"optimized"`
	Presets           []string    `json:"presets"`
	Recurrence        *Recurrence `json:"recurrence"`
	SearchProfiles    *bool       `json:"search_profiles"`
	SearchTags        *bool       `json:"search_tags"`
	Sort              *SearchSort `json:"sort"`
	PageSize          *int        `json:"page_size"`
}

// Search lazily searches events, tags and profiles. Each page has exactly one
// SearchResults item, keeping the three result categories together. Resuming
// a cursor preserves its original page size, matching the upstream SDK.
func (c *Client) Search(p SearchParams) (*pagination.Paginator[models.SearchResults], error) {
	if p.Query == "" {
		return nil, &sdkerrors.UserInputError{Message: "q must be nonempty"}
	}
	if err := validateRecurrence(p.Recurrence); err != nil {
		return nil, err
	}
	if p.Sort != nil {
		switch *p.Sort {
		case SearchByVolume, SearchByVolume24hr, SearchByLiquidity, SearchByCompetitive, SearchByClosedTime, SearchByStartDate, SearchByEndDate:
		default:
			return nil, &sdkerrors.UserInputError{Message: "invalid search sort"}
		}
	}
	params := map[string]any{"q": p.Query}
	paging.QueryOption(params, "ascending", p.Ascending)
	paging.QueryOption(params, "cache", p.Cache)
	paging.QueryOption(params, "events_status", p.EventsStatus)
	paging.QueryOption(params, "keep_closed_markets", p.KeepClosedMarkets)
	paging.QueryOption(params, "optimized", p.Optimized)
	paging.QueryOption(params, "search_profiles", p.SearchProfiles)
	paging.QueryOption(params, "search_tags", p.SearchTags)
	if p.Recurrence != nil {
		params["recurrence"] = string(*p.Recurrence)
	}
	if p.Sort != nil {
		params["sort"] = string(*p.Sort)
	}
	for key, values := range map[string][]string{"events_tag": p.EventsTag, "presets": p.Presets} {
		if len(values) > 0 {
			params[key] = append([]string(nil), values...)
		}
	}
	if len(p.ExcludeTagIDs) > 0 {
		params["exclude_tag_id"] = append([]int64(nil), p.ExcludeTagIDs...)
	}
	size := 10
	if p.PageSize != nil {
		size = *p.PageSize
	}
	if size < 1 {
		return nil, &sdkerrors.UserInputError{Message: "page_size must be positive"}
	}
	fingerprint := paging.QueryFingerprint(params)
	return pagination.New(func(ctx context.Context, cursor *string) (pagination.Page[models.SearchResults], error) {
		pageNumber, effectiveSize := 1, size
		if cursor != nil {
			var cursorErr error
			pageNumber, effectiveSize, cursorErr = decodeSearchCursor(*cursor, fingerprint)
			if cursorErr != nil {
				return pagination.Page[models.SearchResults]{}, cursorErr
			}
		}
		query := paging.QueryValues(params)
		query.Set("page", strconv.Itoa(pageNumber))
		query.Set("limit_per_type", strconv.Itoa(effectiveSize))
		data, err := c.transport.Request(ctx, c.transport.GammaURL, http.MethodGet, "/public-search", query, nil)
		if err != nil {
			return pagination.Page[models.SearchResults]{}, err
		}
		result, more, total, err := parseSearchPage(data)
		if err != nil {
			return pagination.Page[models.SearchResults]{}, &sdkerrors.UnexpectedResponseError{Message: "invalid search response", Cause: err}
		}
		page := pagination.Page[models.SearchResults]{Items: []models.SearchResults{result}, HasMore: more, TotalCount: total}
		if more {
			// Keep overflow visible instead of emitting an unusable negative cursor.
			if pageNumber == int(^uint(0)>>1) {
				return pagination.Page[models.SearchResults]{}, &sdkerrors.UnexpectedResponseError{Message: "search page number exceeds supported range"}
			}
			next := encodeSearchCursor(fingerprint, pageNumber+1, effectiveSize)
			page.NextCursor = &next
		}
		return page, nil
	}, nil), nil
}

func parseSearchPage(data []byte) (models.SearchResults, bool, *int64, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return models.SearchResults{}, false, nil, err
	}
	var more bool
	var total *int64
	if raw, present := fields["pagination"]; present && string(raw) != "null" {
		pageMetadata, paginationErr := codec.Object(raw)
		if paginationErr != nil {
			return models.SearchResults{}, false, nil, paginationErr
		}
		if value, exists := pageMetadata["hasMore"]; exists && string(value) != "null" {
			if err = json.Unmarshal(value, &more); err != nil {
				return models.SearchResults{}, false, nil, err
			}
		}
		if value, exists := pageMetadata["totalResults"]; exists && string(value) != "null" {
			if err = json.Unmarshal(value, &total); err != nil {
				return models.SearchResults{}, false, nil, err
			}
		}
	}
	result, err := codec.DecodeResponse[models.SearchResults](data)
	return result, more, total, err
}

func encodeSearchCursor(fingerprint string, page, size int) string {
	payload := `{"f":` + paging.PythonString(fingerprint) + `,"p":"/public-search","pg":` + strconv.Itoa(page) + `,"s":` + strconv.Itoa(size) + `,"svc":"gamma","v":1}`
	return base64.StdEncoding.EncodeToString([]byte(payload))
}

func decodeSearchCursor(cursor, fingerprint string) (int, int, error) {
	failure := &sdkerrors.UserInputError{Message: "search cursor is invalid or belongs to a different query"}
	if strings.ContainsAny(cursor, "\r\n") {
		return 0, 0, failure
	}
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, 0, failure
	}
	var payload struct {
		Version     json.RawMessage `json:"v"`
		Service     string          `json:"svc"`
		Path        string          `json:"p"`
		Fingerprint string          `json:"f"`
		Page        int             `json:"pg"`
		Size        int             `json:"s"`
	}
	if err = json.Unmarshal(data, &payload); err != nil {
		return 0, 0, failure
	}
	var version float64
	if string(payload.Version) == "true" {
		version = 1
	} else if err = json.Unmarshal(payload.Version, &version); err != nil {
		return 0, 0, failure
	}
	if version != 1 || payload.Service != "gamma" || payload.Path != "/public-search" || payload.Fingerprint != fingerprint || payload.Page < 1 || payload.Size < 1 {
		return 0, 0, failure
	}
	return payload.Page, payload.Size, nil
}
