## MODIFIED Requirements

### Requirement: Content-Type 协商
生成的 handler SHALL 支持基于 Accept header 的 Content-Type 协商。

#### Scenario: protobuf 响应
- **WHEN** 请求 Accept header 为 `application/x-protobuf`
- **THEN** MUST 使用 `c.ProtoBuf()` 返回 proto 二进制

#### Scenario: JSON 响应
- **WHEN** 请求 Accept header 不是 `application/x-protobuf`
- **THEN** MUST 使用 `c.JSON()` 返回 JSON

## MODIFIED Requirements

### Requirement: body 字段映射
插件 SHALL 支持 `body: "field_name"` 将请求体映射到指定字段，并在生成阶段校验字段是否存在。

#### Scenario: body 指定字段名
- **WHEN** `google.api.http` 注解使用 `body: "user"` 且 `put: "/users/{user_id}"`
- **THEN** MUST 在 handler 中先绑定查询参数和路径参数，再使用 `ShouldBindJSON(req.FieldName)` 绑定请求体到指定字段

#### Scenario: body 字段不存在
- **WHEN** `body: "nonexistent"` 但输入 message 没有 `nonexistent` 字段
- **THEN** MUST 在生成阶段报告错误，指明字段名和所在 message 类型

## ADDED Requirements

### Requirement: ExtraImports 排序
插件 SHALL 对额外 import 路径排序后写入生成的代码，确保生成结果幂等。

#### Scenario: 多个额外 import
- **WHEN** 生成代码需要引用多个外部包（如 emptypb）
- **THEN** ExtraImports MUST 按字母序排列

#### Scenario: 重复生成稳定性
- **WHEN** 多次运行 protoc-gen-quix-gin 生成同一 proto 文件
- **THEN** 生成输出 MUST 完全一致（无 diff 噪音）

## ADDED Requirements

### Requirement: Register 函数直接使用传入的 RouterGroup
生成的 `RegisterXxxHTTPService` 函数 SHALL 直接使用传入的 `*gin.RouterGroup` 注册路由，不创建冗余的空路径子 group。

#### Scenario: 路由注册
- **WHEN** 调用 `RegisterGreeterHTTPService(g, svc)`
- **THEN** MUST 直接在 `g` 上注册路由（`g.GET()`、`g.POST()` 等），不创建 `g.Group("")`

## ADDED Requirements

### Requirement: 参数解析 fallback 注释
main.go SHALL 对 ParamFunc 和 fallback 参数解析分别添加注释，说明两者存在的原因。

#### Scenario: 注释完整性
- **WHEN** 查看 main.go 的参数解析代码
- **THEN** MUST 有注释说明为什么需要 fallback 解析（某些环境下 ParamFunc 未被调用）
