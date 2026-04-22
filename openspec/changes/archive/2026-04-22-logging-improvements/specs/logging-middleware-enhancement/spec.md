## MODIFIED Requirements

### Requirement: Logging returns a middleware that logs each HTTP request with structured fields
Logging SHALL 返回一个 middleware，使用结构化字段记录每个 HTTP 请求。日志 fields MUST 包含：method、path、status、latency（字符串）、latency_ms（float64 毫秒数）、client_ip、response_size。当 OTel 启用时，额外包含 trace_id 和 span_id 字段。当 query string 非空时，额外包含 query 字段。当 user_agent 非空时，额外包含 user_agent 字段。当 gin route 可获取（`c.FullPath()` 非 nil）时，额外包含 route 字段（归一化路径如 `/users/:id`）。当 gin context 中存在 `app_error` 且为 `*qerrors.Error` 类型时，额外包含 error_code 字段。

#### Scenario: Request logging with all fields
- **WHEN** 请求到达且路径未被跳过，query 为 `?page=2`，user_agent 为 `TestAgent`
- **THEN** 日志包含 method、path、status、latency、latency_ms、client_ip、response_size、query、user_agent 字段

#### Scenario: Access log includes latency_ms as numeric
- **WHEN** 请求耗时 1.5 秒
- **THEN** `latency_ms` MUST 为 `1500.0`（float64），同时 `latency` 仍为人类可读字符串

#### Scenario: Access log includes normalized route
- **WHEN** 请求路径为 `/users/42`，gin 路由定义为 `/users/:id`
- **THEN** `path` MUST 为 `/users/42`，`route` MUST 为 `/users/:id`

#### Scenario: Access log omits route when FullPath returns nil
- **WHEN** 请求路径未匹配任何 gin 路由（如 404）
- **THEN** 日志 MUST 不包含 `route` 字段

#### Scenario: Access log omits empty query
- **WHEN** 请求 URL 无 query string
- **THEN** 日志 MUST 不包含 `query` 字段

#### Scenario: Access log includes error_code for application errors
- **WHEN** handler 设置了 `*qerrors.Error{Code: "not_found"}` 到 gin context
- **THEN** 日志 MUST 包含 `error_code` 字段，值为 `"not_found"`

#### Scenario: Access log omits error_code for non-qerrors
- **WHEN** handler 设置了普通 Go error 到 gin context
- **THEN** 日志 MUST 不包含 `error_code` 字段

### Requirement: Slow request detection
Logging 中间件 SHALL 支持慢请求检测。通过 `WithSlowThreshold(d time.Duration) LoggingOption` 配置阈值。当请求耗时超过阈值时，MUST 在正常访问日志之外额外输出一条 WARN 级别日志，包含 msg="slow request"、path、latency_ms、threshold_ms 字段。

#### Scenario: Slow request triggers warning
- **WHEN** 配置 `WithSlowThreshold(2 * time.Second)` 且请求耗时 3 秒
- **THEN** MUST 输出额外 WARN 级别日志，msg="slow request"，包含 path、latency_ms=3000.0、threshold_ms=2000.0

#### Scenario: Fast request does not trigger warning
- **WHEN** 配置 `WithSlowThreshold(2 * time.Second)` 且请求耗时 1 秒
- **THEN** MUST 不输出 slow request 日志

#### Scenario: No threshold configured
- **WHEN** 未配置 WithSlowThreshold
- **THEN** MUST 不进行慢请求检测

### Requirement: Logging functional options
Logging 中间件 SHALL 通过选项模式支持自定义行为。`Logging(opts ...LoggingOption) gin.HandlerFunc` 函数 MUST 支持以下选项：

- `WithSkipPaths(paths ...string)` — 设置跳过路径（支持前缀匹配）
- `WithHook(fn LoggingHookFunc)` — 设置自定义 hook 函数
- `WithSlowThreshold(d time.Duration)` — 设置慢请求阈值

#### Scenario: WithSlowThreshold option
- **WHEN** 使用 `Logging(WithSlowThreshold(2 * time.Second))` 配置中间件
- **THEN** 超过 2 秒的请求 MUST 输出 slow request 警告日志
