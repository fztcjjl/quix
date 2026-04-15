## ADDED Requirements

### Requirement: Init creates and registers OTel Providers
`telemetry.Init(ctx, opts...)` SHALL 创建 TracerProvider、MeterProvider（默认），并将其设置为全局默认（`otel.SetTracerProvider`、`otel.SetMeterProvider`）。Init SHALL 返回一个 `func(context.Context) error` 类型的 shutdown 函数，该函数按序 shutdown MeterProvider、TracerProvider。

#### Scenario: Init with default options
- **WHEN** 调用 `Init(ctx)` 不传任何 option
- **THEN** 创建 OTLP gRPC TracerProvider 和 MeterProvider，返回 shutdown func

#### Scenario: Init with TracesEnabled false
- **WHEN** 调用 `Init(ctx, WithTracesEnabled(false))`
- **THEN** 不创建 TracerProvider，仅创建 MeterProvider

#### Scenario: Init with MetricsEnabled false
- **WHEN** 调用 `Init(ctx, WithMetricsEnabled(false))`
- **THEN** 不创建 MeterProvider，仅创建 TracerProvider

#### Scenario: Shutdown flushes all providers in order
- **WHEN** 调用 Init 返回的 shutdown func
- **THEN** 按序 shutdown MeterProvider、TracerProvider

### Requirement: Config provides Option pattern for telemetry configuration
`telemetry` 包 SHALL 提供 `Config` 结构体和 `func(*Config)` 类型的 Option 函数，支持以下配置项：
- `ServiceName`：服务名称，默认 "unknown_service"
- `ServiceVersion`：服务版本号，默认 ""
- `ExporterEndpoint`：OTLP exporter 地址，默认 "localhost:4317"
- `ResourceAttributes`：额外 Resource attributes，默认空
- `TracesEnabled`：是否启用 Traces，默认 true
- `MetricsEnabled`：是否启用 Metrics，默认 true
- `StdoutExporter`：使用 stdout exporter（开发调试），默认 false

#### Scenario: WithServiceName sets service name
- **WHEN** 调用 `Init(ctx, WithServiceName("myapp"))`
- **THEN** TracerProvider 和 MeterProvider 的 Resource 中 service.name 为 "myapp"

#### Scenario: WithExporterEndpoint sets exporter address
- **WHEN** 调用 `Init(ctx, WithExporterEndpoint("collector:4317"))`
- **THEN** OTLP exporter 连接到 "collector:4317"

#### Scenario: WithStdoutExporter switches to stdout
- **WHEN** 调用 `Init(ctx, WithStdoutExporter(true))`
- **THEN** 使用 stdout exporter 输出到 stderr，不连接 OTLP endpoint

#### Scenario: WithResourceAttributes adds custom attributes
- **WHEN** 调用 `Init(ctx, WithResourceAttributes("env", "production", "region", "cn"))`
- **THEN** Resource 包含 env=production 和 region=cn attribute

### Requirement: Init returns error on failure
`Init` SHALL 在 Provider 创建失败时返回非 nil error。失败时不设置全局默认 Provider。

#### Scenario: OTLP connection fails
- **WHEN** OTLP endpoint 不可达且未设置 StdoutExporter
- **THEN** Init 返回非 nil error，全局 Provider 未被修改
