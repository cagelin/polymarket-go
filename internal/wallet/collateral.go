package wallet

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/transport"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/transactions"
)

// PlanCollateralReturn asks the service for an inspectable plan without submitting it.
func (c *Client) PlanCollateralReturn(ctx context.Context) (*models.CollateralReturnPlanResponse, error) {
	if c.config.Kind == models.WalletEOA {
		return nil, input("collateral return requires a contract wallet")
	}
	var headers transport.HeaderResolver
	if c.config.BuilderKey != nil || c.config.RelayerKey != nil {
		headers = c.headers
	}
	raw, err := c.transport.LongRequest(ctx, c.transport.CollateralReturnURL, "POST", "/v1/collateral-return/plan", map[string]any{"wallet": c.config.Wallet}, headers, 120*time.Second)
	if err != nil {
		return nil, err
	}
	result, err := codec.DecodeResponse[models.CollateralReturnPlanResponse](raw)
	return &result, err
}

// ExecuteCollateralReturnPlan submits the exact planned router call without implicit approvals.
func (c *Client) ExecuteCollateralReturnPlan(ctx context.Context, plan models.CollateralReturnPlanResponse) (*transactions.Handle, error) {
	if c.config.Kind == models.WalletEOA {
		return nil, input("collateral return requires a contract wallet")
	}
	if !strings.EqualFold(string(plan.Wallet), string(c.config.Wallet)) || plan.ChainID != 137 {
		return nil, input("plan wallet or chain does not match client")
	}
	if _, err := c.headers("POST", "", nil); err != nil {
		return nil, err
	}
	calls := []models.TransactionCall{{To: plan.RouterCall.To, Data: string(plan.RouterCall.Data), Value: new(big.Int)}}
	if err := c.lock(ctx); err != nil {
		return nil, err
	}
	defer c.unlock()
	for attempt := 0; ; attempt++ {
		envelope, err := c.relayerPayload(ctx, calls, "Collateral return", "")
		if err != nil {
			return nil, err
		}
		raw, err := c.transport.LongRequest(ctx, c.transport.CollateralReturnURL, "POST", "/v1/collateral-return/submit", map[string]any{"plan_hash": plan.PlanHash, "envelope": envelope}, c.headers, 120*time.Second)
		if nonce, _ := rejectedNonce(err); nonce != "" && c.config.Kind == models.WalletDeposit {
			envelope, err = c.relayerPayload(ctx, calls, "Collateral return", nonce)
			if err != nil {
				return nil, err
			}
			raw, err = c.transport.LongRequest(ctx, c.transport.CollateralReturnURL, "POST", "/v1/collateral-return/submit", map[string]any{"plan_hash": plan.PlanHash, "envelope": envelope}, c.headers, 120*time.Second)
		}
		if err == nil {
			return c.submitted(raw)
		}
		if attempt == 10 || !retryable(err) {
			return nil, err
		}
		if err = wait(ctx, c.config.PollInterval); err != nil {
			return nil, err
		}
	}
}
