package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/models"
)

type rawTestTransport func(*http.Request) (*http.Response, error)

func (r rawTestTransport) RoundTrip(request *http.Request) (*http.Response, error) { return r(request) }

type brokenRawBody struct{ sent bool }

func (b *brokenRawBody) Read(buffer []byte) (int, error) {
	if b.sent {
		return 0, io.ErrUnexpectedEOF
	}
	b.sent = true
	return copy(buffer, "part"), nil
}
func (*brokenRawBody) Close() error { return nil }

func TestRawPartialBodyAndAuthenticationExclusion(t *testing.T) {
	var records []models.RawHTTPResponse
	client, err := New(Config{HTTPClient: &http.Client{Transport: rawTestTransport(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: &brokenRawBody{}, Request: request}, nil
	})}, RawHTTPHandler: func(_ context.Context, response models.RawHTTPResponse) error {
		records = append(records, response)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	_, err = client.Request(t.Context(), client.GammaURL, http.MethodGet, "/events", nil, nil)
	if !errors.Is(err, io.ErrUnexpectedEOF) || len(records) != 1 || records[0].Complete || string(records[0].Body) != "part" || records[0].ReadError == "" {
		t.Fatalf("partial raw response = %+v, %v", records, err)
	}
	_, err = client.AuthenticatedRequest(t.Context(), client.CLOBURL, http.MethodGet, "/data/orders", nil, nil, func(_, _ string, _ []byte) (http.Header, error) { return http.Header{"Authorization": {"test"}}, nil })
	if !errors.Is(err, io.ErrUnexpectedEOF) || len(records) != 1 {
		t.Fatalf("authenticated response observed: %+v, %v", records, err)
	}
}

func TestRawMutationDoesNotAffectDecodingAndBatchBody(t *testing.T) {
	client, err := New(Config{HTTPClient: &http.Client{Transport: rawTestTransport(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"price":"0.5"}`)), Request: request}, nil
	})}, RawHTTPHandler: func(_ context.Context, response models.RawHTTPResponse) error {
		if response.Method != http.MethodPost || string(response.RequestBody) != `[{"token_id":"1"}]` {
			t.Errorf("batch request = %+v", response)
		}
		response.Body[0] = '!'
		response.RequestBody[0] = '!'
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	payload, err := client.Request(t.Context(), client.CLOBURL, http.MethodPost, "/prices", nil, []map[string]string{{"token_id": "1"}})
	if err != nil || string(payload) != `{"price":"0.5"}` {
		t.Errorf("mutated raw changed body = %s, %v", payload, err)
	}
}

func TestRawCallbackRateLimitErrorDoesNotRetry(t *testing.T) {
	calls := 0
	failure := &sdkerrors.RateLimitError{}
	client, err := New(Config{HTTPClient: &http.Client{Transport: rawTestTransport(func(request *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{}`)), Request: request}, nil
	})}, RawHTTPHandler: func(context.Context, models.RawHTTPResponse) error { return failure }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	_, err = client.DataRead(t.Context(), "/v2/trades", nil)
	if !errors.Is(err, failure) || calls != 1 {
		t.Errorf("handler 429 = %v, calls=%d; want cause and one call", err, calls)
	}
}

func TestRawCancellationBeforeResponse(t *testing.T) {
	records := 0
	client, err := New(Config{HTTPClient: &http.Client{Transport: rawTestTransport(func(request *http.Request) (*http.Response, error) { return nil, request.Context().Err() })}, RawHTTPHandler: func(context.Context, models.RawHTTPResponse) error { records++; return nil }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = client.Request(ctx, client.GammaURL, http.MethodGet, "/events", nil, nil)
	if !errors.Is(err, context.Canceled) || records != 0 {
		t.Errorf("cancelled request = %v, records=%d", err, records)
	}
}
