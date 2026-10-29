package wallet

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/cagelin/polymarket-go/models"
	"github.com/cagelin/polymarket-go/transactions"
)

// TestPythonTokenOperations compares the submitted call, not just its ABI helper.
func TestPythonTokenOperations(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/wallet-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Operation, Token, Recipient, To, Data, Error string
			Amount                                       *string
			Approved                                     bool
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for i, row := range fixture.Cases {
		if row.Token == "" {
			continue
		}
		t.Run(row.Operation+"/"+strconv.Itoa(i), func(subtest *testing.T) {
			var submits, requests atomic.Int32
			client := testClient(subtest, models.WalletDeposit, func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				switch r.URL.Path {
				case "/":
					rpcReply(subtest, w, r, map[string]any{"eth_getBlockByNumber": map[string]string{"timestamp": "0x66850a00"}})
				case "/v1/account/transactions/params":
					_, _ = fmt.Fprint(w, `{"address":"0x1111111111111111111111111111111111111111","nonce":"19"}`)
				case "/submit":
					submits.Add(1)
					var body struct {
						Metadata string
						Params   struct {
							Calls []struct{ Target, Data, Value string }
						} `json:"depositWalletParams"`
					}
					if decodeErr := json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
						subtest.Error(decodeErr)
						w.WriteHeader(400)
						return
					}
					if len(body.Params.Calls) != 1 {
						subtest.Errorf("submitted %d calls, want 1", len(body.Params.Calls))
					} else if call := body.Params.Calls[0]; call.Target != row.To || call.Data != row.Data || call.Value != "0" {
						subtest.Errorf("submitted %+v, Python wants to=%s data=%s value=0", call, row.To, row.Data)
					}
					if body.Metadata != "parity" {
						subtest.Errorf("metadata = %q, want parity", body.Metadata)
					}
					_, _ = fmt.Fprint(w, `{"transactionID":"token-1","state":"STATE_NEW"}`)
				case "/v1/account/transactions/token-1":
					_, _ = fmt.Fprint(w, `{"transaction_id":"token-1","transaction_hash":"0x123","state":"STATE_CONFIRMED"}`)
				default:
					subtest.Errorf("unexpected request %s", r.URL.Path)
					w.WriteHeader(500)
				}
			})
			var amount *big.Int
			if row.Amount != nil {
				var ok bool
				amount, ok = new(big.Int).SetString(*row.Amount, 10)
				if !ok {
					subtest.Fatal("invalid fixture amount")
				}
			}
			var handle *transactions.Handle
			var operationErr error
			switch row.Operation {
			case "approve_erc20":
				handle, operationErr = client.ApproveERC20(subtest.Context(), models.EVMAddress(row.Token), models.EVMAddress(row.Recipient), amount, "parity")
			case "transfer_erc20":
				handle, operationErr = client.TransferERC20(subtest.Context(), models.EVMAddress(row.Token), models.EVMAddress(row.Recipient), amount, "parity")
			case "approve_erc1155_for_all":
				handle, operationErr = client.ApproveERC1155ForAll(subtest.Context(), models.EVMAddress(row.Token), models.EVMAddress(row.Recipient), row.Approved, "parity")
			default:
				subtest.Fatal("unknown token fixture operation")
			}
			if row.Error != "" {
				if operationErr == nil || requests.Load() != 0 {
					subtest.Fatalf("invalid input: err=%v requests=%d, want rejection before network", operationErr, requests.Load())
				}
				return
			}
			if operationErr != nil {
				subtest.Fatal(operationErr)
			}
			outcome, waitErr := handle.Wait(subtest.Context())
			if waitErr != nil || outcome == nil || outcome.TransactionHash != "0x123" || submits.Load() != 1 {
				subtest.Fatalf("confirmation = %+v, %v; submissions=%d, want confirmed once", outcome, waitErr, submits.Load())
			}
		})
	}
}
