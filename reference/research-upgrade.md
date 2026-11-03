# Public research data upgrade

Version v0.3.0 adds Gamma fields, raw public HTTP responses, CLOB price history, and event counts. Amount filters and Data history timestamps no longer pass through floating-point values. [中文说明](research-upgrade.zh-CN.md).

## Caller migration

`ListTradesParams.FilterAmount`, `ListPositionsParams.FilterAmount`, and `ListMarketHoldersParams.MinBalance` change from `*float64` to `*json.Number`. Construct them from decimal text:

Construct `amount := json.Number("0.123456789012345678901")`, then assign `&amount` to `FilterAmount`.

Existing callers that assign float pointers need source changes. Amounts also participate in cursor fingerprints; start a new traversal after changing filters. The upstream trade `filter_amount` minimum is 0.01, including when zero is supplied. Smaller trades are omitted.

## Gamma and resolution fields

| SDK field                                 | Wire field                 | Meaning                                                   |
| ----------------------------------------- | -------------------------- | --------------------------------------------------------- |
| `Market.CreatedAt`, `UpdatedAt`           | `createdAt`, `updatedAt`   | Creation and update times                                 |
| `Market.State.AcceptingOrdersTimestamp`   | `acceptingOrdersTimestamp` | Order acceptance time, not first trade time               |
| `Market.State.EventStartTime`             | `eventStartTime`           | Upstream event start, without a local-date interpretation |
| `Market.GroupItemIndex`                   | `groupItemThreshold`       | Original index string, not a temperature threshold        |
| `Market.Resolution.UMAEndDate`            | `umaEndDate`               | Upstream UMA timestamp                                    |
| `Market.Resolution.UMAResolutionStatuses` | `umaResolutionStatuses`    | Strings from an array or JSON-encoded array               |
| `Market.Resolution.AutomaticallyResolved` | `automaticallyResolved`    | Upstream automatic resolution flag                        |
| `Event.Schedule.DeployingTimestamp`       | `deployingTimestamp`       | Upstream deployment timestamp                             |
| `Event.Series[].Recurrence`, `SeriesType` | `recurrence`, `seriesType` | Series recurrence and type                                |
| `Event.Tags[].ForceShow`, `ForceHide`     | `forceShow`, `forceHide`   | Tag display flags                                         |

New timestamps use `*TimestampValue`: `Raw` retains the string, `Time` is UTC on success, and `ParseError` records invalid or empty strings. Missing and null values both produce nil pointers; the HTTP body distinguishes them. Existing timestamp types remain unchanged.

`Event.Markets` contains binary markets. `Event.SkippedMarkets` records excluded child JSON and the reason. Normalization can reformat that JSON; only the HTTP callback preserves the complete original body.

`Resolution.RawPrices` retains the original JSON values of `proposed_price`, `reproposed_price`, and `price`, including null, empty strings, and `"69"`. Missing keys stay absent. Interpreted fields retain the existing 69 sentinel behavior and now treat empty strings as absent. Raw values retain upstream units.

## Raw public HTTP responses

`ClientConfig.RawHTTPHandler` receives a `RawHTTPResponse` before status handling or decoding. Records contain service, method, request URL/body, response status/headers, receive time, and response bytes. Coverage includes unsigned Gamma, Data, and CLOB GETs and CLOB batch reads at `/books`, `/prices`, `/midpoints`, `/spreads`, and `/last-trades-prices`.

- Bytes come directly from `net/http`, which may automatically decompress the body. They are never reconstructed from models.
- Non-2xx and malformed JSON responses are observable. Body read failures retain the received prefix; size-limit failures retain at most the configured limit. Both mark `Complete=false`, populate `ReadError`, and fail the call.
- Attempts without an HTTP response produce no record. Each upstream retry has its own record.
- Callbacks run synchronously and can run concurrently across calls. Honor the provided context and synchronize shared state.
- Body slices and response headers are independent copies that the callback can retain or modify.
- A callback error stops decoding and retrying. Its cause remains available through `errors.Is`.
- Authenticated requests and trading writes are excluded. Request headers are not exposed.

## History and counts

`GetCLOBPriceHistory(ctx, PriceHistoryQuery)` reads `/prices-history`. `Market` is an outcome token ID. Provide `Interval`, or both Unix-second `StartTs` and `EndTs`. These selectors are mutually exclusive. The SDK limits explicit windows to 15 days; split longer ranges in the caller. `Fidelity` is positive integer minutes.

`CLOBPriceHistoryPoint.Timestamp` and `Price` use `json.Number` to retain integer and decimal text. This method has no pagination and makes no coverage guarantee. Keep these samples separate from the Data buckets returned by `ListPriceHistory`. Neither source represents order books or individual trades.

`CountEvents(ctx, ListEventsParams)` reads `totalResults` from `/events/pagination` with `limit=1` and `offset=0`. It shares `ListEvents` filters and its `closed=false` default, and ignores `PageSize`. Counts describe one server response; use `ListEvents` for enumeration. The offset endpoint's 2000 limit does not affect this count-only request.

CLOB parameters follow the [official history reference](https://docs.polymarket.com/api-reference/markets/get-prices-history). The 15-day constraint and Gamma count route come from the R6 research observations. Empty responses do not establish complete or unavailable history.

## Validation scope

Regression tests use synthetic responses and local HTTP servers for field preservation, invalid timestamps, resolution originals, precision, query boundaries, callback failures, truncation, cancellation, retries, and concurrency. Pinned Python comparisons still check existing fields; dedicated tests check SDK additions. Fixture difference `SDK-001` retains Python's original error for empty resolution prices.

No live collection or account operation is included. Full CLOB market parameters, freshness, batch history, and stream changes awaiting R4 evidence remain outside this update.
