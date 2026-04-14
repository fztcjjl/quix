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

### Requirement: LoggingWith functional options
Logging 中间件 SHALL 提供 `LoggingWith(opts ...LoggingOption) gin.HandlerFunc` 函数，支持通过选项模式自定义中间件行为。MUST 支持以下选项：

- `WithSkipPaths(paths ...string)` — 设置跳过路径（支持前缀匹配）
- `WithHook(fn LoggingHookFunc)` — 设置自定义 hook 函数

#### Scenario: WithSkipPaths option
- **WHEN** 使用 `LoggingWith(WithSkipPaths("/healthz", "/metrics/"))` 配置中间件
- **THEN** MUST 跳过 `/healthz`（精确）和 `/metrics/*`（前缀）的日志记录

#### Scenario: WithHook receives log fields
- **WHEN** 使用 `LoggingWith(WithHook(fn))` 配置中间件，请求完成后
- **THEN** `fn` MUST 被调用，接收的参数 MUST 包含 method、path、status、latency、clientIP、responseSize 等标准字段

#### Scenario: Hook can add custom fields
- **WHEN** Hook 函数向传入的字段 map 追加 `"custom_key", "custom_val"`
- **THEN** 日志输出 MUST 包含 `custom_key=custom_val` 字段

#### Scenario: Original Logging function still works
- **WHEN** 使用原有的 `Logging(skipPaths...)` 函数
- **THEN** 行为 MUST 保持不变（向后兼容）

### Requirement: LoggingHookFunc type
中间件 SHALL 导出 `LoggingHookFunc` 类型，签名为 `func(c *gin.Context, fields map[string]any)`，在请求完成后、日志写入前被调用。

#### Scenario: Hook function signature
- **WHEN** 开发者查看 `LoggingHookFunc` 类型定义
- **THEN** 签名 MUST 为 `type LoggingHookFunc func(c *gin.Context, fields map[string]any)`
