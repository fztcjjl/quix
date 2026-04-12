## ADDED Requirements

### Requirement: Recovery middleware
框架 SHALL 提供 Recovery 中间件，捕获 handler panic 防止服务崩溃，并通过框架 Logger 输出错误信息。

#### Scenario: Panic is recovered
- **WHEN** handler 中发生 panic
- **THEN** 中间件 MUST 捕获 panic，返回 HTTP 500，服务不崩溃

#### Scenario: Panic logged with stack trace
- **WHEN** handler 中发生 panic
- **THEN** MUST 通过 `log.Error()` 输出 panic 信息和堆栈

#### Scenario: Recovery usage
- **WHEN** 用户调用 `app.Use(middleware.Recovery())`
- **THEN** MUST 返回 `gin.HandlerFunc`

### Requirement: RequestID middleware
框架 SHALL 提供对 `gin-contrib/requestid` 的便捷封装，统一 `middleware.RequestID()` 调用风格。

#### Scenario: Generate RequestID
- **WHEN** 用户调用 `middleware.RequestID()`
- **THEN** MUST 返回基于 `gin-contrib/requestid` 的 `gin.HandlerFunc`

#### Scenario: RequestID in response header
- **WHEN** 使用 RequestID 中间件处理请求
- **THEN** 响应头 MUST 包含 `X-Request-ID`

### Requirement: CORS middleware
框架 SHALL 提供对 `gin-contrib/cors` 的便捷封装。

#### Scenario: Default CORS
- **WHEN** 用户调用 `middleware.CORS()`
- **THEN** MUST 使用 `cors.Default()` 配置（允许所有 Origin）

#### Scenario: Custom CORS configuration
- **WHEN** 用户调用 `middleware.WithCORSConfig(cfg)`
- **THEN** MUST 使用自定义 `cors.Config` 创建 CORS 中间件

#### Scenario: Preflight request
- **WHEN** 收到 OPTIONS 预检请求
- **THEN** MUST 返回正确的 CORS 响应头

### Requirement: Default middleware mounting
App SHALL 默认挂载 Recovery 和 RequestID 中间件到 HTTP Server。

#### Scenario: Default middleware mounted automatically
- **WHEN** 用户调用 `quix.New()` 未传入 `WithDefaultMiddleware(false)`
- **THEN** HTTP Server MUST 自动挂载 Recovery 和 RequestID 中间件

#### Scenario: Disable default middleware
- **WHEN** 用户调用 `quix.New(quix.WithDefaultMiddleware(false))`
- **THEN** HTTP Server MUST 不挂载任何默认中间件

### Requirement: Middleware usage examples
框架 SHALL 在 `examples/middleware/` 下提供可运行的示例。

#### Scenario: Recovery example
- **WHEN** 开发者查看 `examples/middleware/recovery/main.go`
- **THEN** SHALL 演示 Recovery 中间件捕获 panic 的效果

#### Scenario: Example is runnable
- **WHEN** 执行 `go run examples/middleware/recovery/main.go`
- **THEN** MUST 编译通过并正常启动
