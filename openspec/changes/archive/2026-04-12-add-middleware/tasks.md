## 1. Recovery 中间件

- [x] 1.1 创建 `middleware/recovery.go`，实现 Recovery() gin.HandlerFunc
- [x] 1.2 捕获 panic，通过 `log.Error()` 输出堆栈，返回 500
- [x] 1.3 编写 Recovery 中间件测试

## 2. RequestID 中间件

- [x] 2.1 添加 `github.com/gin-contrib/requestid` 依赖
- [x] 2.2 创建 `middleware/requestid.go`，封装 `middleware.RequestID()` 便捷函数
- [x] 2.3 编写 RequestID 中间件测试

## 3. CORS 中间件

- [x] 3.1 添加 `github.com/gin-contrib/cors` 依赖
- [x] 3.2 创建 `middleware/cors.go`，封装 `middleware.CORS()` 和 `middleware.WithCORSConfig()`
- [x] 3.3 编写 CORS 中间件测试

## 4. 默认中间件挂载

- [x] 4.1 新增 `WithDefaultMiddleware(bool)` Option（quix 包）
- [x] 4.2 新增 `WithDefaultMiddleware(bool)` Option（server 包）
- [x] 4.3 NewServer 中根据配置挂载 Recovery + RequestID
- [x] 4.4 编写默认中间件挂载测试

## 5. 示例代码

- [x] 5.1 编写 `examples/middleware/recovery/main.go`

## 6. 验证与归档

- [x] 6.1 运行 `go fmt ./...` 和 `golangci-lint run ./...` 验证
- [x] 6.2 同步 specs 到 openspec/specs/
