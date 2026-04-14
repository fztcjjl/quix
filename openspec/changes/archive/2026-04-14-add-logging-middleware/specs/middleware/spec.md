## ADDED Requirements

### Requirement: Logging middleware
框架 SHALL 提供 Logging 中间件，为每个 HTTP 请求输出结构化 access log。

#### Scenario: Log request with key fields
- **WHEN** 收到 HTTP 请求并完成处理
- **THEN** MUST 输出包含 `method`、`path`、`status`、`latency`、`request_id`、`client_ip`、`response_size` 字段的结构化日志

#### Scenario: Log level by status code
- **WHEN** 请求响应状态码为 2xx 或 3xx
- **THEN** MUST 使用 Info 级别输出日志

#### Scenario: 4xx uses Warn level
- **WHEN** 请求响应状态码为 4xx
- **THEN** MUST 使用 Warn 级别输出日志

#### Scenario: 5xx uses Error level
- **WHEN** 请求响应状态码为 5xx
- **THEN** MUST 使用 Error 级别输出日志

#### Scenario: Logging usage
- **WHEN** 用户调用 `middleware.Logging()`
- **THEN** MUST 返回 `gin.HandlerFunc`

### Requirement: Skip paths
Logging 中间件 SHALL 支持跳过指定路径，不输出日志。

#### Scenario: Skip exact path
- **WHEN** 用户调用 `middleware.Logging("/healthz")` 且请求路径为 `/healthz`
- **THEN** MUST 不输出任何日志

#### Scenario: Non-skipped path logged
- **WHEN** 用户调用 `middleware.Logging("/healthz")` 且请求路径为 `/api/users`
- **THEN** MUST 正常输出日志

#### Scenario: No skip paths configured
- **WHEN** 用户调用 `middleware.Logging()` 不传入跳过路径
- **THEN** MUST 对所有请求输出日志

## MODIFIED Requirements

### Requirement: Default middleware mounting
App SHALL 默认挂载 Recovery、RequestID、Logging 和 Response 中间件到 HTTP Server。

#### Scenario: Default middleware mounted automatically
- **WHEN** 用户调用 `quix.New()` 未传入 `WithDefaultMiddleware(false)`
- **THEN** HTTP Server MUST 自动挂载 Recovery、RequestID、Logging 和 Response 中间件，顺序为 `Recovery → RequestID → Logging → Response`

#### Scenario: Disable default middleware
- **WHEN** 用户调用 `quix.New(quix.WithDefaultMiddleware(false))`
- **THEN** HTTP Server MUST 不挂载任何默认中间件
