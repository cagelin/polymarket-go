package transactions_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/cagelin/polymarket-go/transactions"
)

type receiptRPC struct{}

func (receiptRPC) Request(context.Context, string, []any) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"0x1"}`), nil
}

type statusRelayer struct{}

func (statusRelayer) RelayerEndpoint() string { return "https://example.invalid" }
func (statusRelayer) AuthenticatedRequest(context.Context, string, string, string, url.Values, any, func(string, string, []byte) (http.Header, error)) ([]byte, error) {
	return []byte(`{"state":"STATE_CONFIRMED","transaction_id":"id","transaction_hash":"0x123"}`), nil
}
func TestPublicConstructorsAcceptExternalImplementations(t *testing.T) {
	eoa := transactions.NewEOA(receiptRPC{}, "0x123", 1, time.Second)
	outcome, err := eoa.Wait(t.Context())
	if err != nil || outcome.TransactionHash != "0x123" {
		t.Fatalf("NewEOA(external RPC).Wait() = %v, %v", outcome, err)
	}
	relayed := transactions.NewRelayed(statusRelayer{}, nil, "id", nil, 1, time.Second)
	outcome, err = relayed.Wait(t.Context())
	if err != nil || outcome.TransactionHash != "0x123" {
		t.Errorf("NewRelayed(external relayer).Wait() = %v, %v", outcome, err)
	}
}

func TestTypedNilDependenciesReturnErrors(t *testing.T) {
	var rpc *receiptRPC
	var relayer *statusRelayer
	for _, handle := range []*transactions.Handle{
		transactions.NewEOA(rpc, "hash", 1, time.Second),
		transactions.NewRelayed(relayer, nil, "id", nil, 1, time.Second),
	} {
		if _, err := handle.Wait(t.Context()); err == nil {
			t.Error("Wait(typed nil dependency) succeeded, want input error")
		}
	}
}
