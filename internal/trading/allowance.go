package trading

import (
	"context"
	"errors"
	"strings"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/signing"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/transactions"
)

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
		handle, err = c.onchain.ApproveERC20(ctx, "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB", spender, nil)
	} else {
		token := models.EVMAddress("0x4D97DCd97eC945f40cF65F87097ACe5EA0476045")
		if signing.V2Position(string(order.AssetID)) {
			//nolint:gosec // Public Polygon position-manager contract.
			token = "0x006F54F7f9A22e0000CC2AB60031000000ae9fEF"
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
