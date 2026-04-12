## ADDED Requirements

### Requirement: Transport Server interface
quix 框架 SHALL 在 `core/transport/` 包中定义 `Server` 接口，作为所有服务类型的抽象。

#### Scenario: Interface method signatures
- **WHEN** 开发者查看 Server 接口定义
- **THEN** 接口 SHALL 包含以下方法签名：
  - `Start() error`
  - `Stop(ctx context.Context) error`

### Requirement: HTTP Server implementation
框架 SHALL 在 `core/transport/http/` 包中提供 `Server` 实现 `Server` 接口，内部封装 `*gin.Engine`。

#### Scenario: Server satisfies Server interface
- **WHEN** 创建 `http.NewServer(addr, opts...)`
- **THEN** 返回值 MUST 实现 `transport.Server` 接口

#### Scenario: Access underlying gin.Engine
- **WHEN** 调用 `server.Engine()`
- **THEN** MUST 返回 `*gin.Engine`，用户可直接操作 Gin 原生 API

### Requirement: App manages Server lifecycle
App SHALL 持有默认 Server（HTTP Server）和可选的额外 Server 列表。App.Run() MUST 启动所有 Server，App.Shutdown() MUST 关闭所有 Server。

#### Scenario: Run starts all servers
- **WHEN** 调用 `app.Run(":8080")` 且已通过 `app.AddServer()` 添加额外 Server
- **THEN** MUST 启动默认 HTTP Server 和所有额外 Server

#### Scenario: Shutdown stops all servers
- **WHEN** 调用 `app.Shutdown(ctx)`
- **THEN** MUST 关闭所有已启动的 Server

### Requirement: App convenience methods proxy to HTTP Server
App SHALL 提供 GET/POST/PUT/DELETE/PATCH/GROUP/USE 等方法，代理到默认 HTTP Server，保持简洁的 API。

#### Scenario: Route registration via App
- **WHEN** 用户调用 `app.GET("/hello", handler)`
- **THEN** MUST 注册 GET 路由到默认 HTTP Server，行为与 Gin 一致

#### Scenario: Middleware via App
- **WHEN** 用户调用 `app.Use(middleware)`
- **THEN** MUST 挂载中间件到默认 HTTP Server

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

### Requirement: Config integration
App.Run() SHALL 支持从 Config 读取服务端口。当 addr 为空时，SHALL 从 `server.port` 配置项读取。

#### Scenario: Address from config
- **WHEN** 调用 `app.Run("")` 且 Config 中 `server.port` 为 `3000`
- **THEN** MUST 在 `:3000` 端口启动

#### Scenario: Explicit address takes precedence
- **WHEN** 调用 `app.Run(":8080")`
- **THEN** MUST 使用 `:8080`，忽略配置

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
