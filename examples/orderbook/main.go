// Command orderbook prints the current book for the supplied public asset ID.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cagelin/polymarket-go"
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
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	book, err := client.GetOrderBook(ctx, polymarket.Asset(polymarket.ClobAssetID(os.Args[1])))
	if err != nil {
		return fmt.Errorf("read order book: %w", err)
	}
	fmt.Printf("asset=%s bids=%d asks=%d tick=%s\n", book.AssetID, len(book.Bids), len(book.Asks), book.TickSize)
	if len(book.Bids) > 0 {
		// The SDK preserves ascending bids and descending asks: best levels are last.
		bestBid := book.Bids[len(book.Bids)-1]
		fmt.Printf("best bid=%s size=%s\n", bestBid.Price, bestBid.Size)
	} else {
		fmt.Println("no resting bids")
	}
	if len(book.Asks) > 0 {
		bestAsk := book.Asks[len(book.Asks)-1]
		fmt.Printf("best ask=%s size=%s\n", bestAsk.Price, bestAsk.Size)
	} else {
		fmt.Println("no resting asks")
	}
	return nil
}
