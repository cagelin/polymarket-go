package models

import (
	"encoding/json"

	"github.com/cagelam/polymarket-go/internal/codec"
)

// normalizeEvent groups the flat Gamma wire fields into the public event model.
func normalizeEvent(data []byte) ([]byte, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return nil, err
	}
	if _, normalized := fields["state"]; normalized {
		return data, nil
	}
	result := selectGammaFields(fields, map[string]string{
		"id": "id", "version": "version", "parent_event_id": "parentEventId", "ticker": "ticker", "slug": "slug", "title": "title", "subtitle": "subtitle", "description": "description", "category": "category", "subcategory": "subcategory", "image": "image", "icon": "icon", "featured_image": "featuredImage", "created_at": "createdAt", "updated_at": "updatedAt", "published_at": "published_at", "metadata": "eventMetadata",
	})
	groups := map[string]map[string]string{
		"state":      {"active": "active", "closed": "closed", "archived": "archived", "new": "new", "featured": "featured", "restricted": "restricted", "cyom": "cyom", "live": "live", "ended": "ended", "automatically_active": "automaticallyActive", "comments_enabled": "commentsEnabled", "requires_translation": "requiresTranslation"},
		"schedule":   {"start_date": "startDate", "creation_date": "creationDate", "end_date": "endDate", "closed_time": "closedTime", "start_time": "startTime", "event_date": "eventDate", "event_week": "eventWeek", "finished_at": "finishedTimestamp"},
		"metrics":    {"liquidity": "liquidity", "liquidity_amm": "liquidityAmm", "liquidity_clob": "liquidityClob", "volume": "volume", "volume_24hr": "volume24hr", "volume_1wk": "volume1wk", "volume_1mo": "volume1mo", "volume_1yr": "volume1yr", "open_interest": "openInterest", "competitive": "competitive", "comment_count": "commentCount", "tweet_count": "tweetCount"},
		"display":    {"sort_by": "sortBy", "show_all_outcomes": "showAllOutcomes", "show_market_images": "showMarketImages", "gmp_chart_mode": "gmpChartMode", "color": "color", "featured_order": "featuredOrder", "country_name": "countryName", "election_type": "electionType", "image_optimized": "imageOptimized", "icon_optimized": "iconOptimized", "featured_image_optimized": "featuredImageOptimized"},
		"trading":    {"enable_order_book": "enableOrderBook", "neg_risk": "negRisk", "neg_risk_market_id": "negRiskMarketID", "neg_risk_fee_bips": "negRiskFeeBips", "enable_neg_risk": "enableNegRisk", "neg_risk_augmented": "negRiskAugmented", "cumulative_markets": "cumulativeMarkets"},
		"resolution": {"source": "resolutionSource", "automatically_resolved": "automaticallyResolved"},
		"estimation": {"estimate_value": "estimateValue", "cant_estimate": "cantEstimate", "estimated_value": "estimatedValue"},
		"sports":     {"series_slug": "seriesSlug", "score": "score", "elapsed": "elapsed", "period": "period", "game_status": "gameStatus", "game_id": "gameId", "rescheduled_from_game_id": "rescheduledFromGameId", "sportsradar_match_id": "sportsradarMatchId", "home_team_name": "homeTeamName", "away_team_name": "awayTeamName", "spreads_main_line": "spreadsMainLine", "totals_main_line": "totalsMainLine", "best_lines": "bestLines", "teams": "teams", "sport": "sport", "last_highlight": "lastHighlight", "last_highlight_type": "lastHighlightType", "last_highlight_at": "lastHighlightAt"},
	}
	for group, mapping := range groups {
		nested := selectGammaFields(fields, mapping)
		if group == "sports" {
			for _, key := range []string{"best_lines", "teams"} {
				if !gammaTruthy(nested[key]) {
					nested[key] = json.RawMessage("[]")
				}
			}
		}
		encoded, encodeErr := json.Marshal(nested)
		if encodeErr != nil {
			return nil, encodeErr
		}
		result[group] = encoded
	}
	for key, wireKey := range map[string]string{"partners": "externalPartners", "markets": "markets", "series": "series", "tags": "tags", "creators": "eventCreators"} {
		items, sequenceErr := codec.GammaSequence(fields[wireKey])
		if sequenceErr != nil {
			return nil, sequenceErr
		}
		normalized := make([]json.RawMessage, 0, len(items))
		for _, item := range items {
			entry, entryErr := codec.Object(item)
			if entryErr != nil {
				return nil, entryErr
			}
			if key == "markets" {
				// Upstream skips nonbinary child markets before validating their other fields.
				outcomes, outcomesErr := codec.GammaSequence(entry["outcomes"])
				if outcomesErr != nil {
					return nil, outcomesErr
				}
				if len(outcomes) != 2 {
					continue
				}
			} else {
				entry = normalizeEventRelation(key, entry)
				item, entryErr = json.Marshal(entry)
				if entryErr != nil {
					return nil, entryErr
				}
			}
			normalized = append(normalized, item)
		}
		encoded, encodeErr := json.Marshal(normalized)
		if encodeErr != nil {
			return nil, encodeErr
		}
		result[key] = encoded
	}
	return json.Marshal(result)
}

