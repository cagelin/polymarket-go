# Python / Go 对照测试

对标 `polymarket-client==0.14.0`，提交 `7032ed06175b0bbea77b2d9e898bd9f252af2041`。
生成前验证压缩包版本和 202 个源码文件哈希。普通 Go 测试不需要 Python，也不依赖 `docs/`。

目前有 8,273 条固定对照用例，覆盖公开查询的请求、模型、分页、错误、金额运算、
授权状态，以及实时事件、订阅报文和过滤。另核对 62 个非实验公开客户端成员
（含异步客户端独有的订阅）。本地 WebSocket 测试覆盖断线重连、心跳、取消、
队列溢出和并发关闭。严格 lint 同样检查生成代码。

生成模型和全部固定数据：

```sh
python3 -m venv .venv-parity
.venv-parity/bin/python -m pip install -r tests/parity/requirements.txt
make generate PYTHON=.venv-parity/bin/python
make check
```

Windows 使用 `.venv-parity\Scripts\python.exe`。使用固定 `golangci-lint v2.15.0`
（Go 1.27+ 构建），必要时传入 `GOLANGCI_LINT=/path/to/golangci-lint`。

真实公开查询和订阅：

```sh
POLYMARKET_LIVE_PUBLIC=1 go test -run '^TestLivePublicMilestone$' -v
POLYMARKET_LIVE_STREAMS=1 go test -run '^TestLive(PublicStreams|MarketStream)$' -v
POLYMARKET_LIVE_DISCOVERY=1 go test -run '^TestLiveDiscovery$' -v
POLYMARKET_LIVE_EVENTS=1 go test -run '^TestLiveEvents$' -v
POLYMARKET_LIVE_ASSET_ID=<asset-id> go test -run '^TestLiveCLOB$' -v
go test -run '^$' -bench . -benchmem .
```

真实数据不断变化；实网检查只证明当前响应和消息可用，固定用例负责逐项对照。
评论等低频流在检查时间内可能没有新消息，超时会明确失败。授权查询使用公开排行榜
返回的钱包地址，只读取状态。所有这些检查均无需私钥，不发交易。

Go 不保留重复废弃字段；固定用例仍比较当前字段和服务端实际使用的旧字段名。
错误类别、字段和 cause 按合同验证，错误文案不逐字比较。

已登记的 Python 缺陷修正只有 `UP-001`：固定 pydantic-core 错误解析图片元数据的
负小数 Unix 时间戳。两条用例保留 `python_value`，用 Python 标准库独立计算正确值。
其他输出直接对照 Python SDK 或其默认 Decimal 上下文。

金额测试包含原 apd 无法表示的指数、28 位 half-even、下溢与溢出。
`internal/codec/decimal.go` 在 SDK 内完成舍入，绕开替代库在极小数上的默认舍入缺陷；
期望值不做修正。认证交易的费用和签名测试将在第二阶段补充。

---

Fixtures execute the pinned official Python SDK and default decimal context.
They compare public request contracts, decoded values, error categories, cursors,
approvals, stream frames/filtering, and decimal results. Source hashes are checked
before generation. Ordinary Go tests use committed JSON and require neither Python nor docs.

`make generate` regenerates all types and fixtures, then formats Go code.
`make check` runs strict lint, CGO-free tests, and race detection. The commands above
also enable opt-in public live reads/streams and local benchmarks. Slow live feeds
can time out; a connection alone is not treated as a successful event check.

Deprecated duplicate fields and Python framework-only APIs are omitted. Public
legacy feeds remain supported for capabilities unavailable from newer feeds.
The only corrected Python expectation is `UP-001`, with original output and an
independent standard-library calculation retained in both timestamp fixtures.
The Go decimal adapter preserves Python results, including subnormal rounding.
