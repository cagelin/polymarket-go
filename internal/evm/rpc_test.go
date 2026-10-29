package evm_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cagelin/polymarket-go/internal/evm"
	"github.com/cagelin/polymarket-go/internal/transport"
	"github.com/cagelin/polymarket-go/models"
)

func TestBatchRPCOrderingAndFallback(t *testing.T) {
	for _, mode := range []string{"reordered", "split", "duplicate", "missing", "wrong-id", "rate-limit"} {
		t.Run(mode, func(subtest *testing.T) {
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				var body json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					subtest.Error(err)
					return
				}
				if body[0] != '[' {
					var request struct{ ID uint64 }
					if err := json.Unmarshal(body, &request); err != nil {
						subtest.Error(err)
						return
					}
					_, _ = fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":"0x01"}`, request.ID)
					return
				}
				if mode == "split" {
					w.WriteHeader(503)
					return
				}
				if mode == "rate-limit" {
					w.WriteHeader(429)
					return
				}
				var batch []struct{ ID uint64 }
				if err := json.Unmarshal(body, &batch); err != nil {
					subtest.Error(err)
					return
				}
				if mode == "duplicate" {
					batch[1].ID = batch[0].ID
				}
				if mode == "wrong-id" {
					batch[1].ID = 1000
				}
				replies := []map[string]any{{"jsonrpc": "2.0", "id": batch[1].ID, "result": "0x02"}, {"jsonrpc": "2.0", "id": batch[0].ID, "result": "0x01"}}
				if mode == "missing" {
					replies = replies[:1]
				}
				_ = json.NewEncoder(w).Encode(replies)
			}))
			defer server.Close()
			client, err := transport.New(transport.Config{RPCURL: server.URL})
			if err != nil {
				subtest.Fatal(err)
			}
			subtest.Cleanup(func() { _ = client.Close() })
			rpc := evm.NewRPC(client)
			result, err := rpc.CallBatch(subtest.Context(), []models.TransactionCall{{To: "0x1111111111111111111111111111111111111111", Data: "0x"}, {To: "0x2222222222222222222222222222222222222222", Data: "0x"}})
			switch mode {
			case "reordered":
				if err != nil || strings.Join(result, ",") != "0x01,0x02" {
					subtest.Fatalf("result %v err %v", result, err)
				}
			case "split":
				if err != nil || len(result) != 2 || requests.Load() != 3 {
					subtest.Fatalf("split: %v %v count %d", result, err, requests.Load())
				}
			default:
				if err == nil {
					subtest.Fatal("accepted malformed response")
				}
				if requests.Load() != 1 {
					subtest.Fatal("retried a rejected read unexpectedly")
				}
			}
		})
	}
}
