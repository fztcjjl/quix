## Context

quix 框架基于 Gin，目前开发者需要手工编写路由注册、请求绑定（路径参数、查询参数、请求体）、响应序列化和错误处理代码。通过 protobuf 作为 IDL，配合 `google.api.http` 注解定义 HTTP 路由，自定义 protoc 插件自动生成 Gin 路由注册代码。

参考 [zoo framework](https://github.com/iobrother/zoo) 的 `protoc-gen-zoo-http` 架构模式。

## Goals / Non-Goals

**Goals:**
- 从 proto service + `google.api.http` 注解自动生成 Gin 路由注册代码
- 生成的代码可编译、可直接使用，与 quix 现有错误处理（`core/errors`、`ResponseMiddleware`）兼容
- 支持 GET/POST/PUT/DELETE/PATCH、路径变量、查询参数、请求体绑定
- 兼容 buf 和 protoc

**Non-Goals:**
- 不生成 gRPC 服务端/客户端代码
- 不生成请求验证逻辑（由开发者实现）
- 不生成 OpenAPI/Swagger 文档
- 不支持 proto2 语法
- 不支持自定义 HTTP pattern（custom method）

## Decisions

### 1. 代码生成方式：text/template + embed

选择 text/template（embed 嵌入），模板集中可预览，适合多分支逻辑（body 策略、void 返回等）。

备选：protogen.P() 直接打印 — 对简单结构可行，但复杂分支（body "*" vs body "field"、void vs non-void）可读性差。

### 2. Context 包装器模式

生成代码使用 `runtime.Context{Context: ctx}` 包装 `*gin.Context`，增加 `SetError`/`GetError`/`ShouldBindUri`/`ShouldBindQuery`/`ShouldBindJSON` 方法。

备选：使用 quix 现有的 `qhttp.Handler()` 包装 — 但 proto 生成的接口签名是 `Method(ctx, req) (resp, error)`，不直接操作 gin.Context，需要运行时转换。

### 3. 请求绑定：form decoder + SetTagName("json")

Proto 生成的 Go 结构体只有 `json` tag，没有 `uri`/`form` tag。使用 `go-playground/form/v4` decoder 并设置 `SetTagName("json")` 来解码路径参数和查询参数。

### 4. 错误处理：SetError + ResponseMiddleware

Handler 调用 `c.SetError(err)` 后 return，由 quix 现有的 `ResponseMiddleware` 统一格式化错误响应。`*apperrors.Error` 使用其 StatusCode，其他 error 包装为 500。

### 5. Content-Type 协商

检查请求 Accept header，为 `application/x-protobuf` 返回 proto 二进制，否则返回 JSON。

### 6. 生成文件命名和接口命名

- 文件：`xxx_gin.go`（与 `xxx.pb.go` 同目录）
- 接口：`XxxHTTPService`
- 注册函数：`RegisterXxxHTTPService(g *gin.RouterGroup, svc XxxHTTPService)`
- Handler：`_Xxx_MethodN_HTTP_Handler(svc XxxHTTPService) gin.HandlerFunc`

### 7. paths=source_relative 支持

通过 `ParamFunc` 解析 `paths=source_relative` 参数，使用 `file.GeneratedFilenamePrefix` 构造文件路径，确保与 `protoc-gen-go` 输出到同一目录。

## Risks / Trade-offs

- **[proto 结构体无 uri/form tag]** → 使用 form decoder + SetTagName("json") 绕过，可能无法处理嵌套路径变量（如 `{book.id}`）的精确绑定
- **[Content-Type 协商]** → 目前基于 Accept header 判断，不考虑客户端 Accept 头缺失的情况（默认 JSON）
- **[插件命名]** → `protoc-gen-quix-gin` 命名较长，但确保不与社区插件冲突
