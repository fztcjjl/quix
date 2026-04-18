### Requirement: Error type
框架 SHALL 在 `core/errors/` 包中定义 Error 类型，包含 Code、Message、可选的 Details 字段和 StatusCode 字段。

#### Scenario: Error structure
- **WHEN** 创建 `errors.Error{Code: "param_invalid", Message: "参数验证失败", StatusCode: 400}`
- **THEN** 结构 MUST 包含 `code`（string）、`message`（string）、`details`（any，可选）、`status_code`（int，不序列化到 JSON）

#### Scenario: Error implements error interface
- **WHEN** Error 被作为 Go error 使用
- **THEN** MUST 实现 `error` 接口，`Error()` 返回 Message

#### Scenario: Error implements Unwrap interface
- **WHEN** Error 包含底层错误
- **THEN** MUST 实现 `Unwrap() error` 接口，返回 `cause` 字段，支持 `errors.Is`/`errors.As` 错误链追踪

#### Scenario: Error cause field
- **WHEN** 创建 `errors.Error{Code: "internal_error", Message: "Internal Server Error", cause: originalErr}`
- **THEN** `cause` 字段 MUST 不序列化到 JSON，仅用于错误链追踪

#### Scenario: Details is optional
- **WHEN** 创建 `errors.Error{Code: "not_found", Message: "不存在"}` 不设置 Details
- **THEN** JSON 序列化时 SHALL 省略 details 字段

#### Scenario: Details accepts any type
- **WHEN** 设置 Details 为 `[]map[string]any`、`map[string]string` 或自定义结构体
- **THEN** MUST 正确序列化为 JSON

#### Scenario: StatusCode not serialized
- **WHEN** Error 包含 StatusCode 字段
- **THEN** JSON 序列化时 MUST 不包含 `StatusCode` 字段

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

### Requirement: Predefined error wrap functions
框架 SHALL 为每个预定义错误函数提供 Wrap 变体（如 `BadRequestWrap`、`NotFoundWrap`），支持包装底层错误。Wrap 函数 MUST 返回与原函数相同的 StatusCode 和 Code，并设置 `cause` 字段。

#### Scenario: Wrap function sets cause
- **WHEN** 调用 `errors.NotFoundWrap("user_not_found", "用户不存在", dbErr)`
- **THEN** MUST 返回 `*Error{Code: "user_not_found", Message: "用户不存在", StatusCode: 404, cause: dbErr}`

#### Scenario: errors.Is works with wrapped error
- **WHEN** 使用 `errors.Is(appErr, dbErr)` 检查被 Wrap 的错误
- **THEN** MUST 返回 true
