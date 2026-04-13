## Why

quix 框架目前需要手工编写 Gin 路由注册、请求绑定和响应序列化代码。通过自定义 protoc 插件 `protoc-gen-quix-gin`，可以从 protobuf service 定义 + `google.api.http` 注解自动生成 Gin 路由代码，实现 IDL 驱动的 API 开发，减少样板代码并统一请求处理模式。

## What Changes

- 新增 `cmd/protoc-gen-quix-gin/` 插件，解析 proto service 生成 `_gin.go` 文件
- 新增 `internal/protoc-gen-quix-gin/runtime/` 运行时包，提供 Context 包装器、请求绑定、错误处理
- 生成内容包括：服务接口定义、路由注册函数、handler 函数（含请求绑定 + Content-Type 协商）
- 支持 GET/POST/PUT/DELETE/PATCH、路径变量、查询参数、`body: "*"` 和 `body: "field"`、`additional_bindings`、`google.protobuf.Empty`（204）

## Capabilities

### New Capabilities
- `protoc-gen-quix-gin`: protoc 插件本身（代码生成逻辑、模板、路径解析）
- `protoc-gen-quix-gin-runtime`: 插件运行时（Context 包装器、ShouldBindUri/Query/JSON、SetError/GetError）

### Modified Capabilities

（无现有 spec 级别的行为变更）

## Impact

- **新增依赖**: `google.golang.org/genproto`、`go-playground/form/v4`
- **新增目录**: `cmd/protoc-gen-quix-gin/`、`internal/protoc-gen-quix-gin/runtime/`
- **新增示例**: `examples/proto-api/`（buf 配置 + proto + 生成代码 + 服务实现）
- **与现有系统集成**: 生成的 handler 使用 `runtime.Context{Context: ctx}` 包装器，通过 `c.SetError()` 与现有 `ResponseMiddleware` 配合
