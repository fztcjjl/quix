## Context

quix 是一个基于 Gin 的 Golang HTTP 快速开发框架，目前处于初始阶段（仅有 go.mod）。Logger 是第一个需要实现的组件，所有后续组件（中间件、Metrics、Auth 等）都依赖日志能力。Logger 的接口设计模式也将为其他组件的集成方式建立先例。

## Goals / Non-Goals

**Goals:**
- 定义一个最小、稳定的 Logger 接口，作为框架所有组件的日志抽象
- 提供零依赖的默认实现（slog），开箱即用
- 通过适配器模式支持主流日志库（Zerolog、Zap），满足不同偏好
- 接口设计足够简单，slog 原生满足，无需额外包装

**Non-Goals:**
- 不实现日志轮转、文件输出等高级功能（由底层日志库自身处理）
- 不定义日志格式标准（由具体实现决定）
- 不实现日志级别动态调整（后续按需添加）
- 不集成结构化日志框架的特有功能（如 Zap 的 sampling、hooks）

## Decisions

### D1: 自建最小接口，而非直接使用 slog.Logger

**选择**: 定义独立的 `quix.Logger` 接口（5 个方法），而非直接暴露 `*slog.Logger`。

**替代方案**:
- 直接用 `slog.Logger` 作为类型：简单但锁死了 slog，Zap/Zerolog 用户不自然
- 模仿 `logr` 接口（klog 使用的）：功能更全但过于复杂

**理由**: 5 个方法的接口恰好覆盖框架需求，slog 原生满足零适配，Zap/Zerolog 只需薄适配器。接口是 quix 对外的契约，独立于任何第三方库。

### D2: args 使用 key-value 交替风格

**选择**: `Info(ctx, msg, "key1", val1, "key2", val2)` 而非 `Info(ctx, msg, Field{Key: "k1", Value: v1})`。

**替代方案**:
- 自定义 Field 类型：更安全但增加 API 复杂度
- `...any` 但内部自动检测类型：灵活但隐式行为多

**理由**: key-value 交替风格与 slog 一致，Go 社区逐渐习惯这种写法。

### D3: 组件统一放在 core 子包下

**选择**: `quix/core/logger/` 而非 `quix/logger/`。

**理由**: 框架会有多个基础组件（config、metrics、tracing、auth），统一放在 `core/` 下保持根包干净，结构清晰。后续组件将遵循 `quix/core/<component>/` 的统一路径规范。

### D4: 可选实现放在同包内

**选择**: `core/logger/slog.go`、`core/logger/zerolog.go`、`core/logger/zap.go` 与接口定义同级。

**替代方案**: `core/logger/adapter/zap/` 嵌套子包。

**理由**: 每个适配器文件很小（~50 行），不值得多建一层目录。

### D5: examples/ 目录规范

**选择**: 每个组件在 `examples/<component>/` 下提供可运行的示例代码（`go run` 直接执行）。

**替代方案**: 示例代码放在各组件包内（`core/logger/example_test.go`）：Go 惯例但分散，不适合库使用场景。

**理由**: quix 作为库供他人使用，独立的 `examples/` 目录更直观。每个示例一个文件，用户可直接 `go run` 验证。此规范适用于所有后续组件。

### D6: With 方法返回 Logger 接口

**选择**: `With(args ...any) Logger` 返回接口而非具体类型。

**理由**: 保持接口一致性，用户无需关心底层实现类型。

## Risks / Trade-offs

- **[接口稳定性]** Logger 接口一旦发布，修改就是 breaking change → 接口只保留 5 个核心方法，避免过度设计。需要新方法时通过扩展接口处理
- **[适配器维护成本]** 每新增一个日志库就需要写适配器 → 先只提供 Zap 和 Zerolog，其他按需添加。适配器代码量小（~50 行/个）
- **[性能]** 适配器引入一层间接调用 → 性能影响可忽略（一次方法调用 + slice 处理），日志本身是 I/O 密集型操作
