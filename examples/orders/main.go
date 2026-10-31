// Command orders demonstrates trading entirely through an in-memory mock exchange.
// It uses a public synthetic signer, never reads account configuration and opens no sockets.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/cagelin/polymarket-go"
	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/ericlagergren/decimal"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	reject := flag.Bool("reject", false, "mock a market-order rejection")
	pending := flag.Bool("pending", false, "mock a settlement timeout")
	flag.Parse()
	if flag.NArg() != 0 || (*reject && *pending) {
		return fmt.Errorf("usage: go run ./examples/orders [-reject | -pending]")
	}
	exchange := &mockExchange{rejectMarket: *reject, pendingSettlement: *pending}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := newMockClient(ctx, exchange)
	if err != nil {
		return fmt.Errorf("create mock account: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	return demonstrateOrders(ctx, client)
}

func demonstrateOrders(ctx context.Context, client *polymarket.SecureClient) error {
	fmt.Println("OFFLINE MOCK: no real account, network requests or funds")

	// Create signs after reading metadata; Post is a separate submission step.
	limit, err := client.CreateLimitOrder(ctx, polymarket.CreateLimitOrderParams{
		AssetID: "123", Side: polymarket.Buy, Price: decimal.New(4, 1), Size: decimal.New(10, 0), PostOnly: true,
	})
	if err != nil {
		return fmt.Errorf("sign mock limit order: %w", err)
	}
	fmt.Printf("signed limit: type=%s maker base units=%s taker base units=%s\n", limit.OrderType, limit.MakerAmount, limit.TakerAmount)
	limitResponse, err := client.PostOrder(ctx, *limit)
	if err != nil {
		return fmt.Errorf("post mock limit order: %w", err)
	}
	acceptedLimit, err := acceptedResponse(limitResponse)
	if err != nil {
		return err
	}
	pager, err := client.ListOpenOrders(polymarket.ListOpenOrdersParams{})
	if err != nil {
		return fmt.Errorf("select mock open orders: %w", err)
	}
	page, err := pager.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read mock open orders: %w", err)
	}
	fmt.Printf("limit accepted=%s status=%s open orders=%d\n", acceptedLimit.OrderID, acceptedLimit.Status, len(page.Items))
	cancellation, err := client.CancelOrder(ctx, string(acceptedLimit.OrderID))
	if err != nil {
		return fmt.Errorf("cancel mock limit order: %w", err)
	}
	if !slices.Contains(cancellation.Canceled, acceptedLimit.OrderID) {
		return fmt.Errorf("mock order was not canceled: %v", cancellation.NotCanceled)
	}
	fmt.Printf("canceled=%v\n", cancellation.Canceled)

	// BUY Amount is collateral; SELL would use Shares and MinPrice instead.
	market, err := client.CreateMarketOrder(ctx, polymarket.CreateMarketOrderParams{
		AssetID: "123", Side: polymarket.Buy, Amount: decimal.New(5, 0), MaxPrice: decimal.New(5, 1), OrderType: polymarket.FOK,
	})
	if err != nil {
		return fmt.Errorf("sign mock market order: %w", err)
	}
	marketResponse, err := client.PostOrder(ctx, *market)
	if err != nil {
		return fmt.Errorf("post mock market order: %w", err)
	}
	acceptedMarket, err := acceptedResponse(marketResponse)
	if err != nil {
		return err // An exchange rejection is a typed response even when HTTP succeeds.
	}
	fmt.Printf("market accepted=%s status=%s; waiting for placement fills\n", acceptedMarket.OrderID, acceptedMarket.Status)
	hashes, err := client.WaitForOrderFillSettlement(ctx, *acceptedMarket, time.Second)
	if err != nil {
		if settlementTimeout, pendingSettlement := errors.AsType[*sdkerrors.SettlementTimeoutError](err); pendingSettlement {
			return fmt.Errorf("order %s still has pending trades %v; reconcile without resubmitting: %w",
				acceptedMarket.OrderID, settlementTimeout.PendingTradeIDs, err)
		}
		return fmt.Errorf("wait for mock settlement: %w", err)
	}
	fmt.Printf("mock confirmed transaction hashes=%v\n", hashes)
	return nil
}

func acceptedResponse(response polymarket.OrderResponse) (*polymarket.AcceptedOrder, error) {
	switch result := response.(type) {
	case *polymarket.AcceptedOrder:
		return result, nil
	case *polymarket.RejectedOrder:
		return nil, fmt.Errorf("mock exchange rejected the order: code=%s message=%s", result.Code, result.Message)
	default:
		return nil, fmt.Errorf("unexpected order response %T", response)
	}
}
