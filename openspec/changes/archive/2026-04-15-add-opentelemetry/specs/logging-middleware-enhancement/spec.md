## MODIFIED Requirements

### Requirement: Logging returns a middleware that logs each HTTP request with structured fields
Logging SHALL 返回一个 middleware，使用结构化字段记录每个 HTTP 请求。SkipPaths 指定跳过日志的精确路径（例如 "/healthz"）。当 OTel 启用时，日志 fields SHALL 包含 `trace_id` 字段。

#### Scenario: Request logging with structured fields
- **WHEN** 请求到达且路径未被跳过
- **THEN** 日志包含 method、path、status、latency、client_ip、response_size 字段；当 OTel 启用时额外包含 trace_id 字段

#### Scenario: Skip exact paths
- **WHEN** 请求路径为 "/healthz" 且 skipPaths 包含 "/healthz"
- **THEN** 不输出日志
