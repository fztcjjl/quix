### Requirement: Transport Server interface
quix 框架 SHALL 在 `core/transport/` 包中定义 `Server` 接口，作为所有服务类型的抽象。

#### Scenario: Interface method signatures
- **WHEN** 开发者查看 Server 接口定义
- **THEN** 接口 SHALL 包含以下方法签名：
  - `Start() error`
  - `Stop(ctx context.Context) error`

### Requirement: HTTP Server implementation
框架 SHALL 在 `core/transport/http/server/` 包中提供 `Server` 实现 `Server` 接口，内部嵌入 `*gin.Engine`。

#### Scenario: Server satisfies Server interface
- **WHEN** 创建 `server.NewServer(server.WithAddr(":8080"))`
- **THEN** 返回值 MUST 实现 `transport.Server` 接口

#### Scenario: Access underlying gin.Engine
- **WHEN** 访问 `server.Engine`（嵌入字段）
- **THEN** MUST 返回 `*gin.Engine`，用户可直接操作 Gin 原生 API

### Requirement: App manages HTTP and RPC servers
App SHALL 持有 `httpServer *qhttp.Server` 和 `rpcServer transport.Server`，不支持通用多 server。

#### Scenario: Config-driven server creation
- **WHEN** 配置中有 `http.addr` 或 `http.port`
- **THEN** MUST 创建并启动 HTTP Server

#### Scenario: RPC config triggers RPC server
- **WHEN** 配置中有 `rpc.addr`
- **THEN** MUST 创建并启动 RPC Server（未来实现）

#### Scenario: Default to HTTP when no config
- **WHEN** 配置中既无 `http` 也无 `rpc` 配置
- **THEN** MUST 默认创建 HTTP Server，监听 `:8080`

#### Scenario: Both configured
- **WHEN** 配置中同时有 `http` 和 `rpc` 配置
- **THEN** MUST 同时创建并启动 HTTP Server 和 RPC Server

#### Scenario: Shutdown stops all servers
- **WHEN** 调用 `app.Shutdown(ctx)`
- **THEN** MUST 先关闭 RPC Server，再关闭 HTTP Server

### Requirement: App convenience methods proxy to HTTP Server
App SHALL 提供 GET/POST/PUT/DELETE/PATCH/GROUP/USE 等方法，直接代理到 `httpServer`，无需类型断言。

#### Scenario: Route registration via App
- **WHEN** 用户调用 `app.GET("/hello", handler)`
- **THEN** MUST 注册 GET 路由到 HTTP Server，行为与 Gin 一致

#### Scenario: Middleware via App
- **WHEN** 用户调用 `app.Use(middleware)`
- **THEN** MUST 挂载中间件到 HTTP Server

#### Scenario: Route group via App
- **WHEN** 用户调用 `app.Group("/api")`
- **THEN** MUST 返回 `*gin.RouterGroup`

### Requirement: Graceful shutdown with signal handling
App.Run() SHALL 自动监听 SIGINT 和 SIGTERM 信号，触发所有 Server 的优雅关闭。

#### Scenario: SIGINT triggers graceful shutdown
- **WHEN** 服务运行中收到 SIGINT
- **THEN** MUST 停止接受新请求，等待现有请求完成后关闭所有 Server

#### Scenario: Shutdown timeout
- **WHEN** 优雅关闭超过 5 秒
- **THEN** MUST 强制关闭所有 Server

#### Scenario: Startup and shutdown logs
- **WHEN** 服务启动/关闭
- **THEN** MUST 通过 Logger 输出日志

### Requirement: Config integration for server address
App SHALL 在 `New()` 中从 Config 读取服务地址，通过 `qhttp.WithAddr()` 传递给 HTTP Server。

#### Scenario: Address from http.addr config
- **WHEN** Config 中 `http.addr` 为 `:3000`
- **THEN** MUST 在 `:3000` 端口启动 HTTP Server

#### Scenario: Address from http.port config
- **WHEN** Config 中 `http.port` 为 `3000` 且 `http.addr` 为空
- **THEN** MUST 在 `:3000` 端口启动 HTTP Server

#### Scenario: Default port when no config
- **WHEN** Config 中无 `http.addr` 和 `http.port`
- **THEN** MUST 使用默认端口 `8080`

