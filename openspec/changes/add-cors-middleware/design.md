## Context

quix 框架的默认中间件链当前包含 Recovery、RequestID、Logging 和 Response 中间件。`openspec/specs/middleware/spec.md` 已定义了 CORS 中间件的规格（`middleware.CORS()` 和 `middleware.WithCORSConfig(cfg)`），但尚未实现。Gin 生态中 `github.com/gin-contrib/cors` 是标准的 CORS 解决方案，与框架已使用的 `gin-contrib/requestid` 同属一个组织，保持依赖风格一致。

当前默认中间件挂载顺序（`server.go:117-123`）：
```
Recovery → [otelgin] → RequestID → Logging → ResponseMiddleware
```

## Goals / Non-Goals

**Goals:**
- 实现 `middleware.CORS()` 和 `middleware.WithCORSConfig(cfg)` 便捷函数
- 将 CORS 加入默认中间件链，提供零配置开发体验
- 支持 Option 模式控制 CORS 行为（启用/禁用、自定义配置）

**Non-Goals:**
- 不实现自定义 CORS 逻辑，完全委托给 `gin-contrib/cors`

## Decisions

### 1. 使用 `gin-contrib/cors` 作为底层实现

**选择**: `github.com/gin-contrib/cors`
**替代方案**: 手写 CORS 中间件
**理由**: `gin-contrib/cors` 是 Gin 官方推荐的 CORS 解决方案，功能完善（支持预检缓存、正则匹配 Origin 等），且框架已使用同组织的 `gin-contrib/requestid`，保持依赖风格一致。

### 2. CORS 在默认中间件链中的位置

**选择**: `RequestID → [otelgin] → CORS → Recovery → Logging → ResponseMiddleware`

**理由**: RequestID 最前设置以确保所有中间件都能使用请求标识；otelgin 紧随其后尽早创建 span 捕获完整请求时长；CORS 在 Recovery 之前处理 OPTIONS 预检请求（short-circuit），避免对预检请求执行不必要的 Recovery/Logging/Response 处理；Recovery 保护后续业务逻辑。

### 3. 默认 CORS 配置

**选择**: 使用 `cors.Default()` 作为零配置默认值（允许所有 Origin、常见方法和头部）

**替代方案**: 更严格的默认配置（如仅允许同源）
**理由**: 框架定位为"薄封装"，`cors.Default()` 提供了开箱即用的开发体验。生产环境可通过 `server.WithCORSConfig(cfg)` 自定义。

### 4. CORS 控制机制

**选择**: 在 Server 级别和 App 级别同时提供 CORS 控制：
- Server 级：`WithCORS(enabled bool)` 和 `WithCORSConfig(cfg cors.Config)`
- App 级：`WithCORS(enabled bool)` 和 `WithCORSConfig(cfg cors.Config)`，通过桥接传递到 Server Option

**替代方案**: 仅在 Server 级提供 CORS 控制
**理由**: 与 Telemetry 的 Option 模式一致——App 级 Option 通过桥接传递到 Server 级，用户可以在不直接接触 Server 的情况下配置 CORS。`WithCORS(false)` 支持禁用 CORS。

## Risks / Trade-offs

- **[默认允许所有 Origin]** → 开发阶段方便，但生产环境需要用户主动配置。在文档和日志中提示生产环境应自定义 CORS 配置。
- **[破坏性变更]** → 启用默认中间件时自动挂载 CORS，改变了现有行为。对于已经手动处理 CORS 的项目，可能导致重复 CORS 头。缓解：CORS 中间件不冲突（后续中间件设置的头会覆盖前面的），且用户可通过自定义配置调整。
