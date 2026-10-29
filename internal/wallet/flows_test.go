package wallet

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/internal/transport"
	"github.com/cagelin/polymarket-go/models"
)

func testClient(t *testing.T, kind models.WalletType, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	signer, err := signing.New("0x0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	client, err := transport.New(transport.Config{RPCURL: server.URL, RelayerURL: server.URL, CLOBURL: server.URL, GammaURL: server.URL, CollateralReturnURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	address := signer.Address()
	if kind != models.WalletEOA {
		address = signer.Wallets()["BEACON"]
	}
	wallet, err := New(client, Config{Signer: signer, Wallet: models.EVMAddress(address), Kind: kind, Authority: models.SignerOwner, BuilderKey: &models.BuilderAPIKey{Key: "test", Secret: base64.StdEncoding.EncodeToString([]byte("secret")), Passphrase: "test"}, PollInterval: time.Millisecond, MaxPolls: 3, CLOBHeaders: func(string, string, []byte) (http.Header, error) { return http.Header{}, nil }})
	if err != nil {
		t.Fatal(err)
	}
	wallet.now = func() time.Time { return time.Unix(1720000000, 0) }
	return wallet
}
func rpcReply(t *testing.T, w http.ResponseWriter, r *http.Request, values map[string]any) string {
	t.Helper()
	var request struct {
		ID     uint64
		Method string
		Params []json.RawMessage
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		t.Error(err)
		return ""
	}
	value, ok := values[request.Method]
	if !ok {
		t.Errorf("unexpected RPC %s", request.Method)
		w.WriteHeader(500)
		return request.Method
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": value})
	return request.Method
}
func TestEOASequentialConfirmationAndNoBlindRetry(t *testing.T) {
	for _, mode := range []string{"success", "wrong-chain", "unknown-submit", "reverted", "timeout"} {
		t.Run(mode, func(subtest *testing.T) {
			var writes, receipts atomic.Int32
			wallet := testClient(subtest, models.WalletEOA, func(w http.ResponseWriter, r *http.Request) {
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
					if mode == "wrong-chain" {
						result = "0x1"
					}
				case "eth_getTransactionCount":
					result = "0x0"
				case "eth_gasPrice":
					result = "0x1"
				case "eth_estimateGas":
					result = "0x5208"
				case "eth_sendRawTransaction":
					count := writes.Add(1)
					if count > 1 && receipts.Load() == 0 {
						subtest.Error("broadcast second EOA call before confirming first")
					}
					if mode == "unknown-submit" {
						w.WriteHeader(502)
						return
					}
					result = "0x" + strings.Repeat("ab", 32)
				case "eth_getTransactionReceipt":
					receipts.Add(1)
					switch mode {
					case "timeout":
						result = nil
					case "reverted":
						result = map[string]string{"status": "0x0"}
					default:
						result = map[string]string{"status": "0x1"}
					}
				default:
					subtest.Errorf("unexpected %s", request.Method)
					w.WriteHeader(500)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
			})
			calls := []models.TransactionCall{{To: wallet.config.Wallet, Data: "0x"}, {To: wallet.config.Wallet, Data: "0x"}}
			handle, err := wallet.ExecuteTransaction(subtest.Context(), calls, nil)
			if mode == "success" {
				if err != nil {
					subtest.Fatal(err)
				}
				if _, err = handle.Wait(subtest.Context()); err != nil {
					subtest.Fatal(err)
				}
				if writes.Load() != 2 {
					subtest.Fatal("missing broadcast")
				}
			} else {
				if err == nil {
					subtest.Fatal("expected failure")
				}
				if mode == "wrong-chain" && writes.Load() != 0 {
					subtest.Fatal("broadcast on wrong chain")
				}
				if mode != "wrong-chain" && writes.Load() != 1 {
					subtest.Fatalf("ambiguous write repeated: %d", writes.Load())
				}
			}
		})
	}
}
func TestDepositNonceRepairAndConfirmation(t *testing.T) {
	var submits, polls atomic.Int32
	wallet := testClient(t, models.WalletDeposit, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			rpcReply(t, w, r, map[string]any{"eth_getBlockByNumber": map[string]string{"timestamp": "0x66850a00"}})
		case "/v1/account/transactions/params":
			_, _ = fmt.Fprint(w, `{"address":"0x1111111111111111111111111111111111111111","nonce":"19"}`)
		case "/submit":
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			count := submits.Add(1)
			want := `"19"`
			if count == 2 {
				want = `"20"`
			}
			if string(body["nonce"]) != want {
				t.Error("nonce not rebuilt")
			}
			if count == 1 {
				w.WriteHeader(400)
				_, _ = fmt.Fprint(w, `{"error":"batch nonce 19 does not match on-chain nonce 20"}`)
			} else {
				_, _ = fmt.Fprint(w, `{"transactionID":"tx-1","state":"STATE_NEW"}`)
			}
		case "/v1/account/transactions/tx-1":
			state := "STATE_MINED"
			if polls.Add(1) > 1 {
				state = "STATE_CONFIRMED"
			}
			_, _ = fmt.Fprintf(w, `{"transaction_id":"tx-1","transaction_hash":"0x123","state":%q}`, state)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(500)
		}
	})
	handle, err := wallet.ExecuteTransaction(t.Context(), []models.TransactionCall{{To: wallet.config.Wallet, Data: "0x"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := handle.Wait(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if outcome.TransactionHash != "0x123" || submits.Load() != 2 || polls.Load() != 2 {
		t.Fatalf("unexpected result %+v", outcome)
	}
}
func TestRelayerDoesNotRetryUnknownOutcome(t *testing.T) {
	for _, status := range []int{400, 500, 502} {
		t.Run(fmt.Sprint(status), func(subtest *testing.T) {
			var submits atomic.Int32
			wallet := testClient(subtest, models.WalletDeposit, func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/":
					rpcReply(subtest, w, r, map[string]any{"eth_getBlockByNumber": map[string]string{"timestamp": "0x66850a00"}})
				case "/v1/account/transactions/params":
					_, _ = fmt.Fprint(w, `{"address":"0x1111111111111111111111111111111111111111","nonce":"19"}`)
				case "/submit":
					submits.Add(1)
					w.WriteHeader(status)
					_, _ = fmt.Fprint(w, `{"error":"unknown outcome"}`)
				default:
					subtest.Error(r.URL.Path)
				}
			})
			_, err := wallet.ExecuteTransaction(subtest.Context(), []models.TransactionCall{{To: wallet.config.Wallet, Data: "0x"}}, nil)
			if err == nil || submits.Load() != 1 {
				subtest.Fatalf("err %v submits %d", err, submits.Load())
			}
		})
	}
}
func TestInvalidCallsAndCancellationCannotSubmit(t *testing.T) {
	var requests atomic.Int32
	wallet := testClient(t, models.WalletEOA, func(w http.ResponseWriter, _ *http.Request) { requests.Add(1); w.WriteHeader(500) })
	for _, calls := range [][]models.TransactionCall{nil, {{To: "bad", Data: "0x"}}, {{To: wallet.config.Wallet, Data: "0x"}, {To: wallet.config.Wallet, Data: "0x", Value: big.NewInt(-1)}}} {
		if _, err := wallet.ExecuteTransaction(t.Context(), calls, nil); err == nil {
			t.Fatal("accepted invalid calls")
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := wallet.ExecuteTransaction(ctx, []models.TransactionCall{{To: wallet.config.Wallet, Data: "0x"}}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if requests.Load() != 0 {
		t.Fatal("invalid batch reached network")
	}
}
func TestSessionOwnerRestriction(t *testing.T) {
	wallet := testClient(t, models.WalletEOA, func(_ http.ResponseWriter, _ *http.Request) { t.Error("unexpected request") })
	_, err := wallet.FetchSessionKeys(t.Context())
	if _, ok := errors.AsType[*sdkerrors.UserInputError](err); !ok {
		t.Fatal(err)
	}
}
