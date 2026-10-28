# Polymarket Go SDK

[中文](README.zh-CN.md)

Independent Go SDK targeting official `polymarket-client 0.14.0`, source commit
`7032ed06175b0bbea77b2d9e898bd9f252af2041`. Requires Go 1.27+; CGO is unnecessary.

**Stage 2, accounts and trading, has passed offline acceptance.** A dedicated test
account is not available yet, so authenticated live and funded flows remain unverified.
The module has not been published. Next: wallets/on-chain operations, then full acceptance.

Available capabilities:

- Markets, events, series, tags, sports, comments, search, and profiles.
- Prices, order books, history, market-price estimation, rewards, and public combo markets.
- Public trades, positions, activity, PnL, leaderboards, indexed approvals, and accounting downloads.
- Market, sports, comments, crypto prices/TWAP, and equity streams, with reconnects and heartbeats.
- Authentication, API keys, balances/allowances, notifications, rewards, orders, cancellation, and settlement polling.
- Four wallet signature modes, combo quotes/acceptance/status, user events, and authenticated price streams.

See the [order book example](examples/orderbook/main.go) for complete client setup.

Runnable examples:

```sh
go run ./examples/orderbook <asset-id>
go run ./examples/markets
go run ./examples/search election
go run ./examples/streams
```

Use `NewSecureClient(ctx, SecureClientConfig{...})` for authentication. The
[account example](examples/account/main.go) reads configuration from environment
variables and queries trading restrictions. Construction validates credentials and
may create/derive a key when missing or invalid. `Close` releases resources;
`EndAuthentication` also revokes the account API key.

`CreateLimitOrder`/`CreateMarketOrder` fetch metadata and sign; `PostOrder`/`PostOrders`
submit signed orders; `PlaceLimitOrder`/`PlaceMarketOrder` combine both steps.
`AcceptedOrder` confirms acceptance; use `WaitForOrderFillSettlement` for settlement.
Ordinary submissions do not retry. Combo acceptance retries the same signed payload
once on transport or malformed-response errors, following the pinned Python SDK.

EOAs and already-deployed wallets are supported. Automatic wallet deployment,
on-chain allowance recovery, Session Key management, and persistent RFQ sessions
belong to stage 3. Sending an order heartbeat activates the exchange cancel-all
deadline: send every five seconds; exceeding ten seconds can cancel resting orders.
Closing the client does not disable that deadline.

Amounts use `*decimal.Big` from `github.com/ericlagergren/decimal`. Estimation returns
the marginal limit price, not VWAP, and never submits orders. Public Go values are
mutable; callers must avoid concurrent mutation. The SDK copies request parameters.

Pagination provides `FirstPage`, `Pages`, `Items`, and `FromCursor`; a nil cursor
means completion. Check `LimitReached` when the service caps pagination: results
may be incomplete. Cursors are query-bound and retain upstream continuation rules.

Call `Subscribe(ctx, specs...)`, then consume `Next` or `Events`. Connections are
shared by service. Each returned handle has a 1024-event queue by default, drops
the oldest event when full, and exposes `Dropped()`. Merged specs share that queue;
Python allocates queues per child subscription. Cross-source ordering is unspecified.
Close handles and the client when finished.

Network operations support context cancellation. Clients support concurrent use
and idempotent closure; injected HTTP clients remain caller-owned. Default HTTP
timeout is 10 seconds. Data API reads retry HTTP 429 at most twice under upstream
rules; accounting downloads and other public HTTP requests do not retry automatically.

Deprecated duplicate fields are omitted, while old wire names are handled internally.
Legacy public Binance/Chainlink subscriptions are retained because they provide
unauthenticated access, all-symbol mode, original quote currencies, and 30-second TWAP.

| Directory                                | Responsibility                                             |
| ---------------------------------------- | ---------------------------------------------------------- |
| Root                                     | Client facade, canonical exports, integration tests        |
| `models/`                                | Response types and decoding                                |
| `pagination/`, `streams/`, `errors/`     | Iteration, subscriptions, errors                           |
| `internal/`                              | CLOB, Gamma, Data, RFQ, trading, signing, shared transport |
| `examples/`                              | Runnable examples                                          |
| `tests/parity/`, `testdata/`, `scripts/` | Python comparisons, fixtures, generation                   |

Run `make check` for formatting, 18 linters, tests, and race detection.
Use `golangci-lint v2.15.0` built with Go 1.27+; strict shadow checks apply to
handwritten/generated code and tests. See [parity testing](tests/parity/README.md)
for regeneration and opt-in live reads. CGO-free builds cover Linux/macOS/Windows
on amd64/arm64; runtime verification currently covers macOS arm64.

The SDK is MIT licensed. The go-ethereum signing dependency retains LGPL-3.0-or-later. See [NOTICE](NOTICE) for upstream and dependency attribution.
