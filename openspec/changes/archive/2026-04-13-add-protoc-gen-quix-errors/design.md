## Context

quix 的 `core/errors.Error` 结构体包含 Code、Message、Details、StatusCode 四个字段。当前错误码是手写字符串，如 `errors.NotFound("user_not_found", "用户不存在")`。框架已有 `protoc-gen-quix-gin` 插件从 proto service 生成 Gin 路由代码，新插件需遵循相同的架构模式。

## Goals / Non-Goals

**Goals:**
- 提供 proto 定义方式声明错误码的 Code、Message、HTTP StatusCode（完整声明）
- 生成 Go 错误码常量（避免拼写错误）和无参构造函数（message 由 proto 定义，运行时无需传递）
- 作为独立 protoc 插件，可单独使用或与 protoc-gen-quix-gin 配合

**Non-Goals:**
- 不替代 `core/errors` 的预定义函数（BadRequest/NotFound 等），两者共存
- 不生成 i18n 或文档
- 不处理 gRPC status 映射（纯 HTTP 场景）
- 不生成 Details 结构体类型（用户直接用 proto message 作为 Details）

## Decisions

### 1. Proto 定义模式：enum + 两个独立 field option

**选择**: 在 `google.protobuf.EnumValueOptions` 上扩展两个独立字段：
- `int32 http_status = 84001` — HTTP 状态码
- `string error_message = 84002` — 错误消息

**原因**: enum 天然适合定义一组离散值。两个独立字段比嵌套 message 更简洁直观。这与 proto 生态常见模式一致（grpc-gateway、validate 等）。

**替代方案**: 使用嵌套 message option 包装所有字段 → 被否决，过于冗长。

### 2. 框架 proto 放在 `proto/` 根目录

**选择**: `proto/errdesc/errdesc.proto`（error descriptor），生成的 Go 代码提交到仓库

**原因**: 自定义 option 需要 proto 定义文件。`errdesc` 准确表达"错误描述注解"的含义（而非错误类型定义）。独立目录便于扩展。Go package `errdesc` 简洁直接。

### 3. 直接构造 struct 而非调用预定义函数

**选择**: 生成 `&apperrors.Error{Code: ..., Message: ..., StatusCode: ...}` 而非 `apperrors.NotFound(...)`

**原因**: HTTP 状态码来自 proto 注解，可能是任意值（409、422、429 等），不一定对应现有 5 个预定义函数。直接构造更灵活。

### 4. 无参构造函数

**选择**: 生成 `func UserNotFound() *apperrors.Error`，Code/Message/StatusCode 全部来自 proto

**原因**: 错误的完整语义（码、消息、状态码）都在 proto 中声明，运行时不需要再传参数。这消除了手写消息的不一致性。

**回退**: 如果 proto 未设置 `error_message`，使用 enum 值名称（如 `"USER_NOT_FOUND"`）作为 Message。

### 5. 函数命名：去掉 enum 前缀 + PascalCase

**选择**: `UserError::USER_NOT_FOUND` → 函数名 `UserNotFound`

**算法**:
1. enum 名 `UserError` 去掉 `Error` 后缀 → `User`
2. enum 值 `USER_NOT_FOUND` 去掉前缀 `USER_ERROR_` → `NOT_FOUND`
3. 拼接并转 PascalCase → `UserNotFound`

**风险**: 前缀剥离可能失败 → 回退使用完整值名的 PascalCase + enum 名前缀

### 6. 零值跳过构造函数

**选择**: `_UNSPECIFIED = 0` 只生成常量，不生成构造函数

**原因**: proto 零值是约定俗成的"未知"状态，不应作为真实错误使用。

### 7. WithDetails 变体

**选择**: 为每个错误码额外生成 `XxxWithDetails(details any) *Error`（仅一个参数）

**原因**: Code/Message/StatusCode 已在 proto 中定义，运行时只需传 Details。签名更简洁。

## Risks / Trade-offs

- **自定义 option 需要 proto 依赖**: 用户 proto 文件必须 import `errdesc.proto` → 提供 buf deps 配置和文档说明
- **toPascalCase 重复**: 与 protoc-gen-quix-gin 中的函数重复 → 两个插件是独立的 main package，无法共享代码，可接受；未来可提取到 `internal/protoutil`
- **前缀剥离边缘情况**: enum 命名不符合 `XXX_ERROR_VALUE` 模式时剥离失败 → 回退策略保证不生成错误代码
- **message 硬编码**: 错误消息写死在 proto 中，无法动态修改 → 如需动态消息可直接构造 `apperrors.Error` struct