### Requirement: WithHttpServer and WithRpcServer options
App SHALL 提供 `WithHttpServer(*qhttp.Server)` 和 `WithRpcServer(transport.Server)` Option 函数。

#### Scenario: Custom HTTP server injection
- **WHEN** 使用 `quix.New(quix.WithHttpServer(s))`
- **THEN** MUST 使用注入的 HTTP Server，不创建默认 Server

#### Scenario: Custom RPC server injection
- **WHEN** 使用 `quix.New(quix.WithRpcServer(s))`
- **THEN** MUST 使用注入的 RPC Server

### Requirement: Handler uses gin.Context directly
Handler 函数 SHALL 直接使用 `*gin.Context`，不封装 quix.Context。

#### Scenario: Gin compatibility
- **WHEN** 使用 Gin 生态的第三方中间件或工具
- **THEN** MUST 完全兼容，无需适配

### Requirement: HTTP service example
框架 SHALL 在 `examples/` 中提供完整可运行的 HTTP 服务示例。

#### Scenario: Basic HTTP server example
- **WHEN** 开发者查看 `examples/http/main.go`
- **THEN** SHALL 演示完整 HTTP 服务（路由、中间件、启动、优雅关闭）

#### Scenario: Example is runnable
- **WHEN** 执行 `go run examples/http/main.go`
- **THEN** MUST 启动 HTTP 服务，可通过 curl 访问

### Requirement: Handler wrapper
框架 SHALL 在 `core/transport/http/server/` 包中提供 `Handler()` 函数，将 `func(c *gin.Context) error` 转换为 `gin.HandlerFunc`，自动处理错误响应。

#### Scenario: Handler returns nil
- **WHEN** handler 返回 nil
- **THEN** 请求正常继续，不设置 app_error

#### Scenario: Handler returns *Error
- **WHEN** handler 返回 `errors.NotFound("user_not_found", "用户不存在")`
- **THEN** MUST 将 Error 存入 `c.Get("app_error")`，并设置 HTTP status 为 Error.StatusCode（404）

#### Scenario: Handler returns standard error
- **WHEN** handler 返回 `fmt.Errorf("db connection failed")`
- **THEN** MUST 自动包装为 `*Error{Code: "internal_error", Message: "db connection failed", StatusCode: 500}`，存入 app_error，HTTP status 为 500

#### Scenario: Handler wrapper prevents subsequent handlers
- **WHEN** handler 返回非 nil error
- **THEN** 后续 handler MUST 不执行

### Requirement: ResponseMiddleware
框架 SHALL 提供 ResponseMiddleware，统一格式化错误响应。

#### Scenario: Error response format
- **WHEN** handler 中返回了错误且 ResponseMiddleware 已挂载
- **THEN** 响应体 MUST 为 `{"error": {...}}`，HTTP status 为 Error.StatusCode

#### Scenario: No error skips formatting
- **WHEN** handler 正常执行且未返回错误
- **THEN** ResponseMiddleware MUST 不写入任何响应（成功响应由 handler 直接处理）

### Requirement: HTTP Server default middleware
HTTP Server 创建时 SHALL 默认挂载 Recovery、RequestID 和 ResponseMiddleware 中间件，可通过 Option 关闭。

#### Scenario: Default middleware mounted
- **WHEN** 创建 HTTP Server 且未传入 `server.WithDefaultMiddleware(false)`
- **THEN** MUST 在 Engine 上挂载 Recovery、RequestID 和 ResponseMiddleware 中间件

#### Scenario: Default middleware order
- **WHEN** 创建 HTTP Server 且默认中间件未禁用
- **THEN** 中间件挂载顺序 MUST 为 Recovery → RequestID → ResponseMiddleware

#### Scenario: Disable default middleware
- **WHEN** 创建 HTTP Server 时传入 `server.WithDefaultMiddleware(false)`
- **THEN** MUST 不挂载任何默认中间件

#### Scenario: Default middleware recovers panic
- **WHEN** 使用默认中间件的 Server 处理请求时 handler 发生 panic
- **THEN** MUST 返回 HTTP 500，服务不崩溃
