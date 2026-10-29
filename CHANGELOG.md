# Changelog

Scope update (2026-10-29): this SDK targets new accounts using Deposit Wallets. Legacy Safe/Proxy onboarding, migration, and live validation are excluded from support and release acceptance. Existing legacy implementation and historical tests remain, without a support commitment. EOA signing and other agreed capabilities remain in scope.

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
