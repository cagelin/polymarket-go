package clob

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/models"
)

// PriceHistoryQuery selects an interval or an explicit time window.
// Market is the outcome token ID. StartTs/EndTs are Unix seconds; Fidelity is minutes.
// Explicit windows require both endpoints and are limited to 15 days by this SDK.
type PriceHistoryQuery struct {
	Market   models.ClobAssetID           `json:"market"`
	StartTs  *int64                       `json:"startTs"`
	EndTs    *int64                       `json:"endTs"`
	Interval *models.PriceHistoryInterval `json:"interval"`
	Fidelity *int                         `json:"fidelity"`
}

// GetCLOBPriceHistory reads one CLOB /prices-history response without pagination.
// An empty response does not establish historical coverage or completeness.
func (c *Client) GetCLOBPriceHistory(ctx context.Context, params PriceHistoryQuery) ([]models.CLOBPriceHistoryPoint, error) {
	query, err := clobHistoryQuery(params)
	if err != nil {
		return nil, err
	}
	body, err := c.request(ctx, http.MethodGet, "/prices-history", query, nil)
	if err != nil {
		return nil, err
	}
	fields, err := codec.Object(body)
	if err == nil {
		var points []models.CLOBPriceHistoryPoint
		points, err = codec.Required[[]models.CLOBPriceHistoryPoint](fields, "history")
		if err == nil {
			return points, nil
		}
	}
	return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid CLOB price history", Cause: err}
}

func clobHistoryQuery(p PriceHistoryQuery) (url.Values, error) {
	fail := func(message string) (url.Values, error) { return nil, &sdkerrors.UserInputError{Message: message} }
	if p.Market == "" {
		return fail("market token ID is required")
	}
	query := url.Values{"market": {string(p.Market)}}
	if p.Interval != nil {
		if p.StartTs != nil || p.EndTs != nil {
			return fail("interval cannot be combined with startTs or endTs")
		}
		switch *p.Interval {
		case models.HistoryMax, models.HistoryAll, models.HistoryMonth, models.HistoryWeek, models.HistoryDay, models.HistorySixHours, models.HistoryHour:
		default:
			return fail("invalid CLOB history interval")
		}
		query.Set("interval", string(*p.Interval))
	} else {
		if p.StartTs == nil || p.EndTs == nil {
			return fail("provide interval or both startTs and endTs")
		}
		if *p.StartTs <= 0 || *p.EndTs > 253402300799 || *p.EndTs < *p.StartTs || *p.EndTs-*p.StartTs > 15*86400 {
			return fail("CLOB history window must be ordered, positive and at most 15 days")
		}
		query.Set("startTs", strconv.FormatInt(*p.StartTs, 10))
		query.Set("endTs", strconv.FormatInt(*p.EndTs, 10))
	}
	if p.Fidelity != nil {
		if *p.Fidelity < 1 {
			return fail("fidelity must be positive minutes")
		}
		query.Set("fidelity", strconv.Itoa(*p.Fidelity))
	}
	return query, nil
}
