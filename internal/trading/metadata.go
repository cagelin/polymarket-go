package trading

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/signing"
	"github.com/cagelam/polymarket-go/models"
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
	gate       chan struct{}
	conditions map[string]string
	markets    map[string]marketInfo
	builders   map[string]builderInfo
}

func (m *metadataCache) lock(ctx context.Context) error {
	m.once.Do(func() {
		m.gate = make(chan struct{}, 1)
		m.conditions = map[string]string{}
		m.markets = map[string]marketInfo{}
		m.builders = map[string]builderInfo{}
	})
	select {
	case m.gate <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (m *metadataCache) unlock() { <-m.gate }
func (c *Client) market(ctx context.Context, asset string, force bool) (marketInfo, error) {
	cache := &c.metadata
	if err := cache.lock(ctx); err != nil {
		return marketInfo{}, err
	}
	defer cache.unlock()
	condition := cache.conditions[asset]
	if condition == "" {
		raw, err := c.transport.Request(ctx, c.transport.CLOBURL, "GET", "/markets-by-token/"+url.PathEscape(asset), nil, nil)
		if err != nil {
			return marketInfo{}, err
		}
		fields, err := codec.Object(raw)
		if err != nil {
			return marketInfo{}, response(err)
		}
		condition, err = codec.Required[string](fields, "condition_id")
		if err != nil {
			return marketInfo{}, response(err)
		}
		condition, err = codec.Condition(condition, "feed")
		if err != nil {
			return marketInfo{}, response(err)
		}
		cache.conditions[asset] = condition
	}
	value, exists := cache.markets[condition]
	if force || !exists || !c.now().Before(value.expires) {
		raw, err := c.transport.Request(ctx, c.transport.CLOBURL, "GET", "/clob-markets/"+url.PathEscape(condition), nil, nil)
		if err != nil {
			return marketInfo{}, err
		}
		value, err = parseMarket(raw)
		if err != nil {
			return marketInfo{}, response(err)
		}
		value.expires = c.now().Add(10 * time.Minute)
		cache.markets[condition] = value
	}
	if !value.assets[asset] {
		delete(cache.conditions, asset)
		delete(cache.markets, condition)
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
	if err := cache.lock(ctx); err != nil {
		return nil, err
	}
	defer cache.unlock()
	value, exists := cache.builders[code]
	if !exists || !c.now().Before(value.expires) {
		rates, err := c.GetBuilderFeeRates(ctx, code)
		if err != nil {
			return nil, err
		}
		value = builderInfo{rates: *rates, expires: c.now().Add(10 * time.Minute)}
		cache.builders[code] = value
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
