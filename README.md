# Polymarket Go SDK

[中文](README.zh-CN.md) · [API reference](https://pkg.go.dev/github.com/cagelin/polymarket-go) · [Releases](https://github.com/cagelin/polymarket-go/releases)

An independent Go SDK for Polymarket market data, real-time streams, authenticated
trading, and wallet operations. The implementation is compared against
`polymarket-client 0.14.0`, pinned to source commit
`7032ed06175b0bbea77b2d9e898bd9f252af2041`.

## Features

- Markets, events, series, tags, sports, comments, search, and profiles.
- Prices, order books, history, market-price estimation, rewards, and public combo markets.
- Public trades, positions, activity, PnL, leaderboards, indexed approvals, and accounting downloads.
- Market, sports, comments, crypto prices/TWAP, and equity streams, with reconnects and heartbeats.
- Authentication, API keys, balances/allowances, notifications, rewards, orders, cancellation, and settlement polling.
- Deposit Wallet order signing, combo quotes/acceptance/status, user events, and authenticated price streams.
- Deposit Wallet deployment, EOA broadcasting, relayed batches, approvals, and transfers.
- Position split/merge/redemption, session keys, collateral return plans, and persistent RFQ quote sessions.

## Installation

Requires **Go 1.27 or later**. The SDK builds without CGO.

```sh
go get github.com/cagelin/polymarket-go@v0.2.0
```

## Quick start

The [order book example](examples/orderbook/main.go) demonstrates public client setup,
context cancellation, and resource cleanup. Public queries do not require account credentials.

From a checkout of this repository:

```sh
go run ./examples/orderbook <asset-id>
go run ./examples/markets
go run ./examples/events
go run ./examples/search election
go run ./examples/streams
go run ./examples/wallet <wallet-address>
```

See the [examples guide](examples/README.md) for discovery, quotes/history, public
trades and portfolio data, cursor resumption, typed/raw streams, account queries,
and an offline order lifecycle demonstration.

## Usage notes

### Authentication

Use `NewSecureClient(ctx, SecureClientConfig{...})` for authentication. The
[account example](examples/account/main.go) reads configuration from environment
variables and queries trading restrictions, balances, open orders, fills and notifications;
it also supports authenticated user/price subscriptions. Construction validates credentials and
may create/derive a key when missing or invalid. `Close` releases resources;
`EndAuthentication` also revokes the account API key.

### Orders and settlement

`CreateLimitOrder`/`CreateMarketOrder` fetch metadata and sign; `PostOrder`/`PostOrders`
submit signed orders; `PlaceLimitOrder`/`PlaceMarketOrder` combine both steps.
`AcceptedOrder` confirms acceptance; use `WaitForOrderFillSettlement` for settlement.
Ordinary submissions do not retry. Combo acceptance retries the same signed payload
once on transport or malformed-response errors, following the pinned Python SDK.

An order heartbeat activates the exchange cancel-all deadline: renew every five seconds;
missing ten seconds may cancel all resting orders. Closing the client does not disable this rule.

### Wallet operations

Wallet operations return a `TransactionHandle`; `Wait(ctx)` requires confirmation. Timeouts
retain transaction identifiers and do not trigger resubmission. Chain amounts use `*big.Int`
in base units; nil merge/approval amounts select the maximum. `SetupTradingApprovals`
waits for all missing grants.

Constructing a default Deposit Wallet client deploys the wallet when absent, using `BuilderKey`
or `RelayerKey`. Specify an existing wallet for read-only account workflows.
`PlaceLimitOrder`/`PlaceMarketOrder` recover a confirmed allowance shortfall after an explicit
server rejection, then submit the same order once more. `PostOrder` does not perform recovery.

### RFQ and collateral return

`OpenRFQSession(ctx)` exposes `Next`, `Quote`, `CancelQuote`, and `RespondToConfirmation`.
Disconnection fails pending operations without replay; reopen a session explicitly.
RFQ sessions probe the peer with WebSocket Ping/Pong (10-second interval, 30-second pong budget).
`StreamStats().UnsupportedMessages` distinguishes new event variants from malformed payloads.
`PlanCollateralReturn` only plans; `ExecuteCollateralReturnPlan` submits the exact router call.

HTTP response bodies are limited to 32 MiB, including accounting snapshots. Set
`ClientConfig.MaxResponseBodyBytes` to a positive byte limit for larger downloads. Held and
long requests share a connection pool; an injected HTTP client retains its own transport
header deadlines. Service endpoints must not contain credentials, query strings, or fragments.

### Precision and pagination

Amounts use `*decimal.Big` from `github.com/ericlagergren/decimal`. Estimation returns
the marginal limit price, not VWAP, and never submits orders. Public Go values are
mutable; callers must avoid concurrent mutation. The SDK copies request parameters.

Pagination provides `FirstPage`, `Pages`, `Items`, and `FromCursor`; a nil cursor
means completion. Check `LimitReached` when the service caps pagination: results
may be incomplete. Cursors are query-bound and retain upstream continuation rules.

### Streaming

Call `Subscribe(ctx, specs...)`, then consume `Next` or `Events`. Connections are
shared by service. Each returned handle has a 1024-event queue by default, drops
the oldest event when full, and exposes `Dropped()`. Merged specs share that queue;
Python allocates queues per child subscription. Cross-source ordering is unspecified.
For market/user/sports/RTDS subscriptions, an event matching several selections in one
handle is delivered once per matching selection. A single selection matching several
assets receives one copy. Close handles and the client when finished.

### Raw market messages and connection observations

For byte-preserving collectors, construct `streams.New(streams.Config{...})` from
`github.com/cagelin/polymarket-go/streams`. Set `RawHandler` to receive every complete
market message before decoding, including unknown, malformed and binary messages;
text `PONG` is excluded. The SDK does not reuse the returned buffer. Retained `Data`
must remain immutable because the default typed decoder may still read it.
`RawOnly: true` skips market typed decoding and requires `RawHandler`; subscription
handles still own subscription lifetimes. Other services keep their typed behavior.

`LifecycleHandler` observes market/user/sports/RTDS connections and application
heartbeats, including `read_limit` when an oversized message closes a connection.
It does not observe price sessions or deliver partial oversized data. `ReadLimit`
applies to all stream sockets, including price sessions; zero keeps the 10 MiB default.
Callbacks must return promptly, avoid blocking client methods, and provide their own
bounded handoff. Lifecycle callbacks may run concurrently; `Close` waits for callbacks
and must be called outside them. Observation IDs are local to each client, and shutdown
also emits `disconnected`; observations do not prove complete data coverage.

Repeated short connections retain exponential retry backoff (250 ms cap initially,
30 s maximum, each wait between half its cap and the cap). A connection must last
30 s before resetting that backoff. Default typed APIs remain available.

### Transport and concurrency

Network operations support context cancellation. Clients support concurrent use
and idempotent closure; injected HTTP clients remain caller-owned. Default HTTP
timeout is 10 seconds. Data API reads retry HTTP 429 at most twice under upstream
rules; accounting downloads and other public HTTP requests do not retry automatically.

## Compatibility and validation

Trading support targets new accounts using **Deposit Wallets**. EOA signing and
broadcasting are supported; legacy Safe/Proxy onboarding and migration are outside
the supported scope. Retained legacy implementations and tests do not imply support.

Deprecated duplicate fields are omitted; older wire names are handled internally.
Public Binance/Chainlink subscriptions are retained for unauthenticated access,
all-symbol mode, original quote currencies, and 30-second TWAP.

Release validation covers CGO-free builds for Linux, macOS, and Windows on amd64
and arm64. Runtime verification covers macOS arm64. Offline comparisons and live
validation are recorded separately; implemented features are not all live-validated.
See the [acceptance guide](tests/acceptance/README.md) for verified scenarios and
known gaps, including collateral-return authorization, RFQ settlement, and session keys.

During the v0.1.0 live checks, the relayer required maximum ERC20 exchange allowances
and rejected bounded allowances and zero-value revocations. Direct EOA transactions
and sponsored wallet calls can have different service restrictions.

## Development

Use Go 1.27+ and `golangci-lint v2.15.0` built with Go 1.27+.

```sh
make check
# Check canonical facade exports without modifying files.
make verify-generated
```

This runs formatting checks, static analysis, Go and Python acceptance tests, and
race detection. Python 3 is required for the acceptance-report tests; the SDK itself
has no Python dependency. See [AGENTS.md](AGENTS.md) for coding and repository rules.

- [Parity testing](tests/parity/README.md): pinned Python environment, fixture generation, and opt-in live reads.
- [Release validation](tests/acceptance/README.md): reproducibility, coverage mapping, platform builds, and benchmarks.
- [Changelog](CHANGELOG.md): release history and scope.

### Repository layout

| Path                                                          | Purpose                                                                      |
| ------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Root                                                          | Public clients, exports, and integration tests                               |
| `models/`                                                     | Data types and decoding                                                      |
| `pagination/`, `streams/`, `rfq/`, `transactions/`, `errors/` | Iteration, subscriptions, RFQ sessions, transaction confirmation, and errors |
| `internal/`                                                   | API clients, trading, wallet operations, signing, and transport              |
| `examples/`                                                   | Runnable examples                                                            |
| `tests/`, `testdata/`, `scripts/`                             | Acceptance tests, parity fixtures, and generation tools                      |

## License

The SDK is [MIT licensed](LICENSE). The go-ethereum signing dependency retains
LGPL-3.0-or-later. See [NOTICE](NOTICE) for upstream and dependency attribution.
