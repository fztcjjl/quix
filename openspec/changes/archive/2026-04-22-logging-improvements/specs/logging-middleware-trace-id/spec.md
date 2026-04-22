## MODIFIED Requirements

### Requirement: Logging middleware outputs trace_id in log fields
Logging middleware SHALL 从 `request.Context()` 中提取 OTel trace_id，并作为 `trace_id` 字段写入日志。当 OTel 未启用（context 中无 trace）时，trace_id 字段 SHALL 不输出。同时 SHALL 提取 OTel span_id 并作为 `span_id` 字段写入日志，与 trace_id 行为一致。

#### Scenario: Logging with trace_id and span_id
- **WHEN** 请求经过 otelgin middleware 且 Logging middleware 记录日志
- **THEN** 日志 fields 中包含 `trace_id` 字段（32 位十六进制字符串）和 `span_id` 字段（16 位十六进制字符串）

#### Scenario: Logging without trace_id and span_id
- **WHEN** 请求未经过 otelgin middleware（OTel 未启用）
- **THEN** 日志 fields 中不包含 `trace_id` 和 `span_id` 字段

### Requirement: Logging middleware does not directly import OTel packages
Logging middleware SHALL 通过函数变量 `extractTraceID` 和 `extractSpanID` 解耦 OTel 依赖，这些变量在 telemetry 初始化时被设置。middleware 包本身 SHALL NOT import `go.opentelemetry.io/otel`。

#### Scenario: No OTel import in middleware package
- **WHEN** 检查 `core/transport/http/server/middleware/` 包的 import 列表
- **THEN** 不包含任何 `go.opentelemetry.io/otel` 依赖

## ADDED Requirements

### Requirement: ExtractSpanID function variable
middleware 包 SHALL 导出 `ExtractSpanID func(ctx context.Context) string` 变量，与 `ExtractTraceID` 对称。telemetry 初始化时 MUST 设置此变量。

#### Scenario: ExtractSpanID is nil by default
- **WHEN** 检查 middleware 包的初始状态
- **THEN** `ExtractSpanID` MUST 为 nil

#### Scenario: ExtractSpanID set during telemetry init
- **WHEN** telemetry 初始化完成
- **THEN** `ExtractSpanID` MUST 被设置为非 nil 函数

### Requirement: telemetry.ExtractSpanID function
`core/telemetry/` 包 SHALL 提供 `ExtractSpanID(ctx context.Context) string` 函数，从 OTel context 提取 span_id（16 位十六进制字符串）。无 trace context 时返回空字符串。

#### Scenario: ExtractSpanID with valid trace context
- **WHEN** context 中包含有效 OTel span context
- **THEN** MUST 返回 span_id 的 16 位十六进制字符串

#### Scenario: ExtractSpanID without trace context
- **WHEN** context 中无 OTel span context
- **THEN** MUST 返回空字符串
