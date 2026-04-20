## 1. 依赖与基础设施

- [x] 1.1 添加 `github.com/gin-contrib/cors` 依赖（`go get github.com/gin-contrib/cors`）

## 2. CORS 中间件实现

- [x] 2.1 创建 `core/transport/http/server/middleware/cors.go`，实现 `CORS()` 函数（使用 `cors.Default()` 配置）
- [x] 2.2 实现 `WithCORSConfig(cfg cors.Config)` 函数，支持自定义 CORS 配置
- [x] 2.3 为 CORS 中间件编写单元测试

## 3. 默认中间件链集成

- [x] 3.1 在 `core/transport/http/server/server.go` 的 options 结构体中增加 `corsConfig` 字段
- [x] 3.2 增加 `WithCORSConfig(cfg cors.Config)` Server Option
- [x] 3.3 更新默认中间件挂载逻辑，在 Recovery 之后挂载 CORS 中间件（使用自定义配置或 `cors.Default()`）

## 4. 示例与文档

- [x] 4.1 在 `examples/middleware/` 下创建 CORS 示例，演示默认配置和自定义配置用法
- [x] 4.2 执行 `go fmt ./...` 格式化代码
- [x] 4.3 执行 `golangci-lint run ./...` 确保代码质量
