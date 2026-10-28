package polymarket

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

// pythonString preserves json.dumps(ensure_ascii=True) for portable cursors.
func pythonString(value string) string {
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
					fmt.Fprintf(&out, `\u%04x\u%04x`, high, low)
				} else {
					fmt.Fprintf(&out, `\u%04x`, r)
				}
			} else {
				out.WriteRune(r)
			}
		}
	}
	out.WriteByte('"')
	return out.String()
}

func pythonFloat(value float64) string {
	if math.IsNaN(value) {
		return "NaN"
	}
	if math.IsInf(value, 1) {
		return "Infinity"
	}
	if math.IsInf(value, -1) {
		return "-Infinity"
	}
	format := byte('f')
	if magnitude := math.Abs(value); magnitude != 0 && (magnitude < 1e-4 || magnitude >= 1e16) {
		format = 'e'
	}
	s := strconv.FormatFloat(value, format, -1, 64)
	if !strings.ContainsAny(s, ".e") {
		s += ".0"
	}
	return s
}

func canonicalQueryValue(value any) string {
	switch v := value.(type) {
	case string:
		return pythonString(v)
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return pythonFloat(v)
	case []string:
		items := make([]string, len(v))
		for i, item := range v {
			items[i] = pythonString(item)
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

func queryFingerprint(params map[string]any) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, pythonString(key)+":"+canonicalQueryValue(params[key]))
	}
	hash := sha256.Sum256([]byte("{" + strings.Join(parts, ",") + "}"))
	return hex.EncodeToString(hash[:])[:12]
}

func queryValues(params map[string]any) url.Values {
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
			values.Set(key, canonicalQueryValue(v))
		}
	}
	return values
}

func encodeKeysetCursor(service, path, fingerprint, serverCursor string) string {
	payload := `{"f":` + pythonString(fingerprint) + `,"k":` + pythonString(serverCursor) + `,"p":` + pythonString(path) + `,"svc":` + pythonString(service) + `,"v":1}`
	return base64.StdEncoding.EncodeToString([]byte(payload))
}

func decodeKeysetCursor(cursor, service, path, fingerprint string) (string, error) {
	failure := &UserInputError{Message: "pagination cursor is invalid or belongs to a different query"}
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

func keysetPaginator[T any](c *PublicClient, service, baseURL, path, cursorParam string, params map[string]any, pageSize int, parse func([]byte) ([]T, *string, error)) *Paginator[T] {
	fingerprint := queryFingerprint(params)
	baseQuery := queryValues(params)
	return &Paginator[T]{fetch: func(ctx context.Context, cursor *string) (Page[T], error) {
		query := make(url.Values, len(baseQuery)+2)
		for key, values := range baseQuery {
			query[key] = append([]string(nil), values...)
		}
		query.Set("limit", strconv.Itoa(pageSize))
		if cursor != nil {
			serverCursor, err := decodeKeysetCursor(*cursor, service, path, fingerprint)
			if err != nil {
				return Page[T]{}, err
			}
			query.Set(cursorParam, serverCursor)
		}
		var data []byte
		var err error
		if service == "data" {
			data, err = c.dataRead(ctx, path, query)
		} else {
			data, err = c.requestAt(ctx, baseURL, http.MethodGet, path, query, nil)
		}
		if err != nil {
			return Page[T]{}, err
		}
		items, next, err := parse(data)
		if err != nil {
			return Page[T]{}, &UnexpectedResponseError{Message: "invalid paginated response", Cause: err}
		}
		page := Page[T]{Items: items, HasMore: next != nil}
		if next != nil {
			encoded := encodeKeysetCursor(service, path, fingerprint, *next)
			page.NextCursor = &encoded
		}
		return page, nil
	}}
}

func serverNextCursor(fields map[string]json.RawMessage) (*string, error) {
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
