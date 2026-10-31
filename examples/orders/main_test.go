package main

import (
	"errors"
	"net/http"
	"testing"

	"github.com/cagelin/polymarket-go"
	sdkerrors "github.com/cagelin/polymarket-go/errors"
)

func mockAccount(t *testing.T, exchange *mockExchange) *polymarket.SecureClient {
	t.Helper()
	client, err := newMockClient(t.Context(), exchange)
	if err != nil {
		t.Fatalf("newMockClient(synthetic credentials) = %v, want successful authentication", err)
	}
	t.Cleanup(func() { _ = client.Close() }) // Close only releases local resources.
	return client
}

func TestOrderLifecycleConfirmed(t *testing.T) {
	t.Parallel()
	exchange := &mockExchange{}
	client := mockAccount(t, exchange)
	if err := demonstrateOrders(t.Context(), client); err != nil {
		t.Errorf("demonstrateOrders(confirmed mock) = %v, want nil", err)
	}
	if exchange.limitOpen || exchange.submissions != 2 {
		t.Errorf("demonstrateOrders(confirmed mock) left open=%t submissions=%d, want false and 2", exchange.limitOpen, exchange.submissions)
	}
}

func TestOrderLifecycleRejected(t *testing.T) {
	t.Parallel()
	exchange := &mockExchange{rejectMarket: true}
	client := mockAccount(t, exchange)
	if err := demonstrateOrders(t.Context(), client); err == nil {
		t.Error("demonstrateOrders(rejected market order) = nil, want rejection error")
	}
	if exchange.limitOpen || exchange.submissions != 2 {
		t.Errorf("demonstrateOrders(rejected market order) left open=%t submissions=%d, want false and 2 without replay", exchange.limitOpen, exchange.submissions)
	}
}

func TestOrderLifecyclePending(t *testing.T) {
	t.Parallel()
	exchange := &mockExchange{pendingSettlement: true}
	client := mockAccount(t, exchange)
	err := demonstrateOrders(t.Context(), client)
	var timeout *sdkerrors.SettlementTimeoutError
	if !errors.As(err, &timeout) {
		t.Fatalf("demonstrateOrders(pending settlement) = %v, want SettlementTimeoutError", err)
	}
	if len(timeout.PendingTradeIDs) != 1 || timeout.PendingTradeIDs[0] != "mock-fill" {
		t.Errorf("demonstrateOrders(pending settlement) pending IDs=%v, want [mock-fill]", timeout.PendingTradeIDs)
	}
	if exchange.limitOpen || exchange.submissions != 2 {
		t.Errorf("demonstrateOrders(pending settlement) left open=%t submissions=%d, want false and 2 without resubmission", exchange.limitOpen, exchange.submissions)
	}
}

func TestMockExchangeRejectsExternalHost(t *testing.T) {
	t.Parallel()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://external.invalid/auth/api-keys", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext(external host) = %v, want nil", err)
	}
	var exchange mockExchange
	response, requestErr := exchange.RoundTrip(request)
	if response != nil {
		if response.Body != nil {
			_ = response.Body.Close() // The assertion below reports the unexpected response.
		}
		t.Errorf("RoundTrip(external host) response=%v, want nil", response)
	}
	if requestErr == nil {
		t.Error("RoundTrip(external host) error=nil, want host rejection before any network transport")
	}
}
