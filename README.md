# Polymarket Go SDK

[中文](README.zh-CN.md)

An independent Go SDK targeting official `polymarket-client 0.14.0`, commit
`7032ed06175b0bbea77b2d9e898bd9f252af2041`.

**Work in progress. This is not yet a complete port.** Currently implemented:
single and batch midpoint, price, spread, order book, last trade price, and
market price estimation. Discovery, history, streaming, trading, and wallets
are still pending.

Requires Go 1.27+. CGO is not required. This module is local and has not yet
been published to GitHub or the Go module proxy.

```go
client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
if err != nil {
    return err
}
defer client.Close()

book, err := client.GetOrderBook(ctx, polymarket.Asset(assetID))
```

Run the complete public-read example from this directory:

```sh
go run ./examples/orderbook <asset-id>
```

Asset identifiers use the string type `ClobAssetID`. Monetary values directly
expose `github.com/cockroachdb/apd/v3`. Estimation returns the marginal limit
price required for an amount, not VWAP, and never submits orders.

Only canonical names such as `ConditionID` and `AssetID` are public. Python's
deprecated fields and aliases are not carried into this new SDK. Exchange
wire fields such as `market` and `token_id` are mapped internally.

All network methods accept a context. Clients support concurrent use and
idempotent Close. Injected HTTP clients remain caller-owned. The default
request timeout is 10 seconds in total; inject an HTTP client to customize it.
There are no automatic retries.

```sh
CGO_ENABLED=0 go test ./...
go test -race ./...
go vet ./...
```

Run `make check` for formatting, lint, tests and race checks. Use
`golangci-lint v2.15.0`, built with Go 1.27+; set `GOLANGCI_LINT` to override
its executable path. Strict shadow checks also apply to tests.

See [Python parity tests](tests/parity/README.md) for fixture regeneration.
Set `POLYMARKET_LIVE_ASSET_ID` explicitly to run live public reads. Tests never
trade or access a wallet.

Known limitation: extreme decimal exponents outside apd's supported range are
rejected. Complete numeric compatibility has not been accepted yet.

MIT licensed. See [NOTICE](NOTICE) for upstream attribution and dependency licenses.
