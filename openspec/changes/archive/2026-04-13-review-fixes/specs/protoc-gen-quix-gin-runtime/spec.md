## MODIFIED Requirements

### Requirement: SetError 错误处理
Context SHALL 提供 `SetError(err error)` 方法，内部委托 `server.SetAppError` 实现，确保与 qhttp.Handler 的错误处理行为一致。

#### Scenario: apperrors.Error
- **WHEN** 调用 `SetError(&apperrors.Error{Code: "not_found", StatusCode: 404})`
- **THEN** MUST 将 error 存入 `c.Set("app_error", err)`，并调用 `c.AbortWithStatus(404)`

#### Scenario: 标准 error
- **WHEN** 调用 `SetError(fmt.Errorf("db failed"))`
- **THEN** MUST 包装为 `*Error{Code: "internal_error", StatusCode: 500}`，存入 `app_error`，并调用 `c.AbortWithStatus(500)`
