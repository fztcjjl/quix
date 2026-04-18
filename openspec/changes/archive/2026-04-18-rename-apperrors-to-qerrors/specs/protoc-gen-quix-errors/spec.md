## MODIFIED Requirements

### Requirement: 错误构造函数生成（无参）
插件 SHALL 为每个非零值 enum 值生成无参错误构造函数，返回 `*qerrors.Error`。Code、Message、StatusCode 均来自 proto 定义。

#### Scenario: 无参构造函数
- **WHEN** enum 值为 `USER_NOT_FOUND = 1 [(quix.errors.http_status) = 404, (quix.errors.error_message) = "用户不存在"]`
- **THEN** MUST 生成函数 `func UserNotFound() *qerrors.Error`，Code 为 `"USER_NOT_FOUND"`，Message 为 `"用户不存在"`，StatusCode 为 404

#### Scenario: 零值不生成构造函数
- **WHEN** enum 值为零值 `USER_ERROR_UNSPECIFIED = 0`
- **THEN** MUST NOT 生成构造函数

#### Scenario: 任意 HTTP 状态码
- **WHEN** enum 值标注 `http_status = 429`
- **THEN** MUST 生成 StatusCode 为 429 的构造函数

#### Scenario: 未设置 error_message
- **WHEN** enum 值有 `http_status` 但未设置 `error_message`
- **THEN** MUST 使用 enum 值名称（如 `"USER_NOT_FOUND"`）作为 Message

### Requirement: WithDetails 变体生成
插件 SHALL 为每个非零值 enum 值额外生成带 details 参数的构造函数变体。

#### Scenario: WithDetails 函数
- **WHEN** enum 值为 `USER_INVALID_INPUT = 4 [(quix.errors.http_status) = 400, (quix.errors.error_message) = "参数验证失败"]`
- **THEN** MUST 生成函数 `func UserInvalidInputWithDetails(details any) *qerrors.Error`，Code、Message、StatusCode 与无参版本一致，Details 为传入值

### Requirement: 生成代码导入 qerrors
生成的代码 SHALL import `github.com/fztcjjl/quix/core/errors` 并使用别名 `qerrors`。

#### Scenario: import 别名
- **WHEN** 查看生成的 `_errors.go` 文件
- **THEN** MUST 包含 `import qerrors "github.com/fztcjjl/quix/core/errors"`
