package polymarket

import (
	"context"
	"iter"
)

// Page contains results and explicit continuation/completeness information.
// LimitReached means completeness is unknown, not that more items exist.
type Page[T any] struct {
	Items        []T     `json:"items"`
	HasMore      bool    `json:"has_more"`
	NextCursor   *string `json:"next_cursor"`
	TotalCount   *int64  `json:"total_count"`
	LimitReached bool    `json:"limit_reached"`
}

// Paginator lazily retrieves pages. Independent iterations are safe to run
// concurrently; no network request occurs until FirstPage, Pages or Items.
type Paginator[T any] struct {
	fetch         func(context.Context, *string) (Page[T], error)
	initialCursor *string
	empty         bool
}

// FirstPage fetches the first page of this paginator's selection.
func (p *Paginator[T]) FirstPage(ctx context.Context) (Page[T], error) {
	if p.empty {
		return Page[T]{Items: []T{}}, nil
	}
	return p.fetch(ctx, p.initialCursor)
}

// FromCursor resumes a selection. A nil cursor returns an empty paginator,
// rather than restarting the selection. The cursor's value is copied.
func (p *Paginator[T]) FromCursor(cursor *string) *Paginator[T] {
	if cursor == nil {
		return &Paginator[T]{empty: true, fetch: func(context.Context, *string) (Page[T], error) { return Page[T]{Items: []T{}}, nil }}
	}
	value := *cursor
	return &Paginator[T]{fetch: p.fetch, initialCursor: &value}
}

// Pages yields pages until completion, a depth limit, cancellation or error.
// Breaking the loop stops immediately without prefetching another page.
func (p *Paginator[T]) Pages(ctx context.Context) iter.Seq2[Page[T], error] {
	return func(yield func(Page[T], error) bool) {
		if p.empty {
			return
		}
		cursor := p.initialCursor
		requested := make(map[string]bool)
		if cursor != nil {
			requested[*cursor] = true
		}
		for {
			page, err := p.fetch(ctx, cursor)
			if err != nil {
				yield(Page[T]{}, err)
				return
			}
			// Snapshot continuation before the caller can mutate the yielded page.
			var next *string
			if page.NextCursor != nil {
				value := *page.NextCursor
				next = &value
			}
			if !yield(page, nil) || page.LimitReached || !page.HasMore {
				return
			}
			if next == nil || requested[*next] {
				yield(Page[T]{}, &UnexpectedResponseError{Message: "continuing page has a missing or repeated cursor"})
				return
			}
			requested[*next] = true
			cursor = next
		}
	}
}

// Items flattens pages. Use Pages to observe LimitReached explicitly.
func (p *Paginator[T]) Items(ctx context.Context) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for page, err := range p.Pages(ctx) {
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