func selectGammaFields(fields map[string]json.RawMessage, mapping map[string]string) map[string]json.RawMessage {
	result := make(map[string]json.RawMessage, len(mapping))
	for key, wireKey := range mapping {
		if value, exists := fields[wireKey]; exists {
			result[key] = value
		}
	}
	return result
}

func normalizeEventRelation(kind string, fields map[string]json.RawMessage) map[string]json.RawMessage {
	switch kind {
	case "partners":
		result := selectGammaFields(fields, map[string]string{"id": "id", "external_id": "externalId", "created_at": "createdAt", "updated_at": "updatedAt"})
		// Nonobject partners are absent upstream; Partner ignores extra fields.
		if _, err := codec.Object(fields["partner"]); err == nil {
			result["partner"] = fields["partner"]
		}
		return result
	case "series":
		return selectGammaFields(fields, map[string]string{"id": "id", "slug": "slug", "title": "title", "subtitle": "subtitle", "description": "description", "image": "image", "icon": "icon", "active": "active", "closed": "closed", "archived": "archived", "volume": "volume", "liquidity": "liquidity", "start_date": "startDate"})
	case "tags":
		return selectGammaFields(fields, map[string]string{"id": "id", "slug": "slug", "label": "label"})
	case "creators":
		result := selectGammaFields(fields, map[string]string{"id": "id", "name": "creatorName", "handle": "creatorHandle", "url": "creatorUrl", "image": "creatorImage", "created_at": "createdAt", "updated_at": "updatedAt"})
		if !gammaTruthy(result["url"]) {
			result["url"] = json.RawMessage("null")
			if fallback, exists := fields["creatorURL"]; exists {
				result["url"] = fallback
			}
		}
		return result
	default:
		return fields
	}
}

func gammaTruthy(value json.RawMessage) bool {
	switch string(value) {
	case "", "null", "false", "0", "0.0", `""`, "[]", "{}":
		return false
	default:
		return true
	}
}

func normalizeProfile(data []byte) ([]byte, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return nil, err
	}
	if _, exists := fields["wallet"]; !exists {
		if wallet, present := fields["proxyWallet"]; present {
			fields["wallet"] = wallet
		}
	}
	return json.Marshal(fields)
}

func normalizeSearchResults(data []byte) ([]byte, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return nil, err
	}
	for _, key := range []string{"events", "tags", "profiles"} {
		if value, exists := fields[key]; !exists || string(value) == "null" {
			fields[key] = json.RawMessage("[]")
		}
	}
	return json.Marshal(fields)
}

func normalizePublicProfile(data []byte) ([]byte, error)  { return normalizeProfile(data) }
func normalizeCommentProfile(data []byte) ([]byte, error) { return normalizeProfile(data) }
func normalizeComment(data []byte) ([]byte, error) {
	fields, err := codec.Object(data)
	if err != nil {
		return nil, err
	}
	for _, key := range []string{"parentEntityID", "parentCommentID"} {
		value, exists := fields[key]
		if !exists || string(value) == "null" || value[0] == '"' {
			continue
		}
		text := string(value)
		if text == "true" {
			text = "True"
		}
		if text == "false" {
			text = "False"
		}
		fields[key], err = json.Marshal(text)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(fields)
}
