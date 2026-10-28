package signing

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/cagelam/polymarket-go/models"
)

func TestPythonSigning(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/signing-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Operation, Key, Secret, Method, Path, Exchange, Error string
			Timestamp, Nonce                                      int64
			Body                                                  *string
			Session                                               bool
			Order                                                 models.SignedOrder
			Expected                                              json.RawMessage
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, test := range fixture.Cases {
		t.Run(test.Operation, func(subtest *testing.T) {
			var value any
			var operationErr error
			if test.Operation == "hmac" {
				var body []byte
				if test.Body != nil {
					body = []byte(*test.Body)
				}
				value, operationErr = HMAC(test.Secret, test.Timestamp, test.Method, test.Path, body)
			} else {
				signer, createErr := New(test.Key)
				if createErr != nil {
					subtest.Fatal(createErr)
				}
				switch test.Operation {
				case "l1":
					headers, signErr := signer.L1(test.Timestamp, test.Nonce)
					operationErr = signErr
					flat := map[string]string{}
					for key, entries := range headers {
						flat[key] = entries[0]
					}
					value = flat
				case "wallets":
					value = signer.Wallets()
				case "order":
					authority := models.SignerOwner
					if test.Session {
						authority = models.SignerSessionKey
					}
					value, operationErr = signer.Order(test.Order, test.Exchange, authority)
				default:
					subtest.Fatal(test.Operation)
				}
			}
			if (operationErr != nil) != (test.Error != "") {
				subtest.Fatalf("%s error=%v; Python=%s", test.Operation, operationErr, test.Error)
			}
			if operationErr != nil {
				return
			}
			encoded, encodeErr := json.Marshal(value)
			if encodeErr != nil {
				subtest.Fatal(encodeErr)
			}
			var got, want any
			if err = json.Unmarshal(encoded, &got); err != nil {
				subtest.Fatal(err)
			}
			if err = json.Unmarshal(test.Expected, &want); err != nil {
				subtest.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				subtest.Errorf("%s got=%s; Python=%s", test.Operation, encoded, test.Expected)
			}
		})
	}
}

func BenchmarkL1Signature(b *testing.B) {
	signer, err := New("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err = signer.L1(1720000000, 0); err != nil {
			b.Fatal(err)
		}
	}
}
