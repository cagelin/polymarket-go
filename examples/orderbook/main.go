// Command orderbook prints the current book for the supplied public asset ID.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	polymarket "github.com/cagelam/polymarket-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: go run ./examples/orderbook <asset-id>")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return err
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	book, err := client.GetOrderBook(ctx, polymarket.Asset(polymarket.ClobAssetID(os.Args[1])))
	if err != nil {
		return err
	}
	fmt.Printf("asset=%s bids=%d asks=%d tick=%s\n", book.AssetID, len(book.Bids), len(book.Asks), book.TickSize)
	if len(book.Bids) > 0 {
		fmt.Printf("best bid=%s\n", book.Bids[len(book.Bids)-1].Price)
	}
	if len(book.Asks) > 0 {
		fmt.Printf("best ask=%s\n", book.Asks[len(book.Asks)-1].Price)
	}
	return nil
}
