# 可运行示例

[English](README.md) · [SDK 文档](../README.zh-CN.md)

使用 Go 1.27+，从仓库根目录执行以下命令。公开示例需要网络，但无需账户凭据。
HTTP 示例设置了 10 或 30 秒的 context 截止时间；订阅示例默认运行 30 秒，也可用 Ctrl+C 退出。
`orders` 为离线示例：使用公开的合成签名者与内存 HTTP transport，不读取环境变量，也不打开网络连接。

| 示例                                   | 展示内容                                                 | 命令                                                      |
|----------------------------------------|----------------------------------------------------------|-----------------------------------------------------------|
| [events](events/main.go)               | 事件列表、按 ID／slug／URL 查询、关联市场与事件标签      | `go run ./examples/events`                                |
| [markets](markets/main.go)             | 开放市场、YES／NO 资产 ID、十进制价格历史与 `Items` 迭代 | `go run ./examples/markets`                               |
| [market](market/main.go)               | 按 ID／slug／URL 查询单市场元数据与标签                  | `go run ./examples/market -id <市场ID>`                   |
| [catalog](catalog/main.go)             | 标签、系列、体育及球队目录                               | `go run ./examples/catalog -kind tags`                    |
| [comments](comments/main.go)           | 事件评论与评论线程                                       | `go run ./examples/comments <事件ID>`                     |
| [search](search/main.go)               | 事件搜索及关联标签、用户资料                             | `go run ./examples/search presidential election`          |
| [orderbook](orderbook/main.go)         | 公开订单簿快照、最优报价及份额数量                       | `go run ./examples/orderbook <资产ID>`                    |
| [prices](prices/main.go)               | 批量 BUY／SELL 报价、中间价、价差与最近成交              | `go run ./examples/prices <资产ID> [资产ID ...]`          |
| [history](history/main.go)             | 按周期、时间窗口或历史时点查询价格                       | `go run ./examples/history -interval 1d <资产ID>`         |
| [estimate](estimate/main.go)           | FOK 所需限价、精确十进制与流动性错误类型                 | `go run ./examples/estimate <资产ID> BUY 10.50`           |
| [pagination](pagination/main.go)       | 有限页数的 `Pages` 遍历、完整性检查与 `FromCursor` 续页  | `go run ./examples/pagination -pages 2`                   |
| [trades](trades/main.go)               | 按钱包、condition 与方向筛选公开成交                     | `go run ./examples/trades -limit 10`                      |
| [activity](activity/main.go)           | 钱包活动类型分支与未知类型保留                           | `go run ./examples/activity <钱包地址>`                   |
| [positions](positions/main.go)         | 钱包公开索引持仓与盈亏的第一页                           | `go run ./examples/positions <钱包地址>`                  |
| [portfolio](portfolio/main.go)         | 公开资料、估值、统计、成交量与 PnL 序列                  | `go run ./examples/portfolio <钱包地址>`                  |
| [leaderboard](leaderboard/main.go)     | 按时间窗口及 PnL／成交量查看交易者排名                   | `go run ./examples/leaderboard -window week`              |
| [market-data](market-data/main.go)     | 市场持有人与索引未平仓量                                 | `go run ./examples/market-data <condition-ID>`            |
| [wallet](wallet/main.go)               | 索引中的 ERC20／ERC1155 交易授权状态                     | `go run ./examples/wallet <钱包地址>`                     |
| [streams](streams/main.go)             | 公开 Binance／Chainlink／TWAP、股票、体育及评论订阅      | `go run ./examples/streams -duration 15s btcusdt ethusdt` |
| [market-stream](market-stream/main.go) | 市场订单簿／报价类型事件、类型分支与队列诊断             | `go run ./examples/market-stream -duration 15s <资产ID>`  |
| [raw-stream](raw-stream/main.go)       | 原始市场字节、连接生命周期与有界回调队列                 | `go run ./examples/raw-stream -duration 15s <资产ID>`     |
| [account](account/main.go)             | 限制、余额／授权、挂单、成交、通知及用户／价格订阅       | `go run ./examples/account`                               |
| [orders](orders/main.go)               | 离线限价／市价签名、提交、拒绝、撤单与成交确认           | `go run ./examples/orders`                                |

