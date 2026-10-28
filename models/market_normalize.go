package models

import (
	"encoding/json"
	"fmt"

	"github.com/cagelam/polymarket-go/internal/codec"
)

func normalizeMarket(data []byte) ([]byte, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return nil, err
	}
	if _, normalized := fields["state"]; normalized {
		return data, nil
	}
	labels, err := codec.GammaStrings(fields["outcomes"])
	if err != nil {
		return nil, err
	}
	if len(labels) != 2 {
		return nil, fmt.Errorf("expected binary market outcomes")
	}
	tokens, err := codec.GammaStrings(fields["clobTokenIds"])
	if err != nil {
		return nil, err
	}
	positions, err := codec.GammaStrings(fields["positionIds"])
	if err != nil {
		return nil, err
	}
	prices, err := codec.GammaSequence(fields["outcomePrices"])
	if err != nil {
		return nil, err
	}
	// Validate all prices, including surplus values, as the Python normalizer does.
	for _, price := range prices {
		if _, priceErr := codec.NumericDecimal(map[string]json.RawMessage{"price": price}, "price", false); priceErr != nil {
			return nil, priceErr
		}
	}
	outcomes := make(map[string]any, 2)
	for i, side := range []string{"yes", "no"} {
		outcome := map[string]any{"label": labels[i]}
		if i < len(tokens) {
			outcome["token_id"] = tokens[i]
		}
		if i < len(positions) {
			outcome["position_id"] = positions[i]
		}
		if i < len(prices) {
			outcome["price"] = prices[i]
		}
		outcomes[side] = outcome
	}
	result := make(map[string]any)
	for _, key := range []string{"id", "version", "slug", "question", "description", "category", "image", "icon"} {
		if value, exists := fields[key]; exists {
			result[key] = value
		}
	}
	if value, exists := fields["groupItemTitle"]; exists {
		result["group_item_title"] = value
	}
	if value, exists := fields["conditionId"]; exists && string(value) != `""` {
		result["condition_id"] = value
	}
	result["outcomes"] = outcomes
	groups := map[string]map[string]string{
		"state":      {"active": "active", "closed": "closed", "archived": "archived", "combo_status": "comboStatus", "accepting_orders": "acceptingOrders", "enable_order_book": "enableOrderBook", "neg_risk": "negRisk", "start_date": "startDate", "end_date": "endDate", "closed_time": "closedTime"},
		"metrics":    {"volume": "volume", "volume_num": "volumeNum", "volume_24hr": "volume24hr", "volume_1wk": "volume1wk", "volume_1mo": "volume1mo", "volume_1yr": "volume1yr", "volume_amm": "volumeAmm", "volume_clob": "volumeClob", "liquidity": "liquidity", "liquidity_num": "liquidityNum", "liquidity_clob": "liquidityClob"},
		"prices":     {"best_bid": "bestBid", "best_ask": "bestAsk", "last_trade_price": "lastTradePrice", "spread": "spread", "one_hour_price_change": "oneHourPriceChange", "one_day_price_change": "oneDayPriceChange", "one_week_price_change": "oneWeekPriceChange", "one_month_price_change": "oneMonthPriceChange", "one_year_price_change": "oneYearPriceChange"},
		"trading":    {"minimum_order_size": "orderMinSize", "minimum_tick_size": "orderPriceMinTickSize", "seconds_delay": "secondsDelay", "fees_enabled": "feesEnabled", "fee_type": "feeType", "fee_schedule": "feeSchedule"},
		"resolution": {"onchain_event_id": "onchainEventId", "request_id": "requestId", "resolution_status": "resolutionStatus", "question_id": "questionID", "neg_risk_request_id": "negRiskRequestID", "uma_resolution_status": "umaResolutionStatus", "source": "resolutionSource", "resolved_by": "resolvedBy"},
		"rewards":    {"clob_rewards": "clobRewards", "rewards_min_size": "rewardsMinSize", "rewards_max_spread": "rewardsMaxSpread", "holding_rewards_enabled": "holdingRewardsEnabled"},
		"sports":     {"sports_market_type": "sportsMarketType", "line": "line", "game_id": "gameId", "game_start_time": "gameStartTime"},
	}
	for group, mapping := range groups {
		nested := make(map[string]json.RawMessage)
		for key, wireKey := range mapping {
			if value, exists := fields[wireKey]; exists {
				nested[key] = value
			}
		}
		result[group] = nested
	}
	for _, key := range []string{"events", "tags"} {
		items, sequenceErr := codec.GammaSequence(fields[key])
		if sequenceErr != nil {
			return nil, sequenceErr
		}
		// Only objects are accepted, including in JSON-encoded relationship arrays.
		for _, item := range items {
			if _, objectErr := codec.Object(item); objectErr != nil {
				return nil, objectErr
			}
		}
		result[key] = items
	}
	return json.Marshal(result)
}
