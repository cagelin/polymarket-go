// Command estimate calculates a market order's marginal price without signing or submitting it.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
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
	if len(os.Args) != 4 {
		return fmt.Errorf("usage: go run ./examples/estimate <asset-id> <BUY|SELL> <amount-or-shares>")
	}
	side := polymarket.OrderSide(strings.ToUpper(os.Args[2]))
	if side != polymarket.Buy && side != polymarket.Sell {
		return fmt.Errorf("side must be BUY or SELL")
	}
	quantity, valid := new(decimal.Big).SetString(os.Args[3])
	if !valid || !quantity.IsFinite() || quantity.Sign() <= 0 {
		return fmt.Errorf("amount or shares must be a positive finite decimal")
	}
	params := polymarket.EstimateMarketPriceParams{
		AssetQuery: polymarket.Asset(polymarket.ClobAssetID(os.Args[1])),
		Side:       side,
		OrderType:  polymarket.FOK,
	}
	if side == polymarket.Buy {
		params.Amount = quantity // BUY uses collateral units, not shares.
	} else {
		params.Shares = quantity // SELL uses shares, not collateral units.
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	price, err := client.EstimateMarketPrice(ctx, params)
	if err != nil {
		if _, insufficient := errors.AsType[*sdkerrors.InsufficientLiquidityError](err); insufficient {
			return fmt.Errorf("requested FOK quantity cannot be filled from this book: %w", err)
		}
		return fmt.Errorf("estimate market price: %w", err)
	}
	fmt.Printf("side=%s quantity=%s marginal limit price=%s (not VWAP; no order submitted)\n", side, quantity, price)
	return nil
}
