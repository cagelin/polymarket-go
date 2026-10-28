package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
)

func callCatalogParity(ctx context.Context, c *PublicClient, operation string, input json.RawMessage, cursor *string) (any, error) {
	switch operation {
	case "get_series":
		var p struct {
			ID     SeriesID
			Locale *string
		}
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetSeries(ctx, p.ID, p.Locale)
	case "get_tag":
		var p TagQuery
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetTag(ctx, p)
	case "get_related_tags":
		var p RelatedTagsQuery
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetRelatedTags(ctx, p)
	case "get_related_tag_resources":
		var p RelatedTagResourcesQuery
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetRelatedTagResources(ctx, p)
	case "get_sports":
		return c.GetSports(ctx)
	case "get_sports_market_types":
		return c.GetSportsMarketTypes(ctx)
	case "get_public_profile":
		var p struct{ Address EVMAddress }
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetPublicProfile(ctx, p.Address)
	case "get_comment_thread":
		var p struct {
			ID           CommentID
			GetPositions *bool `json:"get_positions"`
		}
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		return c.GetCommentThread(ctx, p.ID, p.GetPositions)
	case "list_series":
		var p ListSeriesParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListSeries(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_tags":
		var p ListTagsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListTags(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_teams":
		var p ListTeamsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListTeams(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_comments":
		var p ListCommentsParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListComments(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	case "list_comments_by_user_address":
		var p ListCommentsByUserAddressParams
		if err := json.Unmarshal(input, &p); err != nil {
			return nil, err
		}
		pager, err := c.ListCommentsByUserAddress(p)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pager = pager.FromCursor(cursor)
		}
		return pager.FirstPage(ctx)
	default:
		return nil, fmt.Errorf("unknown operation %s", operation)
	}
}
