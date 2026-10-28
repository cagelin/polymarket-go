package trading

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/cagelam/polymarket-go/internal/signing"
	"github.com/cagelam/polymarket-go/internal/transport"
	"github.com/cagelam/polymarket-go/models"
)

type accountInput struct {
	AssetID       *string   `json:"asset_id"`
	ID            *string   `json:"id"`
	Market        *string   `json:"market"`
	MakerAddress  *string   `json:"maker_address"`
	After         *string   `json:"after"`
	Before        *string   `json:"before"`
	OrderID       string    `json:"order_id"`
	OrderIDs      []string  `json:"order_ids"`
	IDs           []string  `json:"ids"`
	AssetType     AssetType `json:"asset_type"`
	HeartbeatID   string    `json:"heartbeat_id"`
	Date          string    `json:"date"`
	Sponsored     *bool     `json:"sponsored"`
	NoCompetition *bool     `json:"no_competition"`
	OrderBy       *string   `json:"order_by"`
	Position      *string   `json:"position"`
	PageSize      *int      `json:"page_size"`
}

func callAccount(ctx context.Context, c *Client, op string, p accountInput, cursor *string) (any, error) {
	switch op {
	case "fetch_api_keys":
		return c.FetchAPIKeys(ctx)
	case "delete_api_key":
		return nil, c.DeleteAPIKey(ctx)
	case "create_builder_api_key":
		return c.CreateBuilderAPIKey(ctx)
	case "fetch_builder_api_keys":
		return c.FetchBuilderAPIKeys(ctx)
	case "get_closed_only_mode":
		return c.GetClosedOnlyMode(ctx)
	case "list_open_orders":
		pages, err := c.ListOpenOrders(ListOpenOrdersParams{AssetID: p.AssetID, ID: p.ID, ConditionID: p.Market})
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pages = pages.FromCursor(cursor)
		}
		return pages.FirstPage(ctx)
	case "get_order":
		return c.GetOrder(ctx, p.OrderID)
	case "list_account_trades":
		pages, err := c.ListAccountTrades(ListAccountTradesParams{AssetID: p.AssetID, ID: p.ID, ConditionID: p.Market, MakerAddress: p.MakerAddress, After: p.After, Before: p.Before})
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pages = pages.FromCursor(cursor)
		}
		return pages.FirstPage(ctx)
	case "get_notifications":
		return c.GetNotifications(ctx)
	case "drop_notifications":
		return nil, c.DropNotifications(ctx, p.IDs)
	case "get_balance_allowance":
		return c.GetBalanceAllowance(ctx, BalanceAllowanceParams{AssetType: p.AssetType, AssetID: p.AssetID})
	case "send_order_heartbeat":
		return c.SendOrderHeartbeat(ctx, p.HeartbeatID)
	case "cancel_order":
		return c.CancelOrder(ctx, p.OrderID)
	case "cancel_orders":
		return c.CancelOrders(ctx, p.OrderIDs)
	case "cancel_all":
		return c.CancelAll(ctx)
	case "cancel_market_orders":
		return c.CancelMarketOrders(ctx, CancelMarketOrdersParams{ConditionID: p.Market, AssetID: p.AssetID})
	case "get_order_scoring":
		return c.GetOrderScoring(ctx, p.OrderID)
	case "get_orders_scoring":
		return c.GetOrdersScoring(ctx, p.OrderIDs)
	case "list_user_earnings_for_day":
		pages, err := c.ListUserEarningsForDay(p.Date)
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pages = pages.FromCursor(cursor)
		}
		return pages.FirstPage(ctx)
	case "get_total_earnings_for_user_for_day":
		return c.GetTotalEarningsForUserForDay(ctx, GetTotalEarningsForUserForDayParams{Date: p.Date, Sponsored: p.Sponsored})
	case "list_user_earnings_and_markets_config":
		pages, err := c.ListUserEarningsAndMarketsConfig(ListUserEarningsAndMarketsConfigParams{Date: p.Date, NoCompetition: p.NoCompetition, OrderBy: p.OrderBy, Position: p.Position, PageSize: p.PageSize})
		if err != nil {
			return nil, err
		}
		if cursor != nil {
			pages = pages.FromCursor(cursor)
		}
		return pages.FirstPage(ctx)
	case "get_reward_percentages":
		return c.GetRewardPercentages(ctx)
	default:
		panic("unknown test operation")
	}
}
func equalJSON(t *testing.T, actual any, expected json.RawMessage) {
	t.Helper()
	raw, err := json.Marshal(actual)
	if err != nil {
		t.Fatal(err)
	}
	var got, want any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err = decoder.Decode(&got); err != nil {
		t.Fatal(err)
	}
	decoder = json.NewDecoder(bytes.NewReader(expected))
	decoder.UseNumber()
	if err = decoder.Decode(&want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %s want %s", raw, expected)
	}
}
func TestPythonAccountQueries(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/account-queries-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, Operation string
			Input           accountInput
			Cursor          *string
			Response        json.RawMessage
			Expected        struct {
				Value json.RawMessage
				Error string
			}
			Request *struct {
				Method, Path string
				Query        map[string][]string
				Body         json.RawMessage
			}
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, tc := range fixture.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			var calls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if tc.Request == nil {
					t.Error("request after invalid input")
					w.WriteHeader(500)
					return
				}
				if r.Method != tc.Request.Method || r.URL.Path != tc.Request.Path {
					t.Errorf("got %s %s want %s %s", r.Method, r.URL.Path, tc.Request.Method, tc.Request.Path)
				}
				if !reflect.DeepEqual(map[string][]string(r.URL.Query()), tc.Request.Query) {
					t.Errorf("query %v want %v", r.URL.Query(), tc.Request.Query)
				}
				body, readErr := io.ReadAll(r.Body)
				if readErr != nil {
					t.Error(readErr)
				}
				if len(body) == 0 {
					body = []byte("null")
				}
				equalJSON(t, json.RawMessage(body), tc.Request.Body)
				if r.Header.Get("POLY_ADDRESS") == "" || r.Header.Get("POLY_API_KEY") != "key" || r.Header.Get("POLY_SIGNATURE") == "" {
					t.Error("missing authentication")
				}
				w.Header().Set("Content-Type", "application/json")
				if _, writeErr := w.Write(tc.Response); writeErr != nil {
					t.Error(writeErr)
				}
			}))
			defer server.Close()
			tr, newErr := transport.New(transport.Config{CLOBURL: server.URL})
			if newErr != nil {
				t.Fatal(newErr)
			}
			defer func() { _ = tr.Close() }()
			signer, signErr := signing.New("0000000000000000000000000000000000000000000000000000000000000001")
			if signErr != nil {
				t.Fatal(signErr)
			}
			//nolint:gosec // Deliberately public synthetic credentials for an in-memory server.
			//noinspection SpellCheckingInspection
			client := &Client{transport: tr, signer: signer, walletType: models.WalletEOA, credentials: models.APIKeyCreds{Key: "key", Secret: "c2VjcmV0", Passphrase: "pass"}, now: time.Now}
			result, callErr := callAccount(t.Context(), client, tc.Operation, tc.Input, tc.Cursor)
			if (callErr != nil) != (tc.Expected.Error != "") {
				t.Fatalf("error %v; Python %s", callErr, tc.Expected.Error)
			}
			if callErr == nil {
				equalJSON(t, result, tc.Expected.Value)
			}
			if tc.Request != nil && calls != 1 {
				t.Errorf("request count %d", calls)
			}
		})
	}
}
