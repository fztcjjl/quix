### Requirement: WithRequestLogger middleware creates request-scoped Logger
Logging 包 SHALL 提供 `WithRequestLogger(opts ...WithRequestLoggerOption) gin.HandlerFunc` 中间件。该中间件 MUST 从 context 提取 trace_id、span_id、request_id，创建携带这些字段的 child logger，并通过 `log.NewContext` 注入 context，再通过 `c.Request = c.Request.WithContext(newCtx)` 替换请求 context。

#### Scenario: WithRequestLogger injects enriched logger
- **WHEN** 请求经过 WithRequestLogger 中间件
- **THEN** `log.FromContext(c.Request.Context())` MUST 返回一个携带 trace_id、span_id、request_id 字段的 child logger

#### Scenario: WithRequestLogger falls back gracefully when no trace
- **WHEN** OTel 未启用，context 中无 trace context
- **THEN** child logger MUST 仅携带 request_id 字段，不携带 trace_id 和 span_id

#### Scenario: WithRequestLogger falls back to default Logger
- **WHEN** context 中无 Logger（WithRequestLogger 是第一个中间件）
- **THEN** MUST 使用 `log.Default()` 作为基础 logger 创建 child logger

### Requirement: WithRequestLogger extracts trace_id and span_id via function variables
WithRequestLogger MUST 通过 `ExtractTraceID` 和 `ExtractSpanID` 函数变量提取 trace 信息，不直接 import otel 包。当函数变量为 nil 时，对应字段 MUST 不注入。

#### Scenario: ExtractTraceID is nil
- **WHEN** `ExtractTraceID` 为 nil
- **THEN** child logger MUST 不包含 trace_id 字段

#### Scenario: ExtractSpanID is nil
- **WHEN** `ExtractSpanID` 为 nil
- **THEN** child logger MUST 不包含 span_id 字段

### Requirement: WithRequestLogger extracts request_id from gin context
WithRequestLogger MUST 从 gin context 的 `X-Request-Id` key 提取 request_id。当不存在时，request_id 字段 MUST 不注入。

#### Scenario: request_id present in gin context
- **WHEN** requestid 中间件已在 WithRequestLogger 之前运行
- **THEN** child logger MUST 包含 `request_id` 字段

#### Scenario: request_id absent
- **WHEN** gin context 中无 `X-Request-Id`
- **THEN** child logger MUST 不包含 `request_id` 字段

### Requirement: WithRequestLogger in default middleware chain
Server 的默认中间件链 MUST 包含 WithRequestLogger，位于 requestid 和 otelgin 之后、Recovery 之前。顺序 MUST 为：requestid → [otelgin] → WithRequestLogger → Recovery → CORS → Logging → Response。

#### Scenario: Default middleware chain order
- **WHEN** 创建 Server 且 `defaultMiddleware` 为 true
- **THEN** WithRequestLogger MUST 在 Recovery 之前被挂载
