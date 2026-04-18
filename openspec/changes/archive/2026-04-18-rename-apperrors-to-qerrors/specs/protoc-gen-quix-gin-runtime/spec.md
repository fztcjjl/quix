## MODIFIED Requirements

### Requirement: SetError 错误处理
Context SHALL 提供 `SetError(err error)` 方法，内部委托 `server.SetAppError` 实现，确保与 qhttp.Handler 的错误处理行为一致。

#### Scenario: qerrors.Error
- **WHEN** 调用 `SetError(&qerrors.Error{Code: "not_found", StatusCode: 404})`
- **THEN** MUST 将 error 存入 `c.Set("app_error", err)`，并调用 `c.AbortWithStatus(404)`

#### Scenario: 标准 error
- **WHEN** 调用 `SetError(fmt.Errorf("db failed"))`
- **THEN** MUST 包装为 `*Error{Code: "internal_error", StatusCode: 500}`，存入 `app_error`，并调用 `c.AbortWithStatus(500)`

### Requirement: GetError 获取错误
Context SHALL 提供 `GetError() *qerrors.Error` 方法。

#### Scenario: 有错误
- **WHEN** 之前调用了 `SetError(err)`
- **THEN** MUST 返回存储的 `*qerrors.Error`

#### Scenario: 无错误
- **WHEN** 未调用 `SetError`
- **THEN** MUST 返回 nil
