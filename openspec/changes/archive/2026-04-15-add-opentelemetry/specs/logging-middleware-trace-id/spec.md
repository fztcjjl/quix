## ADDED Requirements

### Requirement: Logging middleware outputs trace_id in log fields
Logging middleware SHALL 从 `request.Context()` 中提取 OTel trace_id，并作为 `trace_id` 字段写入日志。当 OTel 未启用（context 中无 trace）时，trace_id 字段 SHALL 不输出。

#### Scenario: Logging with trace_id
- **WHEN** 请求经过 otelgin middleware 且 Logging middleware 记录日志
- **THEN** 日志 fields 中包含 `trace_id` 字段，值为 OTel trace_id（32 位十六进制字符串）

#### Scenario: Logging without trace_id
- **WHEN** 请求未经过 otelgin middleware（OTel 未启用）
- **THEN** 日志 fields 中不包含 `trace_id` 字段

### Requirement: Logging middleware does not directly import OTel packages
Logging middleware SHALL 通过函数变量 `extractTraceID` 解耦 OTel 依赖，该变量在 telemetry 初始化时被设置。middleware 包本身 SHALL NOT import `go.opentelemetry.io/otel`。

#### Scenario: No OTel import in middleware package
- **WHEN** 检查 `core/transport/http/server/middleware/` 包的 import 列表
- **THEN** 不包含任何 `go.opentelemetry.io/otel` 依赖
