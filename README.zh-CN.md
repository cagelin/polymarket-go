# Polymarket Go SDK

[English](README.md)

独立 Go SDK，对标官方 `polymarket-client 0.14.0`。**开发中，尚未完整覆盖官方功能。**

目前可用：市场查询与筛选分页、历史价格，以及单个／批量价格、中间价、价差、订单簿、最近成交价和市价估算。事件查询、全文搜索、其他历史数据、实时订阅、交易和钱包功能仍待实现。

需要 Go 1.27 或以上，默认无需 CGO。代码位于本地，尚未发布到 GitHub 或 Go 模块代理。

```go
client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
if err != nil {
    return err
}
defer client.Close()

book, err := client.GetOrderBook(ctx, polymarket.Asset(assetID))
```

完整可运行示例：

```sh
go run ./examples/orderbook <资产ID>
go run ./examples/markets
```

资产 ID 使用字符串类型 `ClobAssetID`；金额直接使用 `github.com/cockroachdb/apd/v3` 的十进制类型。市价估算返回满足指定数量所需的限价，不是平均成交价，也不会下单。

新 SDK 只公开 `ConditionID`、`AssetID` 等当前名称，不继承 Python 的废弃字段和别名。服务端 `market`、`token_id` 的收发转换在内部完成。市场结果中的 `TokenID` 与 `PositionID` 分别表示不同协议的资产编号，不是重复旧别名，因此保留。

`GetMarket` 支持 ID、slug 或 Polymarket 链接；`ListMarkets` 保留完整筛选参数，只返回二元市场。分页器支持 `FirstPage`、`FromCursor`、`Pages` 和 `Items`，提前退出循环不会多发请求；`FromCursor(nil)` 表示没有后续数据。`ListPriceHistory` 支持预设时间段、起止时间和指定时刻查询。

网络方法都接受 `context.Context`。客户端可并发使用；`Close` 可重复调用。传入的 `HTTPClient` 由调用方管理。默认请求总超时 10 秒，可通过自定义 HTTP 客户端调整。价格和市场查询不自动重试；历史价格仅对 429 限流按官方规则最多重试两次，等待可取消。

测试：

```sh
CGO_ENABLED=0 go test ./...
go test -race ./...
go vet ./...
```

统一执行 `make check`：格式检查、golangci、测试和并发检查。使用 `golangci-lint v2.15.0`（需用 Go 1.27 或以上构建）；可通过 `GOLANGCI_LINT=/path/to/golangci-lint` 指定位置。严格检查变量遮蔽，测试代码同样适用。

[Python 对照测试说明](tests/parity/README.md)介绍如何重新生成测试数据。真实公开接口测试需显式设置 `POLYMARKET_LIVE_ASSET_ID`，不会交易或访问钱包。

当前限制：极端十进制指数超出 apd 支持范围时会报错，完整金额兼容性尚未验收。

MIT 许可证；上游版权及依赖许可见 [NOTICE](NOTICE)。
