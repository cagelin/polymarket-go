package trading

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/cagelam/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

func dec(value string) *decimal.Big { result, _ := new(decimal.Big).SetString(value); return result }
func TestPythonOrderMath(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/order-math-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Operation, Tick, Price, Amount, Side, Rate, Exponent, Spend, Builder, Error string
			Expected                                                                    []string
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for i, tc := range fixture.Cases {
		t.Run(fmt.Sprintf("%s/%d", tc.Operation, i), func(t *testing.T) {
			var actual []string
			var failure error
			switch tc.Operation {
			case "limit", "market":
				fn := limitAmounts
				if tc.Operation == "market" {
					fn = marketAmounts
				}
				maker, taker, mathErr := fn(dec(tc.Price), dec(tc.Amount), dec(tc.Tick), models.OrderSide(tc.Side))
				failure = mathErr
				if failure == nil {
					actual = []string{maker.String(), taker.String()}
				}
			case "fees":
				value, mathErr := adjustBuyAmount(dec(tc.Amount), dec(tc.Price), dec(tc.Spend), dec(tc.Rate), dec(tc.Exponent), dec(tc.Builder))
				failure = mathErr
				if failure == nil {
					actual = []string{value.String()}
				}
			case "grid":
				failure = validatePrice(dec(tc.Price), dec(tc.Tick))
				actual = []string{tc.Price}
			}
			if (failure != nil) != (tc.Error != "") {
				t.Fatalf("error %v; Python %s", failure, tc.Error)
			}
			if failure != nil {
				return
			}
			for j, value := range actual {
				if dec(value).Cmp(dec(tc.Expected[j])) != 0 {
					t.Fatalf("got %v want %v", actual, tc.Expected)
				}
			}
		})
	}
}
