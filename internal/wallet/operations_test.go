package wallet

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cagelin/polymarket-go/internal/evm"
	"github.com/cagelin/polymarket-go/models"
)

func TestApprovalsSkipExistingAndConfirmOnlyMissing(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(fmt.Sprint(missing), func(subtest *testing.T) {
			var submits, polls atomic.Int32
			wallet := testClient(subtest, models.WalletDeposit, func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/":
					raw, err := io.ReadAll(r.Body)
					if err != nil {
						subtest.Error(err)
						return
					}
					if raw[0] != '[' {
						var request struct{ ID uint64 }
						if err = json.Unmarshal(raw, &request); err != nil {
							subtest.Error(err)
							return
						}
						_ = json.NewEncoder(w).Encode(map[string]any{"id": request.ID, "result": map[string]string{"timestamp": "0x66850a00"}})
						return
					}
					var requests []struct{ ID uint64 }
					if err = json.Unmarshal(raw, &requests); err != nil {
						subtest.Error(err)
						return
					}
					if len(requests) != 17 {
						subtest.Errorf("approval catalog: %d", len(requests))
					}
					results := make([]map[string]any, 0, len(requests))
					for i, request := range slices.Backward(requests) {
						value := big.NewInt(1)
						if i < 7 {
							value = evm.MaxUint256()
						}
						if missing && i == 0 {
							value = new(big.Int)
						}
						results = append(results, map[string]any{"id": request.ID, "result": fmt.Sprintf("0x%064x", value)})
					}
					_ = json.NewEncoder(w).Encode(results)
				case "/v1/account/transactions/params":
					_, _ = fmt.Fprint(w, `{"address":"0x1111111111111111111111111111111111111111","nonce":"19"}`)
				case "/submit":
					submits.Add(1)
					var body struct {
						Params struct{ Calls []struct{ Data string } } `json:"depositWalletParams"`
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						subtest.Error(err)
						return
					}
					if len(body.Params.Calls) != 1 || !strings.HasPrefix(body.Params.Calls[0].Data, "0x095ea7b3") {
						subtest.Error("wrong missing approval calls")
					}
					_, _ = fmt.Fprint(w, `{"transactionID":"approve-1","state":"STATE_NEW"}`)
				case "/v1/account/transactions/approve-1":
					polls.Add(1)
					_, _ = fmt.Fprint(w, `{"transaction_id":"approve-1","transaction_hash":"0x123","state":"STATE_CONFIRMED"}`)
				default:
					subtest.Error(r.URL.Path)
					w.WriteHeader(500)
				}
			})
			if err := wallet.SetupTradingApprovals(subtest.Context()); err != nil {
				subtest.Fatal(err)
			}
			expected := int32(0)
			if missing {
				expected = 1
			}
			if submits.Load() != expected || polls.Load() != expected {
				subtest.Fatal("unnecessary approvals or missing confirmation")
			}
		})
	}
}
func TestDefaultDepositDeployment(t *testing.T) {
	var submits atomic.Int32
	wallet := testClient(t, models.WalletDeposit, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/deployed":
			_, _ = fmt.Fprint(w, `{"deployed":false}`)
		case "/submit":
			submits.Add(1)
			var body struct{ Type, To string }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			if body.Type != "WALLET-CREATE" || body.To != depositFactory {
				t.Error("wrong deployment envelope")
			}
			_, _ = fmt.Fprint(w, `{"transactionID":"deploy-1","state":"STATE_NEW"}`)
		case "/v1/account/transactions/deploy-1":
			_, _ = fmt.Fprint(w, `{"transaction_id":"deploy-1","transaction_hash":"0x123","state":"STATE_CONFIRMED"}`)
		default:
			t.Error(r.URL.Path)
			w.WriteHeader(500)
		}
	})
	if err := wallet.EnsureReady(t.Context()); err != nil {
		t.Fatal(err)
	}
	if submits.Load() != 1 {
		t.Fatal("missing deployment")
	}
	wallet.config.Wallet = "0x1111111111111111111111111111111111111111"
	if err := wallet.EnsureReady(t.Context()); err == nil {
		t.Fatal("deployed a wallet not derived from this owner")
	}
	if submits.Load() != 1 {
		t.Fatal("unexpected deployment")
	}
}
func TestPositionMergeReadsBalanceAndPreventsOverspend(t *testing.T) {
	for _, excess := range []bool{false, true} {
		t.Run(fmt.Sprint(excess), func(subtest *testing.T) {
			var submits atomic.Int32
			condition := "0x03" + strings.Repeat("ab", 30)
			wallet := testClient(subtest, models.WalletDeposit, func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/markets/keyset":
					_ = json.NewEncoder(w).Encode(map[string]any{"markets": []any{map[string]any{"id": "7", "version": "v2", "conditionId": condition, "outcomes": []string{"Yes", "No"}, "positionIds": []string{"123", "124"}}}, "next_cursor": nil})
				case "/":
					var request struct {
						ID     uint64
						Method string
					}
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
						subtest.Error(err)
						return
					}
					var result any
					switch request.Method {
					case "eth_getBlockByNumber":
						result = map[string]string{"timestamp": "0x66850a00"}
					case "eth_call":
						encoded, encodeErr := evm.Pack("ignored()", []string{"uint256[]"}, []*big.Int{big.NewInt(3000000), big.NewInt(2000000)})
						if encodeErr != nil {
							subtest.Error(encodeErr)
							return
						}
						result = "0x" + encoded[10:]
					default:
						subtest.Error(request.Method)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"id": request.ID, "result": result})
				case "/v1/account/transactions/params":
					_, _ = fmt.Fprint(w, `{"address":"0x1111111111111111111111111111111111111111","nonce":"19"}`)
				case "/submit":
					submits.Add(1)
					_, _ = fmt.Fprint(w, `{"transactionID":"merge-1","state":"STATE_NEW"}`)
				default:
					subtest.Error(r.URL.Path)
					w.WriteHeader(500)
				}
			})
			amount := big.NewInt(1000000)
			if excess {
				amount = big.NewInt(2000001)
			}
			_, err := wallet.MergePositions(subtest.Context(), models.MergePositionsParams{ConditionID: models.ConditionID(condition), Amount: amount})
			if excess {
				if err == nil || submits.Load() != 0 {
					subtest.Fatalf("overspend: %v", err)
				}
			} else {
				if err != nil || submits.Load() != 1 {
					subtest.Fatalf("merge: %v", err)
				}
			}
		})
	}
}
func TestCollateralPlanIdentityAndStalePlan(t *testing.T) {
	var submits atomic.Int32
	wallet := testClient(t, models.WalletDeposit, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			rpcReply(t, w, r, map[string]any{"eth_getBlockByNumber": map[string]string{"timestamp": "0x66850a00"}})
		case "/v1/account/transactions/params":
			_, _ = fmt.Fprint(w, `{"address":"0x1111111111111111111111111111111111111111","nonce":"19"}`)
		case "/v1/collateral-return/submit":
			submits.Add(1)
			w.WriteHeader(409)
			_, _ = fmt.Fprint(w, `{"error":"stale plan"}`)
		default:
			t.Error(r.URL.Path)
			w.WriteHeader(500)
		}
	})
	plan := models.CollateralReturnPlanResponse{Wallet: "0x1111111111111111111111111111111111111111", ChainID: 137, RouterCall: models.CollateralReturnRouterCall{To: router, Data: "0x1234"}}
	if _, err := wallet.ExecuteCollateralReturnPlan(t.Context(), plan); err == nil {
		t.Fatal("accepted wrong wallet")
	}
	plan.Wallet = wallet.config.Wallet
	if _, err := wallet.ExecuteCollateralReturnPlan(t.Context(), plan); err == nil {
		t.Fatal("accepted stale plan")
	}
	if submits.Load() != 1 {
		t.Fatal("retried stale plan")
	}
}
