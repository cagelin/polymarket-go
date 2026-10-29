package signing_test

import (
	"encoding/json"
	"math/big"
	"os"
	"strconv"
	"testing"

	"github.com/cagelin/polymarket-go/internal/evm"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
)

func TestPythonTransactions(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/transactions-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Key, Wallet, Nonce, Deadline string
			Calls                        []struct{ To, Data, Value string }
			ProxyData                    string `json:"proxy_data"`
			SafeData                     string `json:"safe_data"`
			DepositSignature             string `json:"deposit_signature"`
			SessionSignature             string `json:"session_signature"`
			SafeSignature                string `json:"safe_signature"`
			ProxySignature               string `json:"proxy_signature"`
			LegacyTransaction            string `json:"legacy_transaction"`
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for i, row := range fixture.Cases {
		t.Run(strconv.Itoa(i), func(subtest *testing.T) {
			signer, signErr := signing.New(row.Key)
			if signErr != nil {
				subtest.Fatal(signErr)
			}
			calls := make([]models.TransactionCall, len(row.Calls))
			for j, call := range row.Calls {
				amount, ok := new(big.Int).SetString(call.Value, 10)
				if !ok {
					subtest.Fatal("invalid fixture integer")
				}
				calls[j] = models.TransactionCall{To: models.EVMAddress(call.To), Data: call.Data, Value: amount}
			}
			check := func(name, actual, want string, operationErr error) {
				subtest.Helper()
				if operationErr != nil {
					subtest.Fatalf("%s: %v", name, operationErr)
				}
				if actual != want {
					subtest.Fatalf("%s differs from Python\ngot %s\nwant %s", name, actual, want)
				}
			}
			actual, operationErr := evm.ProxyCalls(calls)
			check("proxy ABI", actual, row.ProxyData, operationErr)
			actual, operationErr = evm.SafeMultisend(calls)
			check("safe ABI", actual, row.SafeData, operationErr)
			actual, operationErr = signer.SignDepositBatch(row.Wallet, calls, row.Nonce, row.Deadline, models.SignerOwner)
			check("deposit", actual, row.DepositSignature, operationErr)
			actual, operationErr = signer.SignDepositBatch(row.Wallet, calls, row.Nonce, row.Deadline, models.SignerSessionKey)
			check("session", actual, row.SessionSignature, operationErr)
			actual, operationErr = signer.SignSafe(row.Wallet, string(calls[0].To), row.SafeData, calls[0].Value, 1, row.Nonce)
			check("safe", actual, row.SafeSignature, operationErr)
			actual, operationErr = signer.SignProxy(string(calls[0].To), row.ProxyData, "200000", row.Nonce, "0xD216153c06E857cD7f72665E0aF1d7D82172F494", row.Wallet)
			check("proxy", actual, row.ProxySignature, operationErr)
			actual, operationErr = signer.SignLegacyTransaction(calls[0], 19, 200000, big.NewInt(30000000000))
			check("legacy", actual, row.LegacyTransaction, operationErr)
		})
	}
}
