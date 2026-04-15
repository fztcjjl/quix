## Context

quix 是基于 Gin 的 Go 快速开发框架，当前缺乏可观测性能力。生产环境需要 Traces/Metrics/Logs 三支柱来监控系统健康状态和排查问题。OpenTelemetry 是业界标准，提供统一的 API/SDK 和 OTLP 协议。

当前状态：
- `core/log/` 提供 Logger 统一接口，slog/zerolog/zap 适配器
- `core/transport/http/server/middleware/logging.go` 提供 HTTP 访问日志
- `core/transport/http/server/server.go` 中间件链：Recovery → RequestID → Logging → ResponseMiddleware
- `quix.App` 通过 Option 模式配置，Shutdown 按序停止 RPC → HTTP

约束：
- 不改变现有 Logger interface
- 默认关闭 OTel，WithTelemetry() 不传则零 OTel 依赖启动
- 符合 quix 的 Option 模式和 core/<component>/ 目录结构

## Goals / Non-Goals

**Goals:**
- 提供 `core/telemetry/` 组件统一管理 OTel Provider 生命周期
- 通过 `WithTelemetry` Option 集成到 quix.App，自动配置 Traces + Metrics
- otelgin 中间件自动提供每请求 span 和 HTTP 基础指标
- Logging middleware 输出 trace_id 关联应用日志与调用链（日志采集由外部组件负责）
- 提供 stdout exporter 方便本地开发调试

**Non-Goals:**
- 不做 OTel Logs 桥接（日志采集由外部组件 Promtail/Filebeat 负责，Traces/Metrics 走 OTLP，Logs 走 stdout）
- 不提供业务指标 API（用户直接使用 OTel SDK 的 Meter API）
- 不实现自定义 SpanProcessor 或 MetricReader
- 不提供 OTel Config 与 koanf 配置系统的集成（后续按需）
- 不实现 gRPC server 的 OTel instrumentation（暂无 RPC transport）

## Decisions

### 1. core/telemetry/ 包结构：单文件设计

**决策**: `core/telemetry/telemetry.go` 单文件包含 Config、Option、Init、Shutdown。

**替代方案**: 按支柱拆分文件（traces.go、metrics.go、logs.go）。

**选择理由**: Init 是唯一入口，三个 Provider 的创建逻辑耦合（共享 Resource、共享 Exporter Endpoint），拆分后反而增加跨文件引用复杂度。单文件更符合 quix 小包的风格。

### 2. otelgin 而非 otelhttp

**决策**: 使用 `go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin`。

**替代方案**: 使用 `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`。

**选择理由**: quix 基于 Gin，otelgin 直接作为 gin.HandlerFunc 注入中间件链，无需 HTTP Handler 包装。otelgin 自动创建 root span、传播 context、产出 HTTP metrics，与现有 middleware 链无缝集成。

### 3. otelgin 放置在中间件链中的位置：Recovery 之后、RequestID 之前

**决策**: 中间件链顺序：Recovery → otelgin → RequestID → Logging → ResponseMiddleware。

**替代方案**: otelgin 放在链的最前面。

**选择理由**: otelgin 必须在 Recovery 之后，确保 panic 不产生未结束的 span。otelgin 在 RequestID 之前，确保 span 能捕获 request_id attribute（otelgin 不依赖 request_id，但后续 Logging middleware 需要 trace_id）。otelgin 在 Logging 之前，确保 Logging 能从 context 读取 trace_id。

### 4. OTLP gRPC 为默认 exporter，stdout 为开发选项

**决策**: 默认 OTLP gRPC exporter（port 4317），`WithStdoutExporter()` 开关切换到 stdout。

**替代方案**: 仅提供 OTLP exporter，stdout 由用户自行配置。

**选择理由**: stdout exporter 零外部依赖（`go.opentelemetry.io/otel/exporters/stdout/stdouttrace`），本地开发调试非常方便，避免用户在开发阶段启动 Collector。

### 5. Logging middleware trace_id 通过 context 读取

**决策**: Logging middleware 从 `c.Request.Context()` 读取 OTel trace_id（由 otelgin 注入），不需要直接 import OTel 包。

**选择理由**: otelgin 将 trace context 注入到 `request.Context()`，Logging middleware 只需使用 `trace.SpanContextFromContext()` 即可读取，无需引入 OTel API 依赖到 middleware 包。通过函数变量（`extractTraceID`）解耦，保持 middleware 包不直接依赖 OTel。

### 6. Init 返回统一的 shutdown func

**决策**: `func Init(ctx context.Context, opts ...Option) (func(context.Context) error, error)` 返回单个 shutdown func，内部按序 flush。

**替代方案**: 返回 Telemetry struct 暴露 Shutdown() 方法。

**选择理由**: 与 `App.Shutdown` 的 `func(ctx) error` 签名一致，直接赋值即可。用户不直接操作 Provider，减少误用可能。

## Risks / Trade-offs

**[OTel SDK 体积大]** → OTel 全家桶 go.sum 增加约 100+ 依赖。Mitigation: 默认关闭，用户显式 `go get` 相关包。

**[otelgin 中间件性能开销]** → 每请求创建 span 有一定开销。Mitigation: 生产环境通过采样率控制（`WithSampler` option），默认使用 parent-based sampler。

**[不做 OTel Logs 桥接]** → Logs 支柱仅通过 Logging middleware 输出 trace_id 实现与调用链的关联，日志采集依赖外部组件。Mitigation: stdout + Promtail 是成熟方案，满足绝大部分场景。

**[shutdown 顺序]** → telemetry 必须在 server stop 之后 flush，否则丢失最后的数据。Mitigation: App.Shutdown 严格按照 RPC → HTTP → telemetry 顺序。

**[OTel SDK 版本更新]** → Go OTel SDK 迭代较快，API 可能变化。Mitigation: 锁定版本，使用 stable API（non-experimental）。
