## ADDED Requirements

### Requirement: ShouldBindUriConflictCheck 路径变量冲突检测
Context SHALL 提供 `ShouldBindUriConflictCheck(req any, pathVars []string) error` 方法，绑定路径参数到 req，并在 body 已设置同名字段且值不一致时返回错误。

#### Scenario: body 未传路径字段（无冲突）
- **WHEN** 调用 `ShouldBindUriConflictCheck(req, ["id"])` 且 `req.Id` 为零值，路径变量 `id=123`
- **THEN** MUST 将 `123` 绑定到 `req.Id`，返回 nil

#### Scenario: body 传了路径字段且值一致（无冲突）
- **WHEN** 调用 `ShouldBindUriConflictCheck(req, ["id"])` 且 `req.Id = "123"`（来自 JSON body），路径变量 `id=123`
- **THEN** MUST 将 `123` 绑定到 `req.Id`，返回 nil

#### Scenario: body 传了路径字段且值不一致（冲突）
- **WHEN** 调用 `ShouldBindUriConflictCheck(req, ["id"])` 且 `req.Id = "999"`（来自 JSON body），路径变量 `id=123`
- **THEN** MUST 返回错误，错误信息包含路径变量名、body 值和 URI 值

#### Scenario: 非 struct 指针
- **WHEN** 调用 `ShouldBindUriConflictCheck(nonStruct, pathVars)`
- **THEN** MUST 返回 nil（不做任何操作）
