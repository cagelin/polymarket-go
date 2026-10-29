// Command wallet reads indexed token approvals without private keys or transactions.
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
		return fmt.Errorf("usage: go run ./examples/wallet <wallet-address>")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }() // Close releases local resources.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	state, err := client.GetTradingApprovalsState(ctx, polymarket.EVMAddress(os.Args[1]))
	if err != nil {
		return err
	}
	fmt.Printf("fully approved=%t; missing ERC20=%d ERC1155=%d\n", state.IsFullyApproved, len(state.Missing.ERC20), len(state.Missing.ERC1155))
	return nil
}
