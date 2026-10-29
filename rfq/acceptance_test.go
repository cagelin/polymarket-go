package rfq

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
	"github.com/coder/websocket"
	"github.com/ericlagergren/decimal"
)

func TestQuoterCancelConfirmAndDecline(t *testing.T) {
	for _, decision := range []string{"CANCEL", "CONFIRM", "DECLINE"} {
		t.Run(decision, func(subtest *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				connection, acceptErr := websocket.Accept(writer, request, nil)
				if acceptErr != nil {
					subtest.Error(acceptErr)
					return
				}
				defer func() { _ = connection.CloseNow() }()
				ctx, cancel := context.WithTimeout(request.Context(), 5*time.Second)
				defer cancel()
				readFrame := func() map[string]any {
					_, raw, readErr := connection.Read(ctx)
					if readErr != nil {
						subtest.Error(readErr)
						return nil
					}
					var body map[string]any
					if decodeErr := json.Unmarshal(raw, &body); decodeErr != nil {
						subtest.Error(decodeErr)
					}
					return body
				}
				sendFrame := func(raw string) {
					if writeErr := connection.Write(ctx, websocket.MessageText, []byte(raw)); writeErr != nil {
						subtest.Error(writeErr)
					}
				}
				if readFrame()["type"] != "auth" {
					subtest.Error("missing authentication")
					return
				}
				sendFrame(`{"type":"auth","success":true}`)
				sendFrame(`{"type":"RFQ_REQUEST","rfq_id":"r","requestor_public_id":"peer","leg_position_ids":["123","456"],"condition_id":"0x03` + strings.Repeat("ab", 30) + `","yes_position_id":"123","no_position_id":"124","direction":"BUY","side":"YES","requested_size":{"unit":"notional","value_e6":"50000"},"submission_deadline":2000000000}`)
				quote := readFrame()
				if quote["type"] != "RFQ_QUOTE" || quote["price_e6"] != "500000" || quote["size_e6"] != "100000" {
					subtest.Errorf("wrong bounded quote: %v", quote)
					return
				}
				sendFrame(`{"type":"ACK_RFQ_QUOTE","rfq_id":"r","quote_id":"q"}`)
				response := readFrame()
				if response["rfq_id"] != "r" || response["quote_id"] != "q" {
					subtest.Error("lost quote correlation")
					return
				}
				acknowledgement := "ACK_RFQ_CONFIRMATION_RESPONSE"
				if decision == "CANCEL" {
					acknowledgement = "ACK_RFQ_QUOTE_CANCEL"
					if response["type"] != "RFQ_QUOTE_CANCEL" {
						subtest.Error("wrong cancel message")
					}
				} else if response["type"] != "RFQ_CONFIRMATION_RESPONSE" || response["decision"] != decision {
					subtest.Error("wrong confirmation decision")
				}
				sendFrame(fmt.Sprintf(`{"type":%q,"rfq_id":"r","quote_id":"q"}`, acknowledgement))
				status := "FAILED"
				if decision == "CONFIRM" {
					status = "CONFIRMED"
				}
				sendFrame(fmt.Sprintf(`{"type":"RFQ_EXECUTION_UPDATE","rfq_id":"r","quote_id":"q","status":%q,"tx_hash":"0x%s"}`, status, strings.Repeat("a", 64)))
				_, _, _ = connection.Read(ctx)
			}))
			defer server.Close()
			signer, err := signing.New("0x" + strings.Repeat("0", 63) + "1")
			if err != nil {
				subtest.Fatal(err)
			}
			session, err := Open(subtest.Context(), Config{URL: "ws" + strings.TrimPrefix(server.URL, "http"), Signer: signer, Wallet: models.EVMAddress(signer.Address()), WalletType: models.WalletEOA, Credentials: models.APIKeyCreds{Key: "test", Secret: base64.StdEncoding.EncodeToString([]byte("test")), Passphrase: "test"}})
			if err != nil {
				subtest.Fatal(err)
			}
			defer func() { _ = session.Close() }()
			event, err := session.Next(subtest.Context())
			if err != nil {
				subtest.Fatal(err)
			}
			reference, err := session.Quote(subtest.Context(), event, QuoteParams{Price: decimal.New(5, 1), Size: decimal.New(1, 1), Source: "collateral"})
			if err != nil {
				subtest.Fatal(err)
			}
			var acknowledged QuoteReference
			if decision == "CANCEL" {
				acknowledged, err = session.CancelQuote(subtest.Context(), reference)
			} else {
				acknowledged, err = session.RespondToConfirmation(subtest.Context(), reference, decision)
			}
			if err != nil || acknowledged != reference {
				subtest.Fatalf("acknowledgement=%+v err=%v", acknowledged, err)
			}
			result, err := session.Next(subtest.Context())
			if err != nil {
				subtest.Fatal(err)
			}
			if (result.Status == "CONFIRMED") != (decision == "CONFIRM") {
				subtest.Fatalf("decision=%s result=%s", decision, result.Status)
			}
		})
	}
}
