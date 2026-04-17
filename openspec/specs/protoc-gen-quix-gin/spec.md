### Requirement: Plugin 入口
插件 SHALL 在 `cmd/protoc-gen-quix-gin/` 中提供 `main.go`，使用 protogen 框架作为 protoc 插件入口。

#### Scenario: 插件发现
- **WHEN** protoc 或 buf 使用 `--quix-gin_out` 或 `local: protoc-gen-quix-gin` 配置
- **THEN** MUST 找到 `protoc-gen-quix-gin` 可执行文件并执行

#### Scenario: 解析 proto 文件
- **WHEN** 插件接收到包含 service 定义的 proto 文件
- **THEN** MUST 遍历所有 `Generate == true` 的文件，解析 `google.api.http` 注解并生成 `_gin.go` 文件

### Requirement: paths=source_relative 支持
插件 SHALL 支持 `paths=source_relative` 参数，将生成文件放在与 proto 源文件对应的相对路径下。

#### Scenario: paths=source_relative
- **WHEN** 传入 `paths=source_relative` 参数
- **THEN** MUST 将 `xxx_gin.go` 生成到 proto 文件对应的目录（与 `protoc-gen-go` 的 `xxx.pb.go` 同目录）

#### Scenario: 默认路径
- **WHEN** 未传入 `paths=source_relative` 参数
- **THEN** MUST 使用 Go import path 构造输出文件路径

### Requirement: HTTP 方法支持
插件 SHALL 支持 GET、POST、PUT、DELETE、PATCH 五种 HTTP 方法。

#### Scenario: GET 方法
- **WHEN** `google.api.http` 注解使用 `get: "/path/{var}"`
- **THEN** MUST 生成 `r.GET()` 路由注册，handler 中使用 `ShouldBindQuery` + `ShouldBindUri`

#### Scenario: POST 方法 with body
- **WHEN** `google.api.http` 注解使用 `post: "/path" body: "*"`
- **THEN** MUST 生成 `r.POST()` 路由注册，handler 中使用 `ShouldBindJSON`

#### Scenario: PUT/DELETE/PATCH
- **WHEN** `google.api.http` 注解使用 put/delete/patch
- **THEN** MUST 生成对应的路由注册和 handler

### Requirement: body 字段映射
插件 SHALL 支持 `body: "field_name"` 将请求体映射到指定字段，并在生成阶段校验字段是否存在。

#### Scenario: body 指定字段名
- **WHEN** `google.api.http` 注解使用 `body: "user"` 且 `put: "/users/{user_id}"`
- **THEN** MUST 在 handler 中先绑定查询参数和路径参数，再使用 `ShouldBindJSON(req.FieldName)` 绑定请求体到指定字段

#### Scenario: body 字段不存在
- **WHEN** `body: "nonexistent"` 但输入 message 没有 `nonexistent` 字段
- **THEN** MUST 在生成阶段报告错误，指明字段名和所在 message 类型

### Requirement: additional_bindings 支持
插件 SHALL 支持 `additional_bindings` 为同一个 RPC 生成多条路由。

#### Scenario: 多条路由绑定
- **WHEN** `google.api.http` 注解包含 `additional_bindings`
- **THEN** MUST 为每个 binding 生成独立的 handler 函数和路由注册

### Requirement: google.protobuf.Empty 处理
插件 SHALL 识别 `google.protobuf.Empty` 作为空响应类型。

#### Scenario: Empty 返回
- **WHEN** RPC 方法的返回类型为 `google.protobuf.Empty`
- **THEN** 接口签名 MUST 只返回 `error`（无响应参数），handler MUST 返回 204 No Content

### Requirement: 路径模板转换
插件 SHALL 将 proto 路径模板 `{variable}` 转换为 Gin 风格 `:variable`。

#### Scenario: 路径转换
- **WHEN** proto 路径为 `/v1/users/{user_id}`
- **THEN** MUST 生成 Gin 路径 `/v1/users/:user_id`

### Requirement: 生成文件命名
插件 SHALL 将生成文件命名为 `xxx_gin.go`，其中 `xxx` 为 proto 文件名前缀。

#### Scenario: 文件命名
- **WHEN** proto 文件为 `greeter.proto`
- **THEN** MUST 生成 `greeter_gin.go`

### Requirement: Content-Type 协商
生成的 handler SHALL 支持基于 Accept header 的 Content-Type 协商。

#### Scenario: protobuf 响应
- **WHEN** 请求 Accept header 为 `application/x-protobuf`
- **THEN** MUST 使用 `c.ProtoBuf()` 返回 proto 二进制

#### Scenario: JSON 响应
- **WHEN** 请求 Accept header 不是 `application/x-protobuf`
- **THEN** MUST 使用 `c.JSON()` 返回 JSON

### Requirement: buf 兼容
插件 SHALL 完全兼容 buf 的 `buf generate` 命令。

#### Scenario: buf generate
- **WHEN** `buf.gen.yaml` 配置 `local: protoc-gen-quix-gin`
- **THEN** MUST 正常生成代码

### Requirement: ExtraImports 排序
插件 SHALL 对额外 import 路径排序后写入生成的代码，确保生成结果幂等。

#### Scenario: 多个额外 import
- **WHEN** 生成代码需要引用多个外部包（如 emptypb）
- **THEN** ExtraImports MUST 按字母序排列

#### Scenario: 重复生成稳定性
- **WHEN** 多次运行 protoc-gen-quix-gin 生成同一 proto 文件
- **THEN** 生成输出 MUST 完全一致（无 diff 噪音）

### Requirement: Register 函数直接使用传入的 RouterGroup
生成的 `RegisterXxxHTTPService` 函数 SHALL 直接使用传入的 `*gin.RouterGroup` 注册路由，不创建冗余的空路径子 group。

#### Scenario: 路由注册
- **WHEN** 调用 `RegisterGreeterHTTPService(g, svc)`
- **THEN** MUST 直接在 `g` 上注册路由（`g.GET()`、`g.POST()` 等），不创建 `g.Group("")`

### Requirement: 参数解析 fallback 注释
main.go SHALL 对 ParamFunc 和 fallback 参数解析分别添加注释，说明两者存在的原因。

#### Scenario: 注释完整性
- **WHEN** 查看 main.go 的参数解析代码
- **THEN** MUST 有注释说明为什么需要 fallback 解析（某些环境下 ParamFunc 未被调用）

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
