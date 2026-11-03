package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/cagelin/polymarket-go/internal/endpoint"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/models"
)

// Config configures public requests. Its zero value uses production URLs.
type Config struct {
	// RawHTTPHandler observes public Gamma/Data/CLOB responses before decoding.
	// It runs synchronously, may run concurrently, and must honor ctx cancellation.
	// An error stops this call without retrying or decoding. Authenticated calls are excluded.
	RawHTTPHandler      func(context.Context, models.RawHTTPResponse) error
	RFQWSURL            string
	UserWSURL           string
	PriceWSURL          string
	RPCURL              string
	CollateralReturnURL string
	RelayerURL          string
	BuilderGatewayURL   string
	MarketWSURL         string
	SportsWSURL         string
	RTDSWSURL           string
	StreamQueueSize     int
	CLOBURL             string
	GammaURL            string
	DataURL             string
	RFQURL              string
	// MaxResponseBodyBytes bounds each decoded HTTP response, including snapshots.
	// Zero selects 32 MiB; callers downloading larger snapshots may raise it.
	MaxResponseBodyBytes int64
	// HTTPClient is borrowed and is never closed or modified by the SDK.
	HTTPClient *http.Client
	// OnRateLimitUpdate may run concurrently. Returned errors are ignored.
	OnRateLimitUpdate func(sdkerrors.RateLimitUpdate) error
}

// Client performs unauthenticated queries and supports concurrent use.
type Client struct {
	RPCURL               string
	CollateralReturnURL  string
	RelayerURL           string
	BuilderGatewayURL    string
	CLOBURL              string
	GammaURL             string
	DataURL              string
	RFQURL               string
	customRPC            bool
	maxResponseBodyBytes int64
	http                 *http.Client
	owned                *http.Transport
	held                 *http.Transport
	closed               atomic.Bool
	rawHTTPHandler       func(context.Context, models.RawHTTPResponse) error
	onRateLimitUpdate    func(sdkerrors.RateLimitUpdate) error
}

// New constructs a client without making any network requests.
func New(config Config) (*Client, error) {
	if config.MaxResponseBodyBytes < 0 || config.MaxResponseBodyBytes == math.MaxInt64 {
		return nil, &sdkerrors.UserInputError{Message: "invalid response body limit"}
	}
	if config.MaxResponseBodyBytes == 0 {
		config.MaxResponseBodyBytes = 32 << 20
	}
	clobURL, err := serviceURL(config.CLOBURL, "https://clob.polymarket.com")
	if err != nil {
		return nil, err
	}
	gammaURL, err := serviceURL(config.GammaURL, "https://gamma-api.polymarket.com")
	if err != nil {
		return nil, err
	}
	dataURL, err := serviceURL(config.DataURL, "https://data-api.polymarket.com")
	if err != nil {
		return nil, err
	}
	rfqURL, err := serviceURL(config.RFQURL, "https://combos-rfq-api.polymarket.com")
	if err != nil {
		return nil, err
	}
	relayerURL, err := serviceURL(config.RelayerURL, "https://relayer-v2.polymarket.com")
	if err != nil {
		return nil, err
	}
	builderURL, err := serviceURL(config.BuilderGatewayURL, "https://combos-rfq-gateway-builder.polymarket.com")
	if err != nil {
		return nil, err
	}
	rpcURL, err := serviceURL(config.RPCURL, "https://polygon.drpc.org")
	if err != nil {
		return nil, err
	}
	collateralURL, err := serviceURL(config.CollateralReturnURL, "https://combos-rfq-collateral-return.polymarket.com")
	if err != nil {
		return nil, err
	}
	c := &Client{customRPC: config.RPCURL != "", maxResponseBodyBytes: config.MaxResponseBodyBytes, CollateralReturnURL: collateralURL,
		RPCURL:     rpcURL,
		RelayerURL: relayerURL, BuilderGatewayURL: builderURL,
		RFQURL:            rfqURL,
		CLOBURL:           clobURL,
		GammaURL:          gammaURL,
		DataURL:           dataURL,
		http:              config.HTTPClient,
		rawHTTPHandler:    config.RawHTTPHandler,
		onRateLimitUpdate: config.OnRateLimitUpdate,
	}
	if c.http == nil {
		c.owned = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          20,
			MaxIdleConnsPerHost:   20,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: time.Second,
		}
		c.http = &http.Client{
			Transport:     c.owned,
			Timeout:       10 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}
	}
	if c.owned != nil {
		c.held = c.owned.Clone()
		// The request budget bounds header and body waits on this shared pool.
		c.held.ResponseHeaderTimeout = 0
	}
	return c, nil
}

func serviceURL(value, fallback string) (string, error) {
	if value == "" {
		value = fallback
	}
	return endpoint.Parse(value, "http", "https")
}

// Close releases owned idle connections and prevents new SDK requests.
// It is idempotent. Cancel each request's context to stop in-flight requests.
func (c *Client) Close() error {
	if !c.closed.Swap(true) && c.owned != nil {
		c.owned.CloseIdleConnections()
		c.held.CloseIdleConnections()
	}
	return nil
}

// Request performs one HTTP request without retrying.
func (c *Client) Request(ctx context.Context, baseURL, method, path string, query url.Values, body any) ([]byte, error) {
	return c.AuthenticatedRequest(ctx, baseURL, method, path, query, body, nil)
}

// HeaderResolver signs the exact bytes written to the network; it is invoked per attempt.
type HeaderResolver = func(method, path string, body []byte) (http.Header, error)

