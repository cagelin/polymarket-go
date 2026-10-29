package trading

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/internal/paging"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
	"github.com/cagelin/polymarket-go/pagination"
)

func (c *Client) signatureQuery() url.Values {
	return url.Values{"signature_type": {strconv.Itoa(signing.SignatureType(c.walletType))}}
}
func nonempty(values ...string) error {
	if slices.Contains(values, "") {
		return input("value must be nonempty")
	}
	return nil
}
func requiredBool(raw []byte, key string) (bool, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return false, response(err)
	}
	value, err := codec.Required[bool](fields, key)
	return value, response(err)
}
func objectRequest[T any](ctx context.Context, c *Client, method, path string, query url.Values, body any) (*T, error) {
	raw, err := c.request(ctx, method, path, query, body)
	if err != nil {
		return nil, err
	}
	value, err := codec.DecodeResponse[T](raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
func listRequest[T any](ctx context.Context, c *Client, method, path string, query url.Values, body any) ([]T, error) {
	raw, err := c.request(ctx, method, path, query, body)
	if err != nil {
		return nil, err
	}
	return codec.ParseList[T](raw)
}
func expectOK(raw []byte, err error) error {
	if err != nil {
		return err
	}
	var value string
	if decodeErr := json.Unmarshal(raw, &value); decodeErr != nil || value != "OK" {
		return response(fmt.Errorf("expected OK response"))
	}
	return nil
}

// DeleteAPIKey revokes the credentials used by this client.
func (c *Client) DeleteAPIKey(ctx context.Context) error {
	return expectOK(c.request(ctx, "DELETE", "/auth/api-key", nil, nil))
}

// CreateBuilderAPIKey creates a new builder credential.
func (c *Client) CreateBuilderAPIKey(ctx context.Context) (*models.BuilderAPIKey, error) {
	return objectRequest[models.BuilderAPIKey](ctx, c, "POST", "/auth/builder-api-key", nil, nil)
}

// FetchBuilderAPIKeys returns builder credential identities and revocation times.
func (c *Client) FetchBuilderAPIKeys(ctx context.Context) ([]models.BuilderAPIKeyInfo, error) {
	raw, err := c.request(ctx, "GET", "/auth/builder-api-key", nil, nil)
	if err != nil {
		return nil, err
	}
	rows, err := codec.ParseList[json.RawMessage](raw)
	if err != nil {
		return nil, err
	}
	result := make([]models.BuilderAPIKeyInfo, len(rows))
	for i, row := range rows {
		if len(row) > 0 && row[0] == '"' {
			row = append(append([]byte(`{"key":`), row...), '}')
		}
		if err = json.Unmarshal(row, &result[i]); err != nil {
			return nil, response(err)
		}
	}
	return result, nil
}

// RevokeBuilderAPIKey revokes the builder credential supplied during construction.
func (c *Client) RevokeBuilderAPIKey(ctx context.Context) error {
	if c.builderKey == nil {
		return input("a builder API key is required")
	}
	resolver := func(method, path string, body []byte) (http.Header, error) {
		timestamp := c.now().Unix()
		signature, err := signing.HMAC(c.builderKey.Secret, timestamp, method, path, body)
		if err != nil {
			return nil, err
		}
		return http.Header{"POLY_BUILDER_API_KEY": {c.builderKey.Key}, "POLY_BUILDER_PASSPHRASE": {c.builderKey.Passphrase}, "POLY_BUILDER_TIMESTAMP": {strconv.FormatInt(timestamp, 10)}, "POLY_BUILDER_SIGNATURE": {signature}}, nil
	}
	return expectOK(c.transport.AuthenticatedRequest(ctx, c.transport.CLOBURL, "DELETE", "/auth/builder-api-key", nil, nil, resolver))
}

// GetClosedOnlyMode checks the account's trading restriction.
func (c *Client) GetClosedOnlyMode(ctx context.Context) (bool, error) {
	raw, err := c.request(ctx, "GET", "/auth/ban-status/closed-only", nil, nil)
	if err != nil {
		return false, err
	}
	return requiredBool(raw, "closed_only")
}

// ListOpenOrdersParams filters orders owned by this account. Nil filters are omitted.
type ListOpenOrdersParams struct {
	AssetID     *string `json:"asset_id"`
	ID          *string `json:"id"`
	ConditionID *string `json:"market"`
}

// ListAccountTradesParams filters fills belonging to the authenticated account.
type ListAccountTradesParams struct {
	ListOpenOrdersParams
	MakerAddress *string `json:"maker_address"`
	After        *string `json:"after"`
	Before       *string `json:"before"`
}

func orderFilters(p ListOpenOrdersParams) (url.Values, error) {
	query := url.Values{}
	for key, value := range map[string]*string{"asset_id": p.AssetID, "id": p.ID, "market": p.ConditionID} {
		if value != nil {
			if err := nonempty(*value); err != nil {
				return nil, err
			}
			query.Set(key, *value)
		}
	}
	return query, nil
}
func accountPages[T any](c *Client, path string, query url.Values, total bool) *pagination.Paginator[T] {
	return pagination.New(func(ctx context.Context, cursor *string) (pagination.Page[T], error) {
		params := url.Values{}
		for key, values := range query {
			params[key] = append([]string(nil), values...)
		}
		if cursor != nil {
			if err := nonempty(*cursor); err != nil {
				return pagination.Page[T]{}, err
			}
			params.Set("next_cursor", *cursor)
		}
		raw, err := c.request(ctx, "GET", path, params, nil)
		if err != nil {
			return pagination.Page[T]{}, err
		}
		fields, err := codec.Object(raw)
		if err != nil {
			return pagination.Page[T]{}, response(err)
		}
		items, err := codec.ParseList[T](fields["data"])
		if err != nil {
			return pagination.Page[T]{}, err
		}
		next, err := paging.ServerNextCursor(fields)
		if err != nil {
			return pagination.Page[T]{}, response(err)
		}
		if next != nil && *next == "LTE=" {
			next = nil
		}
		page := pagination.Page[T]{Items: items, HasMore: next != nil, NextCursor: next}
		if total && len(fields["total_count"]) > 0 && string(fields["total_count"]) != "null" {
			count, parseErr := codec.IntegerField(fields, "total_count")
			if parseErr != nil {
				return pagination.Page[T]{}, response(parseErr)
			}
			page.TotalCount = &count
		}
		return page, nil
	}, nil)
}

// ListOpenOrders lazily retrieves authenticated open orders.
func (c *Client) ListOpenOrders(p ListOpenOrdersParams) (*pagination.Paginator[models.OpenOrder], error) {
	q, err := orderFilters(p)
	if err != nil {
		return nil, err
	}
	return accountPages[models.OpenOrder](c, "/data/orders", q, false), nil
}

// GetOrder retrieves a single authenticated order.
func (c *Client) GetOrder(ctx context.Context, orderID string) (*models.OpenOrder, error) {
	if err := nonempty(orderID); err != nil {
		return nil, err
	}
	return objectRequest[models.OpenOrder](ctx, c, "GET", "/data/order/"+url.PathEscape(orderID), nil, nil)
}

// ListAccountTrades lazily retrieves authenticated account fills.
func (c *Client) ListAccountTrades(p ListAccountTradesParams) (*pagination.Paginator[models.ClobTrade], error) {
	q, err := orderFilters(p.ListOpenOrdersParams)
	if err != nil {
		return nil, err
	}
	for key, value := range map[string]*string{"maker_address": p.MakerAddress, "after": p.After, "before": p.Before} {
		if value != nil {
			if err = nonempty(*value); err != nil {
				return nil, err
			}
			q.Set(key, *value)
		}
	}
	return accountPages[models.ClobTrade](c, "/data/trades", q, false), nil
}

// GetNotifications retrieves typed account notifications.
func (c *Client) GetNotifications(ctx context.Context) ([]models.Notification, error) {
	raw, err := c.request(ctx, "GET", "/notifications", c.signatureQuery(), nil)
	if err != nil {
		return nil, err
	}
	items, err := models.ParseNotifications(raw)
	return items, response(err)
}

// DropNotifications deletes the selected notification IDs. Integer IDs can be formatted as decimal strings.
func (c *Client) DropNotifications(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return input("notification IDs must be nonempty")
	}
	if err := nonempty(ids...); err != nil {
		return err
	}
	q := c.signatureQuery()
	q.Set("ids", strings.Join(ids, ","))
	_, err := c.request(ctx, "DELETE", "/notifications", q, nil)
	return err
}

// AssetType identifies collateral or a conditional asset protocol.
type AssetType string

// Supported asset balance categories.
const (
	AssetCollateral    AssetType = "COLLATERAL"
	AssetConditional   AssetType = "CONDITIONAL"
	AssetConditionalV2 AssetType = "CONDITIONAL-V2"
)

// BalanceAllowanceParams selects an asset balance and allowance.
type BalanceAllowanceParams struct {
	AssetType AssetType
	AssetID   *string
}

// GetBalanceAllowance returns integer base units for the selected asset.
func (c *Client) GetBalanceAllowance(ctx context.Context, p BalanceAllowanceParams) (*models.BalanceAllowance, error) {
	if p.AssetType != AssetCollateral && p.AssetType != AssetConditional && p.AssetType != AssetConditionalV2 {
		return nil, input("invalid asset type")
	}
	q := c.signatureQuery()
	q.Set("asset_type", string(p.AssetType))
	if p.AssetID != nil {
		if err := nonempty(*p.AssetID); err != nil {
			return nil, err
		}
		q.Set("token_id", *p.AssetID)
	}
	return objectRequest[models.BalanceAllowance](ctx, c, "GET", "/balance-allowance", q, nil)
}

// SendOrderHeartbeat activates the exchange cancel-all deadline and returns the next ID.
// Send every five seconds after activation; missing the ten-second deadline cancels open orders.
// Close does not disable the exchange deadline. Mismatched IDs are returned without retry.
func (c *Client) SendOrderHeartbeat(ctx context.Context, heartbeatID string) (*models.OrderHeartbeat, error) {
	return objectRequest[models.OrderHeartbeat](ctx, c, "POST", "/v1/heartbeats", nil, map[string]string{"heartbeat_id": heartbeatID})
}