运行前替换尖括号中的占位符。钱包地址是公开的 `0x` EVM 地址；如果出资钱包与签名者不同，
应使用出资钱包地址。

## 事件及关联市场

`events` 默认按成交量降序展示前五个开放事件，可用 `-limit` 调整分页大小，或用 `-closed`
查询已关闭事件。输出事件 ID、slug、标题，以及关联市场 ID、问题和 YES／NO 资产 ID。
列表只读取第一页，可能为空，并打印续页／完整性信息。

从列表复制事件 ID 或 slug，也可使用事件页面 URL，通过 `GetEvent` 和 `GetEventTags`
查询单个事件及其标签：

```sh
go run ./examples/events -limit 3
go run ./examples/events -closed -limit 3
go run ./examples/events -tag '<标签slug>'
go run ./examples/events -series '<系列ID>'
go run ./examples/events -id '<事件ID>'
go run ./examples/events -slug '<事件slug>'
go run ./examples/events -url 'https://polymarket.com/event/<事件slug>'
```

三个单事件查询选项互斥；`-limit`、`-closed`、`-tag` 与 `-series` 仅作用于列表模式。
一个事件包含多个市场，事件 ID 和 slug 不是结果代币 ID。
查询订单簿或订阅行情时，使用输出的关联市场资产 ID。
完整选项见 `go run ./examples/events -h`。

`market` 通过 `GetMarket` 和 `GetMarketTags` 查询单个二元市场，JSON 包含标准化状态、
交易／结算元数据、condition ID 和结果代币。`-id`、`-slug`、`-url` 三选一。
多市场事件应使用关联子市场的 slug 或完整子市场页面 URL。

`catalog -kind tags|series|teams` 展示第一页，默认五条；`catalog -kind sports` 返回不分页的
体育目录。标签 slug 和系列 ID 可用于 `events` 的筛选参数。
`comments <事件ID>` 查询事件评论第一页；`comments -thread <评论ID>` 查询线程及回复。
评论列表中除了请求数量的根评论，还可能包含回复。

## 从发现资产到查询行情

先运行 `go run ./examples/markets`，复制输出中的 `YES asset-id` 或 `NO asset-id`，
再用于 `orderbook`、`estimate`、`market-stream` 或 `raw-stream`。这里的 ID 是结果代币 ID，
不是 Gamma 市场 ID、事件 ID 或 condition ID。列表中的市场可能没有挂单或无法查询订单簿，
示例会展示结果或请求错误。

SDK 保留交易所买盘升序、卖盘降序的排列方式，`orderbook` 因此取两侧切片的最后一档作为最优报价。
价格和数量使用十进制类型，不转换成 `float64`。

估算时，BUY 的数量单位是抵押品，SELL 的数量单位是份额：

```sh
go run ./examples/estimate <资产ID> BUY 10.50
go run ./examples/estimate <资产ID> SELL 5
```

返回当前订单簿满足需求所需的边际限价，不是成交均价，也不会签名或提交订单。
空订单簿或不足以完成 FOK 的流动性会返回错误；后续操作前订单簿可能变化。

`prices` 展示单资产报价的批量接口，同时查询两个 `OrderSide`。按资产 ID 匹配映射和最近成交记录：
无成交的资产可能被省略，响应顺序也不保证与请求一致。四次查询各自独立，不构成同一时点的原子快照。

`history` 默认查询一天的价格历史，第一页最多 20 条；也支持 RFC3339 时间窗口和历史时点，
内部转换为 Unix 秒：

```sh
go run ./examples/history -interval 1w <资产ID>
go run ./examples/history -start '2026-10-20T00:00:00Z' -end '2026-10-21T00:00:00Z' <资产ID>
go run ./examples/history -as-of '2026-10-20T12:00:00Z' <资产ID>
```

