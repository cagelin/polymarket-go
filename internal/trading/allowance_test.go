package trading

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/internal/wallet"
	"github.com/cagelin/polymarket-go/models"
)

func TestAllowanceRecoveryRetriesOnlyAfterConfirmedApproval(t *testing.T) {
	for _, sufficient := range []bool{false, true} {
		t.Run(fmt.Sprint(sufficient), func(subtest *testing.T) {
			var orders, broadcasts, confirmed, refreshed atomic.Int32
			var firstOrder []byte
			spender := signing.Exchange("123", false)
			client := fixtureClient(subtest, func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/order":
					raw, err := io.ReadAll(r.Body)
					if err != nil {
						subtest.Error(err)
						return
					}
					if orders.Add(1) == 1 {
						firstOrder = raw
						w.WriteHeader(400)
						_, _ = fmt.Fprint(w, `{"error":"allowance is not enough"}`)
						return
					}
					if !bytes.Equal(raw, firstOrder) || confirmed.Load() != 1 || refreshed.Load() != 1 {
						subtest.Error("retried before confirmed approval and refresh or changed order")
					}
					writeResponse(subtest, w, map[string]any{"errorMsg": "", "makingAmount": "1", "takingAmount": "2", "orderID": "order", "success": true, "status": "live"})
				case "/balance-allowance":
					allowance := "0"
					if sufficient || refreshed.Load() > 0 {
						allowance = "1000000000000000"
					}
					writeResponse(subtest, w, map[string]any{"balance": "2000000", "allowances": map[string]string{spender: allowance}})
				case "/balance-allowance/update":
					refreshed.Add(1)
					_, _ = fmt.Fprint(w, "{}")
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
					case "eth_chainId":
						result = "0x89"
					case "eth_getTransactionCount":
						result = "0x0"
					case "eth_gasPrice":
						result = "0x1"
					case "eth_estimateGas":
						result = "0x10000"
					case "eth_sendRawTransaction":
						broadcasts.Add(1)
						result = "0x" + strings.Repeat("ab", 32)
					case "eth_getTransactionReceipt":
						confirmed.Add(1)
						result = map[string]string{"status": "0x1"}
					default:
						subtest.Error(request.Method)
					}
					writeResponse(subtest, w, map[string]any{"id": request.ID, "result": result})
				default:
					subtest.Error(r.URL.Path)
					w.WriteHeader(500)
				}
			})
			client.transport.RPCURL = client.transport.CLOBURL
			onchain, err := wallet.New(client.transport, wallet.Config{Signer: client.signer, Wallet: client.wallet, Kind: models.WalletEOA, Authority: models.SignerOwner, PollInterval: time.Millisecond})
			if err != nil {
				subtest.Fatal(err)
			}
			client.onchain = onchain
			if err = client.metadata.lock(subtest.Context()); err != nil {
				subtest.Fatal(err)
			}
			client.metadata.conditions["123"] = "condition"
			client.metadata.markets["condition"] = marketInfo{assets: map[string]bool{"123": true}, expires: client.now().Add(time.Hour)}
			client.metadata.unlock()
			order := models.SignedOrder{Salt: big.NewInt(1), MakerAmount: big.NewInt(1000000), TakerAmount: big.NewInt(2000000), Maker: client.wallet, Signer: client.wallet, AssetID: "123", Side: models.Buy, OrderType: models.GTC}
			_, err = client.postWithAllowanceRecovery(subtest.Context(), order)
			if sufficient {
				if err == nil || orders.Load() != 1 || broadcasts.Load() != 0 {
					subtest.Fatalf("unexpected recovery %v", err)
				}
			} else {
				if err != nil || orders.Load() != 2 || broadcasts.Load() != 1 {
					subtest.Fatalf("recovery %v, orders %d, broadcasts %d", err, orders.Load(), broadcasts.Load())
				}
			}
		})
	}
}
