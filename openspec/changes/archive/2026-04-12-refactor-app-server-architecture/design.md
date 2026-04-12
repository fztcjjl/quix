## Context

当前 App 通过 `server transport.Server`（默认 server）和 `servers []transport.Server`（额外 server）支持任意数量的 transport server。但 quix 的定位明确为 HTTP API 框架，未来仅整合 RPC，不需要通用多 server 能力。这导致了：

- `WithServer(transport.Server)` 缺乏类型安全，注入错误类型在编译期无法发现
- 路由代理方法需要 `(*qhttp.Server)` 类型断言
- `AddServer` 管理一个可变列表，增加不必要的复杂度
- `SetAddr()` 使 Server 可变，配置无法在构造时一次性传入
- 配置 key `server.addr`/`server.port` 在多 server 场景下含义模糊

## Goals / Non-Goals

**Goals:**
- App 显式持有 `httpServer` 和 `rpcServer`，类型安全
- 配置驱动服务创建：http/rpc 配置决定启动哪些服务
- 地址在 `New()` 中从配置读取并传递，消除 `SetAddr()`
- 配置 key 按服务类型命名：`http.addr`/`http.port`，`rpc.addr`

**Non-Goals:**
- RPC transport 实现（仅预留接口和配置）
- HTTP server 内部实现变更（仅调整 App 侧的集成方式）

## Decisions

### 1. App 结构体改为显式字段

```
// Before
server  transport.Server
servers []transport.Server

// After
httpServer *qhttp.Server
rpcServer  transport.Server
```

**理由**：quix 最多支持 HTTP + RPC 两种 server，无需泛型列表。`*qhttp.Server` 具体类型消除了路由代理中的类型断言。`rpcServer` 保持 `transport.Server` 接口，因为 RPC 实现尚未确定。

### 2. Option 函数拆分

```
// Before
WithServer(s transport.Server) Option

// After
WithHttpServer(s *qhttp.Server) Option
WithRpcServer(s transport.Server) Option
```

**理由**：具体类型提供编译期类型检查。移除 `AddServer()` 方法。

### 3. 配置驱动的服务创建

```
hasHttpConfig := config.String("http.addr") != "" || config.Int("http.port") != 0
hasRpcConfig  := config.String("rpc.addr") != ""

if hasHttpConfig || !hasRpcConfig → 创建 HTTP server
// RPC server 创建预留 TODO
```

**理由**：零配置时默认启动 HTTP（quix 是 HTTP 框架）。配置文件中声明了哪个服务就启动哪个，两者都配置则都启动。

### 4. 移除 SetAddr，地址在构造时传入

**理由**：Server 地址应该在创建时确定，不应在运行时变更。`New()` 中从配置读取地址，通过 `qhttp.WithAddr()` 传递给 `NewServer()`。

### 5. Run() 不再接受 addr 参数

**理由**：地址已由配置或 Option 在构造时确定。测试需要特定地址时使用 `WithHttpServer(qhttp.NewServer(qhttp.WithAddr(...)))` 。

## Risks / Trade-offs

- [Breaking API] → 所有使用 `WithServer`、`AddServer`、`Run(addr)`、`SetAddr()` 的代码需要迁移。作为早期框架，现在变更成本最低。
- [HTTP 硬编码] → `httpServer` 字段为具体类型 `*qhttp.Server`，未来如果需要替换 HTTP 实现（如切换到其他框架）需要修改 App 结构体。→ quix 基于 Gin 是框架核心决策，接受此绑定。
