package trading

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
)

func TestPythonComboParity(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/combo-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Request map[string]any
		Cases   []struct {
			Name, Operation, Wallet string
			WalletType              models.WalletType `json:"wallet_type"`
			Input                   json.RawMessage
			Expected                struct {
				Value json.RawMessage
				Error string
			}
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	// Match the concrete request types used in production, rather than JSON's []any.
	legs := []string{"123", "456"}
	fixture.Request["leg_position_ids"] = legs
	for _, tc := range fixture.Cases {
		t.Run(tc.Name, func(subtest *testing.T) {
			signer, signErr := signing.New("0000000000000000000000000000000000000000000000000000000000000001")
			if signErr != nil {
				subtest.Fatal(signErr)
			}
			client := &Client{signer: signer, wallet: models.EVMAddress(signer.Address()), walletType: models.WalletEOA, signerType: models.SignerOwner, now: func() time.Time { return time.Unix(1720000000, 0) }, salt: func(int) (*big.Int, error) { return big.NewInt(12345), nil }}
			var result any
			var callErr error
			switch tc.Operation {
			case "request":
				var p struct {
					LegPositionIDs []string `json:"leg_position_ids"`
					Direction      models.OrderSide
					Amount, Size   *string
					Side           *string
				}
				if err = json.Unmarshal(tc.Input, &p); err != nil {
					subtest.Fatal(err)
				}
				params := RequestComboQuoteParams{LegPositionIDs: p.LegPositionIDs, Direction: p.Direction, Side: p.Side}
				if p.Amount != nil {
					params.Amount = dec(*p.Amount)
				}
				if p.Size != nil {
					params.Size = dec(*p.Size)
				}
				result, callErr = client.comboRequest(params)
			case "quote":
				result, callErr = parseComboQuote(tc.Input, fixture.Request)
			case "model":
				var quote models.ComboQuote
				callErr = json.Unmarshal(tc.Input, &quote)
				result = quote
			case "status":
				result, callErr = parseRFQStatus(tc.Input, "rfq-1")
			case "accept":
				client.wallet = models.EVMAddress(tc.Wallet)
				client.walletType = tc.WalletType
				var quote models.ComboQuote
				if err = json.Unmarshal(tc.Input, &quote); err != nil {
					subtest.Fatal(err)
				}
				result, callErr = client.comboAcceptanceBody(quote)
			}
			if (callErr != nil) != (tc.Expected.Error != "") {
				subtest.Fatalf("error %v; Python %s", callErr, tc.Expected.Error)
			}
			if callErr == nil {
				equalJSON(subtest, result, tc.Expected.Value)
			}
		})
	}
}
