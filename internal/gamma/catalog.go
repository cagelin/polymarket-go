package gamma

import (
	"context"
	"errors"
	"net/http"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/paging"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
)

// TagQuery selects exactly one ID or slug; nil means absent, empty is invalid.
type TagQuery struct {
	ID              *models.TagID `json:"id"`
	Slug            *string       `json:"slug"`
	IncludeTemplate *bool         `json:"include_template"`
	Locale          *string       `json:"locale"`
}

// RelatedTagsQuery controls related tag relationships.
type RelatedTagsQuery struct {
	ID        *models.TagID `json:"id"`
	Slug      *string       `json:"slug"`
	OmitEmpty *bool         `json:"omit_empty"`
	Status    *string       `json:"status"`
}

// RelatedTagResourcesQuery requests full related tag resources.
type RelatedTagResourcesQuery struct {
	RelatedTagsQuery
	Locale *string `json:"locale"`
}

// GetSeries fetches a series by identifier.
func (c *Client) GetSeries(ctx context.Context, id models.SeriesID, locale *string) (*models.Series, error) {
	if id == "" {
		return nil, &sdkerrors.UserInputError{Message: "id must be nonempty"}
	}
	params := make(map[string]any)
	paging.QueryOption(params, "locale", locale)
	return getCatalog[models.Series](ctx, c, "/series/"+escapeSegment(string(id)), params)
}

// GetTag fetches one tag by ID or slug.
func (c *Client) GetTag(ctx context.Context, p TagQuery) (*models.Tag, error) {
	if p.Slug != nil && p.IncludeTemplate != nil {
		return nil, &sdkerrors.UserInputError{Message: "include_template only supports ID lookups"}
	}
	path, err := tagPath(p.ID, p.Slug)
	if err != nil {
		return nil, err
	}
	params := make(map[string]any)
	paging.QueryOption(params, "include_template", p.IncludeTemplate)
	paging.QueryOption(params, "locale", p.Locale)
	return getCatalog[models.Tag](ctx, c, path, params)
}

// GetRelatedTags fetches related tag relationships.
func (c *Client) GetRelatedTags(ctx context.Context, p RelatedTagsQuery) ([]models.RelatedTag, error) {
	if p.Slug != nil && (p.OmitEmpty != nil || p.Status != nil) {
		return nil, &sdkerrors.UserInputError{Message: "omit_empty and status only support ID lookups"}
	}
	path, err := tagPath(p.ID, p.Slug)
	if err != nil {
		return nil, err
	}
	params := make(map[string]any)
	paging.QueryOption(params, "omit_empty", p.OmitEmpty)
	paging.QueryOption(params, "status", p.Status)
	return listCatalog[models.RelatedTag](ctx, c, path+"/related-tags", params)
}

// GetRelatedTagResources fetches the full related tag resources.
func (c *Client) GetRelatedTagResources(ctx context.Context, p RelatedTagResourcesQuery) ([]models.Tag, error) {
	path, err := tagPath(p.ID, p.Slug)
	if err != nil {
		return nil, err
	}
	params := make(map[string]any)
	paging.QueryOption(params, "omit_empty", p.OmitEmpty)
	paging.QueryOption(params, "status", p.Status)
	paging.QueryOption(params, "locale", p.Locale)
	return listCatalog[models.Tag](ctx, c, path+"/related-tags/tags", params)
}

// GetSports fetches sports metadata.
func (c *Client) GetSports(ctx context.Context) ([]models.SportsMetadata, error) {
	return listCatalog[models.SportsMetadata](ctx, c, "/sports", nil)
}

// GetSportsMarketTypes fetches available sports market types.
func (c *Client) GetSportsMarketTypes(ctx context.Context) (*models.SportsMarketTypes, error) {
	return getCatalog[models.SportsMarketTypes](ctx, c, "/sports/market-types", nil)
}

// GetPublicProfile returns nil when the service reports that the wallet has no profile.
func (c *Client) GetPublicProfile(ctx context.Context, address models.EVMAddress) (*models.PublicProfile, error) {
	if address == "" {
		return nil, &sdkerrors.UserInputError{Message: "address must be nonempty"}
	}
	profile, err := getCatalog[models.PublicProfile](ctx, c, "/public-profile", map[string]any{"address": string(address)})
	var rejected *sdkerrors.RequestRejectedError
	if errors.As(err, &rejected) && rejected.Status == http.StatusNotFound {
		return nil, nil
	}
	return profile, err
}

// GetCommentThread fetches a comment and its replies.
func (c *Client) GetCommentThread(ctx context.Context, id models.CommentID, getPositions *bool) ([]models.Comment, error) {
	if id == "" {
		return nil, &sdkerrors.UserInputError{Message: "id must be nonempty"}
	}
	params := make(map[string]any)
	paging.QueryOption(params, "get_positions", getPositions)
	return listCatalog[models.Comment](ctx, c, "/comments/"+escapeSegment(string(id)), params)
}

