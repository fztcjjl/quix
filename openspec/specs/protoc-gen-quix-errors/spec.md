## ADDED Requirements

### Requirement: 插件入口
插件 SHALL 在 `cmd/protoc-gen-quix-errors/` 中提供 `main.go`，使用 protogen 框架作为 protoc 插件入口。

#### Scenario: 插件发现
- **WHEN** protoc 或 buf 使用 `--quix-errors_out` 或 `local: protoc-gen-quix-errors` 配置
- **THEN** MUST 找到 `protoc-gen-quix-errors` 可执行文件并执行

#### Scenario: 解析 proto 文件
- **WHEN** 插件接收到包含 enum 定义的 proto 文件
- **THEN** MUST 遍历所有 `Generate == true` 的文件，提取带 `http_status` 注解的 enum 值并生成 `_errors.go` 文件

### Requirement: paths=source_relative 支持
插件 SHALL 支持 `paths=source_relative` 参数，将生成文件放在与 proto 源文件对应的相对路径下。

#### Scenario: paths=source_relative
- **WHEN** 传入 `paths=source_relative` 参数
- **THEN** MUST 将 `_errors.go` 生成到 proto 文件对应的目录

#### Scenario: 默认路径
- **WHEN** 未传入 `paths=source_relative` 参数
- **THEN** MUST 使用 Go import path 构造输出文件路径

### Requirement: 错误码常量生成
插件 SHALL 为每个带 `http_status` 注解的 enum 值生成 Go string 常量。

#### Scenario: 常量生成
- **WHEN** enum 值为 `USER_NOT_FOUND = 1 [(quix.errors.http_status) = 404]`
- **THEN** MUST 生成常量 `UserNotFoundCode = "USER_NOT_FOUND"`

#### Scenario: 零值也生成常量
- **WHEN** enum 零值为 `USER_ERROR_UNSPECIFIED = 0 [(quix.errors.http_status) = 400]`
- **THEN** MUST 生成常量 `UserErrorCodeUnspecified = "USER_ERROR_UNSPECIFIED"`

### Requirement: 错误构造函数生成（无参）
插件 SHALL 为每个非零值 enum 值生成无参错误构造函数，返回 `*apperrors.Error`。Code、Message、StatusCode 均来自 proto 定义。

#### Scenario: 无参构造函数
- **WHEN** enum 值为 `USER_NOT_FOUND = 1 [(quix.errors.http_status) = 404, (quix.errors.error_message) = "用户不存在"]`
- **THEN** MUST 生成函数 `func UserNotFound() *apperrors.Error`，Code 为 `"USER_NOT_FOUND"`，Message 为 `"用户不存在"`，StatusCode 为 404

#### Scenario: 零值不生成构造函数
- **WHEN** enum 零值为 `USER_ERROR_UNSPECIFIED = 0`
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
- **THEN** MUST 生成函数 `func UserInvalidInputWithDetails(details any) *apperrors.Error`，Code、Message、StatusCode 与无参版本一致，Details 为传入值

### Requirement: 函数命名规则
插件 SHALL 将 enum 值转换为 PascalCase 函数名，去掉 enum 前缀。

#### Scenario: 前缀剥离
- **WHEN** enum 名为 `UserError`，enum 值为 `USER_NOT_FOUND`
- **THEN** MUST 去掉前缀 `USER_ERROR_`，转 PascalCase，生成函数名 `UserNotFound`

#### Scenario: Error 后缀去除
- **WHEN** enum 名以 `Error` 结尾
- **THEN** MUST 去掉 `Error` 后缀再拼接，如 `UserError::USER_ALREADY_EXISTS` → `UserAlreadyExists`

#### Scenario: 常量命名
- **WHEN** enum 值为 `USER_NOT_FOUND`
- **THEN** MUST 生成常量名 `UserNotFoundCode`（PascalCase + `Code` 后缀）

### Requirement: 无注解 enum 不生成输出
插件 SHALL 跳过不包含 `http_status` 注解的 enum，不生成任何文件。

#### Scenario: 普通 enum 忽略
- **WHEN** proto 文件中有 enum `RegularEnum` 但没有值带 `http_status` 注解
- **THEN** MUST 不为该 enum 生成输出文件

#### Scenario: 混合 enum 处理
- **WHEN** 同一 proto 文件有带注解和不带注解的 enum
- **THEN** MUST 只为带注解的 enum 生成文件

### Requirement: 生成文件命名
插件 SHALL 将每个带注解的 enum 生成独立文件，命名为 `snake_case(enum_name)_errors.go`。

#### Scenario: 文件命名
- **WHEN** enum 名为 `UserError`
- **THEN** MUST 生成文件 `user_error_errors.go`

#### Scenario: 多 enum 多文件
- **WHEN** 同一 proto 文件有 `UserError` 和 `OrderError` 两个带注解的 enum
- **THEN** MUST 生成两个文件 `user_error_errors.go` 和 `order_error_errors.go`

### Requirement: 生成代码导入 apperrors
生成的代码 SHALL import `github.com/fztcjjl/quix/core/errors` 并使用别名 `apperrors`。

#### Scenario: import 别名
- **WHEN** 查看生成的 `_errors.go` 文件
- **THEN** MUST 包含 `import apperrors "github.com/fztcjjl/quix/core/errors"`

### Requirement: buf 兼容
插件 SHALL 完全兼容 buf 的 `buf generate` 命令。

#### Scenario: buf generate
- **WHEN** `buf.gen.yaml` 配置 `local: protoc-gen-quix-errors`
- **THEN** MUST 正常生成代码

### Requirement: 使用示例
框架 SHALL 在 `examples/proto-errors/` 中提供完整可运行的示例。

#### Scenario: 示例可运行
- **WHEN** 执行 `cd examples/proto-errors && buf generate && go run main.go`
- **THEN** MUST 展示生成错误码常量和构造函数的完整用法
