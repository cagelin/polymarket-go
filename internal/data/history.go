package data

import (
	"encoding/json"
	"fmt"
	"time"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/paging"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
)

// PriceHistoryParams selects exactly one of Interval, Start or AsOf.
// Times are Unix seconds; use Time.Unix() for a time.Time input.
type PriceHistoryParams struct {
	AssetID       models.ClobAssetID           `json:"asset_id"`
	Interval      *models.PriceHistoryInterval `json:"interval"`
	Start         *int64                       `json:"start"`
	End           *int64                       `json:"end"`
	AsOf          *int64                       `json:"as_of"`
	BucketSeconds *int                         `json:"bucket_seconds"`
	PageSize      *int                         `json:"page_size"` // Nil defaults to 10000.
}

// ListPriceHistory returns a lazy oldest-first history paginator.
// Explicit start/end windows are limited to 15 days. Data reads retry HTTP 429
// at most twice when the suggested delay is no more than five seconds.
func (c *Client) ListPriceHistory(params PriceHistoryParams) (*pagination.Paginator[models.PriceHistoryPoint], error) {
	query, pageSize, err := historyQuery(params, time.Now())
	if err != nil {
		return nil, err
	}
	return paging.KeysetPaginator(c.transport, "data", c.transport.DataURL, "/v2/prices-history", "cursor", query, pageSize, parseHistoryPage), nil
}

func historyQuery(p PriceHistoryParams, now time.Time) (map[string]any, int, error) {
	fail := func(message string) (map[string]any, int, error) {
		return nil, 0, &sdkerrors.UserInputError{Message: message}
	}
	if p.AssetID == "" {
		return fail("asset_id must be nonempty")
	}
	selectors := 0
	if p.Interval != nil {
		selectors++
	}
	if p.Start != nil {
		selectors++
	}
	if p.AsOf != nil {
		selectors++
	}
	if selectors != 1 {
		return fail("provide exactly one of interval, start or as_of")
	}
	if p.End != nil && p.Start == nil {
		return fail("end requires start")
	}
	query := map[string]any{"token_id": string(p.AssetID)}
	if p.Interval != nil {
		switch *p.Interval {
		case models.HistoryMax, models.HistoryAll, models.HistoryMonth, models.HistoryWeek, models.HistoryDay, models.HistorySixHours, models.HistoryHour:
		default:
			return fail("invalid history interval")
		}
		query["interval"] = string(*p.Interval)
	}
	for key, value := range map[string]*int64{"start": p.Start, "end": p.End, "as_of": p.AsOf} {
		if value != nil {
			if *value <= 0 || *value > 253402300799 {
				return fail("timestamp is outside the supported range")
			}
			query[key] = *value
		}
	}
	if p.Start != nil {
		until := now.Unix()
		if p.End != nil {
			until = *p.End
		}
		if until < *p.Start || until-*p.Start > 15*86400 {
			return fail("price history window must not exceed 15 days or end before start")
		}
	}
	if p.AsOf != nil && (p.BucketSeconds != nil || p.PageSize != nil) {
		return fail("as_of forbids bucket_seconds and page_size")
	}
	if p.BucketSeconds != nil {
		minimum := 60
		if p.Interval != nil {
			switch *p.Interval {
			case models.HistoryMax, models.HistoryAll, models.HistoryMonth:
				minimum = 600
			case models.HistoryWeek:
				minimum = 300
			}
		}
		if *p.BucketSeconds < minimum || *p.BucketSeconds > 86400 {
			return fail("bucket_seconds is outside the interval's supported range")
		}
		query["bucket_seconds"] = *p.BucketSeconds
	}
	pageSize := 10000
	if p.PageSize != nil {
		pageSize = *p.PageSize
	}
	if pageSize < 1 || pageSize > 10000 {
		return fail("page_size must be between 1 and 10000")
	}
	return query, pageSize, nil
}

func parseHistoryPage(data []byte) ([]models.PriceHistoryPoint, *string, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return nil, nil, err
	}
	items, err := codec.Required[[]models.PriceHistoryPoint](fields, "data")
	if err != nil {
		return nil, nil, err
	}
	_, next, err := parseHistoryPageMetadata(data)
	return items, next, err
}

func parseHistoryPageMetadata(data []byte) (map[string]json.RawMessage, *string, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return nil, nil, err
	}
	pageMetadata, err := codec.Required[map[string]json.RawMessage](fields, "pagination")
	if err != nil {
		return nil, nil, err
	}
	hasMore, err := codec.Required[bool](pageMetadata, "has_more")
	if err != nil {
		return nil, nil, err
	}
	if _, ok := pageMetadata["next_cursor"]; !ok {
		return nil, nil, fmt.Errorf("missing next_cursor")
	}
	next, err := paging.ServerNextCursor(pageMetadata)
	if err != nil {
		return nil, nil, err
	}
	if hasMore != (next != nil) {
		return nil, nil, fmt.Errorf("has_more and next_cursor disagree")
	}
	return pageMetadata, next, nil
}
