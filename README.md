# Polymarket Go SDK

[中文](README.zh-CN.md)

Independent Go SDK targeting official `polymarket-client 0.14.0`, source commit
`7032ed06175b0bbea77b2d9e898bd9f252af2041`. Requires Go 1.27+; CGO is unnecessary.

**Stage 3, wallet and on-chain operations, has passed offline acceptance.** The dedicated
EOA funded checks and Deposit Wallet deployment, batching, transfers, and post-only order cancellation
have passed live validation. Session-key authorization returned HTTP 403; Safe/Proxy wallets, filled-order
settlement, and RFQ execution remain unverified. The module is unpublished; full acceptance is pending.

The live service requires Deposit Wallets for new orders. Its relayer requires maximum ERC20 allowance
to the exchange and rejects both bounded allowances and zero-value revocations. An RPC transaction
from an EOA and a sponsored wallet call can therefore have different service restrictions.

Available capabilities:

- Markets, events, series, tags, sports, comments, search, and profiles.
- Prices, order books, history, market-price estimation, rewards, and public combo markets.
- Public trades, positions, activity, PnL, leaderboards, indexed approvals, and accounting downloads.
- Market, sports, comments, crypto prices/TWAP, and equity streams, with reconnects and heartbeats.
- Authentication, API keys, balances/allowances, notifications, rewards, orders, cancellation, and settlement polling.
- Four wallet signature modes, combo quotes/acceptance/status, user events, and authenticated price streams.
- Deposit Wallet deployment, EOA broadcasting, relayed batches, approvals, and transfers.
- Position split/merge/redemption, session keys, collateral return plans, and persistent RFQ quote sessions.

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

Wallet operations return a `TransactionHandle`; `Wait(ctx)` requires confirmation. Timeouts
retain transaction identifiers and do not trigger resubmission. Chain amounts use `*big.Int`
in base units; nil merge/approval amounts select the maximum. `SetupTradingApprovals`
waits for all missing grants.

Constructing a default Deposit Wallet client deploys the wallet when absent, using `BuilderKey`
or `RelayerKey`. Specify an existing wallet for read-only account workflows.
`PlaceLimitOrder`/`PlaceMarketOrder` recover a confirmed allowance shortfall after an explicit
server rejection, then submit the same order once more. `PostOrder` does not perform recovery.

`OpenRFQSession(ctx)` exposes `Next`, `Quote`, `CancelQuote`, and `RespondToConfirmation`.
Disconnection fails pending operations without replay; reopen a session explicitly.
`PlanCollateralReturn` only plans; `ExecuteCollateralReturnPlan` submits the exact router call.

An order heartbeat activates the exchange cancel-all deadline: renew every five seconds;
missing ten seconds may cancel all resting orders. Closing the client does not disable this rule.

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

| Directory                                                     | Responsibility                                             |
| ------------------------------------------------------------- | ---------------------------------------------------------- |
| Root                                                          | Client facade, canonical exports, integration tests        |
| `models/`                                                     | Response types and decoding                                |
| `pagination/`, `streams/`, `rfq/`, `transactions/`, `errors/` | Iteration, subscriptions, errors                           |
| `internal/`                                                   | CLOB, Gamma, Data, RFQ, trading, signing, shared transport |
| `examples/`                                                   | Runnable examples                                          |
| `tests/parity/`, `testdata/`, `scripts/`                      | Python comparisons, fixtures, generation                   |

Run `make check` for formatting, variable shadowing, 18 linters, tests, and race detection.
Use `golangci-lint v2.15.0` built with Go 1.27+. New declarations must not hide
enclosing names or Go predeclared identifiers, including callback parameters and
variables of different types. Rename by purpose; do not suppress diagnostics or
make callbacks mutate outer variables to avoid a warning. The rule applies to
handwritten/generated code and tests; fix generators along with their output.
Run `make shadow` separately or as part of `make lint`. Its tool lives in the
independent `tools/shadow/` module and adds no SDK dependencies.
See [parity testing](tests/parity/README.md)
for regeneration and opt-in live reads. CGO-free builds cover Linux/macOS/Windows
on amd64/arm64; runtime verification currently covers macOS arm64.

The SDK is MIT licensed. The go-ethereum signing dependency retains LGPL-3.0-or-later. See [NOTICE](NOTICE) for upstream and dependency attribution.
