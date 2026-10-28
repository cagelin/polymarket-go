# Polymarket Go SDK

[English](README.md)

独立 Go SDK，对标官方 `polymarket-client 0.14.0`，源码提交
`7032ed06175b0bbea77b2d9e898bd9f252af2041`。需要 Go 1.27+，默认无需 CGO。

**当前完成第一阶段：公开数据与实时行情。** 认证交易和链上钱包尚未实现；尚未发布。
后续按“账户与交易 → 钱包与链上操作 → 完整验收”推进。

可用能力：

- 市场、事件、系列、标签、体育、评论、搜索和用户资料。
- 价格、订单簿、历史行情、市价估算、奖励和公开组合市场。
- 公开成交、持仓、活动、盈亏、排行榜、授权状态和账务下载。
- 市场、体育、评论、加密价格及均价、股票价格实时订阅；支持重连、心跳和取消。

```go
client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
if err != nil {
    return err
}
defer client.Close()

book, err := client.GetOrderBook(ctx, polymarket.Asset(assetID))
```

可运行示例：

```sh
go run ./examples/orderbook <资产ID>
go run ./examples/markets
go run ./examples/search election
go run ./examples/streams
```

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

| 目录 | 内容 |
|---|---|
| 根目录 | 客户端入口、统一公开名称及集成测试 |
| `models/` | 数据类型与解析 |
| `pagination/`、`streams/`、`errors/` | 分页、实时订阅、错误 |
| `internal/` | CLOB、Gamma、Data、RFQ 及共享网络实现 |
| `examples/` | 可运行示例 |
| `tests/parity/`、`testdata/`、`scripts/` | Python 对照、固定数据、生成工具 |

执行 `make check`，统一检查格式、18 项 lint、测试和 race。使用 `golangci-lint v2.15.0`
（由 Go 1.27+ 构建），严格禁止变量遮蔽，生成代码与测试同样检查。

[对照测试说明](tests/parity/README.md)包含重新生成和实网检查命令。
六个平台组合已通过无 CGO 构建；实际运行验证目前在 macOS arm64。

MIT 许可证；版权与依赖许可见 [NOTICE](NOTICE)。
