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
	"strings"
	"sync/atomic"
	"time"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
)

// Config configures public requests. Its zero value uses production URLs.
type Config struct {
	UserWSURL         string
	PriceWSURL        string
	RelayerURL        string
	BuilderGatewayURL string
	MarketWSURL       string
	SportsWSURL       string
	RTDSWSURL         string
	StreamQueueSize   int
	CLOBURL           string
	GammaURL          string
	DataURL           string
	RFQURL            string
	// HTTPClient is borrowed and is never closed or modified by the SDK.
	HTTPClient *http.Client
	// OnRateLimitUpdate may run concurrently. Returned errors are ignored.
	OnRateLimitUpdate func(sdkerrors.RateLimitUpdate) error
}

// Client performs unauthenticated queries and supports concurrent use.
type Client struct {
	RelayerURL        string
	BuilderGatewayURL string
	CLOBURL           string
	GammaURL          string
	DataURL           string
	RFQURL            string
	http              *http.Client
	owned             *http.Transport
	held              *http.Transport
	closed            atomic.Bool
	onRateLimitUpdate func(sdkerrors.RateLimitUpdate) error
}

// New constructs a client without making any network requests.
func New(config Config) (*Client, error) {
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
	c := &Client{
		RelayerURL: relayerURL, BuilderGatewayURL: builderURL,
		RFQURL:            rfqURL,
		CLOBURL:           clobURL,
		GammaURL:          gammaURL,
		DataURL:           dataURL,
		http:              config.HTTPClient,
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
		c.held.ResponseHeaderTimeout = 30 * time.Second
	}
	return c, nil
}

func serviceURL(value, fallback string) (string, error) {
	if value == "" {
		value = fallback
	}
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", &sdkerrors.UserInputError{Message: "service URL must be an absolute HTTP(S) URL without credentials, query or fragment"}
	}
	return strings.TrimRight(value, "/"), nil
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
type HeaderResolver func(method, path string, body []byte) (http.Header, error)

// AuthenticatedRequest applies optional authentication only to this request.
func (c *Client) AuthenticatedRequest(ctx context.Context, baseURL, method, path string, query url.Values, body any, headers HeaderResolver) ([]byte, error) {
	return c.request(ctx, baseURL, method, path, query, body, headers, false)
}

// HeldRequest permits a thirty-second response wait for RFQ competition and acceptance.
func (c *Client) HeldRequest(ctx context.Context, baseURL, method, path string, body any, headers HeaderResolver) ([]byte, error) {
	return c.request(ctx, baseURL, method, path, nil, body, headers, true)
}
func (c *Client) request(ctx context.Context, baseURL, method, path string, query url.Values, body any, headers HeaderResolver, held bool) ([]byte, error) {
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
	endpoint := baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
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
	if held {
		requestClient := *client
		requestClient.Timeout = 30 * time.Second
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
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, &sdkerrors.TransportError{Cause: err}
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
