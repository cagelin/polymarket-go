// Command account reads account state or subscribes to authenticated user/price events.
// Construction may create or derive API credentials if the supplied key is unavailable.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/cagelin/polymarket-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	stream := flag.String("stream", "", "subscribe to user or prices events instead of reading account state")
	duration := flag.Duration("duration", 30*time.Second, "account request or subscription lifetime")
	flag.Parse()
	if flag.NArg() != 0 || *duration <= 0 || (*stream != "" && *stream != "user" && *stream != "prices") {
		return fmt.Errorf("usage: go run ./examples/account [-stream user|prices] [-duration 30s]")
	}
	names := []string{"POLYMARKET_PRIVATE_KEY", "POLYMARKET_WALLET", "POLYMARKET_API_KEY", "POLYMARKET_API_SECRET", "POLYMARKET_API_PASSPHRASE"}
	for _, name := range names {
		if os.Getenv(name) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(signalContext, *duration)
	defer cancel()
	client, err := polymarket.NewSecureClient(ctx, polymarket.SecureClientConfig{
		PrivateKey: os.Getenv("POLYMARKET_PRIVATE_KEY"),
		Wallet:     os.Getenv("POLYMARKET_WALLET"),
		Credentials: &polymarket.APIKeyCreds{
			Key:        os.Getenv("POLYMARKET_API_KEY"),
			Secret:     os.Getenv("POLYMARKET_API_SECRET"),
			Passphrase: os.Getenv("POLYMARKET_API_PASSPHRASE"),
		},
	})
	if err != nil {
		return fmt.Errorf("authenticate existing wallet: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	if *stream != "" {
		return subscribeAccount(ctx, client, *stream)
	}
	restricted, err := client.GetClosedOnlyMode(ctx)
	if err != nil {
		return fmt.Errorf("read trading restriction: %w", err)
	}
	fmt.Printf("closed-only trading: %t\n", restricted)
	balance, err := client.GetBalanceAllowance(ctx, polymarket.BalanceAllowanceParams{AssetType: polymarket.AssetCollateral})
	if err != nil {
		return fmt.Errorf("read collateral balance and allowances: %w", err)
	}
	fmt.Printf("collateral balance (base units)=%s allowance entries=%d\n", balance.Balance, len(balance.Allowances))
	orders, err := client.ListOpenOrders(polymarket.ListOpenOrdersParams{})
	if err != nil {
		return fmt.Errorf("select open orders: %w", err)
	}
	orderPage, err := orders.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read open orders: %w", err)
	}
	fmt.Printf("open orders on first page=%d has_more=%t\n", len(orderPage.Items), orderPage.HasMore)
	for _, order := range orderPage.Items {
		fmt.Printf("order=%s asset=%s side=%s price=%s status=%s\n", order.ID, order.AssetID, order.Side, order.Price, order.Status)
	}
	trades, err := client.ListAccountTrades(polymarket.ListAccountTradesParams{})
	if err != nil {
		return fmt.Errorf("select account trades: %w", err)
	}
	tradePage, err := trades.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read account trades: %w", err)
	}
	fmt.Printf("account trades on first page=%d has_more=%t\n", len(tradePage.Items), tradePage.HasMore)
	notifications, err := client.GetNotifications(ctx)
	if err != nil {
		return fmt.Errorf("read notifications: %w", err)
	}
	fmt.Printf("notifications=%d\n", len(notifications))
	return nil
}

func subscribeAccount(ctx context.Context, client *polymarket.SecureClient, feed string) error {
	var spec polymarket.Spec = polymarket.UserSpec{}
	if feed == "prices" {
		spec = polymarket.CryptoPriceSpec{Symbols: []string{"btcusd"}, Provider: "chainlink"}
	}
	handle, err := client.Subscribe(ctx, spec)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("subscribe to account %s events: %w", feed, err)
	}
	defer func() {
		_ = handle.Close() // Close unsubscribes and always returns nil.
		log.Printf("typed queue dropped=%d stats=%+v", handle.Dropped(), client.StreamStats())
	}()
	encoder := json.NewEncoder(os.Stdout)
	for event, readErr := range handle.Events(ctx) {
		if ctx.Err() != nil {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read account stream: %w", readErr)
		}
		if encodeErr := encoder.Encode(event); encodeErr != nil {
			return fmt.Errorf("write account event: %w", encodeErr)
		}
	}
	return nil
}
