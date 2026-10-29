package trading

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

type marketInfo struct {
	tick, rate, exponent *decimal.Big
	negativeRisk         bool
	assets               map[string]bool
	expires              time.Time
}
type builderInfo struct {
	rates   models.BuilderFeeRates
	expires time.Time
}
type metadataCache struct {
	once       sync.Once
	mu         sync.Mutex
	keyMu      sync.Mutex
	keys       map[string]*metadataGate
	conditions map[string]string
	markets    map[string]marketInfo
	builders   map[string]builderInfo
}

type metadataGate struct {
	semaphore chan struct{}
	users     int
}

func (m *metadataCache) lock(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.once.Do(func() {
		m.conditions = map[string]string{}
		m.markets = map[string]marketInfo{}
		m.builders = map[string]builderInfo{}
	})
	m.mu.Lock()
	return nil
}
func (m *metadataCache) unlock() { m.mu.Unlock() }

// lockKey merges concurrent loads without blocking unrelated metadata keys.
func (m *metadataCache) lockKey(ctx context.Context, key string) (func(), error) {
	m.keyMu.Lock()
	if m.keys == nil {
		m.keys = make(map[string]*metadataGate)
	}
	entry := m.keys[key]
	if entry == nil {
		entry = &metadataGate{semaphore: make(chan struct{}, 1)}
		m.keys[key] = entry
	}
	entry.users++
	m.keyMu.Unlock()
	drop := func() {
		m.keyMu.Lock()
		entry.users--
		if entry.users == 0 {
			delete(m.keys, key)
		}
		m.keyMu.Unlock()
	}
	select {
	case entry.semaphore <- struct{}{}:
		return func() { <-entry.semaphore; drop() }, nil
	case <-ctx.Done():
		drop()
		return nil, ctx.Err()
	}
}
func (c *Client) marketCondition(ctx context.Context, asset string) (string, error) {
	cache := &c.metadata
	release, err := cache.lockKey(ctx, "asset:"+asset)
	if err != nil {
		return "", err
	}
	defer release()
	if err = cache.lock(ctx); err != nil {
		return "", err
	}
	condition := cache.conditions[asset]
	cache.unlock()
	if condition != "" {
		return condition, nil
	}
	raw, err := c.transport.Request(ctx, c.transport.CLOBURL, "GET", "/markets-by-token/"+url.PathEscape(asset), nil, nil)
	if err != nil {
		return "", err
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return "", response(err)
	}
	condition, err = codec.Required[string](fields, "condition_id")
	if err != nil {
		return "", response(err)
	}
	condition, err = codec.Condition(condition, "feed")
	if err != nil {
		return "", response(err)
	}
	if err = cache.lock(ctx); err != nil {
		return "", err
	}
	cache.conditions[asset] = condition
	cache.unlock()
	return condition, nil
}
func (c *Client) market(ctx context.Context, asset string, force bool) (marketInfo, error) {
	condition, err := c.marketCondition(ctx, asset)
	if err != nil {
		return marketInfo{}, err
	}
	cache := &c.metadata
	release, err := cache.lockKey(ctx, "market:"+condition)
	if err != nil {
		return marketInfo{}, err
	}
	defer release()
	if err = cache.lock(ctx); err != nil {
		return marketInfo{}, err
	}
	value, exists := cache.markets[condition]
	cache.unlock()
	if force || !exists || !c.now().Before(value.expires) {
		raw, requestErr := c.transport.Request(ctx, c.transport.CLOBURL, "GET", "/clob-markets/"+url.PathEscape(condition), nil, nil)
		if requestErr != nil {
			return marketInfo{}, requestErr
		}
		value, err = parseMarket(raw)
		if err != nil {
			return marketInfo{}, response(err)
		}
		value.expires = c.now().Add(10 * time.Minute)
		if err = cache.lock(ctx); err != nil {
			return marketInfo{}, err
		}
		cache.markets[condition] = value
		cache.unlock()
	}
	if !value.assets[asset] {
		if err = cache.lock(ctx); err != nil {
			return marketInfo{}, err
		}
		delete(cache.conditions, asset)
		delete(cache.markets, condition)
		cache.unlock()
		return marketInfo{}, response(fmt.Errorf("market does not contain requested asset"))
	}
	return value, nil
}
func parseMarket(raw []byte) (marketInfo, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return marketInfo{}, err
	}
	tick, err := marketDecimal(fields, "mts")
	if err != nil {
		return marketInfo{}, err
	}
	if _, err = tickRounding(tick); err != nil {
		return marketInfo{}, err
	}
	result := marketInfo{tick: tick, rate: decimal.New(0, 0), exponent: decimal.New(0, 0), assets: map[string]bool{}}
	if _, ok := fields["nr"]; ok {
		result.negativeRisk, err = codec.Required[bool](fields, "nr")
		if err != nil {
			return marketInfo{}, err
		}
	}
	tokens, err := codec.ParseList[json.RawMessage](fields["t"])
	if err != nil {
		return marketInfo{}, err
	}
	for _, token := range tokens {
		item, parseErr := codec.Object(token)
		if parseErr != nil {
			return marketInfo{}, parseErr
		}
		asset, parseErr := codec.Required[string](item, "t")
		if parseErr != nil {
			return marketInfo{}, parseErr
		}
		if asset == "" {
			return marketInfo{}, fmt.Errorf("empty market asset")
		}
		result.assets[asset] = true
	}
	if fee := fields["fd"]; len(fee) > 0 && string(fee) != "null" {
		params, parseErr := codec.Object(fee)
		if parseErr != nil {
			return marketInfo{}, parseErr
		}
		for key, target := range map[string]**decimal.Big{"r": &result.rate, "e": &result.exponent} {
			if number := params[key]; len(number) > 0 && string(number) != "null" {
				*target, err = marketDecimal(params, key)
				if err != nil {
					return marketInfo{}, err
				}
			}
		}
	}
	return result, nil
}
func validateBuilder(code string) error {
	if len(code) != 66 || !codec.ValidHex(code) {
		return input("builder code must be a 0x-prefixed 32-byte hex string")
	}
	return nil
}

