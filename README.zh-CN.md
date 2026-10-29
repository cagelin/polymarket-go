# Polymarket Go SDK

[English](README.md) · [API 文档](https://pkg.go.dev/github.com/cagelin/polymarket-go) · [版本发布](https://github.com/cagelin/polymarket-go/releases)

独立的 Polymarket Go SDK，提供市场数据、实时订阅、认证交易及钱包操作。
实现与 `polymarket-client 0.14.0` 对照验证，上游源码固定为
`7032ed06175b0bbea77b2d9e898bd9f252af2041`。

## 功能

- 市场、事件、系列、标签、体育、评论、搜索和用户资料。
- 价格、订单簿、历史行情、市价估算、奖励和公开组合市场。
- 公开成交、持仓、活动、盈亏、排行榜、授权状态和账务下载。
- 市场、体育、评论、加密价格及均价、股票价格实时订阅；支持重连、心跳和取消。
- 账户认证、密钥管理、余额／授权查询、通知、奖励、订单创建／提交／撤单及成交确认。
- Deposit Wallet 订单签名、组合报价／接受／状态跟踪、用户订单和认证价格订阅。
- Deposit Wallet 自动部署、EOA 链上提交、合约钱包批量调用、授权和转账。
- 头寸拆分／合并／赎回、会话密钥、资金回收计划，以及长期 RFQ 报价会话。

## 安装

需要 **Go 1.27 或更高版本**，SDK 构建无需 CGO。

```sh
go get github.com/cagelin/polymarket-go@v0.1.0
```

## 快速开始

[订单簿示例](examples/orderbook/main.go)展示公开客户端初始化、context 取消和资源释放。
公开查询无需账户凭据。克隆本仓库后，可运行以下示例：

```sh
go run ./examples/orderbook <资产ID>
go run ./examples/markets
go run ./examples/search election
go run ./examples/streams
go run ./examples/wallet <钱包地址>
```

## 使用说明

### 账户认证

认证入口为 `NewSecureClient(ctx, SecureClientConfig{...})`。完整环境变量示例见
[examples/account](examples/account/main.go)，只展示账户限制查询。构造客户端会校验密钥；
密钥缺失或失效时可能创建／派生新密钥。`Close` 释放资源，`EndAuthentication` 还会撤销账户 API 密钥。

### 订单与成交确认

`CreateLimitOrder`／`CreateMarketOrder` 查询必要行情并签名；`PostOrder`／`PostOrders`
提交已签名订单；`PlaceLimitOrder`／`PlaceMarketOrder` 合并两步。返回 `AcceptedOrder`
只表示订单被接受，成交确认另用 `WaitForOrderFillSettlement`。普通订单提交不会自动重试。
组合报价接受遇到传输或响应格式错误时，按 Python 行为最多重发一次同一份签名订单。

首次发送订单心跳后，服务端启用超时撤单：每 5 秒续发，超过 10 秒可能撤销全部挂单；
关闭客户端不会解除该规则。

### 钱包操作

钱包操作返回 `TransactionHandle`，调用 `Wait(ctx)` 等待确认。超时保留交易标识，
不代表失败，也不会自动重发未知结果。`*big.Int` 表示链上最小单位；合并／授权的
nil 金额表示最大值。`SetupTradingApprovals` 会等待缺失的授权全部确认。

构造默认 Deposit Wallet 客户端时，如钱包不存在，会使用 `BuilderKey` 或 `RelayerKey`
部署并等待确认；只读账户示例应显式指定已有钱包。`PlaceLimitOrder`／`PlaceMarketOrder`
只在服务端明确拒绝且确认授权不足时补授权，再重交原订单一次；`PostOrder` 不做该恢复。

### RFQ 与资金回收

`OpenRFQSession(ctx)` 返回长期会话，使用 `Next`、`Quote`、`CancelQuote`、
`RespondToConfirmation`；`Close` 关闭会话。断线后不自动重放报价，可重新打开会话。
`PlanCollateralReturn` 只生成计划，`ExecuteCollateralReturnPlan` 才提交计划中的调用。

### 精度与分页

金额使用 `github.com/ericlagergren/decimal` 的 `*decimal.Big`，保留十进制精度。
市价估算返回所需限价，不是平均成交价，也不会下单。公开 Go 数据可以修改，调用方
应避免并发写入共享结果。SDK 内部复制请求参数。

分页使用 `FirstPage`、`Pages`、`Items` 和 `FromCursor`；`FromCursor(nil)` 表示结束。
达到服务端分页上限时，检查 `LimitReached`，不能视为数据已取全。游标绑定查询条件。
搜索和评论的续页规则保留 Python 行为，不要自行修改游标。

### 实时订阅

实时订阅通过 `Subscribe(ctx, specs...)` 创建，使用 `Next` 或 `Events` 接收。
同类订阅共用连接；每个返回的 handle 默认缓存 1024 条，满时丢弃最旧消息，
`Dropped()` 返回丢弃数量。多个 spec 合并后共用这一缓存；Python 按子订阅分配缓存。
因此不承诺跨数据源消息顺序。`Close` 结束订阅并释放资源。

### 网络与并发

网络操作支持 context 取消，客户端可以并发使用，`Close` 可重复调用。
传入的 `HTTPClient` 由调用方管理。普通 HTTP 请求默认超时 10 秒；Data API 查询仅对
429 按上游规则最多重试两次，账务下载不重试。其他公开 HTTP 查询不自动重试。

## 兼容性与验证范围

交易支持面向使用 **Deposit Wallet** 的新账户，同时提供 EOA 签名和链上提交能力。
旧 Safe／Proxy 钱包接入与迁移不在支持范围内；保留的旧实现和测试不代表支持承诺。

仅公开当前字段名，旧协议字段在内部转换。保留公开 Binance／Chainlink 订阅，
以提供无需认证、全品种、原报价币种及 30 秒均价等能力。

发布验证覆盖 Linux、macOS、Windows 的 amd64 和 arm64 无 CGO 构建；
实际运行验证覆盖 macOS arm64。离线对照与实网验证分别记录，已实现不代表全部经过实网验证。
已验证场景和已知缺口见[验收说明](tests/acceptance/README.md)，其中包括资金回收服务授权、
RFQ 结算和会话密钥等限制。

v0.1.0 实网检查期间，中继服务要求交易所 ERC20 最大额度授权，并拒绝有限额度及零值撤销。
EOA 直接交易与钱包中继调用可能受不同的服务限制。

## 开发

使用 Go 1.27+ 和由 Go 1.27+ 构建的 `golangci-lint v2.15.0`。

```sh
make check
```

该命令执行格式检查、静态分析、Go 与 Python 验收测试及 race 检查。
验收报告测试需要 Python 3，SDK 本身不依赖 Python。
编码及目录规范见 [AGENTS.md](AGENTS.md)。

- [对照测试说明](tests/parity/README.md)：固定 Python 环境、数据生成及可选的实网只读检查。
- [发布验收说明](tests/acceptance/README.md)：可复现性、测试覆盖映射、平台编译和性能基线。
- [更新日志](CHANGELOG.md)：版本历史与范围。

### 仓库结构

| 路径 | 职责 |
| --- | --- |
| 根目录 | 公开客户端、导出类型及集成测试 |
| `models/` | 数据类型与解析 |
| `pagination/`、`streams/`、`rfq/`、`transactions/`、`errors/` | 分页、订阅、RFQ 会话、交易确认及错误 |
| `internal/` | API 客户端、交易、钱包操作、签名及网络实现 |
| `examples/` | 可运行示例 |
| `tests/`、`testdata/`、`scripts/` | 验收测试、对照数据及生成工具 |

## 许可

SDK 使用 [MIT 许可](LICENSE)，签名依赖 go-ethereum 保留 LGPL-3.0-or-later 许可。
上游及依赖归属见 [NOTICE](NOTICE)。