// AuthenticatedRequest applies optional authentication only to this request.
func (c *Client) AuthenticatedRequest(ctx context.Context, baseURL, method, path string, query url.Values, body any, headers HeaderResolver) ([]byte, error) {
	return c.request(ctx, baseURL, method, path, query, body, headers, 0)
}

// HeldRequest permits a thirty-second response wait for RFQ competition and acceptance.
func (c *Client) HeldRequest(ctx context.Context, baseURL, method, path string, body any, headers HeaderResolver) ([]byte, error) {
	return c.request(ctx, baseURL, method, path, nil, body, headers, 30*time.Second)
}

// LongRequest uses an explicit response budget for slow wallet workflows.
func (c *Client) LongRequest(ctx context.Context, baseURL, method, path string, body any, headers HeaderResolver, timeout time.Duration) ([]byte, error) {
	return c.request(ctx, baseURL, method, path, nil, body, headers, timeout)
}
func (c *Client) request(ctx context.Context, baseURL, method, path string, query url.Values, body any, headers HeaderResolver, timeout time.Duration) ([]byte, error) {
	if c.closed.Load() {
		return nil, &sdkerrors.UserInputError{Message: "client is closed"}
	}
	var reader io.Reader
	var payload []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, &sdkerrors.UserInputError{Message: fmt.Sprintf("encode request: %v", err)}
		}
		payload = encoded
		reader = bytes.NewReader(encoded)
	}
	requestURL := baseURL + path
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL, reader)
	if err != nil {
		return nil, &sdkerrors.TransportError{Cause: err}
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if headers != nil {
		signed, headerErr := headers(method, path, payload)
		if headerErr != nil {
			return nil, headerErr
		}
		for key, values := range signed {
			req.Header[key] = append([]string(nil), values...)
		}
	}
	client := c.http
	if timeout > 0 {
		requestClient := *client
		requestClient.Timeout = timeout
		if c.held != nil {
			requestClient.Transport = c.held
		}
		client = &requestClient
	}
	if headers != nil {
		requestClient := *client
		requestClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		client = &requestClient
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, &sdkerrors.TransportError{Cause: err}
	}
	defer func() { _ = res.Body.Close() }() // Reading below reports errors; closing cannot recover a failed response.
	data, err := io.ReadAll(io.LimitReader(res.Body, c.maxResponseBodyBytes+1))
	if c.rawHTTPHandler != nil && headers == nil && observablePublicRequest(baseURL, method, path, c) {
		service := "clob"
		switch baseURL {
		case c.GammaURL:
			service = "gamma"
		case c.DataURL:
			service = "data"
		}
		record := models.RawHTTPResponse{Service: service, Method: method, URL: req.URL.String(), RequestBody: bytes.Clone(payload), StatusCode: res.StatusCode, Header: res.Header.Clone(), ReceivedAt: time.Now().UTC(), Complete: err == nil && int64(len(data)) <= c.maxResponseBodyBytes}
		if err != nil {
			record.ReadError = err.Error()
		}
		if int64(len(data)) > c.maxResponseBodyBytes {
			record.Body = bytes.Clone(data[:c.maxResponseBodyBytes])
			record.ReadError = "HTTP response exceeds configured body limit"
		} else {
			record.Body = bytes.Clone(data)
		}
		if handlerErr := c.rawHTTPHandler(ctx, record); handlerErr != nil {
			return nil, &responseHandlerError{cause: handlerErr}
		}
	}

	if err != nil {
		return nil, &sdkerrors.TransportError{Cause: err}
	}
	if int64(len(data)) > c.maxResponseBodyBytes {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "HTTP response exceeds configured body limit"}
	}
	if c.onRateLimitUpdate != nil {
		if update := parseRateLimit(res.Header); update != nil {
			_ = c.onRateLimitUpdate(*update)
		}
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, responseError(res, data)
	}
	return data, nil
}

// DataRead retries short 429 delays according to the Data API read policy.
func (c *Client) DataRead(ctx context.Context, path string, query url.Values) ([]byte, error) {
	for attempt := 0; ; attempt++ {
		data, err := c.Request(ctx, c.DataURL, http.MethodGet, path, query, nil)
		if _, handlerFailed := errors.AsType[*responseHandlerError](err); handlerFailed {
			return nil, err
		}
		var rate *sdkerrors.RateLimitError
		if !errors.As(err, &rate) || attempt >= 2 {
			return data, err
		}
		delay := 1.0
		if rate.RetryAfter != nil {
			delay = *rate.RetryAfter
		}
		if delay > 5 {
			return nil, err
		}
		timer := time.NewTimer(time.Duration(math.Max(0, delay) * float64(time.Second)))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, &sdkerrors.TransportError{Cause: ctx.Err()}
		case <-timer.C:
		}
	}
}

// RelayerEndpoint returns the configured relayer base URL.
func (c *Client) RelayerEndpoint() string { return c.RelayerURL }

// CustomRPC reports whether an RPC endpoint was explicitly configured.
func (c *Client) CustomRPC() bool { return c.customRPC }

// Only unsigned reads are observable; wallet writes and authentication are excluded.
func observablePublicRequest(baseURL, method, path string, c *Client) bool {
	if baseURL != c.GammaURL && baseURL != c.DataURL && baseURL != c.CLOBURL {
		return false
	}
	if method == http.MethodGet {
		return true
	}
	if baseURL != c.CLOBURL || method != http.MethodPost {
		return false
	}
	switch path {
	case "/books", "/prices", "/midpoints", "/spreads", "/last-trades-prices":
		return true
	default:
		return false
	}
}

type responseHandlerError struct{ cause error }

func (e *responseHandlerError) Error() string {
	return "public HTTP response handler: " + e.cause.Error()
}
func (e *responseHandlerError) Unwrap() error { return e.cause }
