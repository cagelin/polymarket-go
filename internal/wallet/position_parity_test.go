package wallet

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cagelin/polymarket-go/internal/evm"
	"github.com/cagelin/polymarket-go/models"
	"github.com/cagelin/polymarket-go/transactions"
)

// TestPythonPositionSubmissions follows each V2 entry point through submission.
func TestPythonPositionSubmissions(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/wallet-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Action, Amount, Condition, To, Data string
			V2                                  bool
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, row := range fixture.Cases {
		if !row.V2 || row.Amount != "1" {
			continue
		}
		actions := []string{row.Action}
		if row.Action == "merge" {
			actions = append(actions, "merge_batch")
		}
		for _, action := range actions {
			t.Run(action, func(subtest *testing.T) {
				var submits atomic.Int32
				client := testClient(subtest, models.WalletDeposit, func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/markets/keyset":
						_ = json.NewEncoder(w).Encode(map[string]any{"markets": []any{map[string]any{"id": "7", "version": "v2", "conditionId": row.Condition, "outcomes": []string{"Yes", "No"}, "positionIds": []string{"123", "124"}}}, "next_cursor": nil})
					case "/":
						balances, encodeErr := evm.Pack("ignored()", []string{"uint256[]"}, []*big.Int{big.NewInt(1), big.NewInt(1)})
						if encodeErr != nil {
							subtest.Error(encodeErr)
							return
						}
						balanceResult := "0x" + balances[10:]
						if action == "redeem" {
							balanceResult = fmt.Sprintf("0x%064x", 1)
						}
						rpcReply(subtest, w, r, map[string]any{"eth_call": balanceResult, "eth_getBlockByNumber": map[string]string{"timestamp": "0x66850a00"}})
					case "/v1/account/transactions/params":
						_, _ = fmt.Fprint(w, `{"address":"0x1111111111111111111111111111111111111111","nonce":"19"}`)
					case "/submit":
						submits.Add(1)
						var body struct {
							Params struct {
								Calls []struct{ Target, Data string }
							} `json:"depositWalletParams"`
						}
						if decodeErr := json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
							subtest.Error(decodeErr)
							return
						}
						if len(body.Params.Calls) != 1 {
							subtest.Errorf("calls=%d, want 1", len(body.Params.Calls))
						} else if call := body.Params.Calls[0]; !strings.EqualFold(call.Target, row.To) || call.Data != row.Data {
							subtest.Errorf("submitted %+v, Python wants %s %s", call, row.To, row.Data)
						}
						_, _ = fmt.Fprint(w, `{"transactionID":"position-1","state":"STATE_NEW"}`)
					case "/v1/account/transactions/position-1":
						_, _ = fmt.Fprint(w, `{"transaction_id":"position-1","transaction_hash":"0x123","state":"STATE_CONFIRMED"}`)
					default:
						subtest.Errorf("unexpected request %s", r.URL.Path)
						w.WriteHeader(500)
					}
				})
				position, ok := new(big.Int).SetString(strings.TrimPrefix(row.Condition, "0x")+"01", 16)
				if !ok {
					subtest.Fatal("invalid condition fixture")
				}
				var handle *transactions.Handle
				var operationErr error
				switch action {
				case "split":
					handle, operationErr = client.SplitPosition(subtest.Context(), models.SplitPositionParams{ConditionID: models.ConditionID(row.Condition), Amount: big.NewInt(1)})
				case "merge":
					handle, operationErr = client.MergePositions(subtest.Context(), models.MergePositionsParams{ConditionID: models.ConditionID(row.Condition)})
				case "merge_batch":
					handle, operationErr = client.MergeMultiplePositions(subtest.Context(), []models.MergePositionRequest{{PositionID: models.PositionID(position.String())}}, nil)
				case "redeem":
					handle, operationErr = client.RedeemPositions(subtest.Context(), models.RedeemPositionsParams{PositionID: models.PositionID(position.String())})
				default:
					subtest.Fatalf("unhandled action %q", action)
				}
				if operationErr != nil {
					subtest.Fatal(operationErr)
				}
				if _, waitErr := handle.Wait(subtest.Context()); waitErr != nil || submits.Load() != 1 {
					subtest.Fatalf("confirmation err=%v, submits=%d, want one confirmed submission", waitErr, submits.Load())
				}
			})
		}
	}
}
