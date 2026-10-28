package wallet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cagelam/polymarket-go/models"
)

func TestSessionAuthorizationIdempotencyAndRegistry(t *testing.T) {
	var submissions, reads atomic.Int32
	var first []byte
	var firstKey string
	var wallet *Client
	wallet = testClient(t, models.WalletDeposit, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			rpcReply(t, w, r, map[string]any{"eth_getBlockByNumber": map[string]string{"timestamp": "0x66850a00"}})
		case "/v1/account/transactions/params":
			_, _ = fmt.Fprint(w, `{"address":"0x1111111111111111111111111111111111111111","nonce":"19"}`)
		case "/v1/session-signers/authorizations":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
				return
			}
			if submissions.Add(1) == 1 {
				first = body
				firstKey = r.Header.Get("Idempotency-Key")
				w.WriteHeader(503)
				return
			}
			if !bytes.Equal(first, body) || firstKey == "" || r.Header.Get("Idempotency-Key") != firstKey {
				t.Error("retry changed authorization payload or operation key")
			}
			var payload map[string]any
			if err = json.Unmarshal(body, &payload); err != nil {
				t.Error(err)
				return
			}
			if payload["validUntil"] != "1735534000" {
				t.Errorf("expiry differs from Python: %v", payload["validUntil"])
			}
			_, _ = fmt.Fprint(w, `{"status":"SUBMITTED","transactionId":"auth-1","transactionHash":"0x`+strings.Repeat("ab", 32)+`"}`)
		case "/v1/account/transactions/auth-1":
			_, _ = fmt.Fprint(w, `{"transaction_id":"auth-1","state":"STATE_CONFIRMED","transaction_hash":"0x`+strings.Repeat("ab", 32)+`"}`)
		case "/v1/user/session-signers":
			keys := make([]any, 0) // The registry returns an empty JSON array, not null.
			if reads.Add(1) > 1 {
				keys = append(keys, map[string]any{"address": "0x1111111111111111111111111111111111111111", "scopes": []string{"CLOB"}, "valid_until": 1735534000})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"wallet": wallet.config.Wallet, "signers": keys})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(500)
		}
	})
	result, err := wallet.AuthorizeSessionKey(t.Context(), models.AuthorizeSessionKeyParams{Address: "0x1111111111111111111111111111111111111111", Scopes: []models.SessionKeyScope{models.SessionScopeCLOB}})
	if err != nil {
		t.Fatal(err)
	}
	if result.SessionKey.ValidUntil.Unix() != 1735534000 || submissions.Load() != 2 || reads.Load() != 2 {
		t.Fatal("authorization returned before registry visibility")
	}
}
func TestSessionRevocationWaitsThroughTransientRegistryFailure(t *testing.T) {
	var reads atomic.Int32
	var wallet *Client
	wallet = testClient(t, models.WalletDeposit, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			rpcReply(t, w, r, map[string]any{"eth_getBlockByNumber": map[string]string{"timestamp": "0x66850a00"}})
		case "/v1/account/transactions/params":
			_, _ = fmt.Fprint(w, `{"address":"0x1111111111111111111111111111111111111111","nonce":"19"}`)
		case "/v1/session-signers/revocations":
			_, _ = fmt.Fprint(w, `{"status":"FENCED","fenced":true,"transactionId":"revoke-1"}`)
		case "/v1/user/session-signers":
			if reads.Add(1) == 1 {
				w.WriteHeader(404)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"wallet": wallet.config.Wallet, "signers": []any{}})
		default:
			t.Errorf("unexpected %s", r.URL.Path)
			w.WriteHeader(500)
		}
	})
	if err := wallet.RevokeSessionKey(t.Context(), models.RevokeSessionKeyParams{Address: "0x1111111111111111111111111111111111111111"}); err != nil {
		t.Fatal(err)
	}
	if reads.Load() != 2 {
		t.Fatal("treated failed registry read as successful revocation")
	}
}
func TestSessionRegistryRejectsAnotherWallet(t *testing.T) {
	wallet := testClient(t, models.WalletDeposit, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"wallet":"0x1111111111111111111111111111111111111111","signers":[]}`)
	})
	if _, err := wallet.FetchSessionKeys(t.Context()); err == nil {
		t.Fatal("accepted registry for another wallet")
	}
}
