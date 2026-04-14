## Why

日志组件存在一致性问题（三个 adapter 对非字符串 key 和奇数参数的处理行为不同）、代码缺陷（参数预处理逻辑分散在三个 adapter 中重复实现）、并发安全隐患（全局变量无保护）、以及功能缺失（缺少级别控制、Fatal 级别、生命周期管理）。此外默认 logger 策略调整：`core/log` 包全局默认从 noopLogger 改为 slog（零依赖、开箱即用），`quix` 应用框架层保持 zerolog 为默认（开发体验好）。

## What Changes

- **BREAKING**: Logger 接口新增 3 个方法（Fatal、SetLevel、Close），所有实现和 mock 需同步更新
- 统一三个 adapter（slog/zap/zerolog）对非字符串 key 和奇数参数的处理策略
- 提取 `normalizeArgs` 统一参数预处理逻辑（快路径零分配）
- `WithLogger` option 同步调用 `log.SetDefault()`，解决中间件全局函数与 App 注入 logger 不一致
- `defaultLogger` 改用 `atomic.Pointer[Logger]` 保证并发安全
- Logger 接口新增 `SetLevel` 方法和 `Fatal` 日志级别
- Logger 接口新增 `Close` 方法用于资源清理
- 重命名 `log.go` → `logger.go`，新增 `logger_test.go`
- 删除各 adapter 重复的参数转换函数（`toSlogArgs`、`toZapFields`、`toZerologFields`），统一使用 `normalizeArgs`
- 删除 `noopLogger`，全局默认改为 `NewSlog()`；删除 `mock.go`，`MockLogger` 只在各包 `_test.go` 内局部定义
- Logging 中间件支持前缀匹配 skipPaths 和自定义 hook
- 新增 `NewWriter(io.Writer)` 适配器

## Capabilities

### New Capabilities

- `log-level-control`: Logger 接口级别控制能力（SetLevel）和 Fatal 级别
- `log-lifecycle`: Logger 接口生命周期管理（Close/Sync）
- `log-writer-adapter`: 基于	io.Writer 的简单 Logger 适配器
- `logging-middleware-enhancement`: Logging 中间件增强（前缀匹配、自定义 hook）

### Modified Capabilities

- `logger`: 统一参数处理策略、并发安全、默认值从 noopLogger 改为 slog、接口扩展（Fatal、SetLevel、Close）、文件重命名

## Impact

- **接口变更**: `core/log.Logger` 接口新增 3 个方法（Fatal、SetLevel、Close），所有实现和 mock 需同步更新
- **行为变更**: 全局默认 Logger 从静默 noop 改为 slog 输出；`WithLogger` 会同步更新全局默认
- **文件变更**: `log.go` → `logger.go`，新增 `logger_test.go`、`writer.go`
- **依赖变更**: `core/log` 包的默认路径零外部依赖（仅 stdlib slog）
- **受影响代码**: `quix.go`（默认 logger 创建）、`option.go`（WithLogger）、中间件（logging.go、recovery.go）、所有 adapter 实现、测试文件
