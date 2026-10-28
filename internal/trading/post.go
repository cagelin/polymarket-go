package trading

import (
	"context"
	"strconv"

	"github.com/cagelam/polymarket-go/models"
)

func orderPayload(order models.SignedOrder, owner string) (map[string]any, error) {
	if order.PostOnly && order.OrderType != models.GTC && order.OrderType != models.GTD {
		return nil, input("post-only requires GTC or GTD")
	}
	if order.Salt == nil || order.MakerAmount == nil || order.TakerAmount == nil {
		return nil, input("signed order amounts and salt are required")
	}
	payload := map[string]any{"deferExec": false, "orderType": order.OrderType, "owner": owner,
		"order": map[string]any{"builder": order.Builder, "expiration": strconv.FormatInt(order.Expiration, 10), "maker": order.Maker, "makerAmount": order.MakerAmount.String(), "metadata": order.Metadata, "salt": order.Salt, "side": order.Side, "signature": order.Signature, "signatureType": order.SignatureType, "signer": order.Signer, "takerAmount": order.TakerAmount.String(), "timestamp": strconv.FormatInt(order.Timestamp, 10), "tokenId": order.AssetID}}
	if order.PostOnly {
		payload["postOnly"] = true
	}
	return payload, nil
}

// PostOrder submits a signed order once. Chain allowance recovery is handled separately.
func (c *Client) PostOrder(ctx context.Context, order models.SignedOrder) (models.OrderResponse, error) {
	payload, err := orderPayload(order, c.credentials.Key)
	if err != nil {
		return nil, err
	}
	raw, err := objectRequest[models.RawOrderResponse](ctx, c, "POST", "/order", nil, payload)
	if err != nil {
		return nil, err
	}
	return models.NormalizeOrderResponse(*raw), nil
}

// PostOrders submits between one and fifteen signed orders.
func (c *Client) PostOrders(ctx context.Context, orders []models.SignedOrder) ([]models.OrderResponse, error) {
	if len(orders) == 0 || len(orders) > 15 {
		return nil, input("orders must contain between 1 and 15 entries")
	}
	payload := make([]map[string]any, len(orders))
	for i, order := range orders {
		value, err := orderPayload(order, c.credentials.Key)
		if err != nil {
			return nil, err
		}
		payload[i] = value
	}
	raw, err := listRequest[models.RawOrderResponse](ctx, c, "POST", "/orders", nil, payload)
	if err != nil {
		return nil, err
	}
	result := make([]models.OrderResponse, len(raw))
	for i, item := range raw {
		result[i] = models.NormalizeOrderResponse(item)
	}
	return result, nil
}
