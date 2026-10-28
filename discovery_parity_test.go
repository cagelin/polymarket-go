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
	data, readErr := os.ReadFile("testdata/discovery-parity.json")
	if readErr != nil {
		t.Fatal(readErr)
	}
	var fixture struct {
		Version string `json:"upstream_version"`
		Commit  string `json:"upstream_commit"`
		Cases   []struct {
			Name      string
			Operation string
			Input     json.RawMessage
			Response  json.RawMessage
			Cursor    *string
			Request   *struct {
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
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if test.Request == nil {
					t.Error("invalid input sent a request")
					w.WriteHeader(500)
					return
				}
				if r.Method != "GET" || r.URL.EscapedPath() != test.Request.Path {
					t.Errorf("request = %s %s; Python = GET %s", r.Method, r.URL.EscapedPath(), test.Request.Path)
				}
				if !reflect.DeepEqual(map[string][]string(r.URL.Query()), test.Request.Query) {
					t.Errorf("query = %#v; Python = %#v", r.URL.Query(), test.Request.Query)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(test.Response)
			}))
			defer server.Close()
			client, err := NewPublicClient(ClientConfig{GammaURL: server.URL, DataURL: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			var result any
			switch test.Operation {
			case "get_market":
				var query MarketQuery
				if err = json.Unmarshal(test.Input, &query); err != nil {
					t.Fatal(err)
				}
				result, err = client.GetMarket(t.Context(), query)
			case "list_markets":
				var params ListMarketsParams
				if err = json.Unmarshal(test.Input, &params); err != nil {
					t.Fatal(err)
				}
				var pager *Paginator[Market]
				pager, err = client.ListMarkets(params)
				if err == nil {
					if test.Cursor != nil {
						pager = pager.FromCursor(test.Cursor)
					}
					result, err = pager.FirstPage(t.Context())
				}
			case "history":
				var params PriceHistoryParams
				if err = json.Unmarshal(test.Input, &params); err != nil {
					t.Fatal(err)
				}
				var pager *Paginator[PriceHistoryPoint]
				pager, err = client.ListPriceHistory(params)
				if err == nil {
					if test.Cursor != nil {
						pager = pager.FromCursor(test.Cursor)
					}
					result, err = pager.FirstPage(t.Context())
				}
			default:
				t.Fatal("unknown operation")
			}
			if got := errorClass(err); got != test.Expected.Error {
				t.Fatalf("error = %s (%v), Python = %s", got, err, test.Expected.Error)
			}
			wantCalls := int32(1)
			if test.Request == nil {
				wantCalls = 0
			}
			if calls.Load() != wantCalls {
				t.Errorf("request count = %d, want %d", calls.Load(), wantCalls)
			}
			if err == nil {
				encoded, encodeErr := json.Marshal(result)
				if encodeErr != nil {
					t.Fatal(encodeErr)
				}
				compareJSON(t, encoded, test.Expected.Value)
			}
		})
	}
}
