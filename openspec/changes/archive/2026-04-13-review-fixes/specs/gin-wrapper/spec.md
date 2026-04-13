## MODIFIED Requirements

### Requirement: ResponseMiddleware
框架 SHALL 提供 ResponseMiddleware，统一格式化错误响应。ResponseMiddleware MUST 对 `app_error` 进行安全类型断言（comma-ok），防止非 `*apperrors.Error` 类型导致 panic。

#### Scenario: Error response format
- **WHEN** handler 中返回了错误且 ResponseMiddleware 已挂载
- **THEN** 响应体 MUST 为 `{"error": {...}}`，HTTP status 为 Error.StatusCode

#### Scenario: No error skips formatting
- **WHEN** handler 正常执行且未返回错误
- **THEN** ResponseMiddleware MUST 不写入任何响应（成功响应由 handler 直接处理）

#### Scenario: Non-apperrors.Error in context
- **WHEN** `app_error` 存储了非 `*apperrors.Error` 类型的值
- **THEN** MUST 返回 HTTP 500 且不 panic（而非直接断言崩溃）

## MODIFIED Requirements

### Requirement: Default middleware recovers panic
- **WHEN** 使用默认中间件的 Server 处理请求时 handler 发生 panic
- **THEN** MUST 返回 HTTP 500 JSON 响应 `{"error": {"code": "internal_error", "message": "Internal Server Error"}}`，服务不崩溃

#### Scenario: Recovery logs with request context
- **WHEN** panic 发生时
- **THEN** MUST 使用 `c.Request.Context()` 记录日志（保留链路追踪），并包含 `request_id` 字段

## MODIFIED Requirements

### Requirement: Handler wrapper
框架 SHALL 在 `core/transport/http/server/` 包中提供 `Handler()` 函数和 `SetAppError()` 函数。`SetAppError` SHALL 是错误处理的核心实现，`Handler()` 委托调用 `SetAppError`。

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

## ADDED Requirements

### Requirement: SetAppError 公共错误处理函数
框架 SHALL 在 `core/transport/http/server/errors.go` 中提供 `SetAppError(c *gin.Context, err error)` 函数，作为统一的错误处理入口。

#### Scenario: SetAppError with apperrors.Error
- **WHEN** 调用 `SetAppError(c, &apperrors.Error{Code: "not_found", StatusCode: 404})`
- **THEN** MUST 将 error 存入 `c.Set("app_error", err)`，并调用 `c.AbortWithStatus(404)`

#### Scenario: SetAppError with standard error
- **WHEN** 调用 `SetAppError(c, fmt.Errorf("db failed"))`
- **THEN** MUST 包装为 `*Error{Code: "internal_error", StatusCode: 500}`，存入 app_error，并调用 `c.AbortWithStatus(500)`

## ADDED Requirements

### Requirement: ReadHeaderTimeout 可配置
HTTP Server SHALL 提供 `WithReadHeaderTimeout(d time.Duration)` Option，允许自定义 ReadHeaderTimeout。默认值为 5 秒。

#### Scenario: Custom timeout
- **WHEN** 使用 `NewServer(WithReadHeaderTimeout(10 * time.Second))`
- **THEN** MUST 设置 `http.Server.ReadHeaderTimeout` 为 10 秒

#### Scenario: Default timeout
- **WHEN** 未设置 `WithReadHeaderTimeout`
- **THEN** MUST 使用默认值 5 秒
