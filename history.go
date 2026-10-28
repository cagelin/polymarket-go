package polymarket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cockroachdb/apd/v3"
)

// PriceHistoryInterval selects a historical window relative to server time.
type PriceHistoryInterval string

// Supported history windows. Max and all follow the upstream server semantics.
const (
	HistoryMax      PriceHistoryInterval = "max"
	HistoryAll      PriceHistoryInterval = "all"
	HistoryMonth    PriceHistoryInterval = "1m"
	HistoryWeek     PriceHistoryInterval = "1w"
	HistoryDay      PriceHistoryInterval = "1d"
	HistorySixHours PriceHistoryInterval = "6h"
	HistoryHour     PriceHistoryInterval = "1h"
)

// PriceHistoryParams selects exactly one of Interval, Start or AsOf.
// Times are Unix seconds; use Time.Unix() for a time.Time input.
type PriceHistoryParams struct {
	AssetID       ClobAssetID           `json:"asset_id"`
	Interval      *PriceHistoryInterval `json:"interval"`
	Start         *int64                `json:"start"`
	End           *int64                `json:"end"`
	AsOf          *int64                `json:"as_of"`
	BucketSeconds *int                  `json:"bucket_seconds"`
	PageSize      *int                  `json:"page_size"` // Nil defaults to 10000.
}

// PriceHistoryPoint is a price per share; resolution zero denotes an exact tick.
type PriceHistoryPoint struct {
	Timestamp         time.Time    `json:"timestamp"`
	Price             *apd.Decimal `json:"price"`
	ResolutionSeconds int64        `json:"resolution_seconds"`
}

// ListPriceHistory returns a lazy oldest-first history paginator.
// Explicit start/end windows are limited to 15 days. Data reads retry HTTP 429
// at most twice when the suggested delay is no more than five seconds.
func (c *PublicClient) ListPriceHistory(params PriceHistoryParams) (*Paginator[PriceHistoryPoint], error) {
	query, pageSize, err := historyQuery(params, time.Now())
	if err != nil {
		return nil, err
	}
	return keysetPaginator(c, "data", c.dataURL, "/v2/prices-history", "cursor", query, pageSize, parseHistoryPage), nil
}

func historyQuery(p PriceHistoryParams, now time.Time) (map[string]any, int, error) {
	fail := func(message string) (map[string]any, int, error) { return nil, 0, &UserInputError{Message: message} }
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
		case HistoryMax, HistoryAll, HistoryMonth, HistoryWeek, HistoryDay, HistorySixHours, HistoryHour:
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
			case HistoryMax, HistoryAll, HistoryMonth:
				minimum = 600
			case HistoryWeek:
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

// UnmarshalJSON validates required prices, timestamps and resolutions.
func (p *PriceHistoryPoint) UnmarshalJSON(data []byte) error {
	fields, err := object(data)
	if err != nil {
		return err
	}
	raw, err := required[json.RawMessage](fields, "timestamp")
	if err != nil {
		return err
	}
	var value string
	if len(raw) > 0 && raw[0] == '"' {
		if err = json.Unmarshal(raw, &value); err != nil {
			return err
		}
	} else {
		value = string(raw)
	}
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds < -62135596800 || seconds >= 253402300800 {
		return fmt.Errorf("invalid epoch seconds")
	}
	whole, fraction := math.Modf(seconds)
	result := PriceHistoryPoint{Timestamp: time.Unix(int64(whole), int64(math.RoundToEven(fraction*1e6))*1000).UTC()}
	if result.Price, err = numericDecimal(fields, "price", false); err != nil {
		return err
	}
	if result.ResolutionSeconds, err = integerField(fields, "resolution_seconds"); err != nil {
		return err
	}
	*p = result
	return nil
}

func parseHistoryPage(data []byte) ([]PriceHistoryPoint, *string, error) {
	fields, err := object(data)
	if err != nil {
		return nil, nil, err
	}
	items, err := required[[]PriceHistoryPoint](fields, "data")
	if err != nil {
		return nil, nil, err
	}
	pagination, err := required[map[string]json.RawMessage](fields, "pagination")
	if err != nil {
		return nil, nil, err
	}
	hasMore, err := required[bool](pagination, "has_more")
	if err != nil {
		return nil, nil, err
	}
	if _, ok := pagination["next_cursor"]; !ok {
		return nil, nil, fmt.Errorf("missing next_cursor")
	}
	next, err := serverNextCursor(pagination)
	if err != nil {
		return nil, nil, err
	}
	if hasMore != (next != nil) {
		return nil, nil, fmt.Errorf("has_more and next_cursor disagree")
	}
	return items, next, nil
}

func (c *PublicClient) dataRead(ctx context.Context, path string, query url.Values) ([]byte, error) {
	for attempt := 0; ; attempt++ {
		data, err := c.requestAt(ctx, c.dataURL, http.MethodGet, path, query, nil)
		var rate *RateLimitError
		if !errors.As(err, &rate) || attempt >= 2 {
			return data, err
		}
		delay := 1.0
		if rate.RetryAfter != nil {
			delay = *rate.RetryAfter
		}
		if delay > 5 {
			return nil, err
		}
		timer := time.NewTimer(time.Duration(math.Max(0, delay) * float64(time.Second)))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, &TransportError{Cause: ctx.Err()}
		case <-timer.C:
		}
	}
}
