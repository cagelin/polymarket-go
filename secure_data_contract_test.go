package polymarket_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	sdk "github.com/cagelin/polymarket-go"
)

func TestSecureDataUserSelection(t *testing.T) {
	operations := []struct {
		name, path string
		call       func(context.Context, *sdk.SecureClient, string) error
	}{
		{name: "ListTrades", path: "/v2/trades", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.ListTradesParams{}
			if user != "" {
				params.User = &user
			}
			pages, err := client.ListTrades(params)
			if err != nil {
				return err
			}
			_, err = pages.FirstPage(ctx)
			return err
		}},
		{name: "ListActivity", path: "/v2/activity", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.ListActivityParams{}
			params.User = user
			pages, err := client.ListActivity(params)
			if err != nil {
				return err
			}
			_, err = pages.FirstPage(ctx)
			return err
		}},
		{name: "ListComboActivity", path: "/v2/activity/combos", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.ListComboActivityParams{}
			params.User = user
			pages, err := client.ListComboActivity(params)
			if err != nil {
				return err
			}
			_, err = pages.FirstPage(ctx)
			return err
		}},
		{name: "ListPositions", path: "/v2/positions", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.ListPositionsParams{}
			if user != "" {
				params.User = &user
			}
			pages, err := client.ListPositions(params)
			if err != nil {
				return err
			}
			_, err = pages.FirstPage(ctx)
			return err
		}},
		{name: "ListComboPositions", path: "/v2/positions/combos", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.ListComboPositionsParams{}
			params.User = user
			pages, err := client.ListComboPositions(params)
			if err != nil {
				return err
			}
			_, err = pages.FirstPage(ctx)
			return err
		}},
		{name: "GetPortfolioValue", path: "/v2/value", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.GetPortfolioValueParams{}
			params.User = user
			_, err := client.GetPortfolioValue(ctx, params)
			return err
		}},
		{name: "GetUserStats", path: "/v2/user-stats", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.GetUserStatsParams{}
			params.User = user
			_, err := client.GetUserStats(ctx, params)
			return err
		}},
		{name: "GetUserPnL", path: "/v2/user-pnl", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.GetUserPnLParams{}
			params.User = user
			_, err := client.GetUserPnL(ctx, params)
			return err
		}},
		{name: "GetUserVolume", path: "/v2/user-volume", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.GetUserVolumeParams{}
			params.User = user
			_, err := client.GetUserVolume(ctx, params)
			return err
		}},
		{name: "GetTraderLeaderboardStanding", path: "/v2/leaderboard", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.GetTraderLeaderboardStandingParams{}
			params.User = user
			_, err := client.GetTraderLeaderboardStanding(ctx, params)
			return err
		}},
		{name: "DownloadAccountingSnapshot", path: "/v1/accounting/snapshot", call: func(ctx context.Context, client *sdk.SecureClient, user string) error {
			params := sdk.DownloadAccountingSnapshotParams{}
			params.User = user
			_, err := client.DownloadAccountingSnapshot(ctx, params)
			return err
		}},
	}
	for _, operation := range operations {
		for _, selectedUser := range []string{"", "0x1111111111111111111111111111111111111111"} {
			t.Run(operation.name+"/"+selectedUser, func(subtest *testing.T) {
				var reads atomic.Int32
				expectedUser := selectedUser
				if expectedUser == "" {
					expectedUser = testSigner
				}
				server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if request.URL.Path == "/auth/api-keys" {
						writeTestJSON(subtest, writer, map[string]any{"apiKeys": []string{"synthetic-key"}})
						return
					}
					reads.Add(1)
					if request.Method != http.MethodGet || request.URL.Path != operation.path || request.URL.Query().Get("user") != expectedUser {
						subtest.Errorf("%s(user=%q) request = %s %s, want GET %s with user=%q", operation.name, selectedUser, request.Method, request.URL, operation.path, expectedUser)
					}
					if request.Header.Get("POLY_API_KEY") != "" || request.Header.Get("POLY_SIGNATURE") != "" {
						subtest.Errorf("%s public Data request included authentication headers", operation.name)
					}
					writer.WriteHeader(http.StatusServiceUnavailable)
				}))
				subtest.Cleanup(server.Close)
				client, err := sdk.NewSecureClient(subtest.Context(), secureConfig(server.URL))
				if err != nil {
					subtest.Fatalf("NewSecureClient(local server) = %v, want nil", err)
				}
				subtest.Cleanup(func() {
					if closeErr := client.Close(); closeErr != nil {
						subtest.Error(closeErr)
					}
				})
				err = operation.call(subtest.Context(), client, selectedUser)
				rejection, ok := errors.AsType[*sdk.RequestRejectedError](err)
				if !ok || rejection.Status != http.StatusServiceUnavailable {
					subtest.Errorf("%s(user=%q) error = %v, want preserved HTTP 503", operation.name, selectedUser, err)
				}
				if reads.Load() != 1 {
					subtest.Errorf("%s(user=%q) reads = %d, want 1 without retry", operation.name, selectedUser, reads.Load())
				}
			})
		}
	}
}
