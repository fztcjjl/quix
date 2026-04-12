## ADDED Requirements

### Requirement: Handler wrapper
框架 SHALL 在 `core/transport/http/server/` 包中提供 `Handler()` 函数，将 `func(c *gin.Context) error` 转换为 `gin.HandlerFunc`，自动处理错误响应。

#### Scenario: Handler returns nil
- **WHEN** handler 返回 nil
- **THEN** 请求正常继续，不设置 app_error

#### Scenario: Handler returns *Error
- **WHEN** handler 返回 `errors.NotFound("user_not_found", "用户不存在")`
- **THEN** MUST 将 Error 存入 `c.Get("app_error")`，并设置 HTTP status 为 Error.StatusCode（404）

#### Scenario: Handler returns standard error
- **WHEN** handler 返回 `fmt.Errorf("db connection failed")`
- **THEN** MUST 自动包装为 `*Error{Code: "internal_error", Message: "db connection failed", StatusCode: 500}`，存入 app_error，HTTP status 为 500

#### Scenario: Handler wrapper prevents subsequent handlers
- **WHEN** handler 返回非 nil error
- **THEN** 后续 handler MUST 不执行

### Requirement: ResponseMiddleware in default middleware
HTTP Server 的默认中间件列表 SHALL 包含 ResponseMiddleware。

#### Scenario: Default middleware order
- **WHEN** 创建 HTTP Server 且默认中间件未禁用
- **THEN** 中间件挂载顺序 MUST 为 Recovery → RequestID → ResponseMiddleware

### Requirement: ResponseMiddleware
框架 SHALL 提供 ResponseMiddleware，统一格式化错误响应。

#### Scenario: Error response format
- **WHEN** handler 中返回了错误且 ResponseMiddleware 已挂载
- **THEN** 响应体 MUST 为 `{"error": {...}}`，HTTP status 为 Error.StatusCode

#### Scenario: No error skips formatting
- **WHEN** handler 正常执行且未返回错误
- **THEN** ResponseMiddleware MUST 不写入任何响应（成功响应由 handler 直接处理）

### Requirement: Default middleware mounting
ResponseMiddleware SHALL 默认挂载到 HTTP Server，顺序为 Recovery → RequestID → ResponseMiddleware。

#### Scenario: Default middleware includes ResponseMiddleware
- **WHEN** 用户调用 `quix.New()` 未禁用默认中间件
- **THEN** HTTP Server MUST 挂载 Recovery、RequestID、ResponseMiddleware

#### Scenario: Disable skips ResponseMiddleware
- **WHEN** 用户调用 `quix.New(quix.WithDefaultMiddleware(false))`
- **THEN** HTTP Server MUST 不挂载 ResponseMiddleware
