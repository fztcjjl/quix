## Context

当前 quix 的可观测性命名存在两个重复/冗余问题：

1. `middleware/logging.go` 导出 `ExtractTraceID` / `ExtractSpanID` 函数变量，用于在运行时解耦 OTel 依赖。`core/telemetry/` 包有同名导出函数。`quix.New()` 在 telemetry 初始化后将 `telemetry.ExtractTraceID` 赋值给 `middleware.ExtractTraceID`。这种间接层在框架内部是不必要的——middleware 本身属于框架，可以直接 import telemetry 包。

2. `server.go` 持有 `telemetryServiceName` / `telemetryTracesEnabled` 字段和对应的 Option，`quix.New()` 在 telemetry 初始化后将 `telCfg.ServiceName` / `telCfg.TracesEnabled` 复制到 server options。这造成配置冗余：同一个值存在于 `telemetry.Config` 和 `server.options` 两处。

## Goals / Non-Goals

**Goals:**
- 消除 `ExtractTraceID` / `ExtractSpanID` 的重复定义，middleware 直接调用 telemetry 包
- 去除 server config 上的遥测冗余字段，otelgin 注入由 quix.New() 统一管理
- 保持现有行为不变（功能等价）

**Non-Goals:**
- 不改变 `TracesEnabled` / `MetricsEnabled` 命名
- 不改变 middleware 包的整体架构
- 不引入新的公开 API

## Decisions

### Decision 1: middleware 直接 import telemetry 包，删除函数变量

**选择**: 删除 `middleware.ExtractTraceID` / `middleware.ExtractSpanID` 导出变量，改为在 `AccessLog()` 和 `WithRequestLogger()` 内部直接调用 `telemetry.ExtractTraceID(ctx)` / `telemetry.ExtractSpanID(ctx)`。

**替代方案**: 保持函数变量但改为包内未导出变量——没有必要，因为 middleware 是框架内部包，不对外暴露。

**理由**: 原设计意图是 middleware 不 import OTel 包。但现在 `telemetry` 是框架内部的薄封装，middleware 依赖它是合理的。这消除了 `quix.New()` 中手动赋值函数变量的步骤。

### Decision 2: otelgin 由 quix.New() 挂载到 engine，去除 server 遥测字段

**选择**: 删除 `server.go` 的 `telemetryServiceName`、`telemetryTracesEnabled`、`WithTelemetryServiceName()`、`WithTelemetryTracesEnabled()`。otelgin middleware 在 `quix.New()` 中创建 server 之后、返回之前直接挂载到 `engine.Use()`。

**实现方式**: `quix.New()` 在调用 `qhttp.NewServer()` 创建 server 后，如果 telemetry 已启用且 traces 已启用，直接执行 `app.httpServer.Use(otelgin.Middleware(telCfg.ServiceName))`。otelgin 的挂载位置应在 RequestID 之后、Recovery 之前（与当前 server.go 中的顺序一致）。

**替代方案**: 将 telemetry.Config 传递给 server.NewServer() —— 会增加 server 对 telemetry 包的耦合，且 server 不需要了解遥测细节。

**理由**: otelgin 是应用级别的关注点（由 WithTelemetry 控制），不是 server 级别的配置。由 quix.New() 统一挂载更清晰。

### Decision 3: middleware 挂载位置调整

当前 server.go 的默认中间件链顺序为：`RequestID → [otelgin] → RequestLogger → Recovery → CORS → AccessLog → Response`。otelgin 需要在 RequestID 之后才能正确关联 request_id。

otelgin 由 quix.New() 挂载后，需要确保挂载顺序正确。实现方式：`quix.New()` 在创建 server 后调用 `app.httpServer.Use(otelgin.Middleware(...))` 时，otelgin 会追加到中间件链末尾。为保持语义等价，需要改为在 server 创建时让 defaultMiddleware 跳过 otelgin，由 quix.New() 在 server 创建后插入。

更简洁的做法：server.go 的 defaultMiddleware 链保留一个"slot"位置给 otelgin，通过一个未导出的 hook 或直接由 quix.New() 操作 engine。最简方案是 quix.New() 在 server 创建后、返回前直接调用 `app.httpServer.Use(otelgin.Middleware(...))`，并确认默认中间件链中不再包含 otelgin 相关逻辑。

## Risks / Trade-offs

- **[middleware 包新增对 telemetry 的 import]** → 可接受：两者都是框架内部包，且 telemetry 是轻量级接口封装，不会引入重依赖
- **[otelgin 挂载顺序可能变化]** → 需在实现时验证中间件链顺序与当前行为一致，通过测试确认
- **[BREAKING API 变更]** → `middleware.ExtractTraceID` 和 `qhttp.WithTelemetryServiceName` 被删除。迁移：直接调用 `telemetry.ExtractTraceID()`；otelgin 由 quix.New() 自动管理，无需手动配置
