## Why

quix 框架当前缺乏可观测性能力。生产环境中的微服务需要 Traces（调用链追踪）、Metrics（指标监控）、Logs（日志关联）三支柱来定位性能瓶颈、排查故障、监控系统健康状态。OpenTelemetry 是业界标准，提供统一的 API/SDK 和 OTLP 协议对接各种后端（Jaeger、Tempo、Prometheus 等）。

## What Changes

- 新增 `core/telemetry/` 组件，统一管理 TracerProvider、MeterProvider 生命周期，提供 Config + Option 模式配置
- 新增 `WithTelemetry` Option 集成到 `quix.App`，App.Shutdown 中按序 flush 所有 provider
- 注入 `otelgin.Middleware` 到 Gin 默认中间件链，自动提供每请求 Traces + 基础 HTTP Metrics
- 增强 Logging middleware：从 gin.Context 读取 otelgin 注入的 trace_id 写入日志，实现应用日志与调用链关联（日志采集由外部组件 Promtail/Filebeat 负责，不做 OTel Logs 桥接）
- 提供 stdout exporter 作为开发调试选项（零外部依赖）

## Capabilities

### New Capabilities
- `telemetry-provider`: OTel Provider 生命周期管理（TracerProvider、MeterProvider）、Config/Option 配置、统一 Init/Shutdown
- `app-telemetry`: WithTelemetry Option 集成到 quix.App，Shutdown 顺序保证
- `otel-gin-middleware`: otelgin 中间件注入 Gin 中间件链，自动 Traces + HTTP Metrics
- `logging-middleware-trace-id`: Logging middleware 输出 trace_id，关联应用日志与调用链

### Modified Capabilities
- `logging-middleware-enhancement`: 新增 trace_id 字段输出（从 gin.Context 读取 otelgin 注入的 trace_id）

## Impact

- **新增依赖**: go.opentelemetry.io/otel 全家桶（otel、sdk、exporters/otlp、contrib/gin/otelgin）
- **新增代码**: `core/telemetry/` 包、`core/telemetry/telemetry.go`、`core/telemetry/telemetry_test.go`
- **修改代码**: `option.go`（新增 WithTelemetry）、`quix.go`（Shutdown 增加 telemetry flush）、`core/transport/http/server/middleware/logging.go`（trace_id 输出）
- **不改变现有 Logger interface**: 日志采集由外部组件（Promtail/Filebeat）负责，quix 不做 OTel Logs 桥接
- **默认关闭**: 不调用 WithTelemetry() 则零 OTel 依赖启动