func tagPath(id *models.TagID, slug *string) (string, error) {
	if (id == nil) == (slug == nil) {
		return "", &sdkerrors.UserInputError{Message: "provide exactly one tag id or slug"}
	}
	if id != nil {
		if *id == "" {
			return "", &sdkerrors.UserInputError{Message: "id must be nonempty"}
		}
		return "/tags/" + escapeSegment(string(*id)), nil
	}
	if *slug == "" {
		return "", &sdkerrors.UserInputError{Message: "slug must be nonempty"}
	}
	return "/tags/slug/" + escapeSegment(*slug), nil
}
func getCatalog[T any](ctx context.Context, c *Client, path string, params map[string]any) (*T, error) {
	raw, err := c.transport.Request(ctx, c.transport.GammaURL, http.MethodGet, path, paging.QueryValues(params), nil)
	if err != nil {
		return nil, err
	}
	result, err := codec.DecodeResponse[T](raw)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func listCatalog[T any](ctx context.Context, c *Client, path string, params map[string]any) ([]T, error) {
	raw, err := c.transport.Request(ctx, c.transport.GammaURL, http.MethodGet, path, paging.QueryValues(params), nil)
	if err != nil {
		return nil, err
	}
	return codec.ParseList[T](raw)
}

// ListSeriesParams controls the complete upstream catalog filter set.
type ListSeriesParams struct {
	Ascending     *bool       `json:"ascending"`
	Closed        *bool       `json:"closed"`
	ExcludeEvents *bool       `json:"exclude_events"`
	Locale        *string     `json:"locale"`
	Order         *string     `json:"order"`
	Recurrence    *Recurrence `json:"recurrence"`
	Slugs         []string    `json:"slug"`
	PageSize      *int        `json:"page_size"`
}

// ListSeries lazily retrieves series with resumable offset pagination.
func (c *Client) ListSeries(p ListSeriesParams) (*pagination.Paginator[models.Series], error) {
	if err := validateRecurrence(p.Recurrence); err != nil {
		return nil, err
	}
	params := make(map[string]any)
	paging.QueryOption(params, "ascending", p.Ascending)
	paging.QueryOption(params, "closed", p.Closed)
	paging.QueryOption(params, "exclude_events", p.ExcludeEvents)
	paging.QueryOption(params, "locale", p.Locale)
	paging.QueryOption(params, "order", p.Order)
	if p.Recurrence != nil {
		params["recurrence"] = string(*p.Recurrence)
	}
	if len(p.Slugs) > 0 {
		params["slug"] = append([]string(nil), p.Slugs...)
	}
	size := 20
	if p.PageSize != nil {
		size = *p.PageSize
	}
	return paging.OffsetPaginator(c.transport, "gamma", c.transport.GammaURL, "/series", params, paging.OffsetOptions[models.Series]{PageSize: size, MaxPageSize: 50})
}

// ListTagsParams controls the complete upstream catalog filter set.
type ListTagsParams struct {
	Ascending       *bool   `json:"ascending"`
	IncludeTemplate *bool   `json:"include_template"`
	IsCarousel      *bool   `json:"is_carousel"`
	Locale          *string `json:"locale"`
	Order           *string `json:"order"`
	PageSize        *int    `json:"page_size"`
}

// ListTags lazily retrieves tags with resumable offset pagination.
func (c *Client) ListTags(p ListTagsParams) (*pagination.Paginator[models.Tag], error) {
	params := make(map[string]any)
	paging.QueryOption(params, "ascending", p.Ascending)
	paging.QueryOption(params, "include_template", p.IncludeTemplate)
	paging.QueryOption(params, "is_carousel", p.IsCarousel)
	paging.QueryOption(params, "locale", p.Locale)
	paging.QueryOption(params, "order", p.Order)
	size := 20
	if p.PageSize != nil {
		size = *p.PageSize
	}
	return paging.OffsetPaginator(c.transport, "gamma", c.transport.GammaURL, "/tags", params, paging.OffsetOptions[models.Tag]{PageSize: size, MaxPageSize: 100})
}

// ListTeamsParams controls the complete upstream catalog filter set.
type ListTeamsParams struct {
	Abbreviations []string `json:"abbreviation"`
	Ascending     *bool    `json:"ascending"`
	Leagues       []string `json:"league"`
	Names         []string `json:"name"`
	Order         *string  `json:"order"`
	ProviderIDs   []int64  `json:"provider_ids"`
	PageSize      *int     `json:"page_size"`
}

// ListTeams lazily retrieves teams with resumable offset pagination.
func (c *Client) ListTeams(p ListTeamsParams) (*pagination.Paginator[models.Team], error) {
	params := make(map[string]any)
	if len(p.Abbreviations) > 0 {
		params["abbreviation"] = append([]string(nil), p.Abbreviations...)
	}
	paging.QueryOption(params, "ascending", p.Ascending)
	if len(p.Leagues) > 0 {
		params["league"] = append([]string(nil), p.Leagues...)
	}
	if len(p.Names) > 0 {
		params["name"] = append([]string(nil), p.Names...)
	}
	paging.QueryOption(params, "order", p.Order)
	if len(p.ProviderIDs) > 0 {
		params["provider_id"] = append([]int64(nil), p.ProviderIDs...)
	}
	size := 20
	if p.PageSize != nil {
		size = *p.PageSize
	}
	return paging.OffsetPaginator(c.transport, "gamma", c.transport.GammaURL, "/teams", params, paging.OffsetOptions[models.Team]{PageSize: size, MaxPageSize: 100})
}