周期、起始时间与历史时点三种选择互斥。显式时间窗口最多 15 天，省略结束时间时使用当前时间。
历史时点查询不设置分页大小或分桶大小；空历史属于正常结果。

## 公开钱包与市场数据

`trades` 默认展示公开成交第一页，可用 `-wallet`、`-condition` 和 `-side BUY|SELL` 筛选。
`activity` 展示钱包活动的类型分支，并在 JSON 中保留所有变体。这些公开索引记录与
`account` 使用的认证账户成交接口 `ListAccountTrades` 不同。

`portfolio` 无需凭据即可查询资料、估值、统计、累计 PnL 与成交量；用
`-interval 1w -fidelity 1d` 控制 PnL 序列。不存在的用户资料或统计输出为 null。
独立响应的更新时间可能不同，组合估值不等同于可用抵押品余额。
`leaderboard` 支持 `-window day|week|month|all`、可选 `-category` 和 `-sort PNL|VOLUME`。

`market-data` 使用 `market` 元数据中的 **condition ID** 查询持有人第一页与未平仓量。
该标识不同于事件 ID、Gamma 市场 ID 和 CLOB 资产 ID。JSON 分页结果保留
`has_more`、`next_cursor` 与 `limit_reached`，第一页不能当成全部历史或汇总。

## 分页与完整性

`pagination` 默认最多读取两页开放市场，每页五条。如有后续数据，会打印续页游标。
原样复制该值并加引号：

```sh
go run ./examples/pagination -pages 3 -cursor '<输出的游标>'
```

游标绑定原查询条件，包括筛选条件和分页大小；续页必须使用相同查询，游标也不会冻结上游数据。
`FromCursor(nil)` 表示结束，而不是重新开始。示例检查 `LimitReached`，遇到无法确认完整性的
服务端分页上限会返回错误。提前退出 `Pages` 不会预取下一页。
`positions` 与 `search` 只展示第一页，并打印续页／完整性信息，不能当作全账户汇总。

## 类型事件与原始行情

`streams` 默认订阅 Binance `btcusdt`，可通过位置参数选择其他品种，`-feed` 切换公开数据源：

```sh
go run ./examples/streams -feed chainlink -duration 15s btc/usd
go run ./examples/streams -feed twap -duration 15s btc/usd
go run ./examples/streams -feed equity -duration 15s AAPL
go run ./examples/streams -feed sports -duration 15s
go run ./examples/streams -feed comments -duration 15s <数字事件ID>
```

Chainlink 及其 30 秒 TWAP 使用 `btc/usd` 等带斜杠品种名；下方认证价格示例使用 `btcusd`
等标准 USD 对。股票代码需在上游目录中存在，暂时无事件不能证明订阅无效。
体育订阅接收全部公开赛事结果，评论订阅接收指定事件的新评论。

`market-stream` 与 `raw-stream` 支持一个或多个资产 ID。选项需放在位置参数之前。
两个类型事件示例使用 256 条队列，退出时向 stderr 打印 `Dropped()` 与流诊断统计。

`raw-stream` 直接使用 `streams` 包，设置 `RawOnly: true`，保留不可修改的
`RawFrame.Data` 缓冲区。接收完整消息，包括畸形数据和二进制消息，不含文本 `PONG`。
JSON 输出中的 `frame.Data` 为 **base64**，可保留任意字节，不将其解释成 JSON 或 UTF-8。
连接观测单独输出为 `lifecycle` 记录。例如：

```sh
go run ./examples/raw-stream -duration 15s <资产ID> > /tmp/polymarket-raw.jsonl
```

回调只入队，输出在 WebSocket 读取回调之外完成。示例将单条消息限制为 1 MiB，最多缓存
64 条记录，满时丢弃新记录。退出时向 stderr 报告丢弃的原始帧、连接观测及尚未处理的记录数，
不会在退出前排空队列。如果输出写入阻塞，退出可能延迟至该次写入完成。
这是有界示范，不提供持久化或无损采集保证。`read_limit` 表示消息超过读取上限，连接关闭，
不会交付部分字节。

