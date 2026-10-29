package wallet

import (
	"encoding/json"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/cagelin/polymarket-go/models"
)

func TestPythonWalletCalls(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/wallet-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Operation, Action, Amount, Condition, Data, To, Error string
			V2                                                    bool
			Legs, IDs                                             []string
			Scopes, Expected                                      []models.SessionKeyScope
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for i, row := range fixture.Cases {
		t.Run(strconv.Itoa(i), func(subtest *testing.T) {
			switch row.Operation {
			case "call":
				amount, ok := new(big.Int).SetString(row.Amount, 10)
				if !ok {
					subtest.Fatal("invalid fixture")
				}
				p := positionContext{condition: row.Condition, v2: row.V2, adapter: collateralAdapter}
				call, callErr := p.call(row.Action, amount, 1)
				if (callErr != nil) != (row.Error != "") {
					subtest.Fatalf("error %v, Python %s", callErr, row.Error)
				}
				if callErr == nil && (call.Data != row.Data || string(call.To) != row.To) {
					subtest.Fatal("ABI differs from Python")
				}
			case "combo":
				p, callErr := comboContext(row.Legs)
				if (callErr != nil) != (row.Error != "") {
					subtest.Fatalf("error %v, Python %s", callErr, row.Error)
				}
				if callErr == nil {
					ids := []string{p.ids[0].String(), p.ids[1].String()}
					if p.condition != row.Condition || !reflect.DeepEqual(ids, row.IDs) || p.prepare.Data != row.Data {
						subtest.Fatalf("combo derivation differs: %+v", p)
					}
				}
			case "scopes":
				actual, callErr := sessionScopes(row.Scopes)
				if (callErr != nil) != (row.Error != "") {
					subtest.Fatalf("error %v, Python %s", callErr, row.Error)
				}
				if callErr == nil && !reflect.DeepEqual(actual, row.Expected) {
					subtest.Fatal("scope normalization differs")
				}
			case "approve_erc20", "transfer_erc20", "approve_erc1155_for_all":
				// Submitted token calls are checked by TestPythonTokenOperations.
			default:
				subtest.Fatalf("unhandled wallet fixture operation %q", row.Operation)
			}
		})
	}
}
