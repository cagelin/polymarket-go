package trading

import (
	"context"

	"github.com/cagelin/polymarket-go/models"
)

// CancelOrder cancels one account order.
func (c *Client) CancelOrder(ctx context.Context, id string) (*models.CancelOrdersResponse, error) {
	if err := nonempty(id); err != nil {
		return nil, err
	}
	return objectRequest[models.CancelOrdersResponse](ctx, c, "DELETE", "/order", nil, map[string]string{"orderID": id})
}

// CancelOrders cancels between one and 3000 account orders.
func (c *Client) CancelOrders(ctx context.Context, ids []string) (*models.CancelOrdersResponse, error) {
	if len(ids) == 0 || len(ids) > 3000 {
		return nil, input("order IDs must contain between 1 and 3000 entries")
	}
	if err := nonempty(ids...); err != nil {
		return nil, err
	}
	return objectRequest[models.CancelOrdersResponse](ctx, c, "DELETE", "/orders", nil, ids)
}

// CancelAll cancels every resting order for the account.
func (c *Client) CancelAll(ctx context.Context) (*models.CancelOrdersResponse, error) {
	return objectRequest[models.CancelOrdersResponse](ctx, c, "DELETE", "/cancel-all", nil, nil)
}

// CancelMarketOrdersParams selects orders by condition, asset, or both.
type CancelMarketOrdersParams struct {
	ConditionID *string
	AssetID     *string
}

// CancelMarketOrders cancels orders matching at least one supplied filter.
func (c *Client) CancelMarketOrders(ctx context.Context, p CancelMarketOrdersParams) (*models.CancelOrdersResponse, error) {
	if p.ConditionID == nil && p.AssetID == nil {
		return nil, input("condition or asset is required")
	}
	body := map[string]string{}
	for key, value := range map[string]*string{"market": p.ConditionID, "asset_id": p.AssetID} {
		if value != nil {
			if err := nonempty(*value); err != nil {
				return nil, err
			}
			body[key] = *value
		}
	}
	return objectRequest[models.CancelOrdersResponse](ctx, c, "DELETE", "/cancel-market-orders", nil, body)
}