生命周期回调可能并发执行，示例使用并发安全的 channel 与计数器。不要在回调内调用 `Close`
或其他阻塞客户端方法。序号和连接代只在单个客户端内有效，重连观测不能证明数据连续。
`RawOnly` 不交付市场类型事件，因此示例保留订阅句柄控制生命周期，不读取 `handle.Events`。

## 认证账户与订阅

`account` 要求以下五个环境变量，包括显式指定已有钱包。默认查询 closed-only 状态、
抵押品余额／授权、挂单和账户成交第一页，以及通知数量：

```sh
export POLYMARKET_PRIVATE_KEY='<私钥>'
export POLYMARKET_WALLET='<已有钱包地址>'
export POLYMARKET_API_KEY='<API密钥>'
export POLYMARKET_API_SECRET='<API密钥secret>'
export POLYMARKET_API_PASSPHRASE='<API密钥passphrase>'
go run ./examples/account
go run ./examples/account -stream user -duration 15s
go run ./examples/account -stream prices -duration 15s
```

余额和授权使用整数最小单位。用户订阅选择账户的全部市场，认证价格订阅选择 Chainlink `btcusd`。
订阅模式跳过快照查询，输出 JSON Lines，退出时向 stderr 打印队列诊断。
`-duration` 包含认证及后续请求／订阅的总耗时。

通过本地秘密管理方式注入凭据，不要提交到仓库或写入共享 shell 历史。
`NewSecureClient` 会验证凭据；传入密钥失效或不在账户密钥列表中时，可能创建／派生 API 密钥。
显式指定钱包可避免自动选择默认钱包；如 Deposit Wallet 尚未部署，仍可能触发部署准备，
或因缺少中继凭据而失败，因此应使用已部署钱包。构造过程不保证只读。
`Close` 只释放本地资源；`EndAuthentication` 会撤销 API 密钥，本示例不使用它。

## 离线订单生命周期

`orders` 通过自定义 `http.RoundTripper` 在内存中返回模拟响应。它生成公开的合成 EOA 签名者，
使用合成 API 凭据，不打开网络连接，也不读取账户秘密，可直接运行：

```sh
go run ./examples/orders
go run ./examples/orders -reject
go run ./examples/orders -pending
```

成功流程依次创建／签名 post-only GTC 限价单、提交一次、查询挂单、撤单，再签名／提交
带限价保护的 BUY FOK 市价单，等待提交时产生的成交确认。`-reject` 模拟 HTTP 200 中的
`RejectedOrder`；`-pending` 模拟确认超时并保留待确认成交 ID，两者以非零状态退出。
未知结果应使用原标识核对，不能直接重交。`AcceptedOrder` 本身不代表结算完成；确认轮询只覆盖
提交时产生的成交，不等待挂单未来成交。

模拟器展示 SDK 调用和响应处理，不验证真实撮合、余额、签名或链上确认。
生产代码可用 `PlaceLimitOrder`／`PlaceMarketOrder` 合并创建与提交，它们可能恢复已确认的授权不足；
`PostOrder` 只提交一次。示例均不会启用交易所订单心跳的超时撤销全部挂单规则。

钱包部署、授权／转账、拆分／合并／赎回、RFQ／组合交易、会话密钥和资金回收属于进阶流程，
参见 [SDK 使用说明](../README.zh-CN.md)。真实账户／人工资金验证程序及记录应按
[AGENTS.md](../AGENTS.md) 放在忽略的 `.local/` 下。

## 验证

```sh
CGO_ENABLED=0 go build ./examples/...
make check
```

上述检查编译所有示例并运行 SDK 离线检查，包含订单示例的正常、拒单和确认超时回归测试；
单独运行 `orders` 可查看命令输出。
这些检查不会执行网络示例或验证账户凭据。
订阅参数可通过 `go run ./examples/streams -h` 查看，`market-stream` 与 `raw-stream` 同理；
分页参数可通过 `go run ./examples/pagination -h` 查看。
