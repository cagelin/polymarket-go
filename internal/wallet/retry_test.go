package wallet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/models"
)

func TestRelayerRespectsRetryAfter(t *testing.T) {
	var submits atomic.Int32
	client := testClient(t, models.WalletDeposit, func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/":
			rpcReply(t, w, request, map[string]any{"eth_getBlockByNumber": map[string]string{"timestamp": "0x66850a00"}})
		case "/v1/account/transactions/params":
			_, _ = fmt.Fprint(w, `{"address":"0x1111111111111111111111111111111111111111","nonce":"19"}`)
		case "/submit":
			submits.Add(1)
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(429)
		default:
			t.Errorf("unexpected request %s", request.URL.Path)
			w.WriteHeader(500)
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
	defer cancel()
	_, err := client.ExecuteTransaction(ctx, []models.TransactionCall{{To: client.config.Wallet, Data: "0x"}}, nil)
	if !errors.Is(err, context.DeadlineExceeded) || submits.Load() != 1 {
		t.Errorf("ExecuteTransaction(429) = %v, submits=%d; want canceled wait after one submit", err, submits.Load())
	}
}
func TestUnknownRejectionDoesNotAuthorizeRetry(t *testing.T) {
	failure := &sdkerrors.RequestRejectedError{Status: 400, Code: "future_code", Message: "unrecognized rejection"}
	if retryable(failure) {
		t.Error("retryable(unknown rejection) = true, want false")
	}
	if nonce, ok := rejectedNonce(failure); ok || nonce != "" {
		t.Errorf("rejectedNonce(unknown rejection) = %q, %t; want no nonce", nonce, ok)
	}
}
