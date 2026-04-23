## REMOVED Requirements

### Requirement: Logging middleware does not directly import OTel packages
**Reason**: middleware 包直接调用 `telemetry.ExtractTraceID()` / `telemetry.ExtractSpanID()`，不再需要函数变量解耦。middleware 是框架内部包，可以依赖同属框架的 telemetry 包。
**Migration**: 无需迁移，外部代码不受影响（ExtractTraceID/ExtractSpanID 变量仅在框架内部使用）。

### Requirement: ExtractSpanID function variable
**Reason**: 与 ExtractTraceID 一起被移除，改为直接调用 telemetry 包函数。
**Migration**: 无需迁移，该变量为框架内部使用。

## MODIFIED Requirements

### Requirement: Logging middleware outputs trace_id in log fields
Logging middleware SHALL 从 `request.Context()` 中提取 OTel trace_id，并作为 `trace_id` 字段写入日志。当 OTel 未启用（context 中无 trace）时，trace_id 字段 SHALL 不输出。同时 SHALL 提取 OTel span_id 并作为 `span_id` 字段写入日志，与 trace_id 行为一致。middleware SHALL 直接调用 `telemetry.ExtractTraceID(ctx)` 和 `telemetry.ExtractSpanID(ctx)` 获取 trace/span ID。

#### Scenario: Logging with trace_id and span_id
- **WHEN** 请求经过 otelgin middleware 且 Logging middleware 记录日志
- **THEN** 日志 fields 中包含 `trace_id` 字段（32 位十六进制字符串）和 `span_id` 字段（16 位十六进制字符串）

#### Scenario: Logging without trace_id and span_id
- **WHEN** 请求未经过 otelgin middleware（OTel 未启用）
- **THEN** 日志 fields 中不包含 `trace_id` 和 `span_id` 字段
