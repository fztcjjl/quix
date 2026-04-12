## Why

当前 Logger 需要通过 `app.Logger().Info(...)` 获取实例后使用，使用不够便捷。开发者习惯开箱即用的全局日志函数（如 `slog.Info()`、`logrus.Info()`）。此外包名 `logger` 较冗长，`log.Info` 比 `logger.Info` 更简洁。

## What Changes

- **BREAKING** 包名重命名：`core/logger/` → `core/log/`，`package logger` → `package log`
- **BREAKING** 所有 import 路径从 `core/logger` → `core/log`
- **BREAKING** 示例目录 `examples/logger/` → `examples/log/`
- 新增全局默认 Logger 实例（初始为 noopLogger，避免 nil panic）
- 新增包级函数 `log.Info/Error/Warn/Debug/With`，委托给全局默认实例
- 新增 `log.SetDefault(l Logger)` 设置全局默认
- `App.New()` 创建 Logger 后自动调用 `log.SetDefault()`
- 更新现有 logger spec：默认实现从 slog 改为 zerolog，新增全局默认相关需求

## Capabilities

### New Capabilities

### Modified Capabilities
- `logger`: 包名变更、新增全局默认实例和包级函数、默认实现从 slog 改为 zerolog、示例目录变更

## Impact

- `core/logger/` 目录重命名为 `core/log/`（含所有适配器文件）
- `quix.go`、`option.go`：import 路径更新
- `quix_test.go`：import 路径和 mock 定义更新
- `examples/logger/` → `examples/log/`
- `openspec/specs/logger/spec.md`：需求更新
- `CLAUDE.md`：相关描述更新
