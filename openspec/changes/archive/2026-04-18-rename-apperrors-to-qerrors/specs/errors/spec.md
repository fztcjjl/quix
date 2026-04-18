## MODIFIED Requirements

### Requirement: Predefined error functions
框架 SHALL 提供常用预定义错误创建函数，每个函数返回带默认 StatusCode 的 `*Error`。生成的代码 import 时 SHALL 使用别名 `qerrors`（`import qerrors "github.com/fztcjjl/quix/core/errors"`）。

#### Scenario: Available predefined errors
- **WHEN** 开发者调用 `errors.BadRequest()`、`errors.NotFound()`、`errors.Unauthorized()`、`errors.Internal()`、`errors.Forbidden()`
- **THEN** MUST 返回 `*Error` 实例

#### Scenario: Predefined errors with code and message
- **WHEN** 调用 `errors.NotFound("user_not_found", "用户不存在")`
- **THEN** MUST 返回 `*Error{Code: "user_not_found", Message: "用户不存在", StatusCode: 404}`

#### Scenario: Default StatusCode for predefined errors
- **WHEN** 调用 `errors.BadRequest()`、`errors.Unauthorized()`、`errors.Forbidden()`、`errors.NotFound()`、`errors.Internal()`
- **THEN** StatusCode MUST 分别为 400、401、403、404、500

#### Scenario: Import alias in generated code
- **WHEN** 生成的错误码代码引用 `core/errors` 包
- **THEN** MUST 使用别名 `qerrors`（`import qerrors "github.com/fztcjjl/quix/core/errors"`）
