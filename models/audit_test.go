package models

import (
	"encoding/json"
	"testing"
)

func TestUnknownFieldKindsAreRejected(t *testing.T) {
	if _, err := normalizeGammaField(json.RawMessage(`true`), gammaField{name: "flag", kind: "unsupported"}); err == nil {
		t.Error("normalizeGammaField(unsupported) succeeded, want invalid kind error")
	}
	if _, err := dataValue(json.RawMessage(`true`), dataField{name: "flag", kind: "unsupported"}); err == nil {
		t.Error("dataValue(unsupported) succeeded, want invalid kind error")
	}
	if _, err := dataValue(nil, dataField{name: "id", kind: "notification_id"}); err == nil {
		t.Error("dataValue(empty notification_id) succeeded, want error")
	}
	if _, err := normalizeGammaField(nil, gammaField{name: "id", kind: "id"}); err == nil {
		t.Error("normalizeGammaField(empty id) succeeded, want error")
	}
}
