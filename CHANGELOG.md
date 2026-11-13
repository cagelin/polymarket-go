# Changelog

Scope update (2026-10-29): this SDK targets new accounts using Deposit Wallets. Legacy Safe/Proxy onboarding, migration, and live validation are excluded from support and release acceptance. Existing legacy implementation and historical tests remain, without a support commitment. EOA signing and other agreed capabilities remain in scope.

## v0.3.3 — 2026-11-13

- Decode JSON decimal numbers directly from their original text. Prices and quantities retain all digits instead of passing through binary floating point.
- Added regression tests for long fractions, large integers, signed values and exponents outside the float64 range. String decimal inputs keep their existing behavior.

中文：JSON 数字直接按原始文本解析为十进制值，价格和数量保留全部有效数字。新增长小数、大整数、负数及超出 float64 范围的指数回归测试；字符串数值沿用原有行为。

## v0.3.1 — 2026-11-03

- Added offline tests for account-query user selection, credential isolation, RFQ message validation, event discovery, price history and bounded pagination.
- Made the event, history and pagination examples testable with local HTTP endpoints and explicit output writers. They now report output failures and stop further reads.
- Added the existing 80% whole-repository coverage gate to `make check`, shared by local validation and CI. The threshold and measured package scope are unchanged.

中文：补齐账户查询、凭据隔离、RFQ 消息校验和公开研究示例的离线测试。示例在输出失败时返回错误并停止后续读取。原有全仓 80% 覆盖率门槛纳入 `make check`，本地与 CI 使用同一入口；门槛和统计范围不变。

## v0.3.0 — 2026-11-03

- Added Gamma lifecycle timestamps with original text and parse errors, market group indexes, resolution metadata, series/tag fields, and records of skipped nonbinary child markets.
- Added public HTTP response callbacks before decoding, including per-attempt errors, partial reads, bounded bodies, and independent retained bytes.
- Added `GetCLOBPriceHistory` and `CountEvents`; preserved resolution price JSON and removed floating-point conversion from Data history timestamps.
- Changed amount filters from `*float64` to `*json.Number`. This requires caller source changes. See the [migration and API reference](reference/research-upgrade.md).

中文：补齐 Gamma 时间与关系字段，保留异常时间、结算原值和被跳过的非二元市场；新增公共 HTTP 原文回调、CLOB 价格历史与事件总数方法。金额筛选改为 `*json.Number`，调用方须修改旧类型参数。详见[升级说明](reference/research-upgrade.zh-CN.md)。

## v0.2.0 — 2026-10-31

- Added opt-in raw market messages and raw-only consumption, retaining complete unknown, malformed and binary messages without changing default typed delivery.
- Added connection and application-heartbeat observations and configurable message read limits; oversized messages report loss without delivering partial bytes.
- Preserved immutable retained bytes during typed array decoding and retained exponential backoff across repeated short connections.
- Documented callback concurrency, shutdown and byte ownership, client-local identifiers, and the observation scope; added regression coverage and canonical facade exports.

Validation is offline using synthetic local servers, fixed fixtures, race checks and CGO-free builds. This release adds no real-account or live-market acceptance claim.

中文：新增可选的市场原始消息回调和仅接收原始消息模式，保留未知、畸形及二进制消息；增加连接／心跳观测和读取上限，超限不交付截断字节。修复数组类型解析改写保留字节，以及连续短连接失败重连过快的问题。补齐回调并发、关闭、字节归属和客户端内标识约定及回归测试，默认类型接口保持兼容。本次验证使用合成服务、固定样本、race 与无 CGO 构建，不新增真实账户或实网行情验收结论。

## v0.1.1 — 2026-10-29

- Made RFQ signers and transaction-handle dependencies implementable through public interfaces; restored missing notification constants.
- Bounded HTTP response sizes, validated endpoint URLs, reused long-request connections, and added RFQ Ping/Pong liveness checks.
- Honored Retry-After with finite-duration checks, bounded settlement requests by their timeout, and added polling backoff.
- Coordinated metadata loads by key and indexed stream dispatch while preserving overlapping-selection delivery semantics. Dispatch benchmarks record both improvements and regressions; no universal speedup is claimed.
- Consolidated pinned contract addresses and position parsing; rejected malformed decoder/query inputs and distinguished unsupported stream events from malformed messages.
- Added generated-export verification, regression tests, and GitHub Actions checks with an 80% coverage floor and six CGO-free build targets.

The [audit resolution](tests/acceptance/code-audit-2026-10-29.md) records the two-round review consensus and remaining non-blocking improvements. This patch preserves the pinned Python 0.14.0 contract and the v0.1.0 support scope. Validation is offline; no additional real-account or funded acceptance is claimed.

中文：本次修复公开接口可实现性和缺失的通知常量，增加 HTTP 响应上限、端点校验、长请求连接复用及 RFQ Ping/Pong；重试尊重 Retry-After 并防止延迟溢出，结算请求受总超时约束且采用退避。元数据按 key 协调加载，流分发使用索引并保留重叠 selection 语义，基准同时记录收益和退化，不宣称普遍加速。统一固定合约地址及 position 解析，明确非法查询／解码和未知流事件的错误处理，补充生成物门禁、回归测试及六目标无 CGO 构建、80% 覆盖率下限的 CI。

两轮复核结论及非阻断改进见[审查处理清单](tests/acceptance/code-audit-2026-10-29.zh-CN.md)。保持固定 Python 0.14.0 契约和 v0.1.0 支持范围；本次为离线验证，不新增真实账户或资金验收结论。

## v0.1.0 — 2026-10-29

Reference: `polymarket-client 0.14.0`, commit
`7032ed06175b0bbea77b2d9e898bd9f252af2041`. Go versions are independent of Python versions.

- Public queries, pagination, realtime subscriptions, authenticated trading, wallet operations and RFQ sessions.
- Fixed Python comparisons, strict declaration checks, reproducible generation, platform builds and performance records.
- Corrected upstream negative fractional timestamps and rejected confirmed trades without transaction hashes.
- Wallet market lookups explicitly include V1 and V2, fixing V2 position operations against Gamma defaults while preserving closed-market filtering.
- Omitted deprecated duplicate fields and gasless no-op methods; experimental features and Python table conversions remain outside scope.

Release scope: new Deposit Wallet accounts, host macOS arm64 execution, and six
cross-compilation targets. Live validation covered ordinary CTF fills, wallet funding,
asset split/merge/redemption (including V2, negative-risk and combo positions), and
cancellation/heartbeat workflows. Test funds were recovered and temporary asset permissions cleared.

Unverified scenarios are accepted as non-blocking, not marked passed: collateral-return
execution (service rejected authorization; recovery used ordinary wallet merge), RFQ
real settlement, remaining market/order variants, Builder attribution and independent
Builder-key invalidation, notification deletion, stock TWAP updates, and missing historical
responses. Session-key live validation remains paused. See tests/acceptance/README.md.

中文：本次发布面向新 Deposit Wallet 账户，验收覆盖 macOS arm64 本机运行与六平台编译。
普通 CTF 成交、钱包资金、CTF／V2／负风险／组合资产及撤单／心跳流程已完成相关实测，
测试资金已回收，临时资产授权已清理。资金回收服务提交被拒绝，实际通过普通钱包合并回收；
该服务执行、RFQ 真实结算、其余市场／订单变体、Builder 归因及独立失效探测、通知删除、
股票 TWAP 更新、缺失历史响应仍未验证，本次不阻塞。会话密钥实测继续暂挂。
