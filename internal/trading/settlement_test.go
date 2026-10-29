package trading

import (
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/models"
)

func TestSettlementTimeoutBoundsSlowRequests(t *testing.T) {
	client := fixtureClient(t, func(_ http.ResponseWriter, request *http.Request) { <-request.Context().Done() })
	started := time.Now()
	_, err := client.WaitForOrderFillSettlement(t.Context(), models.AcceptedOrder{TradeIDs: []string{"pending"}}, 30*time.Millisecond)
	var timeout *sdkerrors.SettlementTimeoutError
	if !errors.As(err, &timeout) || len(timeout.PendingTradeIDs) != 1 {
		t.Fatalf("WaitForOrderFillSettlement(slow response) = %v, want pending trade timeout", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Errorf("WaitForOrderFillSettlement(30ms) took %s, want bounded HTTP request", elapsed)
	}
}
func TestSettlementRateLimitRespectsRetryAfter(t *testing.T) {
	var attempts atomic.Int32
	client := fixtureClient(t, func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(429)
	})
	_, err := client.WaitForOrderFillSettlement(t.Context(), models.AcceptedOrder{TradeIDs: []string{"pending"}}, 400*time.Millisecond)
	var timeout *sdkerrors.SettlementTimeoutError
	if !errors.As(err, &timeout) || attempts.Load() != 1 {
		t.Errorf("WaitForOrderFillSettlement(429) = %v, requests=%d; want timeout without early retry", err, attempts.Load())
	}
}
