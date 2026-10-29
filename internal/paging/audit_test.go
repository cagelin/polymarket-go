package paging

import (
	"encoding/json"
	"testing"
)

func TestUnsupportedQueryTypesFailBeforeIO(t *testing.T) {
	for _, value := range []any{nil, int32(3), json.Number("7"), struct{}{}} {
		if _, err := CanonicalQueryValue(value); err == nil {
			t.Errorf("CanonicalQueryValue(%T) succeeded, want error", value)
		}
		if _, err := QueryFingerprint(map[string]any{"bad": value}); err == nil {
			t.Errorf("QueryFingerprint(%T) succeeded, want error", value)
		}
		if _, err := QueryValues(map[string]any{"bad": value}); err == nil {
			t.Errorf("QueryValues(%T) succeeded, want error", value)
		}
		pages := KeysetPaginator[string](nil, "data", "", "", "cursor", map[string]any{"bad": value}, 10, func([]byte) ([]string, *string, error) { t.Error("unexpected parse call"); return nil, nil, nil })
		if _, err := pages.FirstPage(t.Context()); err == nil {
			t.Errorf("FirstPage(%T) succeeded, want error before network IO", value)
		}
	}
}