// GetBuilderFeeRates retrieves builder rates without account authentication.
func (c *Client) GetBuilderFeeRates(ctx context.Context, code string) (*models.BuilderFeeRates, error) {
	if err := validateBuilder(code); err != nil {
		return nil, err
	}
	if code == signing.ZeroBytes32 {
		return nil, input("builder code must identify a real builder")
	}
	raw, err := c.transport.Request(ctx, c.transport.CLOBURL, "GET", "/fees/builder-fees/"+code, nil, nil)
	var rejected *sdkerrors.RequestRejectedError
	if errors.As(err, &rejected) && rejected.Status == 404 {
		return nil, input("unknown builder code")
	}
	if err != nil {
		return nil, err
	}
	value, err := codec.DecodeResponse[models.BuilderFeeRates](raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
func (c *Client) builderRate(ctx context.Context, code string) (*decimal.Big, error) {
	if code == "" || code == signing.ZeroBytes32 {
		return decimal.New(0, 0), nil
	}
	cache := &c.metadata
	release, err := cache.lockKey(ctx, "builder:"+code)
	if err != nil {
		return nil, err
	}
	defer release()
	if err = cache.lock(ctx); err != nil {
		return nil, err
	}
	value, exists := cache.builders[code]
	cache.unlock()
	if !exists || !c.now().Before(value.expires) {
		rates, requestErr := c.GetBuilderFeeRates(ctx, code)
		if requestErr != nil {
			return nil, requestErr
		}
		value = builderInfo{rates: *rates, expires: c.now().Add(10 * time.Minute)}
		if err = cache.lock(ctx); err != nil {
			return nil, err
		}
		cache.builders[code] = value
		cache.unlock()
	}
	return new(decimal.Big).Copy(value.rates.Taker), nil
}

func marketDecimal(fields map[string]json.RawMessage, key string) (*decimal.Big, error) {
	raw := fields[key]
	if len(raw) > 0 && raw[0] != '"' {
		var number json.Number
		if err := json.Unmarshal(raw, &number); err != nil {
			return nil, err
		}
		if number == "" {
			return nil, fmt.Errorf("missing number")
		}
		quoted, err := json.Marshal(string(number))
		if err != nil {
			return nil, err
		}
		return codec.DecimalField(map[string]json.RawMessage{key: quoted}, key)
	}
	return codec.DecimalField(fields, key)
}
