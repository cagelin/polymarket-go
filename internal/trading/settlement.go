package trading

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/transport"
	"github.com/cagelin/polymarket-go/models"
)

// WaitForOrderFillSettlement waits for placement fills; it does not wait for later resting-order fills.
// Without placement trade IDs it returns the supplied hashes immediately; an empty result does not prove settlement.
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
	// Signing timestamps may use c.now; timeout budgets must use an advancing real clock.
	bounded, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	pendingError := func() error {
		pending := make([]string, 0, len(ids))
		for _, id := range ids {
			if _, exists := settled[id]; !exists {
				pending = append(pending, id)
			}
		}
		return &sdkerrors.SettlementTimeoutError{PendingTradeIDs: pending}
	}
	interval := 250 * time.Millisecond
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if bounded.Err() != nil {
			return nil, pendingError()
		}
		retryDelay := time.Duration(0)
		for _, id := range ids {
			if _, ok := settled[id]; ok {
				continue
			}
			pages, err := c.ListAccountTrades(ListAccountTradesParams{ID: &id})
			if err != nil {
				return nil, err
			}
			page, err := pages.FirstPage(bounded)
			if err != nil {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				if bounded.Err() != nil {
					return nil, pendingError()
				}
				var rateLimit *sdkerrors.RateLimitError
				if !errors.As(err, &rateLimit) {
					return nil, err
				}
				nextDelay, validDelay := transport.RetryDelay(interval, rateLimit.RetryAfter)
				if !validDelay {
					return nil, err
				}
				retryDelay = nextDelay
				break
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
		//nolint:gosec // Polling jitter is not a security decision.
		delay := max(interval+time.Duration(rand.Int64N(int64(interval/5))), retryDelay)
		interval = min(interval*2, 2*time.Second)
		timer := time.NewTimer(delay)
		select {
		case <-bounded.Done():
			timer.Stop()
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, pendingError()
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
