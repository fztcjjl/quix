## ADDED Requirements

### Requirement: GET/DELETE 禁止声明 body
插件 SHALL 在代码生成阶段校验 HTTP 方法与 body 的兼容性：GET 和 DELETE 方法 MUST NOT 声明 body。

#### Scenario: GET 方法声明 body
- **WHEN** `google.api.http` 注解使用 `get: "/path"` 且 `body: "*"`
- **THEN** MUST 报告错误 `"<method>.<rpc>: GET must not have a body"` 并终止生成

#### Scenario: DELETE 方法声明 body
- **WHEN** `google.api.http` 注解使用 `delete: "/path"` 且 `body: "field"`
- **THEN** MUST 报告错误 `"<method>.<rpc>: DELETE must not have a body"` 并终止生成

#### Scenario: POST 方法声明 body（正常）
- **WHEN** `google.api.http` 注解使用 `post: "/path" body: "*"`
- **THEN** MUST 正常生成代码，不报错

### Requirement: body 星号路径变量同名冲突警告
当 `body: "*"` 且路径变量与 input message 的字段同名时，插件 SHALL 在生成阶段输出警告提示潜在冲突，但 MUST 继续正常生成代码。

#### Scenario: 路径变量与 message 字段同名
- **WHEN** `google.api.http` 注解使用 `post: "/agents/{id}" body: "*"` 且 input message 包含 `string id` 字段
- **THEN** MUST 输出警告，说明路径变量 `id` 与 body 字段 `id` 同名，可能存在冲突，并继续正常生成代码

#### Scenario: 路径变量与 message 字段不同名
- **WHEN** `google.api.http` 注解使用 `post: "/users/{user_id}/items" body: "*"` 且 input message 不包含 `user_id` 字段
- **THEN** MUST 正常生成代码，不输出警告

### Requirement: body 星号有路径变量的模板分支
当 `body: "*"` 且路径包含变量时，插件 SHALL 在生成的 handler 中先调用 `ShouldBindJSON(req)`，再调用 `ShouldBindUriConflictCheck(req, pathVars)` 进行运行时冲突检测和 URI 绑定。

#### Scenario: body 星号有路径变量
- **WHEN** `google.api.http` 注解使用 `post: "/agents/{agent_id}" body: "*"` 且 input message 无 `agent_id` 字段
- **THEN** 生成的 handler MUST 先调用 `ShouldBindJSON(req)`，再调用 `ShouldBindUriConflictCheck(req, []string{"agent_id"})`

#### Scenario: body 星号无路径变量
- **WHEN** `google.api.http` 注解使用 `post: "/agents" body: "*"`
- **THEN** 生成的 handler MUST 只调用 `ShouldBindJSON(req)`（行为不变）
