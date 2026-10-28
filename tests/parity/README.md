# Python parity tests / Python 对照测试

Fixtures execute the official Python SDK's request builders, response parsers,
market-price calculation, and HTTP error handling. Go tests make requests to
a local HTTP server and compare requests, decoded values, and error classes.
Error text is not compared word for word. This covers the implemented CLOB
subset, not the entire SDK.

Python-only deprecated input aliases are outside the Go API. Expected model
output omits the duplicate `market` and `token_id` fields; canonical fields
are still compared. Exchange wire aliases remain covered by response tests.

测试数据由官方 Python SDK 实际执行生成。Go 测试通过本地 HTTP 服务对照请求、
解析结果和错误类别。当前只覆盖已实现的价格与订单簿部分。

Regenerate with Python 3.12+ / 重新生成：

```sh
python3 -m venv .venv-parity
.venv-parity/bin/python -m pip install -r tests/parity/requirements.txt
.venv-parity/bin/python tests/parity/generate.py
go test ./...
```

On Windows use `.venv-parity\Scripts\python.exe` instead.

The SDK archive is pinned by SHA-256. Dependencies are version-pinned.
Before generating fixtures, the script verifies all 202 installed upstream
source files against `tests/parity/upstream.json`. A changed source fails
generation. Generated JSON files are committed inputs for ordinary Go tests;
Python is not required to run them.

SDK 压缩包固定 SHA-256，依赖固定版本。生成前检查 202 个上游源码文件的哈希。
普通 Go 测试直接读取生成的 JSON，不需要安装 Python。

Go 不提供 Python 的废弃输入别名；输出对照剔除重复的旧字段，但仍比较当前字段。
服务端实际使用的旧字段名仍有解析测试。测试不依赖被 Git 忽略的 `docs/` 目录。

Live public reads / 真实公开接口查询：

```sh
POLYMARKET_LIVE_ASSET_ID=<asset-id> go test -run '^TestLiveCLOB$' -v
```

This performs public reads only. Market data changes, so live checks establish
connectivity and response compatibility; fixed fixtures establish deterministic
Python/Go equivalence. No private keys, account credentials, or funds are used.

真实查询检查连通性和响应兼容；固定数据负责逐项比较。全程不使用私钥、账户或资金。
