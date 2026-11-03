# 公开数据接口升级

v0.3.0 补齐研究所需的 Gamma 字段、公共 HTTP 原文、CLOB 价格历史和事件总数。价格与数量筛选参数使用十进制文本，Data 价格历史时间戳不再经过浮点数。

## 调用方需要调整

`ListTradesParams.FilterAmount`、`ListPositionsParams.FilterAmount` 和 `ListMarketHoldersParams.MinBalance` 从 `*float64` 改为 `*json.Number`。用原始十进制文本构造参数：

构造 `amount := json.Number("0.123456789012345678901")`，再把 `&amount` 赋给 `FilterAmount`。

已有旧类型变量的调用方需要修改源码。新的金额参数参与分页游标计算；不要继续使用旧筛选条件生成的游标。成交筛选 `filter_amount` 的上游最小值为 0.01，传 0 也按该下限处理，小于下限的成交不返回。

## Gamma 字段

| SDK 字段                                  | Gamma 原字段               | 含义                                         |
| ----------------------------------------- | -------------------------- | -------------------------------------------- |
| `Market.CreatedAt`、`UpdatedAt`           | `createdAt`、`updatedAt`   | 创建与更新时间                               |
| `Market.State.AcceptingOrdersTimestamp`   | `acceptingOrdersTimestamp` | 开始接受订单的时间，不等同于首次成交         |
| `Market.State.EventStartTime`             | `eventStartTime`           | 上游事件起点，不在 SDK 中解释为当地目标日期  |
| `Market.GroupItemIndex`                   | `groupItemThreshold`       | 原字符串序号，不当作温度阈值                 |
| `Market.Resolution.UMAEndDate`            | `umaEndDate`               | 上游 UMA 时间，不推断其与最终结算的关系      |
| `Market.Resolution.UMAResolutionStatuses` | `umaResolutionStatuses`    | 状态字符串数组，支持数组和 JSON 编码的字符串 |
| `Market.Resolution.AutomaticallyResolved` | `automaticallyResolved`    | 上游自动结算标记                             |
| `Event.Schedule.DeployingTimestamp`       | `deployingTimestamp`       | 上游部署时间                                 |
| `Event.Series[].Recurrence`、`SeriesType` | `recurrence`、`seriesType` | 系列重复周期与类型                           |
| `Event.Tags[].ForceShow`、`ForceHide`     | `forceShow`、`forceHide`   | 标签显示标记                                 |

新增时间字段使用 `*TimestampValue`。`Raw` 保存原字符串，解析成功后 `Time` 为 UTC；解析失败时 `Time` 为空，`ParseError` 说明原因。空字符串也保留并标明解析失败。缺失与 JSON null 都对应 nil，精确区分两者须读取 HTTP 原文。既有时间字段的类型不变。

`Event.Markets` 继续只包含二元市场。非二元子市场进入 `Event.SkippedMarkets`，保留对象 JSON 和跳过原因。该对象 JSON 可能在模型规范化时被重新排版，不能代替完整 HTTP 原文。

`Resolution.RawPrices` 保存 `proposed_price`、`reproposed_price`、`price` 的原始 JSON 值，区分缺失、null、空字符串和 `"69"`。原有解释字段仍把 `69` 当作未设置值，新增空字符串转为空值的处理。原文中的数字仍为原单位，不能直接当作 0 到 1 的概率。

## HTTP 原文回调

`ClientConfig.RawHTTPHandler` 接收 `RawHTTPResponse`，包含服务、方法、请求 URL、请求正文、响应状态、响应头、接收时间和正文原字节。回调在状态处理和模型解码之前运行，覆盖 Gamma、Data、CLOB 的未认证 GET，以及 CLOB `/books`、`/prices`、`/midpoints`、`/spreads`、`/last-trades-prices` 批量读取。

- 正文是 `net/http` 交付的字节，可能已经自动解压；SDK 不解析后重新编码。
- 正常响应、非 2xx 和解析失败响应都能留证。读取失败保留已读部分；超过配置上限只保留上限内的前缀。两种情况均设置 `Complete=false` 和 `ReadError`，调用仍返回错误。
- 未获得 HTTP 响应的连接失败不触发回调。上游重试每次单独触发，不合并尝试。
- 回调同步运行，多请求可并发调用。回调须响应传入 context 的取消，并自行同步共享文件或状态。
- 字节切片和响应头为独立副本，可留存或修改，不影响 SDK 后续解码。
- 回调返回错误时，本次调用立即结束，不重试、不解码，`errors.Is` 能识别回调错误。
- 不观察认证请求和交易写入，不提供请求头。

## CLOB 价格历史与事件总数

`GetCLOBPriceHistory(ctx, PriceHistoryQuery)` 调用 `/prices-history`。`Market` 是 outcome token ID；可选择 `Interval`，或同时提供 Unix 秒整数 `StartTs`、`EndTs`。两种方式互斥；显式时间窗由 SDK 限制为不超过 15 天，较长范围由调用方拆分。`Fidelity` 是正整数分钟。

返回的 `CLOBPriceHistoryPoint.Timestamp`、`Price` 是 `json.Number`，保留数值文本，不经过浮点数。该方法没有分页，不承诺上游返回完整历史。它与 `ListPriceHistory` 的 Data 分桶序列分别保留，不能当作盘口或逐笔成交。

`CountEvents(ctx, ListEventsParams)` 请求 `/events/pagination?limit=1&offset=0` 的 `pagination.totalResults`，复用 `ListEvents` 筛选条件，默认 `closed=false`，忽略 `PageSize`。总数仅代表响应时点；枚举仍使用 `ListEvents`。该偏移接口的 2000 上限不影响只读总数。

CLOB 参数依据 [官方价格历史文档](https://docs.polymarket.com/api-reference/markets/get-prices-history)。15 天限制及 Gamma 总数入口依据 项目 R6 的接口实测记录。接口没有返回数据时不能据此认定历史不存在或采集完整。

## 验证边界

新增回归测试使用合成响应和本地临时 HTTP 服务，覆盖字段保留、异常时间、结算原值、精度、参数边界、回调错误、截断、取消、重试和并发。固定 Python 对照继续核对原有字段；SDK 新增字段单独测试。空结算价格的语义差异登记为 `SDK-001`，固定样本保留 Python 原错误。

本次不新增实网采集或账户操作。完整 CLOB 市场参数、freshness、批量历史及待 R4 确认的流能力不在本次实现范围。
