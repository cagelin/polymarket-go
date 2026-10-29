package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/internal/paging"
	"github.com/cagelin/polymarket-go/pagination"
)

func queryValues(query map[string]any) url.Values { return paging.QueryValues(query) }
func parseList[T any](raw []byte) ([]T, error)    { return codec.ParseList[T](raw) }
func parseObject[T any](optional bool) func([]byte) (*T, error) {
	return func(raw []byte) (*T, error) {
		if string(raw) == "null" {
			if optional {
				return nil, nil
			}
			return nil, fmt.Errorf("data must not be null")
		}
		var value T
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		return &value, nil
	}
}
func getData[T any](ctx context.Context, c *Client, path string, query map[string]any, parse func([]byte) (T, error)) (T, error) {
	var zero T
	raw, err := c.transport.DataRead(ctx, path, queryValues(query))
	if err != nil {
		return zero, err
	}
	result, err := parseEnvelope(raw, parse)
	if err != nil {
		return zero, &sdkerrors.UnexpectedResponseError{Message: "invalid Data API response", Cause: err}
	}
	return result, nil
}
func parseEnvelope[T any](raw []byte, parse func([]byte) (T, error)) (T, error) {
	var zero T
	fields, err := codec.Object(raw)
	if err != nil {
		return zero, err
	}
	data, ok := fields["data"]
	if !ok {
		return zero, fmt.Errorf("missing data")
	}
	return parse(data)
}
func listData[T any](c *Client, path string, query map[string]any, pageSize *int, maximum int, parse func([]byte) ([]T, error)) (*pagination.Paginator[T], error) {
	size := 100
	if pageSize != nil {
		size = *pageSize
	}
	if size < 1 || size > maximum {
		return nil, input(fmt.Sprintf("page_size must be between 1 and %d", maximum))
	}
	return paging.KeysetPaginator(c.transport, "data", c.transport.DataURL, path, "cursor", query, size, func(raw []byte) ([]T, *string, error) {
		items, err := parseEnvelope(raw, parse)
		if err != nil {
			return nil, nil, err
		}
		_, next, err := parseHistoryPageMetadata(raw)
		if err != nil {
			return nil, nil, err
		}
		return items, next, nil
	}), nil
}
