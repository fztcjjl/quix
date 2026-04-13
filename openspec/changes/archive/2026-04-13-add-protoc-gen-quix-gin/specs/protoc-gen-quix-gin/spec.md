## ADDED Requirements

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
插件 SHALL 支持 `body: "field_name"` 将请求体映射到指定字段。

#### Scenario: body 指定字段名
- **WHEN** `google.api.http` 注解使用 `body: "user"` 且 `put: "/users/{user_id}"`
- **THEN** MUST 在 handler 中先绑定查询参数和路径参数，再使用 `ShouldBindJSON(req.FieldName)` 绑定请求体到指定字段

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
