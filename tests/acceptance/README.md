# 完整验收 / Release validation

2026-10-29 范围调整：SDK 面向全新账户，交易钱包仅支持 Deposit Wallet；旧 Safe／Proxy 钱包接入、迁移和实测均排除，不作为验收或发布门槛。现有旧钱包实现与历史测试记录保留，不代表支持承诺。EOA 签名及其他已约定功能的范围不变。

Scope update (2026-10-29): this SDK targets new accounts using Deposit Wallets. Legacy Safe/Proxy onboarding, migration, and live validation are excluded from support and release acceptance. Existing legacy implementation and historical tests remain, without a support commitment. EOA signing and other agreed capabilities remain in scope.

`make validate-release PYTHON=.venv-parity/bin/python` 执行本地检查、固定 Python 数据重建、
六平台编译、本机离线运行和性能基线。需要 Go 1.27+、Make、固定 Python 环境，以及
由 Go 1.27+ 构建的 golangci-lint v2.15.0。结果保存在忽略的 `.local/validation/`。

也可分开执行：

```sh
python scripts/validate_release.py fixtures
python scripts/validate_release.py platforms
python scripts/validate_release.py benchmarks
```

第一项须使用固定 Python 环境，重建结果有差异即失败。每份报告记录源码散列、上游版本、
平台和执行日志；性能记录每项 3 次、每次 300ms，不设速度门槛。
`coverage.json` 将全部客户端成员关联到固定用例和 Go 测试；关联存在不代表每个线上场景通过。
新增成员、遗失用例或失效测试链接会使验收失败。

脚本会清除子进程中的 `POLYMARKET_*` 环境变量，不连接真实账户，也不执行资金测试。
构建通过与实际运行分开记录；未获得对应机器的平台记为 `not_run`。
按用户确认范围验收本机运行与六平台编译；其他平台实际运行不在验收范围内。
会话密钥实测按用户要求暂挂，离线测试继续保留。

---

`make validate-release PYTHON=.venv-parity/bin/python` checks code quality, regenerates
pinned Python fixtures, cross-compiles six targets, runs host tests, and records benchmarks.
Requires Go 1.27+, Make, the pinned Python environment, and golangci-lint v2.15.0 built
with Go 1.27+. Reports and logs are written under ignored `.local/validation/`.

The three commands above run individual steps. `fixtures` requires the pinned environment
and fails on generated drift. Reports identify source hashes, upstream revision and toolchain.
Benchmarks use three 300ms samples with no performance threshold. `coverage.json` links
client members to fixtures and tests; it does not establish full live acceptance.
New members, missing fixtures and broken test links fail validation.

Child processes do not inherit `POLYMARKET_*` credentials or live-test switches.
Compilation and execution are separate results; unavailable runtime targets stay `not_run`.
The agreed scope covers host execution and all six build targets; other runtime
checks are outside the acceptance scope.
Live session-key validation is paused; its offline tests remain enabled.


本目录仅保留 SDK 的自动验收。真实账户验证程序及其专用测试、数据和说明放在忽略的 `.local/stage4/live/`，单独运行，不纳入 `make check`。

This directory contains automated SDK acceptance only. Manual account validation tools and their dedicated tests, fixtures and instructions belong under ignored `.local/stage4/live/` and run separately from `make check`.

## v0.1.0 live scope / 实网验收范围

As of 2026-10-29, the funded checks are closed. Verified scenarios include EOA funding,
Deposit Wallet deployment/batching/transfers, post-only cancellation, market/all cancellation,
heartbeat cancellation, ordinary CTF buy/sell settlement using limit orders with FOK,
and CTF/V2/negative-risk/combo split, merge and redemption, including batch operations.
This does not establish coverage of every market-order entry point or market variant.
Test positions and Deposit cash are zero; pre-existing dust and original exchange approvals remain.
Temporary API invalidation was independently verified; Builder revocation was acknowledged,
but its independent invalidation probe was inconclusive. Private account evidence stays in `.local/`.

The following remain unverified and do not block this release by agreement: collateral-return
execution (`invalid authorization`; principal recovered through ordinary relayer merge),
RFQ settlement with an independent counterparty, remaining trading variants and Builder attribution,
notification deletion, stock TWAP live updates, and missing historical responses.
Session-key live validation is paused. Legacy Safe/Proxy onboarding and migration are out of scope.

截至 2026-10-29，实网资金测试已收尾。已核验 EOA 资金、Deposit Wallet 部署／批量／转账、
挂单撤单、市场／全部撤单、心跳自动撤单、普通 CTF 的限价单 FOK 买卖结算，及
CTF／V2／负风险／组合头寸拆分、合并、赎回和相关批量操作。以上不代表所有市价入口和市场变体均通过。
测试持仓与 Deposit 现金已归零；原有粉尘和交易所授权保留。临时 API 已独立验证失效；
Builder 撤销已收到成功响应，但独立失效探测无法确认。账户证据保存在忽略的 `.local/`。

以下按约定不阻塞本次发布，仍标记未验证：资金回收服务执行（`invalid authorization`，
实际通过普通中继合并回收）、独立对手方 RFQ 结算、其余交易变体与 Builder 归因、
通知删除、股票 TWAP 实时更新及历史响应缺口。会话密钥实测暂挂，旧 Safe／Proxy 接入迁移排除。
