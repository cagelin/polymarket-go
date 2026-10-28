package polymarket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

// ClientConfig configures public requests. Its zero value uses production URLs.
type ClientConfig struct {
	CLOBURL  string
	GammaURL string
	DataURL  string
	// HTTPClient is borrowed and is never closed or modified by the SDK.
	HTTPClient *http.Client
	// OnRateLimitUpdate may run concurrently. Returned errors are ignored.
	OnRateLimitUpdate func(RateLimitUpdate) error
}

// PublicClient performs unauthenticated queries and supports concurrent use.
type PublicClient struct {
	clobURL           string
	gammaURL          string
	dataURL           string
	http              *http.Client
	owned             *http.Transport
	closed            atomic.Bool
	onRateLimitUpdate func(RateLimitUpdate) error
}

// NewPublicClient constructs a client without making any network requests.
func NewPublicClient(config ClientConfig) (*PublicClient, error) {
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
	c := &PublicClient{
		clobURL:           clobURL,
		gammaURL:          gammaURL,
		dataURL:           dataURL,
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
	return c, nil
}

func serviceURL(value, fallback string) (string, error) {
	if value == "" {
		value = fallback
	}
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", &UserInputError{Message: "service URL must be an absolute HTTP(S) URL without credentials, query or fragment"}
	}
	return strings.TrimRight(value, "/"), nil
}

// Close releases owned idle connections and prevents new SDK requests.
// It is idempotent. Cancel each request's context to stop in-flight requests.
func (c *PublicClient) Close() error {
	if !c.closed.Swap(true) && c.owned != nil {
		c.owned.CloseIdleConnections()
	}
	return nil
}

func (c *PublicClient) request(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	return c.requestAt(ctx, c.clobURL, method, path, query, body)
}

func (c *PublicClient) requestAt(ctx context.Context, baseURL, method, path string, query url.Values, body any) ([]byte, error) {
	if c.closed.Load() {
		return nil, &UserInputError{Message: "client is closed"}
	}
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, &UserInputError{Message: fmt.Sprintf("encode request: %v", err)}
		}
		reader = bytes.NewReader(encoded)
	}
	endpoint := baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, &TransportError{Cause: err}
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, &TransportError{Cause: err}
	}
	defer res.Body.Close() //nolint:errcheck // Body is read below; closing after a read failure cannot recover the response.
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, &TransportError{Cause: err}
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

func decodeResponse[T any](data []byte) (T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return result, &UnexpectedResponseError{Message: "response did not match expected shape", Cause: err}
	}
	return result, nil
}
