package paging

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/internal/transport"
	"github.com/cagelin/polymarket-go/pagination"
)

// PythonString preserves json.dumps(ensure_ascii=True) for portable cursors.
func PythonString(value string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, r := range value {
		switch r {
		case '"', '\\':
			out.WriteByte('\\')
			out.WriteRune(r)
		case '\b':
			out.WriteString(`\b`)
		case '\f':
			out.WriteString(`\f`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if r < 32 || r >= 127 {
				if r > 0xffff {
					high, low := utf16.EncodeRune(r)
					_, _ = fmt.Fprintf(&out, `\u%04x\u%04x`, high, low) // Writing to strings.Builder cannot fail.
				} else {
					_, _ = fmt.Fprintf(&out, `\u%04x`, r) // Writing to strings.Builder cannot fail.
				}
			} else {
				out.WriteRune(r)
			}
		}
	}
	out.WriteByte('"')
	return out.String()
}

// CanonicalQueryValue encodes a query value for a stable fingerprint.
func CanonicalQueryValue(value any) string {
	switch v := value.(type) {
	case string:
		return PythonString(v)
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return codec.PythonFloat(v)
	case []string:
		items := make([]string, len(v))
		for i, item := range v {
			items[i] = PythonString(item)
		}
		return "[" + strings.Join(items, ",") + "]"
	case []int64:
		items := make([]string, len(v))
		for i, item := range v {
			items[i] = strconv.FormatInt(item, 10)
		}
		return "[" + strings.Join(items, ",") + "]"
	default:
		panic("unsupported internal query parameter type")
	}
}

// QueryFingerprint binds a cursor to its original filters.
func QueryFingerprint(params map[string]any) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, PythonString(key)+":"+CanonicalQueryValue(params[key]))
	}
	hash := sha256.Sum256([]byte("{" + strings.Join(parts, ",") + "}"))
	return hex.EncodeToString(hash[:])[:12]
}

// QueryValues copies typed filters into HTTP query values.
func QueryValues(params map[string]any) url.Values {
	values := make(url.Values, len(params))
	for key, value := range params {
		switch v := value.(type) {
		case string:
			values.Set(key, v)
		case []string:
			values[key] = append([]string(nil), v...)
		case []int64:
			for _, item := range v {
				values.Add(key, strconv.FormatInt(item, 10))
			}
		default:
			values.Set(key, CanonicalQueryValue(v))
		}
	}
	return values
}

// EncodeKeysetCursor wraps a server cursor with query identity.
func EncodeKeysetCursor(service, path, fingerprint, serverCursor string) string {
	payload := `{"f":` + PythonString(fingerprint) + `,"k":` + PythonString(serverCursor) + `,"p":` + PythonString(path) + `,"svc":` + PythonString(service) + `,"v":1}`
	return base64.StdEncoding.EncodeToString([]byte(payload))
}

// DecodeKeysetCursor validates and unwraps a query-bound cursor.
func DecodeKeysetCursor(cursor, service, path, fingerprint string) (string, error) {
	failure := &sdkerrors.UserInputError{Message: "pagination cursor is invalid or belongs to a different query"}
	// Go's base64 decoder ignores CR/LF; Python's validated decoder rejects them.
	if strings.ContainsAny(cursor, "\r\n") {
		return "", failure
	}
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", failure
	}
	var payload struct {
		Version     json.RawMessage `json:"v"`
		Service     string          `json:"svc"`
		Path        string          `json:"p"`
		Fingerprint string          `json:"f"`
		Cursor      string          `json:"k"`
	}
	if err = json.Unmarshal(data, &payload); err != nil {
		return "", failure
	}
	var version float64
	// Python equality also accepts True and 1.0 as version 1.
	if string(payload.Version) == "true" {
		version = 1
	} else if err = json.Unmarshal(payload.Version, &version); err != nil {
		return "", failure
	}
	if version != 1 || payload.Service != service || payload.Path != path || payload.Fingerprint != fingerprint || payload.Cursor == "" {
		return "", failure
	}
	return payload.Cursor, nil
}

// KeysetPaginator lazily retrieves server-cursor pages.
func KeysetPaginator[T any](c *transport.Client, service, baseURL, path, cursorParam string, params map[string]any, pageSize int, parse func([]byte) ([]T, *string, error)) *pagination.Paginator[T] {
	fingerprint := QueryFingerprint(params)
	baseQuery := QueryValues(params)
	return pagination.New(func(ctx context.Context, cursor *string) (pagination.Page[T], error) {
		query := make(url.Values, len(baseQuery)+2)
		for key, values := range baseQuery {
			query[key] = append([]string(nil), values...)
		}
		query.Set("limit", strconv.Itoa(pageSize))
		if cursor != nil {
			serverCursor, err := DecodeKeysetCursor(*cursor, service, path, fingerprint)
			if err != nil {
				return pagination.Page[T]{}, err
			}
			query.Set(cursorParam, serverCursor)
		}
		var data []byte
		var err error
		if service == "data" {
			data, err = c.DataRead(ctx, path, query)
		} else {
			data, err = c.Request(ctx, baseURL, http.MethodGet, path, query, nil)
		}
		if err != nil {
			return pagination.Page[T]{}, err
		}
		items, next, err := parse(data)
		if err != nil {
			return pagination.Page[T]{}, &sdkerrors.UnexpectedResponseError{Message: "invalid paginated response", Cause: err}
		}
		page := pagination.Page[T]{Items: items, HasMore: next != nil}
		if next != nil {
			encoded := EncodeKeysetCursor(service, path, fingerprint, *next)
			page.NextCursor = &encoded
		}
		return page, nil
	}, nil)
}

// ServerNextCursor validates optional server continuation metadata.
func ServerNextCursor(fields map[string]json.RawMessage) (*string, error) {
	value, ok := fields["next_cursor"]
	if !ok || string(value) == "null" {
		return nil, nil
	}
	var cursor string
	if err := json.Unmarshal(value, &cursor); err != nil {
		return nil, err
	}
	if cursor == "" {
		return nil, fmt.Errorf("next_cursor must be nonempty")
	}
	return &cursor, nil
}

// QueryOption includes a nonnil optional filter.
func QueryOption[T string | bool | int | int64 | float64](params map[string]any, key string, value *T) {
	if value != nil {
		params[key] = *value
	}
}
