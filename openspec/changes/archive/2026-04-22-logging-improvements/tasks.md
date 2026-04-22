## 1. AtomicLevel 与并发安全

- [x] 1.1 在 `core/log/logger.go` 中定义导出的 `AtomicLevel` struct（含 `atomic.Int32`）和 `Enabled()`/`SetLevel()`/`Level()` 方法
- [x] 1.2 修改 `slog.go` — 持有 `*AtomicLevel`，删除独立 `level Level` 字段
- [x] 1.3 修改 `zerolog.go` — 持有 `*AtomicLevel`，同上
- [x] 1.4 修改 `zap.go` — 持有 `*AtomicLevel`，同上
- [x] 1.5 修改 `writer.go` — 持有 `*AtomicLevel`，同上
- [x] 1.6 运行 `go test ./core/log/...` 确认现有测试通过
- [x] 1.7 运行 `go test -race ./core/log/...` 确认无 data race

## 2. Level 扩展

- [x] 2.1 在 `core/log/logger.go` 中添加 `LevelTrace Level = -1` 常量，调整现有常量值（LevelDebug=0, LevelInfo=1, LevelWarn=2, LevelError=3）
- [x] 2.2 在 Logger 接口中添加 `Trace(ctx context.Context, msg string, args ...any)` 方法
- [x] 2.3 在 `core/log/logger.go` 中添加包级 `Trace()` 函数
- [x] 2.4 在 `slog.go` 中实现 `Trace()`，映射为 `slog.Level(-8)`
- [x] 2.5 在 `zerolog.go` 中实现 `Trace()`，使用 `zerolog.TraceLevel`
- [x] 2.6 在 `zap.go` 中实现 `Trace()`
- [x] 2.7 在 `writer.go` 中实现 `Trace()`
- [x] 2.8 为 `Level` 添加 `String() string` 方法和 `ParseLevel(s string) (Level, error)` 函数
- [x] 2.9 编写 Trace 级别和 ParseLevel 的单元测试
- [x] 2.10 运行 `go build ./...` 和 `golangci-lint run ./...`

## 3. zerolog adapter 性能修复

- [x] 3.1 在 `zerolog.go` 中用类型分发函数 `addFieldsToEvent(e *zerolog.Event, args []any)` 替代 `argsToMap()`
- [x] 3.2 覆盖常见类型：string、error、int、int64、float64、bool、time.Duration，default 走 `.Interface()`
- [x] 3.3 删除 `argsToMap` 函数
- [x] 3.4 运行 `go test ./core/log/...` 确认测试通过

## 4. Context 感知日志

- [x] 4.1 在 `core/log/logger.go` 中添加 `contextKey` 类型和 `NewContext()`/`FromContext()` 函数
- [x] 4.2 编写 NewContext/FromContext 的单元测试

## 5. 访问日志中间件增强

- [x] 5.1 在 `logging.go` 中新增 `latency_ms` 数值字段（保留原 `latency` 字符串字段）
- [x] 5.2 新增 `query` 字段（非空时输出）
- [x] 5.3 新增 `user_agent` 字段（非空时输出）
- [x] 5.4 新增 `route` 字段（`c.FullPath()` 非 nil 时输出）
- [x] 5.5 新增 `error_code` 字段（从 gin context 的 `app_error` 提取 `*qerrors.Error.Code`）
- [x] 5.6 新增 `WithSlowThreshold(d time.Duration) LoggingOption` 和慢请求检测逻辑
- [x] 5.7 编写新字段和慢请求检测的单元测试
- [x] 5.8 运行 `go test ./core/transport/http/server/middleware/...` 和 `golangci-lint run ./...`

## 6. span_id 提取

- [x] 6.1 在 `core/telemetry/telemetry.go` 中添加 `ExtractSpanID()` 函数
- [x] 6.2 在 `middleware/logging.go` 中添加 `ExtractSpanID func(ctx context.Context) string` 变量
- [x] 6.3 在访问日志中使用 `ExtractSpanID` 输出 `span_id` 字段
- [x] 6.4 在 `quix.go` 中设置 `middleware.ExtractSpanID = telemetry.ExtractSpanID`
- [x] 6.5 编写 ExtractSpanID 的单元测试
- [x] 6.6 运行 `go test ./core/telemetry/...` 确认通过

## 7. WithRequestLogger 中间件

- [x] 7.1 在 `middleware/logging.go` 中实现 `WithRequestLogger(opts ...WithRequestLoggerOption) gin.HandlerFunc`
- [x] 7.2 在 `core/transport/http/server/server.go` 默认中间件链中插入 WithRequestLogger（位于 otelgin 之后、Recovery 之前）
- [x] 7.3 编写 WithRequestLogger 的集成测试（验证 context 中注入的 logger 携带 trace_id/span_id/request_id）
- [x] 7.4 运行 `go test ./core/transport/http/server/...` 和 `golangci-lint run ./...`

## 8. 收尾

- [x] 8.1 运行 `go test ./...` 全量测试
- [x] 8.2 运行 `golangci-lint run ./...` 全量 lint
- [x] 8.3 运行 `go fmt ./...` 格式化
- [x] 8.4 更新 `core/log/` 下的示例代码（如需要）

## 9. 开发环境 Caller 输出

- [x] 9.1 在 `quix.go` 中，当 `env == EnvDev` 时为默认 zerolog logger 添加 `CallerWithSkipFrameCount(4)` 配置
- [x] 9.2 编写 `TestZerologTimestampField` 测试，验证 `time` 字段存在于 JSON 输出
- [x] 9.3 编写 `TestZerologCallerField` 测试，验证 `caller` 字段存在且指向调用方文件（通过包级 `Info()` 调用）
- [x] 9.4 运行 `go test ./core/log/...` 和 `golangci-lint run ./...`
