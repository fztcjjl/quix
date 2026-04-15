## ADDED Requirements

### Requirement: otelgin middleware injected into Gin default middleware chain
当 `WithTelemetry` 启用时，otelgin middleware SHALL 被注入到 Gin 默认中间件链中。中间件顺序 SHALL 为：Recovery → otelgin → RequestID → Logging → ResponseMiddleware。

#### Scenario: Default middleware chain with telemetry
- **WHEN** App 启用 WithTelemetry 且 defaultMiddleware 为 true
- **THEN** Gin Engine 的中间件链为 Recovery → otelgin → RequestID → Logging → ResponseMiddleware

#### Scenario: Default middleware chain without telemetry
- **WHEN** App 未启用 WithTelemetry 且 defaultMiddleware 为 true
- **THEN** Gin Engine 的中间件链为 Recovery → RequestID → Logging → ResponseMiddleware（与当前行为一致）

### Requirement: otelgin uses configured service name
otelgin middleware SHALL 使用 `telemetry.Config.ServiceName` 作为 middleware 名称参数。

#### Scenario: Service name propagation
- **WHEN** 调用 `WithTelemetry(telemetry.WithServiceName("myapp"))`
- **THEN** otelgin.Middleware("myapp") 被注入到中间件链

### Requirement: otelgin automatically creates spans and HTTP metrics
otelgin middleware SHALL 为每个 HTTP 请求自动创建 root span，并自动产出以下 OTel metrics：
- `http.server.request.duration`（histogram）
- `http.server.active_requests`（up-down counter）

#### Scenario: Request trace span
- **WHEN** 发送 HTTP 请求到任意注册路由
- **THEN** otelgin 创建 root span，span name 为 HTTP method + route pattern，span 包含 HTTP attributes（method、status_code、url、user_agent 等）

#### Scenario: HTTP metrics
- **WHEN** 发送 HTTP 请求
- **THEN** otelgin 自动记录 request duration 和 active requests metric
