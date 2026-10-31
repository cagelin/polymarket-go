# Repository Guidelines / 仓库开发规范

These guidelines apply to source code, tests, examples, and generated files.
本规范适用于源码、测试、示例及生成文件。

## Go declarations / Go 声明

- New declarations must not shadow enclosing variables, parameters, package-level names, imported package names, or Go predeclared identifiers.
  新声明不得遮蔽外层变量、参数、包级名称、导入包名或 Go 预声明标识符。
- Name conflicting identifiers by purpose, such as `requestPath`, `orderResponse`, or `subtest`. Valid reuse within the same scope is allowed. Do not make callbacks mutate outer variables merely to avoid a warning.
  同名时按用途命名，例如 `requestPath`、`orderResponse`、`subtest`；允许同一作用域内合法复用变量，不得为消除告警让回调修改外层变量。
- Do not bypass shadowing checks with `nolint` or IDE suppression. Fix generators together with generated output.
  不得使用 `nolint` 或 IDE 忽略设置绕过遮蔽检查；生成代码有问题时，同步修正生成器。

## Quality checks / 质量检查

- Run `make check` before committing. It checks formatting, static analysis, automated tests, and race detection.
  提交前运行 `make check`，检查格式、静态分析、自动测试和 race。
- Resolve problems reported by GoLand inspections before committing, including errors and warnings. Passing `make check` does not replace this requirement. Do not bypass problems by disabling inspections or adding suppression rules.
  提交前必须解决 GoLand 检查识别出的问题，包括错误和警告。`make check` 通过不能替代此项要求，不得通过关闭检查或添加忽略规则规避问题。
- `make lint` uses golangci-lint's `govet shadow` with `strict: true` and `predeclared`. Do not maintain a separate shadowing checker.
  `make lint` 使用 golangci-lint 的 `govet shadow`（`strict: true`）与 `predeclared`，不维护自写遮蔽检查工具。
- Automated checks do not detect every name conflict. Review callback parameters and variables of different types against the same declaration rules.
  自动检查不能覆盖全部同名声明；回调参数及不同类型变量仍须按声明规范审查。
- Keep the SDK usable without CGO. Development checks must not add SDK runtime dependencies.
  SDK 默认无需 CGO；开发检查不得增加 SDK 运行依赖。

## Repository layout / 目录约定

- Keep automated SDK tests in their packages or under `tests/`. They must not depend on `.local/` programs, accounts, or files.
  SDK 自动测试放在对应包或 `tests/`，不得依赖 `.local/` 的程序、账户或文件。
- Keep real-account and manual funded-validation tools under ignored `.local/`, together with their dedicated tests, configuration, data, and execution records. Do not place them in `tests/` or public `scripts/`.
  真实账户或人工资金验证程序及其专用测试、配置、数据和运行记录放在忽略的 `.local/`，不放入 `tests/` 或公共 `scripts/`。
- `make check` covers the SDK. Run offline tests for local validation tools separately and explicitly. Moving tools must preserve historical records, attempt locks, and account configuration.
  `make check` 检查 SDK；本机验证程序的离线测试单独显式运行。移动程序时保留历史记录、尝试锁及账户配置。

## Documentation and changes / 文档与变更

- Write code comments in English and maintain user documentation in English and Chinese.
  代码注释使用英文，用户文档提供中英版本。
- Keep changes focused on the task. Summarize the result and verification concisely, including any incomplete work.
  修改保持必要范围；简洁说明结果、验证情况及未完成事项。

See [README.md](README.md) for setup, [parity testing](tests/parity/README.md) for
fixture generation, and [release validation](tests/acceptance/README.md) for acceptance procedures.
环境配置见 [README.zh-CN.md](README.zh-CN.md)，数据生成见[对照测试说明](tests/parity/README.md)，
验收流程见[发布验收说明](tests/acceptance/README.md)。
