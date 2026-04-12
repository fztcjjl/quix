## Why

当前 App 采用通用的 `server transport.Server` + `servers []transport.Server` 设计，支持注入任意数量的 transport server。但 quix 定位明确：HTTP API 服务为主，未来整合 RPC。通用多 server 设计增加了不必要的复杂度（类型断言、列表管理），且 `WithServer` 缺乏类型安全。需要将 App 改为显式支持 HTTP 和 RPC 两种 server，配置驱动的服务启动，简化架构。

## What Changes

- **BREAKING** App 结构体字段：`server transport.Server` + `servers []transport.Server` → `httpServer *qhttp.Server` + `rpcServer transport.Server`
- **BREAKING** Option 函数：`WithServer(transport.Server)` → `WithHttpServer(*qhttp.Server)` + `WithRpcServer(transport.Server)`
- **BREAKING** 移除 `AddServer(s transport.Server)` 方法
- **BREAKING** 移除 `SetAddr()` 方法，地址在 `New()` 中从配置读取并传递给 Server
- **BREAKING** `Run()` 不再接受 `addr` 参数
- **BREAKING** 配置 key 变更：`server.addr` / `server.port` → `http.addr` / `http.port`，新增 `rpc.addr`
- 配置驱动的服务创建：有 `http` 配置启动 HTTP，有 `rpc` 配置启动 RPC，都没有时默认启动 HTTP
- 路由代理方法直接使用 `a.httpServer`，不再需要类型断言
- 新增 `HttpServer()` getter 方法

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `quix.go`：App 结构体重构，New()/Run()/Shutdown() 逻辑变更
- `option.go`：Option 函数签名变更
- `quix_test.go`：测试用例适配新 API
- `core/transport/http/server/server.go`：移除 `SetAddr()` 方法
- `examples/http/main.go`：`app.Run()` 调用方式变更
