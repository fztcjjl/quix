## MODIFIED Requirements

### Requirement: SetAppError 公共错误处理函数
框架 SHALL 在 `core/transport/http/server/errors.go` 中提供 `SetAppError(c *gin.Context, err error)` 函数，作为统一的错误处理入口。

#### Scenario: SetAppError with qerrors.Error
- **WHEN** 调用 `SetAppError(c, &qerrors.Error{Code: "not_found", StatusCode: 404})`
- **THEN** MUST 将 error 存入 `c.Set("app_error", err)`，并调用 `c.AbortWithStatus(404)`

#### Scenario: SetAppError with standard error
- **WHEN** 调用 `SetAppError(c, fmt.Errorf("db failed"))`
- **THEN** MUST 包装为 `*Error{Code: "internal_error", StatusCode: 500, cause: err}`，存入 `app_error`，并调用 `c.AbortWithStatus(500)`
  - 当 `HideInternalErrors` 为 true（prod/staging）时：Message 为 "Internal Server Error"
  - 当 `HideInternalErrors` 为 false（dev/test）时：Message 为原始错误消息

### Requirement: ResponseMiddleware
框架 SHALL 提供 ResponseMiddleware，统一格式化错误响应。ResponseMiddleware MUST 对 `app_error` 进行安全类型断言（comma-ok），防止非 `*qerrors.Error` 类型导致 panic。

#### Scenario: Error response format
- **WHEN** handler 中返回了错误且 ResponseMiddleware 已挂载
- **THEN** 响应体 MUST 为 `{"error": {...}}`，HTTP status 为 Error.StatusCode

#### Scenario: No error skips formatting
- **WHEN** handler 正常执行且未返回错误
- **THEN** ResponseMiddleware MUST 不写入任何响应（成功响应由 handler 直接处理）

#### Scenario: Non-qerrors.Error in context
- **WHEN** `app_error` 存储了非 `*qerrors.Error` 类型的值
- **THEN** MUST 返回 HTTP 500 且不 panic（而非直接断言崩溃）
