## Why

quix 框架目前只有 Logger 和 Config 两个基础组件，App 结构体还无法启动 HTTP 服务。Transport 层是框架的核心，它将 App 变成一个真正可运行的服务，为后续的 Middleware、Metrics、Auth 等组件提供挂载点。通过 `core/transport/` 目录设计，App 可以同时管理 HTTP 和将来的 RPC 服务，这是 quix 从"工具包"变成"框架"的关键一步。

## What Changes

- 新增 `core/transport/` 包，定义 `Server` 接口（Start/Stop）
- 新增 `core/transport/http/` 包，实现 Server（封装 gin.Engine）
- App 结构体管理 Server 生命周期，默认使用 HTTP Server
- App 便捷方法（GET/POST/Use 等）代理到默认 HTTP Server
- 实现 `App.Run(addr)` 启动所有 Server，集成优雅关闭（信号监听 + 5 秒超时）
- 实现 `App.AddServer()` 支持添加额外 Server（为将来 RPC 预留）
- 支持通过 Config 读取服务端口
- 不封装 `gin.Context`，直接使用 Gin 原生 Context
- 在 `examples/` 中提供完整可运行的 HTTP 服务示例

## Capabilities

### New Capabilities
- `transport`: Server/Client 接口定义、HTTP Server 实现（Gin 封装）、服务生命周期管理

### Modified Capabilities

## Impact

- 新增 `core/transport/` 和 `core/transport/http/` 子包
- `quix.go` 重构：App 持有 `transport.Server`，代理方法到 HTTPServer
- 后续所有组件（Middleware、Auth 等）将依赖此能力
- 将来 RPC 支持只需新增 `core/transport/rpc/`，无需修改 App
