package trading

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
)

// GetOrderScoring reports whether an order currently earns liquidity rewards.
func (c *Client) GetOrderScoring(ctx context.Context, id string) (bool, error) {
	if err := nonempty(id); err != nil {
		return false, err
	}
	raw, err := c.request(ctx, "GET", "/order-scoring", url.Values{"order_id": {id}}, nil)
	if err != nil {
		return false, err
	}
	return requiredBool(raw, "scoring")
}

// GetOrdersScoring reports reward eligibility for each requested order.
func (c *Client) GetOrdersScoring(ctx context.Context, ids []string) (map[string]bool, error) {
	if len(ids) == 0 {
		return nil, input("order IDs must be nonempty")
	}
	if err := nonempty(ids...); err != nil {
		return nil, err
	}
	raw, err := c.request(ctx, "POST", "/orders-scoring", nil, ids)
	if err != nil {
		return nil, err
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, response(err)
	}
	result := make(map[string]bool, len(fields))
	for key := range fields {
		value, parseErr := codec.Required[bool](fields, key)
		if parseErr != nil {
			return nil, response(parseErr)
		}
		result[key] = value
	}
	return result, nil
}

// ListUserEarningsForDay lists reward earnings for an ISO calendar date.
func (c *Client) ListUserEarningsForDay(date string) (*pagination.Paginator[models.UserEarning], error) {
	q, err := c.dateQuery(date)
	if err != nil {
		return nil, err
	}
	return accountPages[models.UserEarning](c, "/rewards/user", q, false), nil
}

// GetTotalEarningsForUserForDayParams selects a reward date and optional sponsorship category.
type GetTotalEarningsForUserForDayParams struct {
	Date      string
	Sponsored *bool
}

// GetTotalEarningsForUserForDay retrieves the account's reward totals.
func (c *Client) GetTotalEarningsForUserForDay(ctx context.Context, p GetTotalEarningsForUserForDayParams) ([]models.TotalUserEarning, error) {
	q, err := c.dateQuery(p.Date)
	if err != nil {
		return nil, err
	}
	if p.Sponsored != nil {
		q.Set("sponsored", strconv.FormatBool(*p.Sponsored))
	}
	return listRequest[models.TotalUserEarning](ctx, c, "GET", "/rewards/user/total", q, nil)
}

// ListUserEarningsAndMarketsConfigParams selects earned rewards and associated markets.
type ListUserEarningsAndMarketsConfigParams struct {
	Date          string
	NoCompetition *bool
	OrderBy       *string
	Position      *string
	PageSize      *int
}

// ListUserEarningsAndMarketsConfig lists account earnings alongside their reward configurations.
func (c *Client) ListUserEarningsAndMarketsConfig(p ListUserEarningsAndMarketsConfigParams) (*pagination.Paginator[models.UserRewardsEarning], error) {
	q, err := c.dateQuery(p.Date)
	if err != nil {
		return nil, err
	}
	size := 100
	if p.PageSize != nil {
		size = *p.PageSize
	}
	if size < 1 || size > 500 {
		return nil, input("page size must be between 1 and 500")
	}
	q.Set("page_size", strconv.Itoa(size))
	if p.NoCompetition != nil {
		q.Set("no_competition", strconv.FormatBool(*p.NoCompetition))
	}
	for key, value := range map[string]*string{"order_by": p.OrderBy, "position": p.Position} {
		if value != nil {
			if err = nonempty(*value); err != nil {
				return nil, err
			}
			q.Set(key, *value)
		}
	}
	return accountPages[models.UserRewardsEarning](c, "/rewards/user/markets", q, true), nil
}

// GetRewardPercentages returns the account's share of rewards for each condition.
func (c *Client) GetRewardPercentages(ctx context.Context) (map[models.ConditionID]float64, error) {
	raw, err := c.request(ctx, "GET", "/rewards/user/percentages", c.signatureQuery(), nil)
	if err != nil {
		return nil, err
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, response(err)
	}
	result := make(map[models.ConditionID]float64, len(fields))
	for key, value := range fields {
		if string(value) == "null" {
			return nil, response(fmt.Errorf("null reward percentage"))
		}
		var number float64
		if err = json.Unmarshal(value, &number); err != nil {
			return nil, response(err)
		}
		result[models.ConditionID(key)] = number
	}
	return result, nil
}
func (c *Client) dateQuery(date string) (url.Values, error) {
	if !validDate(date) {
		return nil, input("date must be a valid ISO calendar date")
	}
	q := c.signatureQuery()
	q.Set("date", date)
	return q, nil
}

// validDate follows Python date.fromisoformat, including compact and ISO week dates.
func validDate(value string) bool {
	for _, layout := range []string{"2006-01-02", "20060102"} {
		if parsed, err := time.Parse(layout, value); err == nil && parsed.Year() > 0 && parsed.Format(layout) == value {
			return true
		}
	}
	compact := strings.ReplaceAll(value, "-", "")
	if len(compact) != 7 && len(compact) != 8 {
		return false
	}
	if compact[4] != 'W' {
		return false
	}
	year, yearErr := strconv.Atoi(compact[:4])
	week, weekErr := strconv.Atoi(compact[5:7])
	day := 1
	var dayErr error
	if len(compact) == 8 {
		day, dayErr = strconv.Atoi(compact[7:])
	}
	if yearErr != nil || weekErr != nil || dayErr != nil || year < 1 || year > 9999 || week < 1 || week > 53 || day < 1 || day > 7 {
		return false
	}
	// Reject mixed compact and extended spellings accepted by simple hyphen removal.
	expected := fmt.Sprintf("%04d-W%02d", year, week)
	short := fmt.Sprintf("%04dW%02d", year, week)
	if len(compact) == 8 {
		expected += "-" + strconv.Itoa(day)
		short += strconv.Itoa(day)
	}
	if value != expected && value != short {
		return false
	}
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	weekday := (int(jan4.Weekday()) + 6) % 7
	target := jan4.AddDate(0, 0, -weekday+7*(week-1)+day-1)
	actualYear, actualWeek := target.ISOWeek()
	return actualYear == year && actualWeek == week && target.Year() <= 9999
}
