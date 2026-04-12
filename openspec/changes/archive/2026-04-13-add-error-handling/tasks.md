## 1. Error 类型

- [x] 1.1 创建 `core/errors/errors.go`，定义 Error 结构体（Code、Message、Details any、StatusCode int `json:"-"`）
- [x] 1.2 Error 实现 error 接口
- [x] 1.3 编写 Error 类型测试（JSON 序列化、Details 可选、Details 任意类型、StatusCode 不序列化）

## 2. 预定义错误

- [x] 2.1 实现 BadRequest（400）、NotFound（404）、Unauthorized（401）、Internal（500）、Forbidden（403）函数
- [x] 2.2 编写预定义错误测试（默认 StatusCode 正确）

## 3. Handler 包装器（gin-wrapper）

- [x] 3.1 在 `core/transport/http/server/handler.go` 实现 `Handler(fn func(c *gin.Context) error) gin.HandlerFunc`
- [x] 3.2 返回 `*Error` 时设置 app_error 和对应 StatusCode
- [x] 3.3 返回普通 `error` 时包装为 500 内部错误
- [x] 3.4 编写 Handler 包装器测试（nil 返回、*Error 返回、普通 error 返回、中止后续 handler）

## 4. ResponseMiddleware

- [x] 4.1 创建 `core/transport/http/server/middleware/response.go`
- [x] 4.2 检测 app_error 并格式化 `{"error": {...}}` 响应
- [x] 4.3 编写 ResponseMiddleware 测试

## 5. 默认挂载

- [x] 5.1 在 server.go 默认中间件中添加 ResponseMiddleware（Recovery → RequestID → ResponseMiddleware）
- [x] 5.2 编写默认挂载测试（错误响应格式正确）

## 6. 验证与归档

- [x] 6.1 运行 `go fmt ./...` 和 `golangci-lint run ./...` 验证
- [x] 6.2 同步 specs 到 openspec/specs/
