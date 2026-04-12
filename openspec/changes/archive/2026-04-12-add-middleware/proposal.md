## Why

当前 HTTP Server 使用 `gin.New()` 创建，不带任何默认中间件。handler 中的 panic 会直接导致服务崩溃。API 服务需要基础的中间件能力，框架应内置提供开箱即用的中间件。quix 定位为薄封装，RequestID 和 CORS 直接使用成熟的 `gin-contrib` 库，Recovery 自行实现以集成框架 Logger。

## What Changes

- 新增 `middleware/` 包，提供内置中间件便捷函数
- 实现 Recovery 中间件：捕获 panic，通过 `log.Error()` 输出错误信息，返回 500
- 引入 `gin-contrib/requestid`，提供 `middleware.RequestID()` 便捷函数
- 引入 `gin-contrib/cors`，提供 `middleware.CORS()` 便捷函数
- HTTP Server 默认挂载 Recovery + RequestID 中间件
- 新增 `WithDefaultMiddleware(bool)` Option 控制是否挂载默认中间件
- 每个中间件提供可运行示例

## Capabilities

### New Capabilities
- `middleware`: 内置 Gin 中间件（Recovery 自实现、RequestID/CORS 封装 gin-contrib）及默认挂载机制

### Modified Capabilities
- `gin-wrapper`: App 默认挂载 Recovery + RequestID 中间件

## Impact

- 新增 `middleware/` 目录及中间件实现
- `go.mod`：新增 `github.com/gin-contrib/requestid`、`github.com/gin-contrib/cors` 依赖
- `quix.go`：New 中传递默认中间件配置
- `option.go`：新增 `WithDefaultMiddleware(bool)` Option
- `core/transport/http/server/server.go`：NewServer 接收默认中间件配置并挂载
- `examples/middleware/`：新增示例代码
