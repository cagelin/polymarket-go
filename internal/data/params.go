package data

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
)

func put[T any](query map[string]any, key string, value *T) {
	if value != nil {
		query[key] = *value
	}
}
func input(message string) error { return &sdkerrors.UserInputError{Message: message} }

// normalizeQuery applies the endpoint's public filter contract before any request.
func normalizeQuery(operation string, q map[string]any) error {
	if user, ok := q["user"]; ok && user == "" {
		return input("user is required")
	}
	if category, ok := q["category"]; ok && category == "" {
		return input("category must be nonempty")
	}
	if q["condition_id"] != nil && q["event_id"] != nil {
		return input("provide condition_id or event_id, not both")
	}
	grammar := "market"
	if operation == "list_activity" || operation == "list_trades" {
		grammar = "feed"
	}
	if operation == "list_combo_activity" || operation == "list_combo_positions" {
		grammar = "combo"
	}
	if operation == "get_token_references" {
		return tokenSelectors(q)
	}
	for _, key := range []string{"condition_id", "condition_ids"} {
		if values, ok := q[key].([]string); ok {
			normalized, err := conditions(values, grammar)
			if err != nil {
				return err
			}
			delete(q, key)
			q["condition_id"] = normalized
		}
	}
	for _, key := range []string{"event_id", "event_ids"} {
		if values, ok := q[key].([]int64); ok {
			normalized, err := events(values)
			if err != nil {
				return err
			}
			delete(q, key)
			q["event_id"] = normalized
		}
	}
	if operation == "get_resolutions" {
		selectors := 0
		for _, key := range []string{"question_id", "condition_id", "event_id"} {
			if _, ok := q[key]; ok {
				selectors++
			}
		}
		if selectors != 1 || q["question_id"] == "" {
			return input("provide exactly one nonempty resolution selector")
		}
		if values, ok := q["event_id"].([]string); ok && len(values) > 20 {
			return input("event_ids accepts at most 20 distinct values")
		}
	}
	if operation == "get_event_live_volume" && q["event_id"] == nil {
		return input("event_ids is required")
	}
	if operation == "list_market_holders" {
		values, ok := q["condition_id"].([]string)
		if !ok || len(values) == 0 {
			return input("condition_ids is required")
		}
		if q["include_pnl"] == true && len(values) != 1 {
			return input("include_pnl requires exactly one condition")
		}
	}
	if full, _ := q["full_history"].(bool); full {
		if q["start"] != nil || q["end"] != nil {
			return input("full_history cannot be combined with start or end")
		}
		if operation != "list_positions" {
			q["start"] = int64(1)
		}
	}
	delete(q, "full_history")
	for _, key := range []string{"start", "end", "updated_after", "updated_before"} {
		if value, ok := q[key].(int64); ok {
			minValue := int64(1)
			if strings.HasPrefix(key, "updated_") {
				minValue = 0
			}
			if value < minValue || value > 253402300799 {
				return input("timestamp is outside the supported range")
			}
		}
	}
	if after, ok := q["updated_after"].(int64); ok {
		if before, exists := q["updated_before"].(int64); exists && before < after {
			return input("updated_before must be at least updated_after")
		}
	}
	for _, key := range []string{"filter_amount", "min_balance"} {
		if value, ok := q[key].(float64); ok && (math.IsNaN(value) || math.IsInf(value, 0) || value < 0) {
			return input(key + " must be finite and nonnegative")
		}
	}
	choices := map[string]string{"side": "BUY SELL", "filter_type": "CASH TOKENS", "sort_direction": "ASC DESC", "window": "day week month all"}
	switch operation {
	case "list_activity":
		choices["activity_types"] = "TRADE SPLIT MERGE REDEEM REWARD CONVERSION MIGRATION DEPOSIT WITHDRAWAL YIELD MAKER_REBATE TAKER_REBATE REFERRAL_REWARD TIP"
		q["exclude_deposits_withdrawals"] = false
	case "list_positions":
		choices["status"] = "OPEN REDEEMABLE REDEEMABLE_LOST MERGEABLE CLOSED"
		choices["sort_by"] = "CURRENT_VALUE PRICE TOKENS UNREALIZED_PNL REALIZED_PNL TOTAL_PNL TIMESTAMP"
		if q["user"] == nil {
			values, _ := q["condition_id"].([]string)
			if len(values) != 1 || q["event_id"] != nil || q["status"] == "REDEEMABLE_LOST" || q["status"] == "MERGEABLE" {
				return input("provide user or exactly one condition_id; this filter may require user")
			}
		}
		if q["status"] == "CLOSED" && q["include_archived"] == true {
			return input("include_archived is invalid with CLOSED")
		}
		if title, ok := q["title"].(string); ok {
			if strings.TrimSpace(title) == "" {
				delete(q, "title")
			} else if utf8.RuneCountInString(title) > 200 {
				return input("title must contain at most 200 characters")
			}
		}
	case "list_combo_positions":
		choices["status"] = "OPEN REDEEMABLE PARTIAL RESOLVED_PARTIAL RESOLVED_WIN RESOLVED_LOSS"
		choices["sort_by"] = "FIRST_ENTRY ENTRY_COST CURRENT_VALUE UPDATED"
		if values, ok := q["status"].([]string); ok {
			values = distinct(values, false)
			if len(values) == 0 {
				return input("status must be nonempty")
			}
			if len(values) > 1 {
				for _, value := range values {
					if value == "REDEEMABLE" {
						return input("REDEEMABLE must be the only status")
					}
				}
			}
			q["status"] = values
		}
	case "get_user_pnl":
		choices["interval"] = "max all 1m 1w 1d 12h 6h"
		choices["fidelity"] = "1d 18h 12h 3h 1h"
	case "list_trader_leaderboard":
		choices["sort_by"] = "PNL VOLUME"
	case "get_builder_volumes":
		choices["interval"] = "day week month all"
		if value, ok := q["bucket_limit"].(int); ok && (value < 1 || value > 90) {
			return input("bucket_limit must be between 1 and 90")
		}
	}
	for key, allowed := range choices {
		if value, ok := q[key]; ok {
			values, sequence := value.([]string)
			if !sequence {
				text, _ := value.(string)
				values = []string{text}
			}
			for _, item := range values {
				if !strings.Contains(" "+allowed+" ", " "+item+" ") || item == "" {
					return input(fmt.Sprintf("invalid %s", key))
				}
			}
		}
	}
	for from, to := range map[string]string{"activity_types": "type", "window": "time_period", "bucket_limit": "limit"} {
		if value, ok := q[from]; ok {
			delete(q, from)
			q[to] = value
		}
	}
	for key, value := range q {
		if values, ok := value.([]string); ok {
			if len(values) == 0 {
				delete(q, key)
			} else {
				q[key] = strings.Join(values, ",")
			}
		}
	}
	return nil
}

