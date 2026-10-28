package rfq

import (
	"encoding/json"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/cagelam/polymarket-go/internal/signing"
	"github.com/cagelam/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

func TestPythonMakerQuotes(t *testing.T) {
	raw, err := os.ReadFile("../testdata/rfq-session-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Wallet        string
			Kind          models.WalletType
			Request       json.RawMessage
			Price, Source string
			Expected      map[string]any
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	signer, err := signing.New("0x0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range fixture.Cases {
		t.Run(strconv.Itoa(i), func(subtest *testing.T) {
			request, parseErr := parseEvent(row.Request)
			if parseErr != nil {
				subtest.Fatal(parseErr)
			}
			price, ok := new(decimal.Big).SetString(row.Price)
			if !ok {
				subtest.Fatal("invalid fixture")
			}
			session := Session{config: Config{Signer: signer, Wallet: models.EVMAddress(row.Wallet), WalletType: row.Kind}}
			body, signErr := session.signedQuote(request, QuoteParams{Price: price, Source: row.Source}, big.NewInt(19), 1720000000)
			if signErr != nil {
				subtest.Fatal(signErr)
			}
			encoded, encodeErr := json.Marshal(body)
			if encodeErr != nil {
				subtest.Fatal(encodeErr)
			}
			var actual map[string]any
			if decodeErr := json.Unmarshal(encoded, &actual); decodeErr != nil {
				subtest.Fatal(decodeErr)
			}
			if !reflect.DeepEqual(actual, row.Expected) {
				subtest.Fatalf("quote differs from Python\ngot %s\nwant %v", encoded, row.Expected)
			}
		})
	}
}
