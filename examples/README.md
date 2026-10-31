# Runnable examples

[中文](README.zh-CN.md) · [SDK documentation](../README.md)

Run these commands from the repository root with Go 1.27+. Public examples need
network access but no credentials. Each HTTP example has a 10- or 30-second context
deadline. Streaming examples stop after 30 seconds by default, or on Ctrl+C.
`orders` is an offline exception: its in-memory HTTP transport opens no sockets and
uses a public synthetic signer without reading environment variables.

| Example                                | What it demonstrates                                                                | Command                                                    |
| -------------------------------------- | ----------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| [events](events/main.go)               | Event lists, lookup by ID/slug/URL, child markets and event tags                    | `go run ./examples/events`                                 |
| [markets](markets/main.go)             | Open markets, YES/NO asset IDs, decimal price history, `Items` iteration            | `go run ./examples/markets`                                |
| [market](market/main.go)               | Single-market metadata and tags by ID/slug/URL                                      | `go run ./examples/market -id <market-id>`                 |
| [catalog](catalog/main.go)             | Tags, series, sports and teams                                                      | `go run ./examples/catalog -kind tags`                     |
| [comments](comments/main.go)           | Event comments and comment threads                                                  | `go run ./examples/comments <event-id>`                    |
| [search](search/main.go)               | Event search, related tags and profiles                                             | `go run ./examples/search presidential election`           |
| [orderbook](orderbook/main.go)         | Public book snapshot, best prices and share sizes                                   | `go run ./examples/orderbook <asset-id>`                   |
| [prices](prices/main.go)               | Batch BUY/SELL quotes, midpoints, spreads and last trades                           | `go run ./examples/prices <asset-id> [asset-id ...]`       |
| [history](history/main.go)             | Price history by interval, explicit time window or as-of time                       | `go run ./examples/history -interval 1d <asset-id>`        |
| [estimate](estimate/main.go)           | FOK marginal limit price, exact decimals, typed liquidity errors                    | `go run ./examples/estimate <asset-id> BUY 10.50`          |
| [pagination](pagination/main.go)       | Bounded `Pages` iteration, completeness checks, `FromCursor`                        | `go run ./examples/pagination -pages 2`                    |
| [trades](trades/main.go)               | Public trades with wallet, condition and side filters                               | `go run ./examples/trades -limit 10`                       |
| [activity](activity/main.go)           | Wallet activity variants, including unknown types                                   | `go run ./examples/activity <wallet-address>`              |
| [positions](positions/main.go)         | First page of public indexed wallet positions and PnL                               | `go run ./examples/positions <wallet-address>`             |
| [portfolio](portfolio/main.go)         | Public profile, value, statistics, volume and PnL series                            | `go run ./examples/portfolio <wallet-address>`             |
| [leaderboard](leaderboard/main.go)     | Trader rankings by time window and PnL/volume                                       | `go run ./examples/leaderboard -window week`               |
| [market-data](market-data/main.go)     | Market holders and indexed open interest                                            | `go run ./examples/market-data <condition-id>`             |
| [wallet](wallet/main.go)               | Indexed ERC20/ERC1155 trading approval state                                        | `go run ./examples/wallet <wallet-address>`                |
| [streams](streams/main.go)             | Public Binance/Chainlink/TWAP, equity, sports and comment streams                   | `go run ./examples/streams -duration 15s btcusdt ethusdt`  |
| [market-stream](market-stream/main.go) | Typed market book/quote events, type switches, queue diagnostics                    | `go run ./examples/market-stream -duration 15s <asset-id>` |
| [raw-stream](raw-stream/main.go)       | Raw market bytes, connection lifecycle, bounded callback queue                      | `go run ./examples/raw-stream -duration 15s <asset-id>`    |
| [account](account/main.go)             | Restrictions, balances/allowances, orders, fills, notifications, user/price streams | `go run ./examples/account`                                |
| [orders](orders/main.go)               | Offline limit/market signing, submission, rejection, cancellation and settlement    | `go run ./examples/orders`                                 |

Replace angle-bracket placeholders before running a command. Wallet addresses are
public `0x` EVM addresses; use the funded wallet's address when it differs from the signer.

## Events and their markets

`events` lists the first five open events ordered by descending volume. Use
`-limit` to change the page size, or `-closed` to list closed events. It prints each
event's ID, slug and title, plus child market IDs, questions and YES/NO asset IDs.
List output includes continuation/completeness information and covers only the
first page, which may be empty.

Copy an event ID or slug from the list, or use an event page URL, to fetch a single
event and its tags through `GetEvent` and `GetEventTags`:

```sh
go run ./examples/events -limit 3
go run ./examples/events -closed -limit 3
go run ./examples/events -tag '<tag-slug>'
go run ./examples/events -series '<series-id>'
go run ./examples/events -id '<event-id>'
go run ./examples/events -slug '<event-slug>'
go run ./examples/events -url 'https://polymarket.com/event/<event-slug>'
```

