## MODIFIED Requirements

### Requirement: Unified Logger interface
quix 框架 SHALL 在 `core/log/` 包中定义最小化的 Logger 接口，包含 9 个方法：Trace、Info、Error、Warn、Debug、Fatal、With、SetLevel、Close。所有框架内部组件 MUST 通过此接口输出日志。

#### Scenario: Interface method signatures
- **WHEN** 开发者查看 Logger 接口定义
- **THEN** 接口 SHALL 包含以下方法签名：
  - `Trace(ctx context.Context, msg string, args ...any)`
  - `Info(ctx context.Context, msg string, args ...any)`
  - `Error(ctx context.Context, msg string, args ...any)`
  - `Warn(ctx context.Context, msg string, args ...any)`
  - `Debug(ctx context.Context, msg string, args ...any)`
  - `Fatal(ctx context.Context, msg string, args ...any)`
  - `With(args ...any) Logger`
  - `SetLevel(level Level)`
  - `Close() error`

### Requirement: Zerolog default implementation
框架 SHALL 使用 Zerolog 作为 quix App 的默认 Logger 实现。当用户未指定 Logger 时，`quix.New()` MUST 使用 Zerolog 实现。zerolog adapter MUST 使用类型分发（type switch）将 key-value 参数路由到 zerolog 的类型化方法（`.Str()`/`.Int()`/`.Float64()`/`.Dur()` 等），避免每次日志调用分配 map。

#### Scenario: Default logger without configuration
- **WHEN** 用户调用 `quix.New()` 且未传入 `quix.WithLogger()` 选项
- **THEN** 框架 MUST 使用 Zerolog 默认实现作为 App 的 Logger

#### Scenario: Zerolog implementation satisfies Logger interface
- **WHEN** 使用 `log.NewZerolog(zeroLogInstance)` 创建 Logger
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

#### Scenario: Zerolog adapter avoids map allocation for common types
- **WHEN** 使用 zerolog adapter 调用 `logger.Info(ctx, "msg", "name", "alice", "age", 30, "duration", time.Second)`
- **THEN** MUST 通过 zerolog 的 `.Str()`/`.Int()`/`.Dur()` 方法添加字段，不分配 `map[string]any`

#### Scenario: Zerolog adapter falls back to Interface for unknown types
- **WHEN** 使用 zerolog adapter 传入非基础类型的 value（如自定义 struct）
- **THEN** MUST 使用 zerolog 的 `.Interface()` 方法添加字段

### Requirement: slog adapter implementation
框架 SHALL 使用 slog 作为 `core/log` 包的全局默认 Logger 实现。slog 实现使用 Go stdlib `log/slog`，零外部依赖。

#### Scenario: slog as package-level default
- **WHEN** 在 `quix.New()` 调用之前使用 `log.Info(ctx, "msg")`
- **THEN** SHALL 使用 slog 默认实现输出日志到 stderr

### Requirement: Concurrent-safe level control via AtomicLevel
所有 adapter MUST 通过持有 `*AtomicLevel` 指针实现 level 控制。`AtomicLevel` MUST 使用 `atomic.Int32` 存储 level，保证 `SetLevel()` 和日志方法的并发安全。`AtomicLevel` MUST 提供导出方法：`Enabled(l Level) bool`、`SetLevel(l Level)`、`Level() Level`。

#### Scenario: Concurrent SetLevel during logging
- **WHEN** 一个 goroutine 调用 `logger.SetLevel(LevelWarn)` 同时另一个 goroutine 调用 `logger.Info(ctx, "msg")`
- **THEN** MUST 不产生 data race

#### Scenario: AtomicLevel shared across adapters
- **WHEN** 多个 Logger 实例共享同一个 `*AtomicLevel`，调用 `atomicLevel.SetLevel(LevelError)`
- **THEN** 所有 Logger MUST 立即受影响，只输出 Error 及以上级别日志

### Requirement: Level string representation and parsing
`Level` 类型 SHALL 提供 `String() string` 方法，返回小写级别名称（"trace"/"debug"/"info"/"warn"/"error"）。未识别值 MUST 返回 "unknown"。`core/log/` 包 SHALL 提供 `ParseLevel(s string) (Level, error)` 函数，支持大小写不敏感解析。

#### Scenario: Level.String() for known levels
- **WHEN** 调用 `log.LevelInfo.String()`
- **THEN** MUST 返回 `"info"`

#### Scenario: Level.String() for unknown level
- **WHEN** 调用 `log.Level(99).String()`
- **THEN** MUST 返回 `"unknown"`

#### Scenario: ParseLevel for valid levels
- **WHEN** 调用 `log.ParseLevel("INFO")` 或 `log.ParseLevel("info")`
- **THEN** MUST 返回 `(log.LevelInfo, nil)`

#### Scenario: ParseLevel for invalid levels
- **WHEN** 调用 `log.ParseLevel("invalid")`
- **THEN** MUST 返回 `(Level(0), error)`

### Requirement: Timestamp field in default logger output
默认 zerolog logger MUST 在每条日志中输出 `time` 字段（通过 zerolog 的 `.Timestamp()` 配置）。

#### Scenario: Timestamp field present in log output
- **WHEN** 使用默认 zerolog logger 输出日志
- **THEN** JSON 输出 MUST 包含 `time` 字段，值符合 RFC3339 格式

### Requirement: Caller field in dev environment default logger
开发环境（`env == EnvDev`）的默认 zerolog logger MUST 输出 `caller` 字段。生产环境 MUST 不输出 `caller` 字段。`caller` MUST 通过 `log.WithCaller()` ZerologOption 启用，在适配器内部通过 `findCaller()` 遍历调用栈，跳过 `runtime.`、`testing.`、`core/log/` 内部帧，返回第一个用户代码帧。这种方式对 `log.Info()` 和 `log.FromContext(ctx).Info()` 等所有调用路径均正确。

#### Scenario: Caller field in dev environment
- **WHEN** `quix.New()` 在开发环境创建 App 且未传入 `quix.WithLogger()`
- **THEN** 默认 logger MUST 在每条日志中输出 `caller` 字段，值为调用者文件名和行号

#### Scenario: No caller field in production
- **WHEN** `quix.New()` 在生产环境创建 App 且未传入 `quix.WithLogger()`
- **THEN** 默认 logger MUST 不输出 `caller` 字段

#### Scenario: Caller points to user code via package-level function
- **WHEN** 用户通过 `log.Info(ctx, "msg")` 调用日志
- **THEN** `caller` 字段 MUST 指向用户代码位置，而非框架内部文件

#### Scenario: Caller points to user code via FromContext
- **WHEN** 用户通过 `log.FromContext(ctx).Info(ctx, "msg")` 调用日志（WithRequestLogger 注入的 context logger）
- **THEN** `caller` 字段 MUST 指向用户代码位置，而非框架内部文件

#### Scenario: With preserves callerEnabled
- **WHEN** 调用 `zerologLogger.With("key", "value")` 创建 child logger
- **THEN** child logger MUST 继承 `callerEnabled` 配置，后续日志调用也输出 `caller` 字段
