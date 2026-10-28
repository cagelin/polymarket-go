package polymarket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestPythonDiscoveryParity(t *testing.T) {
	data, err := os.ReadFile("testdata/discovery-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	testPythonDiscoveryParity(t, data)
}

func TestPythonEventsParity(t *testing.T) {
	data, err := os.ReadFile("testdata/events-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	testPythonDiscoveryParity(t, data)
}

func TestPythonApprovals(t *testing.T) {
	raw, err := os.ReadFile("testdata/approvals-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	testPythonDiscoveryParity(t, raw)
}

func TestPythonPublicClobQueries(t *testing.T) {
	raw, err := os.ReadFile("testdata/public-clob-queries-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	testPythonDiscoveryParity(t, raw)
}

func TestPythonCatalogQueries(t *testing.T) {
	raw, err := os.ReadFile("testdata/catalog-queries-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	testPythonDiscoveryParity(t, raw)
}

func TestPythonDataQueries(t *testing.T) {
	raw, err := os.ReadFile("testdata/data-queries-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	testPythonDiscoveryParity(t, raw)
}

func testPythonDiscoveryParity(t *testing.T, data []byte) {
	t.Helper()
	var fixture struct {
		Version string `json:"upstream_version"`
		Commit  string `json:"upstream_commit"`
		Cases   []struct {
			Name                    string
			CompatibilityDifference string          `json:"compatibility_difference"`
			PythonValue             json.RawMessage `json:"python_value"`
			Operation               string
			Input                   json.RawMessage
			Response                json.RawMessage
			Cursor                  *string
			Request                 *struct {
				Path  string
				Query map[string][]string
			}
			Expected struct {
				Value json.RawMessage
				Error string
			}
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != UpstreamVersion || fixture.Commit != UpstreamCommit {
		t.Fatal("fixture source mismatch")
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(subtest *testing.T) {
			subtest.Parallel()
			if test.CompatibilityDifference != "" {
				if test.CompatibilityDifference != "UP-001" || len(test.PythonValue) == 0 {
					subtest.Fatalf("unregistered compatibility difference %q", test.CompatibilityDifference)
				}
				if string(test.PythonValue) == string(test.Expected.Value) {
					subtest.Fatal("registered compatibility difference must retain the differing Python result")
				}
			}
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if test.Request == nil {
					subtest.Error("invalid input sent a request")
					w.WriteHeader(500)
					return
				}
				if r.Method != "GET" || r.URL.EscapedPath() != test.Request.Path {
					subtest.Errorf("request = %s %s; Python = GET %s", r.Method, r.URL.EscapedPath(), test.Request.Path)
				}
				if !reflect.DeepEqual(map[string][]string(r.URL.Query()), test.Request.Query) {
					subtest.Errorf("query = %#v; Python = %#v", r.URL.Query(), test.Request.Query)
				}
				w.Header().Set("Content-Type", "application/json")
				if test.Operation == "download_accounting_snapshot" {
					_, _ = w.Write([]byte("accounting data"))
				} else {
					_, _ = w.Write(test.Response)
				}
			}))
			defer server.Close()
			client, err := NewPublicClient(ClientConfig{CLOBURL: server.URL, RFQURL: server.URL, GammaURL: server.URL, DataURL: server.URL})
			if err != nil {
				subtest.Fatal(err)
			}
			defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
			var result any
			switch test.Operation {
			case "get_series", "get_tag", "get_related_tags", "get_related_tag_resources", "get_sports", "get_sports_market_types", "get_public_profile", "get_comment_thread", "list_series", "list_tags", "list_teams", "list_comments", "list_comments_by_user_address":
				result, err = callCatalogParity(subtest.Context(), client, test.Operation, test.Input, test.Cursor)
			case "list_builder_trades", "list_reward_markets", "list_current_rewards", "list_market_rewards", "get_current_rebates", "list_combo_markets":
				result, err = callPublicClobParity(subtest.Context(), client, test.Operation, test.Input, test.Cursor)
			case "get_trading_approvals_state":
				var input struct{ Wallet EVMAddress }
				if err = json.Unmarshal(test.Input, &input); err != nil {
					subtest.Fatal(err)
				}
				result, err = client.GetTradingApprovalsState(subtest.Context(), input.Wallet)
			case "get_event":
				var query EventQuery
				if err = json.Unmarshal(test.Input, &query); err != nil {
					subtest.Fatal(err)
				}
				result, err = client.GetEvent(subtest.Context(), query)
			case "get_event_tags", "get_market_tags":
				var query struct{ ID string }
				if err = json.Unmarshal(test.Input, &query); err != nil {
					subtest.Fatal(err)
				}
				if test.Operation == "get_event_tags" {
					result, err = client.GetEventTags(subtest.Context(), EventID(query.ID))
				} else {
					result, err = client.GetMarketTags(subtest.Context(), MarketID(query.ID))
				}
			case "list_events":
				var params ListEventsParams
				if err = json.Unmarshal(test.Input, &params); err != nil {
					subtest.Fatal(err)
				}
				var pager *Paginator[Event]
				pager, err = client.ListEvents(params)
				if err == nil {
					if test.Cursor != nil {
						pager = pager.FromCursor(test.Cursor)
					}
					result, err = pager.FirstPage(subtest.Context())
				}
			case "search":
				var params SearchParams
				if err = json.Unmarshal(test.Input, &params); err != nil {
					subtest.Fatal(err)
				}
				var pager *Paginator[SearchResults]
				pager, err = client.Search(params)
				if err == nil {
					if test.Cursor != nil {
						pager = pager.FromCursor(test.Cursor)
					}
					result, err = pager.FirstPage(subtest.Context())
				}
			case "get_market":
				var query MarketQuery
				if err = json.Unmarshal(test.Input, &query); err != nil {
					subtest.Fatal(err)
				}
				result, err = client.GetMarket(subtest.Context(), query)
			case "list_markets":
				var params ListMarketsParams
				if err = json.Unmarshal(test.Input, &params); err != nil {
					subtest.Fatal(err)
				}
				var pager *Paginator[Market]
				pager, err = client.ListMarkets(params)
				if err == nil {
					if test.Cursor != nil {
						pager = pager.FromCursor(test.Cursor)
					}
					result, err = pager.FirstPage(subtest.Context())
				}
			case "history":
				var params PriceHistoryParams
				if err = json.Unmarshal(test.Input, &params); err != nil {
					subtest.Fatal(err)
				}
				var pager *Paginator[PriceHistoryPoint]
				pager, err = client.ListPriceHistory(params)
				if err == nil {
					if test.Cursor != nil {
						pager = pager.FromCursor(test.Cursor)
					}
					result, err = pager.FirstPage(subtest.Context())
				}
			default:
				result, err = callDataParity(subtest.Context(), client, test.Operation, test.Input, test.Cursor)
			}
			if got := errorClass(err); got != test.Expected.Error {
				subtest.Fatalf("error = %s (%v), Python = %s", got, err, test.Expected.Error)
			}
			wantCalls := int32(1)
			if test.Request == nil {
				wantCalls = 0
			}
			if calls.Load() != wantCalls {
				subtest.Errorf("request count = %d, want %d", calls.Load(), wantCalls)
			}
			if err == nil {
				encoded, encodeErr := json.Marshal(result)
				if encodeErr != nil {
					subtest.Fatal(encodeErr)
				}
				compareJSON(subtest, encoded, test.Expected.Value)
			}
		})
	}
}
