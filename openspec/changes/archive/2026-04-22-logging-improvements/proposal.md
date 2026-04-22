## Why

日志系统存在实际缺陷（level 数据竞争、zerolog adapter 每 log 分配 map）和关键能力缺失（无 context 感知日志、无 span_id、访问日志字段不完整）。这些问题直接影响生产环境下的可观测性、性能和开发体验。

## What Changes

- **修复 level 数据竞争**：4 个 adapter 的 `level Level` 改为 `atomic.Int32`，通过导出 `AtomicLevel` 消除重复代码
- **添加 Trace 级别**：Logger 接口新增 `Trace()` 方法，常量 `LevelTrace`
- **增强访问日志**：新增 `latency_ms`（数值）、`query`、`user_agent`、`route`（归一化路径）、`span_id`、`error_code` 字段
- **慢请求检测**：新增 `WithSlowThreshold()` 选项
- **Context 感知日志**：新增 `IntoContext/FromContext` 和 `RequestLogger` 中间件，handler 自动携带 trace_id/span_id/request_id
- **开发环境 Caller 输出**：开发环境默认 zerolog logger 启用 `caller` 字段（通过 `CallerWithSkipFrameCount` 跳过框架内部帧），生产环境不加
- **修复 zerolog 性能**：用类型分发替代 `argsToMap()` 的 map 分配
- **Level 字符串化**：新增 `String()` 和 `ParseLevel()` 支持配置驱动级别控制

## Capabilities

### New Capabilities
- `trace-log-level`: Trace 日志级别定义与 Logger 接口扩展
- `context-aware-logging`: Logger 的 IntoContext/FromContext 上下文传播与 RequestLogger 中间件
- `request-logger-middleware`: WithRequestLogger 中间件

### Modified Capabilities
- `logger`: Logger 接口新增 Trace 方法、导出 AtomicLevel、Level String/ParseLevel、zerolog adapter 性能修复、开发环境默认 Caller 输出
- `log-level-control`: level 常量扩展 LevelTrace、并发安全改为 atomic.Int32
- `logging-middleware-enhancement`: 访问日志新增 latency_ms/query/user_agent/route/span_id/error_code 字段、慢请求检测
- `logging-middleware-trace-id`: 扩展 span_id 提取与日志输出

## Impact

- **BREAKING**: Logger 接口新增 `Trace()` 方法，外部实现需同步
- `core/log/` 包：logger.go、slog.go、zerolog.go、zap.go、writer.go 均有修改
- `core/transport/http/server/middleware/logging.go`：访问日志字段扩展
- `core/telemetry/telemetry.go`：新增 `ExtractSpanID()`
- `core/transport/http/server/server.go`：默认中间件链插入 RequestLogger
- 开发环境默认 logger 启用 `caller` 字段（`quix.go`）
- `quix.go`：设置 ExtractSpanID
- 无新增外部依赖
