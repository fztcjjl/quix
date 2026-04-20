### Requirement: Logging returns a middleware that logs each HTTP request with structured fields
Logging SHALL 返回一个 middleware，使用结构化字段记录每个 HTTP 请求。SkipPaths 指定跳过日志的精确路径（例如 "/healthz"）。当 OTel 启用时，日志 fields SHALL 包含 `trace_id` 字段。

#### Scenario: Request logging with structured fields
- **WHEN** 请求到达且路径未被跳过
- **THEN** 日志包含 method、path、status、latency、client_ip、response_size 字段；当 OTel 启用时额外包含 trace_id 字段

#### Scenario: Skip exact paths
- **WHEN** 请求路径为 "/healthz" 且 skipPaths 包含 "/healthz"
- **THEN** 不输出日志

### Requirement: Prefix matching for skipPaths
Logging 中间件 SHALL 支持前缀匹配。当 skipPaths 中的路径以 `/` 结尾时（如 `/metrics/`），MUST 匹配所有以该前缀开头的请求路径。不以 `/` 结尾的路径保持精确匹配。

#### Scenario: Exact match (no trailing slash)
- **WHEN** skipPaths 为 `["/healthz"]`，请求路径为 `/healthz`
- **THEN** MUST 跳过日志记录

#### Scenario: Exact match rejects sub-paths (no trailing slash)
- **WHEN** skipPaths 为 `["/healthz"]`，请求路径为 `/healthz/ready`
- **THEN** MUST NOT 跳过日志记录

#### Scenario: Prefix match (trailing slash)
- **WHEN** skipPaths 为 `["/metrics/"]`，请求路径为 `/metrics/health`
- **THEN** MUST 跳过日志记录

#### Scenario: Prefix match rejects parent path (trailing slash)
- **WHEN** skipPaths 为 `["/metrics/"]`，请求路径为 `/metrics`
- **THEN** MUST NOT 跳过日志记录

### Requirement: Logging functional options
Logging 中间件 SHALL 通过选项模式支持自定义行为。`Logging(opts ...LoggingOption) gin.HandlerFunc` 函数 MUST 支持以下选项：

- `WithSkipPaths(paths ...string)` — 设置跳过路径（支持前缀匹配）
- `WithHook(fn LoggingHookFunc)` — 设置自定义 hook 函数

#### Scenario: WithSkipPaths option
- **WHEN** 使用 `Logging(WithSkipPaths("/healthz", "/metrics/"))` 配置中间件
- **THEN** MUST 跳过 `/healthz`（精确）和 `/metrics/*`（前缀）的日志记录

#### Scenario: WithHook receives log fields
- **WHEN** 使用 `Logging(WithHook(fn))` 配置中间件，请求完成后
- **THEN** `fn` MUST 被调用，接收的参数 MUST 包含 method、path、status、latency、clientIP、responseSize 等标准字段

#### Scenario: Hook can add custom fields
- **WHEN** Hook 函数向传入的字段 map 追加 `"custom_key", "custom_val"`
- **THEN** 日志输出 MUST 包含 `custom_key=custom_val` 字段

#### Scenario: No options defaults to all requests logged
- **WHEN** 使用 `Logging()` 不传入任何选项
- **THEN** MUST 对所有请求输出日志

### Requirement: LoggingHookFunc type
中间件 SHALL 导出 `LoggingHookFunc` 类型，签名为 `func(c *gin.Context, fields map[string]any)`，在请求完成后、日志写入前被调用。

#### Scenario: Hook function signature
- **WHEN** 开发者查看 `LoggingHookFunc` 类型定义
- **THEN** 签名 MUST 为 `type LoggingHookFunc func(c *gin.Context, fields map[string]any)`
