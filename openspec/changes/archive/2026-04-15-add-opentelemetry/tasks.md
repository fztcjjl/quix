## 1. 依赖与基础设施

- [x] 1.1 添加 OTel Go SDK 依赖（otel、sdk、otlptrace/otlptracegrpc、otlpmetric/otlpmetricgrpc、contrib/gin/otelgin、exporters/stdout）
- [x] 1.2 创建 `core/telemetry/` 目录和 `telemetry.go` 文件骨架

## 2. telemetry Provider 管理

- [x] 2.1 实现 `Config` 结构体和 Option 函数（WithServiceName、WithExporterEndpoint、WithResourceAttributes、WithTracesEnabled、WithMetricsEnabled、WithStdoutExporter）
- [x] 2.2 实现 `Init(ctx, opts...)` —— Resource 构建、OTLP gRPC exporter 创建、TracerProvider/MeterProvider 初始化、全局注册
- [x] 2.3 实现 stdout exporter 分支（StdoutExporter=true 时）
- [x] 2.4 实现 shutdown func —— 按序 flush MeterProvider、TracerProvider
- [x] 2.5 实现 `Init` 错误处理 —— OTLP 连接失败返回 error，不修改全局 Provider

## 3. App 集成

- [x] 3.1 实现 `WithTelemetry(opts ...telemetry.Option)` Option 函数，调用 `telemetry.Init` 并存储 shutdown func
- [x] 3.2 修改 `App` 结构体，添加 `telemetryShutdown` 字段
- [x] 3.3 修改 `App.Shutdown`，在 HTTP server stop 后调用 telemetry shutdown、logger close

## 4. otelgin 中间件注入

- [x] 4.1 修改 `core/transport/http/server/server.go`，将 telemetry 状态传递到 Server
- [x] 4.2 实现 otelgin 中间件条件注入 —— WithTelemetry 启用时注入 otelgin.Middleware(serviceName)，中间件顺序：Recovery → otelgin → RequestID → Logging → ResponseMiddleware
- [x] 4.3 实现 `WithTracesEnabled(false)` 时跳过 otelgin 注入

## 5. Logging middleware trace_id 增强

- [x] 5.1 在 middleware 包中添加 `extractTraceID` 函数变量（`func(context.Context) string`），默认为 nil
- [x] 5.2 修改 `LoggingWith` 中的日志 fields 逻辑 —— 当 `extractTraceID != nil` 且返回非空时，添加 `trace_id` 字段
- [x] 5.3 在 `telemetry.Init` 中设置 `middleware.extractTraceID` 为 OTel trace context 提取函数

## 6. 测试

- [x] 6.1 `core/telemetry/telemetry_test.go` —— 测试 Init/Shutdown、Option 配置、stdout exporter、错误处理
- [x] 6.2 修改 `quix_test.go` —— 测试 WithTelemetry Option 和 Shutdown 顺序
- [x] 6.3 修改 `middleware/logging_test.go` —— 测试 extractTraceID 输出 trace_id 字段
- [x] 6.4 修改 `core/transport/http/server/server_test.go` —— 测试中间件链顺序（含/不含 otelgin）

## 7. 示例与收尾

- [x] 7.1 创建 `examples/telemetry/` 可运行示例
- [x] 7.2 执行 `go fmt ./...` 和 `golangci-lint run ./...`
- [x] 7.3 执行 `go test ./...` 确认全部通过
- [x] 7.4 更新 CLAUDE.md 架构说明和开发命令
