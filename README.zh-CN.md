# Polymarket Go SDK

[English](README.md)

独立 Go SDK，对标官方 `polymarket-client 0.14.0`，源码提交
`7032ed06175b0bbea77b2d9e898bd9f252af2041`。需要 Go 1.27+，默认无需 CGO。

**第三阶段：离线验收通过，资金实测进行中。** EOA 资金流程、Deposit Wallet 部署、批量调用、
转账和挂撤单已实测通过。会话密钥授权返回 403；Safe／Proxy、真实成交结算和 RFQ 执行仍待实测。
尚未发布，尚未达到完整验收条件。

当前平台要求新订单使用 Deposit Wallet。中继服务对交易所只接受 ERC20 最大额度授权，
会拒绝有限额度和撤销请求；EOA 直接发送链上交易与钱包中继调用的限制并不相同。

可用能力：

- 市场、事件、系列、标签、体育、评论、搜索和用户资料。
- 价格、订单簿、历史行情、市价估算、奖励和公开组合市场。
- 公开成交、持仓、活动、盈亏、排行榜、授权状态和账务下载。
- 市场、体育、评论、加密价格及均价、股票价格实时订阅；支持重连、心跳和取消。
- 账户认证、密钥管理、余额／授权查询、通知、奖励、订单创建／提交／撤单及成交确认。
- 四类钱包的订单签名、组合报价／接受／状态跟踪、用户订单和认证价格订阅。
- Deposit Wallet 自动部署、EOA 链上提交、合约钱包批量调用、授权和转账。
- 头寸拆分／合并／赎回、会话密钥、资金回收计划，以及长期 RFQ 报价会话。

完整调用方式见[订单簿示例](examples/orderbook/main.go)。

可运行示例：

```sh
go run ./examples/orderbook <资产ID>
go run ./examples/markets
go run ./examples/search election
go run ./examples/streams
```

认证入口为 `NewSecureClient(ctx, SecureClientConfig{...})`。完整环境变量示例见
[examples/account](examples/account/main.go)，只展示账户限制查询。构造客户端会校验密钥；
密钥缺失或失效时可能创建／派生新密钥。`Close` 释放资源，`EndAuthentication` 还会撤销账户 API 密钥。

`CreateLimitOrder`／`CreateMarketOrder` 查询必要行情并签名；`PostOrder`／`PostOrders`
提交已签名订单；`PlaceLimitOrder`／`PlaceMarketOrder` 合并两步。返回 `AcceptedOrder`
只表示订单被接受，成交确认另用 `WaitForOrderFillSettlement`。普通订单提交不会自动重试。
组合报价接受遇到传输或响应格式错误时，按 Python 行为最多重发一次同一份签名订单。

钱包操作返回 `TransactionHandle`，调用 `Wait(ctx)` 等待确认。超时保留交易标识，
不代表失败，也不会自动重发未知结果。`*big.Int` 表示链上最小单位；合并／授权的
nil 金额表示最大值。`SetupTradingApprovals` 会等待缺失的授权全部确认。

构造默认 Deposit Wallet 客户端时，如钱包不存在，会使用 `BuilderKey` 或 `RelayerKey`
部署并等待确认；只读账户示例应显式指定已有钱包。`PlaceLimitOrder`／`PlaceMarketOrder`
只在服务端明确拒绝且确认授权不足时补授权，再重交原订单一次；`PostOrder` 不做该恢复。

`OpenRFQSession(ctx)` 返回长期会话，使用 `Next`、`Quote`、`CancelQuote`、
`RespondToConfirmation`；`Close` 关闭会话。断线后不自动重放报价，可重新打开会话。
`PlanCollateralReturn` 只生成计划，`ExecuteCollateralReturnPlan` 才提交计划中的调用。

首次发送订单心跳后，服务端启用超时撤单：每 5 秒续发，超过 10 秒可能撤销全部挂单；
关闭客户端不会解除该规则。

金额使用 `github.com/ericlagergren/decimal` 的 `*decimal.Big`，保留十进制精度。
市价估算返回所需限价，不是平均成交价，也不会下单。公开 Go 数据可以修改，调用方
应避免并发写入共享结果。SDK 内部复制请求参数。

分页使用 `FirstPage`、`Pages`、`Items` 和 `FromCursor`；`FromCursor(nil)` 表示结束。
达到服务端分页上限时，检查 `LimitReached`，不能视为数据已取全。游标绑定查询条件。
搜索和评论的续页规则保留 Python 行为，不要自行修改游标。

实时订阅通过 `Subscribe(ctx, specs...)` 创建，使用 `Next` 或 `Events` 接收。
同类订阅共用连接；每个返回的 handle 默认缓存 1024 条，满时丢弃最旧消息，
`Dropped()` 返回丢弃数量。多个 spec 合并后共用这一缓存；Python 按子订阅分配缓存。
因此不承诺跨数据源消息顺序。`Close` 结束订阅并释放资源。

网络操作支持 context 取消，客户端可以并发使用，`Close` 可重复调用。
传入的 `HTTPClient` 由调用方管理。普通 HTTP 请求默认超时 10 秒；Data API 查询仅对
429 按上游规则最多重试两次，账务下载不重试。其他公开 HTTP 查询不自动重试。

只公开当前字段名，不保留重复废弃别名。旧协议字段在内部转换。公开 Binance／Chainlink
订阅虽被上游标为旧接口，仍保留：它们提供无需认证、全品种、原报价币种和 30 秒均价等能力。

目录按职责划分：

| 目录                                                          | 内容                                                 |
| ------------------------------------------------------------- | ---------------------------------------------------- |
| 根目录                                                        | 客户端入口、统一公开名称及集成测试                   |
| `models/`                                                     | 数据类型与解析                                       |
| `pagination/`、`streams/`、`rfq/`、`transactions/`、`errors/` | 分页、实时订阅、RFQ 会话、交易确认、错误             |
| `internal/`                                                   | CLOB、Gamma、Data、RFQ、账户交易、签名及共享网络实现 |
| `examples/`                                                   | 可运行示例                                           |
| `tests/parity/`、`testdata/`、`scripts/`                      | Python 对照、固定数据、生成工具                      |

执行 `make check`，统一检查格式、变量遮蔽、18 项 lint、测试和 race。使用 `golangci-lint v2.15.0`
（由 Go 1.27+ 构建）。[编码标准](AGENTS.md)禁止遮蔽外层名称和 Go 预声明标识符，
包括回调参数及不同类型的同名变量；生成代码与测试同样检查。
可用 `make shadow` 单独检查，`make lint` 也会执行。检查工具位于独立模块 `tools/shadow/`，不增加 SDK 依赖。

[对照测试说明](tests/parity/README.md)包含重新生成和实网检查命令。
六个平台组合已通过无 CGO 构建；实际运行验证目前在 macOS arm64。

SDK 使用 MIT；以太坊签名依赖 go-ethereum 使用 LGPL-3.0-or-later，保留各依赖原许可。版权与依赖许可见 [NOTICE](NOTICE)。