The three lookup options are mutually exclusive; `-limit`, `-closed`, `-tag` and
`-series` apply only to list mode. An event groups markets; its ID and slug do not identify an outcome
token. Use the printed child-market asset IDs with the book and stream examples.
Run `go run ./examples/events -h` for flag help.

`market` uses `GetMarket` and `GetMarketTags` for one binary market. Its JSON output
includes canonical state, trading/resolution metadata, condition ID and outcome
tokens. Specify exactly one of `-id`, `-slug` or `-url`. For an event with multiple
markets, use the child market's slug or full child market page URL.

`catalog -kind tags|series|teams` prints the first page (default five items);
`catalog -kind sports` prints the sports catalog without pagination. Tag slugs and
series IDs can be used with `events` filters. `comments <event-id>` prints a first
page of event comments; `comments -thread <comment-id>` reads a thread and replies.
Comment listings may contain replies in addition to the requested root comments.

## Discover an asset, then inspect it

Run `go run ./examples/markets`, copy a `YES asset-id` or `NO asset-id` value, then
use it with `orderbook`, `estimate`, `market-stream`, or `raw-stream`. These are
outcome token IDs, not Gamma market IDs, event IDs, or condition IDs. A listed market
can have an empty or unavailable book; the examples report the result or request error.

`orderbook` uses the last level of each slice because the SDK preserves the
exchange's ascending bids and descending asks. Prices and sizes retain decimal
precision rather than being converted to `float64`.

For estimation, BUY takes collateral units and SELL takes shares:

```sh
go run ./examples/estimate <asset-id> BUY 10.50
go run ./examples/estimate <asset-id> SELL 5
```

The estimate is the marginal limit price required by the current book, not VWAP.
It does not sign or submit an order. An empty book or insufficient FOK liquidity
returns an error. The book may change before any subsequent operation.

`prices` demonstrates the batch equivalents of single-asset quote queries, including
both `OrderSide` values. Match maps and last-trade records by asset ID: responses may
omit assets without trades, and response order need not match request order. The
four requests are independent reads, not one atomic snapshot.

`history` defaults to a one-day interval and a first page of 20 observations. It also
accepts explicit RFC3339 windows or an as-of time, converting dates to Unix seconds:

```sh
go run ./examples/history -interval 1w <asset-id>
go run ./examples/history -start '2026-10-20T00:00:00Z' -end '2026-10-21T00:00:00Z' <asset-id>
go run ./examples/history -as-of '2026-10-20T12:00:00Z' <asset-id>
```

Interval, start and as-of selectors are mutually exclusive. An explicit window may
span at most 15 days; omitting its end uses the current time. As-of queries do not
set page size or bucket size. Empty history is a valid result.

## Public wallet and market data

`trades` defaults to the latest public trade page. `-wallet`, `-condition` and
`-side BUY|SELL` narrow the selection; `activity` demonstrates type switches on
wallet activity while retaining all variants in JSON. These public indexed records
are separate from authenticated `ListAccountTrades` fills in `account`.

`portfolio` reads profile, valuation, statistics, cumulative PnL and volume without
credentials. `-interval 1w -fidelity 1d` controls the PnL series; absent profile or
statistics are emitted as null. Independent responses can reflect different update
times; portfolio value is not an available-collateral balance. `leaderboard` accepts
`-window day|week|month|all`, optional `-category`, and `-sort PNL|VOLUME`.

`market-data` takes a **condition ID** from `market` metadata and reads the first
holder page plus open interest. This is a different identifier from an event ID,
Gamma market ID or CLOB asset ID. All JSON page outputs include `has_more`,
`next_cursor` and `limit_reached`; a first page is not a complete history or total.

## Pagination and completeness

`pagination` reads at most two pages of five open markets by default. If more pages
remain, it prints a continuation cursor. Copy the printed value exactly and quote it:

```sh
go run ./examples/pagination -pages 3 -cursor '<printed-cursor>'
```

Cursors belong to the original query, including filters and page size. Resume with
the same query; a cursor does not freeze changing upstream data. `FromCursor(nil)`
means completion rather than a restart. The example checks `LimitReached` and
reports unknown completeness as an error. Breaking `Pages` does not fetch another
page. `positions` and `search` intentionally show only their first page and print
continuation/completeness information; their output is not an account-wide total.

## Typed and raw streams

`streams` defaults to Binance `btcusdt`; positional arguments select other symbols.
`-feed` selects additional public services:

```sh
go run ./examples/streams -feed chainlink -duration 15s btc/usd
go run ./examples/streams -feed twap -duration 15s btc/usd
go run ./examples/streams -feed equity -duration 15s AAPL
go run ./examples/streams -feed sports -duration 15s
go run ./examples/streams -feed comments -duration 15s <numeric-event-id>
```

Chainlink and its 30-second TWAP use slash-separated symbols such as `btc/usd`;
the authenticated price example below uses canonical USD pairs such as `btcusd`.
Equity symbols must exist in the upstream catalog; a quiet stream is not proof of
an invalid subscription. Sports receives all public sports results, and comments
receives new comments for one event.

