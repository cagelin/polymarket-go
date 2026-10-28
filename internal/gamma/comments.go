package gamma

import (
	"context"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/paging"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
)

// CommentParentEntityType identifies a comment thread's owning resource.
type CommentParentEntityType string

// Supported comment parents.
const (
	CommentParentEvent  CommentParentEntityType = "Event"
	CommentParentSeries CommentParentEntityType = "Series"
)

// ListCommentsParams selects a thread and optional holder/position filters.
type ListCommentsParams struct {
	ParentEntityID   string                  `json:"parent_entity_id"`
	ParentEntityType CommentParentEntityType `json:"parent_entity_type"`
	Ascending        *bool                   `json:"ascending"`
	GetPositions     *bool                   `json:"get_positions"`
	HoldersOnly      *bool                   `json:"holders_only"`
	Order            *string                 `json:"order"`
	PageSize         *int                    `json:"page_size"`
}

// ListCommentsByUserAddressParams selects comments authored by one wallet.
type ListCommentsByUserAddressParams struct {
	Address   models.EVMAddress `json:"address"`
	Ascending *bool             `json:"ascending"`
	Order     *string           `json:"order"`
	PageSize  *int              `json:"page_size"`
}

// ListComments uses unbounded cursor pages where supported. Filtered listings
// stop at the server's depth limit and report LimitReached without claiming completion.
func (c *Client) ListComments(p ListCommentsParams) (*pagination.Paginator[models.Comment], error) {
	if p.ParentEntityID == "" || (p.ParentEntityType != CommentParentEvent && p.ParentEntityType != CommentParentSeries) {
		return nil, &sdkerrors.UserInputError{Message: "a nonempty parent ID and Event or Series parent type are required"}
	}
	params := map[string]any{"parent_entity_id": p.ParentEntityID, "parent_entity_type": string(p.ParentEntityType)}
	paging.QueryOption(params, "ascending", p.Ascending)
	paging.QueryOption(params, "get_positions", p.GetPositions)
	paging.QueryOption(params, "holders_only", p.HoldersOnly)
	paging.QueryOption(params, "order", p.Order)
	size := 20
	if p.PageSize != nil {
		size = *p.PageSize
	}
	maxOffset := 200
	offset, err := paging.OffsetPaginator(c.transport, "gamma", c.transport.GammaURL, "/comments", params, paging.OffsetOptions[models.Comment]{PageSize: size, MaxPageSize: 100, MaxOffset: &maxOffset, Fill: func(items []models.Comment) int {
		roots := 0
		for _, item := range items {
			if item.ParentCommentID == nil {
				roots++
			}
		}
		return roots
	}})
	if err != nil {
		return nil, err
	}
	if (p.GetPositions != nil && *p.GetPositions) || (p.HoldersOnly != nil && *p.HoldersOnly) || (p.Order != nil && *p.Order != "id" && *p.Order != "createdAt") {
		return offset, nil
	}
	order, ascending := "createdAt", false
	if p.Order != nil {
		order = *p.Order
		ascending = true
		if p.Ascending != nil {
			ascending = *p.Ascending
		}
	}
	query := map[string]any{"parent_entity_id": p.ParentEntityID, "parent_entity_type": string(p.ParentEntityType), "order": order, "ascending": ascending}
	keyset := paging.KeysetPaginator(c.transport, "gamma", c.transport.GammaURL, "/comments/keyset", "after_cursor", query, size, parseCommentsPage)
	return pagination.New(func(ctx context.Context, cursor *string) (pagination.Page[models.Comment], error) {
		if cursor == nil {
			return keyset.FirstPage(ctx)
		}
		path, pathErr := paging.CursorPath(*cursor)
		if pathErr != nil {
			return pagination.Page[models.Comment]{}, pathErr
		}
		if path == "/comments" {
			return offset.FromCursor(cursor).FirstPage(ctx)
		}
		return keyset.FromCursor(cursor).FirstPage(ctx)
	}, nil), nil
}

// ListCommentsByUserAddress lists comments up to the service's explicit depth cap.
func (c *Client) ListCommentsByUserAddress(p ListCommentsByUserAddressParams) (*pagination.Paginator[models.Comment], error) {
	if p.Address == "" {
		return nil, &sdkerrors.UserInputError{Message: "address must be nonempty"}
	}
	params := make(map[string]any)
	paging.QueryOption(params, "ascending", p.Ascending)
	paging.QueryOption(params, "order", p.Order)
	size := 20
	if p.PageSize != nil {
		size = *p.PageSize
	}
	maxOffset := 200
	return paging.OffsetPaginator(c.transport, "gamma", c.transport.GammaURL, "/comments/user_address/"+escapeSegment(string(p.Address)), params, paging.OffsetOptions[models.Comment]{PageSize: size, MaxPageSize: 100, MaxOffset: &maxOffset})
}

func parseCommentsPage(raw []byte) ([]models.Comment, *string, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, nil, err
	}
	items, err := codec.Required[[]models.Comment](fields, "comments")
	if err != nil {
		return nil, nil, err
	}
	next, err := paging.ServerNextCursor(fields)
	return items, next, err
}
