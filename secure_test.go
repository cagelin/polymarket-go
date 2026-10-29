package polymarket_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	sdk "github.com/cagelin/polymarket-go"
)

const testSigner = "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf"
const testPrivate = "0000000000000000000000000000000000000000000000000000000000000001"

func testCredentials() *sdk.APIKeyCreds {
	//nolint:gosec // Public synthetic credentials used only with httptest servers.
	//noinspection SpellCheckingInspection
	return &sdk.APIKeyCreds{Key: "synthetic-key", Secret: "c2VjcmV0", Passphrase: "synthetic-pass"}
}
func secureConfig(address string) sdk.SecureClientConfig {
	return sdk.SecureClientConfig{CLOBURL: address, RelayerURL: address, DataURL: address, PrivateKey: testPrivate, Wallet: testSigner, Credentials: testCredentials()}
}
func writeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Error(err)
	}
}
func TestSecureCredentialLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name              string
		probe, create     int
		provided          bool
		derive, wantError bool
	}{
		{name: "existing", probe: 200, provided: true}, {name: "401-create", probe: 401, provided: true, create: 200}, {name: "missing-create", create: 200}, {name: "400-derive", create: 400, derive: true}, {name: "403-not-retried", probe: 403, provided: true, wantError: true}, {name: "500-not-derived", create: 500, wantError: true},
	} {
		t.Run(tc.name, func(subtest *testing.T) {
			var posts, derives, deletes atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method + " " + r.URL.Path {
				case "GET /auth/api-keys":
					if r.Header.Get("POLY_API_KEY") != "synthetic-key" || r.Header.Get("POLY_SIGNATURE") == "" {
						subtest.Error("missing L2 credentials")
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tc.probe)
					writeTestJSON(subtest, w, map[string]any{"apiKeys": []string{"synthetic-key"}})
				case "POST /auth/api-key", "GET /auth/derive-api-key":
					if r.Header.Get("POLY_ADDRESS") != testSigner || r.Header.Get("POLY_NONCE") != "0" || r.Header.Get("POLY_SIGNATURE") == "" || r.Header.Get("POLY_API_KEY") != "" {
						subtest.Error("invalid L1 authentication")
					}
					status := 200
					if r.Method == "POST" {
						posts.Add(1)
						status = tc.create
					} else {
						derives.Add(1)
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					//nolint:gosec // Public synthetic credentials; no funded account.
					//noinspection SpellCheckingInspection
					writeTestJSON(subtest, w, map[string]string{"apiKey": "synthetic-key", "secret": "c2VjcmV0", "passphrase": "synthetic-pass"})
				case "DELETE /auth/api-key":
					deletes.Add(1)
					writeTestJSON(subtest, w, "OK")
				case "GET /midpoint":
					if r.Header.Get("POLY_API_KEY") != "" {
						subtest.Error("credentials leaked into public request")
					}
					writeTestJSON(subtest, w, map[string]string{"mid": "0.5"})
				default:
					subtest.Errorf("unexpected %s %s", r.Method, r.URL)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			cfg := secureConfig(server.URL)
			if !tc.provided {
				cfg.Credentials = nil
			}
			client, err := sdk.NewSecureClient(subtest.Context(), cfg)
			if (err != nil) != tc.wantError {
				subtest.Fatalf("constructor error %v", err)
			}
			if err != nil {
				return
			}
			defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
			if client.Wallet() != testSigner || client.Signer() != testSigner || client.WalletType() != sdk.WalletEOA {
				subtest.Fatal("wrong identity")
			}
			if credentials := client.Credentials(); credentials.Key != "synthetic-key" || credentials.Passphrase != "synthetic-pass" {
				subtest.Fatal("Credentials() differs from the authenticated account")
			}
			if _, err = client.GetMidpoint(subtest.Context(), sdk.Asset("123")); err != nil {
				subtest.Fatal(err)
			}
			public, err := client.EndAuthentication(subtest.Context())
			if err != nil {
				subtest.Fatal(err)
			}
			defer func() { _ = public.Close() }() // Close only releases local resources and always returns nil.
			if deletes.Load() != 1 {
				subtest.Fatal("credential revocation not performed")
			}
			if tc.derive != (derives.Load() == 1) {
				subtest.Fatal("incorrect derive fallback")
			}
			if tc.name == "existing" && posts.Load() != 0 {
				subtest.Fatal("valid credentials recreated")
			}
			if _, err = client.FetchAPIKeys(subtest.Context()); err == nil {
				subtest.Fatal("closed client still usable")
			}
		})
	}
}
func TestSecureRevocationErrorsAndRedaction(t *testing.T) {
	for _, status := range []int{401, 404, 500} {
		t.Run(fmt.Sprint(status), func(subtest *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "DELETE" {
					w.WriteHeader(status)
					return
				}
				writeTestJSON(subtest, w, map[string]any{"apiKeys": []string{"synthetic-key"}})
			}))
			defer server.Close()
			client, err := sdk.NewSecureClient(subtest.Context(), secureConfig(server.URL))
			if err != nil {
				subtest.Fatal(err)
			}
			defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
			public, err := client.EndAuthentication(subtest.Context())
			if status == 500 && err == nil {
				subtest.Fatal("revocation failure suppressed")
			}
			if status != 500 && err != nil {
				subtest.Fatal(err)
			}
			if public != nil {
				defer func() { _ = public.Close() }() // Close only releases local resources and always returns nil.
			}
			if _, err = client.FetchAPIKeys(subtest.Context()); err == nil {
				subtest.Fatal("revocation failure left client open")
			}
		})
	}
	for _, value := range []any{secureConfig("http://localhost"), *testCredentials(), sdk.BuilderAPIKey{Key: "secret-value", Secret: "secret-value", Passphrase: "secret-value"}} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			printed := fmt.Sprintf(format, value)
			if strings.Contains(printed, "synthetic") || strings.Contains(printed, testPrivate) || strings.Contains(printed, "secret-value") {
				t.Fatalf("sensitive format output: %s", printed)
			}
		}
	}
}
func TestAuthenticatedRedirectDoesNotLeak(t *testing.T) {
	var leaked atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { leaked.Store(true); w.WriteHeader(200) }))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	config := secureConfig(source.URL)
	config.HTTPClient = &http.Client{}
	_, err := sdk.NewSecureClient(t.Context(), config)
	if err == nil || leaked.Load() {
		t.Fatalf("redirect forwarded authenticated request: error %v leak %v", err, leaked.Load())
	}
}
func TestSecureDefaultWalletAndUser(t *testing.T) {
	var deployedChecks atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/deployed":
			deployedChecks.Add(1)
			writeTestJSON(t, w, map[string]bool{"deployed": true})
		case "/auth/api-keys":
			writeTestJSON(t, w, map[string]any{"apiKeys": []string{"synthetic-key"}})
		case "/v2/positions":
			if r.URL.Query().Get("user") == "" || r.URL.Query().Get("user") == testSigner {
				t.Error("position query did not default to holding wallet")
			}
			writeTestJSON(t, w, map[string]any{"data": []any{}, "pagination": map[string]any{"next_cursor": nil, "has_more": false}})
		default:
			t.Errorf("unexpected %s", r.URL)
			w.WriteHeader(500)
		}
	}))
	defer server.Close()
	config := secureConfig(server.URL)
	config.Wallet = ""
	client, err := sdk.NewSecureClient(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	if client.WalletType() != sdk.WalletDeposit || client.Wallet() == client.Signer() {
		t.Fatal("wrong default wallet")
	}
	pages, err := client.ListPositions(sdk.ListPositionsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pages.FirstPage(t.Context()); err != nil {
		t.Fatal(err)
	}
	if deployedChecks.Load() != 2 {
		t.Fatal("wallet readiness not checked")
	}
}