`market-stream` and `raw-stream` accept one or more asset IDs. Put flags before
positional arguments. Both typed examples use a 256-event queue and print
`Dropped()` and aggregate stream diagnostics to stderr on exit.

`raw-stream` uses the `streams` package directly with `RawOnly: true`. It retains
immutable `RawFrame.Data` buffers and receives complete messages, including malformed
and binary data, excluding text `PONG`. JSON output encodes `frame.Data` as **base64**,
so arbitrary bytes survive without being interpreted as JSON or UTF-8. Connection
observations appear as separate `lifecycle` records. For example:

```sh
go run ./examples/raw-stream -duration 15s <asset-id> > /tmp/polymarket-raw.jsonl
```

Callbacks only enqueue records; output happens outside the socket reader. The
example limits messages to 1 MiB and queues at most 64 records, dropping new records
when full. It reports dropped frames, dropped observations and records still queued
at exit to stderr. It deliberately does not drain the queue on shutdown. Slow output
can delay shutdown until the current output operation completes. This is a bounded demonstration;
it does not provide durable or lossless capture. A `read_limit` observation means a
message exceeded the configured size and the connection closed without delivering
partial bytes.

Lifecycle callbacks may run concurrently; channel sends and counters are concurrency
safe. Never call `Close` or blocking client methods from these callbacks. Sequence
and generation identifiers are scoped to one client, and reconnection observations
do not prove data continuity. `RawOnly` suppresses typed market delivery, so this
example keeps the subscription handle alive without consuming `handle.Events`.

## Authenticated account and subscriptions

`account` requires all five environment variables, including an explicit existing
wallet. By default, it reads closed-only status, collateral balance/allowances,
the first open-order and account-fill pages, and the notification count:

```sh
export POLYMARKET_PRIVATE_KEY='<private-key>'
export POLYMARKET_WALLET='<existing-wallet-address>'
export POLYMARKET_API_KEY='<api-key>'
export POLYMARKET_API_SECRET='<api-secret>'
export POLYMARKET_API_PASSPHRASE='<api-passphrase>'
go run ./examples/account
go run ./examples/account -stream user -duration 15s
go run ./examples/account -stream prices -duration 15s
```

Balances and allowances use integer base units. The user stream selects all account
markets; the authenticated price stream selects Chainlink `btcusd`. Stream modes
skip the snapshot queries and print JSON Lines, then queue diagnostics to stderr.
`-duration` includes authentication time as well as requests/subscription lifetime.

Supply secrets through your local secret-management workflow and keep them out of
source control and shared shell history. `NewSecureClient` validates credentials and
may create/derive an API key when the supplied key is invalid or absent from the
account's keys. An explicit wallet avoids automatic default-wallet selection; an
undeployed Deposit Wallet can still trigger deployment setup or fail without relayer
credentials. Use an already deployed wallet. Construction is therefore not guaranteed
to be read-only. `Close` releases local resources; `EndAuthentication` would revoke
the API key and is not used here.

## Offline order lifecycle

`orders` uses a custom `http.RoundTripper` with in-memory responses. It generates a
public synthetic EOA signer and supplies synthetic API credentials; it opens no
sockets and never loads account secrets. Run it directly:

```sh
go run ./examples/orders
go run ./examples/orders -reject
go run ./examples/orders -pending
```

The successful flow creates/signs a post-only GTC limit, submits it once, lists it,
cancels it, then signs/submits a price-bounded BUY FOK market order and waits for
placement-fill confirmation. `-reject` returns a `RejectedOrder` from an HTTP 200
response; `-pending` times out with the pending trade IDs. Both failure modes exit
nonzero. Reconcile unknown outcomes using retained identifiers rather than submitting
again. `AcceptedOrder` alone does not prove settlement, and settlement polling covers
placement fills, not future fills of a resting limit order.

The mock demonstrates SDK calls and response handling; it does not validate real
exchange matching, balances, signatures or chain confirmation. Production code may
use `PlaceLimitOrder`/`PlaceMarketOrder` to combine creation and submission; those
methods can recover confirmed allowance shortfalls, whereas `PostOrder` submits
once. No example activates the exchange order-heartbeat cancel-all deadline.

Wallet deployment, approval/transfer, split/merge/redemption, RFQ/combo execution,
session keys and collateral return remain advanced workflows described in the
[SDK guide](../README.md). Real-account/funded validation programs and their records
belong under ignored `.local/`, as required by [AGENTS.md](../AGENTS.md).

## Verification

```sh
CGO_ENABLED=0 go build ./examples/...
make check
```

These checks compile every example and run SDK offline checks, including the
order example's successful, rejected and pending-settlement regression tests.
Execute `orders` separately to see the command output. They do not execute the network
examples or validate account credentials. Streaming flags are documented
by `go run ./examples/streams -h` (likewise `market-stream` and `raw-stream`), and
pagination flags by `go run ./examples/pagination -h`.
