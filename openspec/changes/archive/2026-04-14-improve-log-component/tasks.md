## 1. Logger 接口扩展与全局默认重构

- [x] 1.1 重命名 `core/log/log.go` → `core/log/logger.go`，更新文件头包声明
- [x] 1.2 定义 `Level` 类型和级别常量（`LevelDebug`、`LevelInfo`、`LevelWarn`、`LevelError`）
- [x] 1.3 Logger 接口新增 `Fatal`、`SetLevel`、`Close` 三个方法
- [x] 1.4 删除 `noopLogger`，全局默认改为 `NewSlog()`
- [x] 1.5 全局 `defaultLogger` 改用 `atomic.Pointer[Logger]` 存储
- [x] 1.6 更新包级函数（Info/Error/Warn/Debug/Fatal/With/SetLevel），从 `atomic.Pointer[Logger]` 加载
- [x] 1.7 新增 `core/log/logger_test.go`：测试 SetDefault、全局函数、并发安全

## 2. Adapter 统一参数处理

- [x] 2.1 提取 `normalizeArgs` 统一函数（快路径零分配），删除 `toSlogArgs`、`toZapFields`、`toZerologFields`
- [x] 2.2 `slog.go`、`zap.go`、`zerolog.go`、`writer.go` 改用 `normalizeArgs`
- [x] 2.3 新增 `argsToMap` 辅助函数供 zerolog adapter 使用
- [x] 2.4 更新 `slog_test.go`、`zap_test.go`、`zerolog_test.go`：增加非字符串 key 多个场景的测试用例

## 3. Adapter 实现接口扩展

- [x] 3.1 `slog.go`：实现 `Fatal`、`SetLevel`、`Close` 方法
- [x] 3.2 `zap.go`：实现 `Fatal`、`SetLevel`、`Close` 方法（`Close` 调用 `Sync()`）
- [x] 3.3 `zerolog.go`：实现 `Fatal`、`SetLevel`、`Close` 方法
- [x] 3.4 更新三个 adapter 的测试覆盖新方法

## 4. NewWriter 适配器

- [x] 4.1 新增 `core/log/writer.go`：实现 `NewWriter(w io.Writer) Logger`，基于 `slog.NewJSONHandler`
- [x] 4.2 新增 `core/log/writer_test.go`：测试 NewWriter 基本功能

## 5. App 层集成

- [x] 5.1 `option.go`：`WithLogger` 增加 `log.SetDefault(l)` 调用，同步更新全局默认
- [x] 5.2 `quix.go`：确认 `quix.New()` 默认使用 zerolog Logger 并 `log.SetDefault`
- [x] 5.3 更新 `quix_test.go`：使用真实 Logger 实现（`log.NewSlog()`）

## 6. MockLogger 内部化

- [x] 6.1 删除 `core/log/mock.go`，`MockLogger` 不暴露到生产 API
- [x] 6.2 `quix_test.go` 使用真实 Logger 实现（`log.NewSlog()`）替代 MockLogger
- [x] 6.3 `logging_test.go`、`recovery_test.go` 在测试文件内局部定义 mock

## 7. Logging 中间件增强

- [x] 7.1 新增 `LoggingOption` 类型和 `WithSkipPaths`、`WithHook` 选项函数
- [x] 7.2 新增 `LoggingHookFunc` 类型定义
- [x] 7.3 实现 `LoggingWith(opts ...LoggingOption) gin.HandlerFunc`，支持前缀匹配和 hook
- [x] 7.4 更新原有 `Logging()` 函数，内部委托给 `LoggingWith`
- [x] 7.5 更新 `logging_test.go`：增加前缀匹配和 hook 测试用例

## 8. 文档与收尾

- [x] 8.1 更新 CLAUDE.md：`core/log` 默认 Slog、`quix` 默认 Zerolog 的描述
- [x] 8.2 更新项目架构说明中的文件命名（`logger.go`）
- [x] 8.3 执行 `go fmt ./...`
- [x] 8.4 执行 `golangci-lint run ./...` 修复 lint 问题
- [x] 8.5 执行 `go test ./...` 确认所有测试通过
