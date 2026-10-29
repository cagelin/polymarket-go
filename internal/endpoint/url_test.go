package endpoint

import "testing"

func TestRejectAmbiguousServiceURLs(t *testing.T) {
	for _, endpointURL := range []string{"wss://user:secret@example.com/ws", "wss://example.com/ws?key=secret", "wss://example.com/ws?", "wss://example.com/ws#", "wss://example.com/ws#fragment", "wss://:80/ws", "https://example.com/ws"} {
		if _, err := Parse(endpointURL, "ws", "wss"); err == nil {
			t.Errorf("Parse(%q) succeeded, want invalid endpoint", endpointURL)
		}
	}
	if value, err := Parse("wss://example.com/ws/", "ws", "wss"); err != nil || value != "wss://example.com/ws" {
		t.Errorf("Parse(valid URL) = %q, %v", value, err)
	}
}
