## ADDED Requirements

### Requirement: WithTelemetry Option initializes telemetry
`quix.WithTelemetry(opts ...telemetry.Option)` SHALL 在 App 创建时调用 `telemetry.Init`，并将返回的 shutdown func 存储到 App 结构体中。

#### Scenario: App with telemetry enabled
- **WHEN** 调用 `quix.New(quix.WithTelemetry(telemetry.WithServiceName("myapp")))`
- **THEN** telemetry.Init 被调用，全局 OTel Provider 已设置，App 存储了 shutdown func

#### Scenario: App without telemetry
- **WHEN** 调用 `quix.New()` 不传 WithTelemetry
- **THEN** 不调用 telemetry.Init，无 OTel Provider 创建，无额外依赖加载

### Requirement: App Shutdown flushes telemetry
`App.Shutdown(ctx)` SHALL 在停止所有 server 之后调用 telemetry shutdown func（如存在），确保 OTel 数据 flush。

#### Scenario: Shutdown order
- **WHEN** 调用 `app.Shutdown(ctx)` 且 WithTelemetry 已启用
- **THEN** 按序执行：RPC server stop → HTTP server stop → telemetry shutdown → logger close

#### Scenario: Shutdown without telemetry
- **WHEN** 调用 `app.Shutdown(ctx)` 且未启用 WithTelemetry
- **THEN** 仅执行 server stop，不调用 telemetry shutdown