func distinct(values []string, fold bool) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		key := value
		if fold {
			key = strings.ToLower(key)
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, value)
		}
	}
	return result
}
func conditions(values []string, grammar string) ([]string, error) {
	if len(values) == 0 {
		return nil, input("condition_id must be nonempty")
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		parsed, err := codec.Condition(value, grammar)
		if err != nil {
			return nil, input(err.Error())
		}
		result = append(result, parsed)
	}
	result = distinct(result, true)
	if len(result) > 20 {
		return nil, input("condition_id accepts at most 20 distinct values")
	}
	return result, nil
}
func events(values []int64) ([]string, error) {
	if len(values) == 0 {
		return nil, input("event_ids must be nonempty")
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value < 1 || value > 2147483647 {
			return nil, input("event_ids must contain positive 32-bit integers")
		}
		result = append(result, strconv.FormatInt(value, 10))
	}
	return distinct(result, false), nil
}
func tokenSelectors(q map[string]any) error {
	assets, hasAssets := q["asset_ids"].([]string)
	conditions, hasConditions := q["condition_ids"].([]string)
	if hasAssets == hasConditions {
		return input("provide exactly one of asset_ids or condition_ids")
	}
	values, key, maximum := assets, "token_id", 50
	if hasConditions {
		values, key, maximum = conditions, "condition", 10
	}
	result := make([]string, 0, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if hasAssets {
			if strings.Trim(value, "0123456789") != "" {
				return input("asset_ids must contain decimal digit strings")
			}
			value = strings.TrimLeft(value, "0")
			if value == "" {
				value = "0"
			}
			if len(value) > 78 {
				return input("asset_ids exceeds 78 digits")
			}
		} else {
			if (len(value) != 64 && len(value) != 66) || !codec.ValidHex(value) {
				return input("condition_ids must contain 31-byte or 32-byte hex strings")
			}
			value = strings.ToLower(value)
		}
		result = append(result, value)
	}
	result = distinct(result, false)
	if len(result) == 0 || len(result) > maximum {
		return input("invalid number of distinct token selectors")
	}
	delete(q, "asset_ids")
	delete(q, "condition_ids")
	q[key] = strings.Join(result, ",")
	return nil
}
