## Context

quix 框架已完成 Logger 和 Config 两个基础组件。App 结构体目前只有 logger 和 config 字段，无法启动服务。Gin 已作为依赖存在于 go.mod 中。本次变更建立 Transport 层，让 App 成为真正的服务容器。

## Goals / Non-Goals

**Goals:**
- 定义 `transport.Server` 接口，作为所有服务类型的抽象
- 实现 `transport/http.HTTP Server`，封装 gin.Engine
- App 管理多个 Server 的生命周期（启动、优雅关闭、信号监听）
- App 提供便捷方法代理到默认 HTTP Server，保持简单用法不变
- 纯 HTTP 场景用法与直接用 Gin 一样简洁

**Non-Goals:**
- 不实现 RPC Server（目录结构预留 `core/transport/rpc/`）
- 不实现 HTTP Client（目录结构预留 `core/transport/http/client.go`）
- 不封装 `gin.Context`
- 不实现 HTTPS/TLS
- 不实现多端口监听

## Decisions

### D1: Transport 分层架构

**选择**: `core/transport/` 根目录定义接口，`http/` 子目录放实现。

**替代方案**:
- 直接 `core/server/http/`：层级不够清晰
- 每个协议一个独立顶层包 `core/httpserver/`：包太多

**理由**: "transport" 是传输层的准确语义。HTTP 和 RPC 都是传输协议，Client 和 Server 是协议下的角色。目录结构为将来扩展预留了清晰的命名空间。

### D2: App 管理多个 Server

**选择**: App 持有 `server transport.Server`（默认 HTTP）和 `servers []transport.Server`（额外 Server）。

**替代方案**:
- App 只持有一个 Server：无法同时跑 HTTP + RPC
- App 不持任何 Server，全部由用户注册：纯 HTTP 场景太啰嗦

**理由**: 99% 的场景是纯 HTTP，默认行为必须简单。但架构上必须支持多 Server，否则加 RPC 时要大改。

### D3: 便捷方法代理而非嵌入

**选择**: App 不嵌入 `*gin.Engine`，而是持有 `HTTP Server` 实例，通过方法代理。

**替代方案**:
- 嵌入 `*gin.Engine`：简单但无法扩展 RPC

**理由**: App 需要管理多种 Server 类型，嵌入只能有一个。代理方式保持了简洁 API，同时架构可扩展。

### D4: 优雅关闭集成信号监听

**选择**: `App.Run()` 内部自动监听 SIGINT/SIGTERM，触发所有 Server 优雅关闭。

**理由**: 合理默认值，零配置即可获得优雅关闭。

## Risks / Trade-offs

- **[代理方法维护]** App 需要代理 Gin 的路由方法 → 只需代理常用方法（GET/POST/PUT/DELETE/PATCH/GROUP/USE），其他通过 `HTTP Server.Engine()` 访问
- **[类型断言]** 便捷方法内部需要类型断言 `server.(*HTTP Server)` → 默认 Server 始终是 HTTP Server，不会失败
- **[目录预空]** `core/transport/rpc/` 和 `client.go` 预留但暂不实现 → 用 README 或 .gitkeep 标记即可
