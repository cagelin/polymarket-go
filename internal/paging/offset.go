package paging

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/transport"
	"github.com/cagelam/polymarket-go/pagination"
)

// OffsetOptions defines server limits and the number of rows that fill a page.
type OffsetOptions[T any] struct {
	PageSize    int
	MaxPageSize int
	MaxOffset   *int
	Fill        func([]T) int
}

// OffsetPaginator preserves the original offset and page size on resume.
func OffsetPaginator[T any](client *transport.Client, service, baseURL, path string, params map[string]any, options OffsetOptions[T]) (*pagination.Paginator[T], error) {
	if err := checkWindow(0, options.PageSize, options.MaxPageSize, options.MaxOffset); err != nil {
		return nil, err
	}
	fingerprint := QueryFingerprint(params)
	baseQuery := QueryValues(params)
	return pagination.New(func(ctx context.Context, cursor *string) (pagination.Page[T], error) {
		offset, size := 0, options.PageSize
		if cursor != nil {
			var err error
			offset, size, err = decodeOffset(*cursor, service, path, fingerprint)
			if err != nil {
				return pagination.Page[T]{}, err
			}
		}
		if err := checkWindow(offset, size, options.MaxPageSize, options.MaxOffset); err != nil {
			return pagination.Page[T]{}, err
		}
		query := make(url.Values, len(baseQuery))
		for key, values := range baseQuery {
			query[key] = append([]string(nil), values...)
		}
		query.Set("offset", strconv.Itoa(offset))
		query.Set("limit", strconv.Itoa(size))
		raw, err := client.Request(ctx, baseURL, http.MethodGet, path, query, nil)
		if err != nil {
			return pagination.Page[T]{}, err
		}
		items, err := codec.ParseList[T](raw)
		if err != nil {
			return pagination.Page[T]{}, err
		}
		filled := len(items)
		if options.Fill != nil {
			filled = options.Fill(items)
		}
		more := filled >= size
		page := pagination.Page[T]{Items: items, HasMore: more}
		if more {
			if offset > int(^uint(0)>>1)-size {
				return pagination.Page[T]{}, &sdkerrors.PaginationLimitError{Message: "offset exceeds supported range"}
			}
			next := encodeOffset(service, path, fingerprint, offset+size, size)
			page.NextCursor = &next
			page.LimitReached = options.MaxOffset != nil && offset+size > *options.MaxOffset
		}
		return page, nil
	}, nil), nil
}

func checkWindow(offset, size, maxSize int, maxOffset *int) error {
	if size < 1 || (maxSize > 0 && size > maxSize) {
		return &sdkerrors.UserInputError{Message: "page_size is outside the supported range"}
	}
	if maxOffset != nil && offset > *maxOffset {
		return &sdkerrors.PaginationLimitError{Message: "pagination reached the deepest served page; completeness is unknown"}
	}
	return nil
}

func encodeOffset(service, path, fingerprint string, offset, size int) string {
	raw := `{"f":` + PythonString(fingerprint) + `,"o":` + strconv.Itoa(offset) + `,"p":` + PythonString(path) + `,"s":` + strconv.Itoa(size) + `,"svc":` + PythonString(service) + `,"v":1}`
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func decodeOffset(cursor, service, path, fingerprint string) (int, int, error) {
	failure := &sdkerrors.UserInputError{Message: "offset cursor is invalid or belongs to a different query"}
	fields, err := cursorFields(cursor)
	if err != nil {
		return 0, 0, err
	}
	var payload struct {
		Version     json.RawMessage `json:"v"`
		Service     string          `json:"svc"`
		Path        string          `json:"p"`
		Fingerprint string          `json:"f"`
		Offset      *int            `json:"o"`
		Size        *int            `json:"s"`
	}
	if err = json.Unmarshal(fields, &payload); err != nil {
		return 0, 0, failure
	}
	var version float64
	if string(payload.Version) == "true" {
		version = 1
	} else if err = json.Unmarshal(payload.Version, &version); err != nil {
		return 0, 0, failure
	}
	if version != 1 || payload.Service != service || payload.Path != path || payload.Fingerprint != fingerprint || payload.Offset == nil || *payload.Offset < 0 || payload.Size == nil || *payload.Size < 1 {
		return 0, 0, failure
	}
	return *payload.Offset, *payload.Size, nil
}

func cursorFields(cursor string) ([]byte, error) {
	failure := &sdkerrors.UserInputError{Message: "invalid pagination cursor"}
	if strings.ContainsAny(cursor, "\r\n") {
		return nil, failure
	}
	raw, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, failure
	}
	if _, err = codec.Object(raw); err != nil {
		return nil, failure
	}
	return raw, nil
}

// CursorPath reads a route selector; the chosen paginator still validates every field.
func CursorPath(cursor string) (string, error) {
	raw, err := cursorFields(cursor)
	if err != nil {
		return "", err
	}
	var value struct {
		Path string `json:"p"`
	}
	if err = json.Unmarshal(raw, &value); err != nil || value.Path == "" {
		return "", &sdkerrors.UserInputError{Message: "invalid pagination cursor"}
	}
	return value.Path, nil
}
