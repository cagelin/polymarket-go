# Python / Go 对照测试

对标 `polymarket-client==0.14.0`，提交 `7032ed06175b0bbea77b2d9e898bd9f252af2041`。
生成前验证压缩包版本和 202 个源码文件哈希。普通 Go 测试不需要 Python，也不依赖 `docs/`。

目前有 **13,322 条固定对照用例**，覆盖公开查询、账户请求、模型、分页、金额、
订单签名／提交／成交等待、组合报价及实时事件。公开客户端核对 62 个成员；
认证客户端核对 120 个成员：117 个方法、构造入口，以及明确排除的 2 个废弃空操作。
接口存在检查与行为对照分别执行，不以存在检查代替行为验收。

本地 HTTP／WebSocket 测试覆盖认证回退、重定向隔离、并发缓存、撤单、订阅确认／拒绝、
断线恢复、共享过滤、64 个过滤条件上限、取消和关闭。所有账户与交易测试均用合成密钥
及本地服务；专用 EOA 已通过认证和只读实测，资金操作尚未验收。严格 lint 同样检查生成代码。

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
go test -run '^$' -bench . -benchmem . ./internal/signing
```

真实数据不断变化；实网检查只证明当前响应和消息可用，固定用例负责逐项对照。
评论等低频流在检查时间内可能没有新消息，超时会明确失败。授权查询使用公开排行榜
返回的钱包地址，只读取状态。所有这些检查均无需私钥，不发交易。

Go 不保留重复废弃字段；固定用例仍比较当前字段和服务端实际使用的旧字段名。
错误类别、字段和 cause 按合同验证，错误文案不逐字比较。

登记的上游差异保留原始 Python 输出：

- `UP-001`：图片元数据负小数时间戳被固定 pydantic-core 错误解析；两条向量使用标准库验证正确时间。
- `UP-002`：已确认成交缺少交易哈希时，Python 返回 `(None,)`，与声明的哈希列表不符。
  Go 返回 `UnexpectedResponseError`；原值、修正值及输入保留在 `order-post-parity.json`。

金额测试包含极端指数、28 位 half-even、下溢、溢出和费用计算。其他期望值直接来自
固定 Python SDK／默认 Decimal 上下文。第三阶段增加交易 ABI／签名字节、组合条件推导、
资金回收模型、会话权限参数和 RFQ 报价签名对照。本地服务覆盖部署、授权恢复、确认、
部分失败、nonce 修正、幂等重试、撤权等待和 RFQ 应答关联。

---

Fixtures execute the pinned Python SDK and default decimal context. All 13,322
cases are offline. Source hashes are checked before generation; ordinary Go tests
need neither Python nor docs. Surface checks are separate from behavioral tests.

`make generate` regenerates models and fixtures. `make check` runs strict lint,
CGO-free tests and race detection. Live public checks remain opt-in. Authenticated
live authentication passed for a dedicated EOA; funded flows remain unverified. Stage 3 adds
ABI/signature parity, wallet flows, session keys, collateral plans, and persistent RFQ sessions.

`UP-001` corrects negative fractional timestamps using Python's standard library.
`UP-002` rejects a confirmed trade without a transaction hash rather than returning
Python's `(None,)`. Fixtures preserve both original outputs and explicit corrections.
Deprecated duplicate fields and Python framework-only APIs are omitted; public
legacy feeds remain supported where they provide otherwise unavailable capabilities.

第三阶段补充差异：不保留 `setup_gasless_wallet` / `is_gasless_ready` 两个废弃空操作；
Go 以 `TransactionHandle` 统一 EOA 和 relayer 等待，合并／授权用 nil 表示最大金额。
会话 scope 按固定 Python 源码保留大小写，其文档宣称的自动大写并未实现。
RPC 和 relayer 返回不匹配的请求／交易标识时拒绝响应；代理 gas 估算不吞掉 context 取消。
超时 RFQ 的迟到应答不能确认后续报价；断线后重新建会话，不重放资金操作。
