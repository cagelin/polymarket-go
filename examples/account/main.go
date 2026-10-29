// Command account authenticates and reads the account's trading restriction.
// Construction may create or derive API credentials if the supplied key is unavailable.
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
	names := []string{"POLYMARKET_PRIVATE_KEY", "POLYMARKET_WALLET", "POLYMARKET_API_KEY", "POLYMARKET_API_SECRET", "POLYMARKET_API_PASSPHRASE"}
	for _, name := range names {
		if os.Getenv(name) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		return err
	}
	defer func() { _ = client.Close() }() // Close only releases local resources and always returns nil.
	restricted, err := client.GetClosedOnlyMode(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("closed-only trading: %t\n", restricted)
	return nil
}
