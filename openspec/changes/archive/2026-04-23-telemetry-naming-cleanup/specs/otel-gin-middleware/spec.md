## MODIFIED Requirements

### Requirement: otelgin middleware injected into Gin default middleware chain
当 `WithTelemetry` 启用时，otelgin middleware SHALL 被注入到 Gin 默认中间件链中。otelgin SHALL 由 `quix.New()` 在 server 创建后直接挂载到 engine，不通过 server options 传递。中间件顺序 SHALL 为：RequestID → otelgin → RequestLogger → Recovery → CORS → Logging → ResponseMiddleware。

#### Scenario: Default middleware chain with telemetry
- **WHEN** App 启用 WithTelemetry 且 defaultMiddleware 为 true
- **THEN** Gin Engine 的中间件链为 RequestID → otelgin → RequestLogger → Recovery → CORS → AccessLog → ResponseMiddleware

#### Scenario: Default middleware chain without telemetry
- **WHEN** App 未启用 WithTelemetry 且 defaultMiddleware 为 true
- **THEN** Gin Engine 的中间件链为 RequestID → RequestLogger → Recovery → CORS → AccessLog → ResponseMiddleware（无 otelgin）

## REMOVED Requirements

### Requirement: otelgin uses configured service name
**Reason**: otelgin 现由 quix.New() 直接挂载，不再通过 server options 传递 service name。quix.New() 已持有 telemetry.Config，直接使用 telCfg.ServiceName。
**Migration**: otelgin 由 quix.New() 自动管理，无需通过 qhttp.WithTelemetryServiceName 配置。
