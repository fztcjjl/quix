## MODIFIED Requirements

### Requirement: WithTelemetry Option initializes telemetry
`quix.WithTelemetry(opts ...telemetry.Option)` SHALL 在 App 创建时调用 `telemetry.Init`，并将返回的 shutdown func 存储到 App 结构体中。App SHALL NOT 存储 `telemetryServiceName` / `telemetryTracesEnabled` 字段，这些信息仅保留在 `telemetry.Config` 中。

#### Scenario: App with telemetry enabled
- **WHEN** 调用 `quix.New(quix.WithTelemetry(telemetry.WithServiceName("myapp")))`
- **THEN** telemetry.Init 被调用，全局 OTel Provider 已设置，App 存储了 shutdown func，otelgin middleware 被挂载到 HTTP engine

#### Scenario: App without telemetry
- **WHEN** 调用 `quix.New()` 不传 WithTelemetry
- **THEN** 不调用 telemetry.Init，无 OTel Provider 创建，otelgin 不挂载

## ADDED Requirements

### Requirement: quix.New() directly mounts otelgin middleware
当 telemetry 初始化成功且 `TracesEnabled` 为 true 时，`quix.New()` SHALL 在创建 HTTP server 后直接调用 `app.httpServer.Use(otelgin.Middleware(telCfg.ServiceName))` 挂载 otelgin middleware，不通过 server options 传递遥测配置。

#### Scenario: otelgin mounted after server creation
- **WHEN** `quix.New(quix.WithTelemetry(telemetry.WithServiceName("myapp")))` 创建 App 且 telemetry 初始化成功
- **THEN** otelgin.Middleware("myapp") 被追加到 HTTP engine 的中间件链

#### Scenario: otelgin not mounted when traces disabled
- **WHEN** `quix.New(quix.WithTelemetry(telemetry.WithTracesEnabled(false)))` 创建 App
- **THEN** otelgin middleware 不挂载到 HTTP engine
