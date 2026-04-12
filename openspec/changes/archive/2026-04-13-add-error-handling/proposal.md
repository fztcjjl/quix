## Why

当前 handler 中的错误响应格式不统一，每个开发者需要手动构建错误 JSON 结构。框架应提供统一的错误类型和响应机制，确保所有 API 错误响应格式一致：`{"error": {"code": "...", "message": "...", "details": ...}}`。应用层只需编写返回 `error` 的 handler 函数，框架自动处理错误响应格式化和 HTTP 状态码。request_id 由已有的 RequestID 中间件通过响应头注入。

## What Changes

- 新增 `core/errors/` 包，定义 `Error` 类型（Code、Message、Details any、StatusCode int）和预定义错误函数
- 新增 `Handler()` 包装器（放在 `core/transport/http/server/`），将 `func(c *gin.Context) error` 转为 gin.HandlerFunc，自动处理 `*Error` 和普通 `error`
- 新增 ResponseMiddleware（放在 `core/transport/http/server/middleware/`），检测错误并格式化统一响应
- 成功响应由用户通过 `c.JSON()` 直接处理，框架不干预

## Capabilities

### New Capabilities
- `errors`: 统一的错误类型、预定义错误函数

### Modified Capabilities
- `gin-wrapper`: 新增 Handler 包装器、ResponseMiddleware，默认挂载到 HTTP Server

## Impact

- 新增 `core/errors/` 目录
- 新增 `core/transport/http/server/handler.go`
- 新增 `core/transport/http/server/middleware/response.go`
- `core/transport/http/server/server.go`：默认中间件新增 ResponseMiddleware
