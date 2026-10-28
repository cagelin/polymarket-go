package trading

import (
	"context"
	"fmt"
	"time"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/models"
)

// WaitForOrderFillSettlement waits for placement fills; it does not wait for later resting-order fills.
// A zero timeout selects thirty seconds. Cancellation never cancels the exchange order.
func (c *Client) WaitForOrderFillSettlement(ctx context.Context, order models.AcceptedOrder, timeout time.Duration) ([]models.TransactionHash, error) {
	if len(order.TradeIDs) == 0 {
		return append([]models.TransactionHash{}, order.TransactionsHashes...), nil
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if timeout < 0 {
		return nil, input("settlement timeout must be positive")
	}
	ids := make([]string, 0, len(order.TradeIDs))
	seen := map[string]bool{}
	for _, id := range order.TradeIDs {
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	settled := map[string]models.ClobTrade{}
	deadline := time.Now().Add(timeout)
	for {
		for _, id := range ids {
			if _, ok := settled[id]; ok {
				continue
			}
			pages, err := c.ListAccountTrades(ListAccountTradesParams{ID: &id})
			if err != nil {
				return nil, err
			}
			page, err := pages.FirstPage(ctx)
			if err != nil {
				return nil, err
			}
			for _, trade := range page.Items {
				if seen[trade.ID] && (trade.Status == "CONFIRMED" || trade.Status == "FAILED") {
					settled[trade.ID] = trade
				}
			}
		}
		if len(settled) == len(ids) {
			break
		}
		pending := make([]string, 0)
		for _, id := range ids {
			if _, ok := settled[id]; !ok {
				pending = append(pending, id)
			}
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, &sdkerrors.SettlementTimeoutError{PendingTradeIDs: pending}
		}
		delay := min(250*time.Millisecond, remaining)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	hashes := make([]models.TransactionHash, 0)
	unique := map[string]bool{}
	confirmed := false
	for _, id := range ids {
		trade := settled[id]
		if trade.Status == "FAILED" {
			continue
		}
		confirmed = true
		if trade.TransactionHash == nil {
			return nil, response(fmt.Errorf("confirmed trade %s has no transaction hash", id))
		}
		hash := *trade.TransactionHash
		if !unique[hash] {
			hashes = append(hashes, models.TransactionHash(hash))
			unique[hash] = true
		}
	}
	if !confirmed {
		return nil, &sdkerrors.TransactionFailedError{}
	}
	return hashes, nil
}
