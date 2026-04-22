## Context

quix 框架的日志系统已有 4 个 adapter（slog、zerolog、zap、writer），统一的 Logger 接口，以及 HTTP 访问日志中间件。当前存在两类问题：

1. **正确性和性能缺陷**：4 个 adapter 各自持有 `level Level` 字段存在数据竞争（每个文件都有 TODO 注释）；zerolog adapter 每次日志调用分配 `map[string]any` 抵消了 zerolog 的零分配优势；4 个 adapter 的 level 管理代码完全重复。
2. **生产可观测性缺失**：访问日志缺少 span_id、latency_ms（数值）、query、user_agent、route（归一化路径）等标准字段；无 context 感知日志导致 handler 中的业务日志不携带 trace_id；无慢请求检测。

## Goals / Non-Goals

**Goals:**
- 修复 level 并发安全问题，消除 adapter 间的 level 管理代码重复
- 补全访问日志标准字段，提升 Loki/Grafana 可查询性
- 实现 context 感知日志，handler 自动携带 trace_id/span_id/request_id
- 新增 Trace 级别，支持配置驱动的日志级别控制
- 修复 zerolog adapter 的 map 分配性能问题

**Non-Goals:**
- 不替换 Logger 接口为 `*slog.Logger`（保持适配器无关抽象）
- 不实现自定义 slog.Handler（WithRequestLogger 中间件对所有 adapter 通用）
- 不实现敏感数据脱敏（业务关注点，非框架职责）
- 不实现生产采样、组件级级别控制、热更新日志级别
- 不实现日志行大小限制（基础设施层面处理）

## Decisions

### D1: AtomicLevel 替代重复 level 管理

**选择**：在 `logger.go` 中定义导出的 `AtomicLevel` struct，内含 `atomic.Int32`，提供 `Enabled()`/`SetLevel()`/`Level()` 方法。所有 adapter 持有 `*AtomicLevel` 指针，多个 adapter 可共享同一实例。

**理由**：当前 4 个 adapter 各自声明 `level Level` 并手动做 `if l.level > LevelXxx` 判断，完全重复。导出 `AtomicLevel` 让用户可以共享级别控制，同时与 `zap.AtomicLevel` / `slog.LevelVar` 命名惯例一致。

**替代方案**：使用接口组合（如 `levelSetter`/`levelChecker` 接口）— 过度抽象，增加不必要的间接层。

### D2: Trace 级别直接添加到 Logger 接口

**选择**：`LevelTrace Level = -1`，Logger 接口直接新增 `Trace()` 方法。

**理由**：保持 Level 常量的自然递增序列（-1, 0, 1, 2）。slog adapter 内部映射为 `slog.Level(-8)` 适配 slog 的 Trace 约定。直接加方法简洁明确。

**替代方案**：可选接口 `Tracer` — 增加类型断言复杂度，用户使用不便。

**BREAKING 影响**：外部 Logger 实现需新增 `Trace()` 方法。作为 minor 版本发布。

### D3: zerolog adapter 使用类型分发替代 map

**选择**：用 `switch v := args[i+1].(type)` 按 value 类型分发到 zerolog 的 `.Str()`/`.Int()`/`.Float64()`/`.Dur()` 等方法。

**理由**：`argsToMap()` 每次分配 `map[string]any`，与 zerolog 零分配设计矛盾。类型分发在常见类型上零分配，仅在遇到 `default`（`Interface()`）时才有反射。

### D3.5: 开发环境启用 Caller 字段

**选择**：在 `quix.New()` 中，当 `env == EnvDev` 时，默认 zerolog logger 使用 `CallerWithSkipFrameCount(4)` 启用 caller 字段。生产环境不启用。

**理由**：开发环境 caller 有助于快速定位日志来源。`CallerWithSkipFrameCount(4)` 跳过 zerolog hook 基础设施（2 帧）+ zerologLogger adapter（1 帧）+ log.Info 包级函数（1 帧），使 caller 指向用户代码。生产环境不添加 caller 以减少日志体积和性能开销。用户通过 `WithLogger()` 自定义 logger 时不受影响。

**替代方案**：全局设置 `zerolog.CallerSkipFrameCount` — 影响所有 zerolog 实例，不够隔离。

### D4: Context 感知日志通过中间件 + IntoContext/FromContext

**选择**：`log.NewContext(ctx, childLogger)` 注入 context，`log.FromContext(ctx)` 提取。`WithRequestLogger` 中间件在请求链路早期创建携带 trace_id/span_id/request_id 的 child logger 注入 context。

**理由**：Go 标准模式（`ctx.Value`），简单通用，对所有 adapter 透明。

**替代方案**：自定义 slog.Handler 自动注入 trace_id — 仅 slog adapter 受益，zerolog/zap 需要各自实现。

### D5: WithRequestLogger 在中间件链中的位置

**选择**：`requestid → [otelgin] → WithRequestLogger → Recovery → CORS → Logging → Response`

**理由**：WithRequestLogger 需要在 requestid（提供 request_id）和 otelgin（提供 trace context）之后运行，在业务 handler 之前运行。Recovery 之前确保 panic 日志也携带 context 信息。

### D6: span_id 通过函数变量注入中间件

**选择**：与现有 `ExtractTraceID` 模式对称，新增 `ExtractSpanID func(ctx context.Context) string`。`telemetry.ExtractSpanID()` 负责实际提取。

**理由**：保持 middleware 包不直接 import otel 包的解耦原则。

## Risks / Trade-offs

- **[BREAKING] Logger 接口变更** → 外部 Logger 实现需同步添加 `Trace()` 方法。通过 minor 版本号管理。
- **[性能] zerolog 类型分发 default 分支** → 非基础类型（自定义 struct）仍走 `Interface()` 产生分配。可接受，绝大多数日志使用 string/int/float/error 等基础类型。
- **[兼容] WithRequestLogger 中间件插入默认链** → 现有依赖 `c.Request.Context()` 的代码行为不变，但下游可通过 `log.FromContext(ctx)` 获取增强 logger。完全向后兼容。
