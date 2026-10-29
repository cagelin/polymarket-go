package wallet

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"slices"
	"strconv"
	"sync/atomic"
	"testing"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/models"
)

// The real Gamma endpoint omits V2 unless requested and defaults to open markets.
func TestMarketContextProtocolFilters(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/gamma/wallet-market-lookup.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Markets []json.RawMessage }
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Markets) != 2 {
		t.Fatalf("market fixtures = %d, want 2", len(fixture.Markets))
	}
	for _, marketRaw := range fixture.Markets {
		var market struct {
			ID, ConditionID string
			Closed          bool
			PositionIDs     []string
		}
		if err = json.Unmarshal(marketRaw, &market); err != nil {
			t.Fatal(err)
		}
		for _, selector := range []string{"condition", "market"} {
			t.Run(market.ID+"/"+selector, func(subtest *testing.T) {
				var requests atomic.Int32
				client := testClient(subtest, models.WalletDeposit, func(writer http.ResponseWriter, request *http.Request) {
					requests.Add(1)
					if request.Method != http.MethodGet || request.URL.Path != "/markets/keyset" {
						subtest.Errorf("marketContext request = %s %s, want GET /markets/keyset", request.Method, request.URL.Path)
						writer.WriteHeader(http.StatusBadRequest)
						return
					}
					query := request.URL.Query()
					expectedKey, expectedID := "condition_ids", market.ConditionID
					if selector == "market" {
						expectedKey, expectedID = "id", market.ID
					}
					if query.Get(expectedKey) != expectedID || query.Get("limit") != "1" {
						subtest.Errorf("marketContext(%s) query = %v, want exact ID and limit=1", selector, query)
					}
					var rows []json.RawMessage
					versions := query["version"]
					if slices.Contains(versions, "v2") && query.Get("closed") == strconv.FormatBool(market.Closed) {
						rows = append(rows, marketRaw)
					}
					if !slices.Contains(versions, "v1") {
						subtest.Errorf("marketContext versions = %v, want V1 retained", versions)
					}
					if encodeErr := json.NewEncoder(writer).Encode(map[string]any{"markets": rows, "next_cursor": nil}); encodeErr != nil {
						subtest.Error(encodeErr)
					}
				})
				selectedCondition, selectedMarket := market.ConditionID, ""
				if selector == "market" {
					selectedCondition, selectedMarket = "", market.ID
				}
				actual, lookupErr := client.marketContext(subtest.Context(), selectedCondition, selectedMarket, &market.Closed)
				if lookupErr != nil {
					subtest.Fatalf("marketContext(%s,%s,closed=%t) = %v, want a V2 market", selectedCondition, selectedMarket, market.Closed, lookupErr)
				}
				if !actual.v2 || actual.condition != market.ConditionID || len(actual.ids) != 2 || actual.ids[0].String() != market.PositionIDs[0] || actual.ids[1].String() != market.PositionIDs[1] || requests.Load() != 1 {
					subtest.Errorf("marketContext(%s) = %+v, requests=%d, want matching V2 IDs in one read", market.ID, actual, requests.Load())
				}
			})
		}
	}
}

func TestMarketContextV1AndClosedDefaults(t *testing.T) {
	for _, tc := range []struct {
		name, version string
		closed        bool
		wantError     bool
	}{
		{name: "open-v1", version: "v1"},
		{name: "closed-v2-not-an-open-split", version: "v2", closed: true, wantError: true},
	} {
		t.Run(tc.name, func(subtest *testing.T) {
			var writes atomic.Int32
			client := testClient(subtest, models.WalletDeposit, func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodGet || request.URL.Path != "/markets/keyset" {
					writes.Add(1)
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
				query := request.URL.Query()
				if query.Has("closed") {
					subtest.Errorf("default closed query = %q, want omitted", query.Get("closed"))
				}
				rows := make([]map[string]any, 0)
				if (slices.Contains(query["version"], tc.version) || (len(query["version"]) == 0 && tc.version == "v1")) && !tc.closed {
					rows = append(rows, map[string]any{"id": "1", "conditionId": "0x1234", "version": "v1", "closed": false, "negRisk": false, "outcomes": []string{"Yes", "No"}, "clobTokenIds": []string{"123", "124"}})
				}
				if encodeErr := json.NewEncoder(writer).Encode(map[string]any{"markets": rows, "next_cursor": nil}); encodeErr != nil {
					subtest.Error(encodeErr)
				}
			})
			actual, lookupErr := client.marketContext(subtest.Context(), "0x1234", "", nil)
			if tc.wantError {
				if _, ok := errors.AsType[*sdkerrors.UserInputError](lookupErr); !ok {
					subtest.Errorf("marketContext(closed market) = %v, want UserInputError", lookupErr)
				}
			} else if lookupErr != nil || actual.v2 || actual.condition != "0x1234" {
				subtest.Errorf("marketContext(V1) = %+v,%v, want V1 context", actual, lookupErr)
			}
			if writes.Load() != 0 {
				subtest.Errorf("market lookup writes = %d, want 0", writes.Load())
			}
		})
	}
}
