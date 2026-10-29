package trading

import (
	"context"
	"errors"
	"strings"

	"github.com/cagelin/polymarket-go/internal/chain"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
	"github.com/cagelin/polymarket-go/transactions"
)

// The pinned upstream (7032ed06175b0bbea77b2d9e898bd9f252af2041) exposes this rejection as text.
// Do not infer allowance recovery from unknown codes or generic balance errors.
func (c *Client) postWithAllowanceRecovery(ctx context.Context, order models.SignedOrder) (models.OrderResponse, error) {
	result, err := c.PostOrder(ctx, order)
	var rejected *sdkerrors.RequestRejectedError
	if !errors.As(err, &rejected) || rejected.Status != 400 || !strings.Contains(rejected.Message, "allowance is not enough") {
		return result, err
	}
	original := err
	market, err := c.market(ctx, string(order.AssetID), false)
	if err != nil {
		return result, err
	}
	spender := models.EVMAddress(signing.Exchange(string(order.AssetID), market.negativeRisk))
	params := BalanceAllowanceParams{AssetType: "COLLATERAL"}
	if order.Side == models.Sell {
		params.AssetType = "CONDITIONAL"
		asset := string(order.AssetID)
		params.AssetID = &asset
		if signing.V2Position(string(order.AssetID)) {
			params.AssetType = "CONDITIONAL-V2"
		}
	}
	balance, err := c.GetBalanceAllowance(ctx, params)
	if err != nil {
		return result, err
	}
	for address, amount := range balance.Allowances {
		if strings.EqualFold(address, string(spender)) && amount.Cmp(order.MakerAmount) >= 0 {
			return result, original
		}
	}
	var handle *transactions.Handle
	if order.Side == models.Buy {
		handle, err = c.onchain.ApproveERC20(ctx, chain.Collateral, spender, nil)
	} else {
		token := models.EVMAddress(chain.ConditionalTokens)
		if signing.V2Position(string(order.AssetID)) {
			token = chain.PositionManager
		}
		handle, err = c.onchain.ApproveERC1155ForAll(ctx, token, spender, true)
	}
	if err != nil {
		return result, err
	}
	if _, err = handle.Wait(ctx); err != nil {
		return result, err
	}
	query := c.signatureQuery()
	query.Set("asset_type", string(params.AssetType))
	if params.AssetID != nil {
		query.Set("token_id", *params.AssetID)
	}
	if _, err = c.request(ctx, "GET", "/balance-allowance/update", query, nil); err != nil {
		return result, err
	}
	if _, err = c.GetBalanceAllowance(ctx, params); err != nil {
		return result, err
	}
	return c.PostOrder(ctx, order)
}
