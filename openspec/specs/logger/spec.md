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

### Requirement: Global default Logger
`core/log/` 包 SHALL 提供全局默认 Logger 实例和包级日志函数，支持开箱即用。初始默认值 MUST 为 `NewSlog()`（基于 Go stdlib slog，零外部依赖）。全局默认变量 MUST 使用 `atomic.Pointer[Logger]` 存储以保证并发安全。

#### Scenario: Package-level logging functions
- **WHEN** 开发者在项目任意位置调用 `log.Info(ctx, "msg", "key", val)`
- **THEN** MUST 通过全局默认 Logger 输出日志，无需获取 App 实例或注入 Logger

#### Scenario: Global slog default before App creation
- **WHEN** 在 `quix.New()` 调用之前使用 `log.Info(ctx, "msg")`
- **THEN** SHALL 使用 slog 默认实现输出日志到 stderr

#### Scenario: SetDefault replaces global logger
- **WHEN** 调用 `log.SetDefault(customLogger)`
- **THEN** 全局默认 Logger MUST 替换为 `customLogger`，后续包级函数调用 MUST 委托给 `customLogger`

#### Scenario: Concurrent access to global default Logger
- **WHEN** 多个 goroutine 并发调用包级日志函数（如 `log.Info`）和 `log.SetDefault`
- **THEN** SHALL 不产生 data race，MUST 通过 `atomic.Pointer[Logger]` 保证安全

### Requirement: App sets global default Logger automatically
`App.New()` SHALL 在创建 Logger 后自动调用 `log.SetDefault()` 设置全局默认。

#### Scenario: App.New sets global default
- **WHEN** 用户调用 `quix.New()` 且未传入 `WithLogger()`
- **THEN** MUST 自动调用 `log.SetDefault()` 将 zerolog Logger 设为全局默认

#### Scenario: WithLogger sets global default
- **WHEN** 用户调用 `quix.New(quix.WithLogger(customLogger))`
- **THEN** MUST 自动调用 `log.SetDefault()` 将自定义 Logger 设为全局默认

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

### Requirement: Zerolog adapter implementation
框架 SHALL 提供可选的 Zerolog 适配器实现。Zerolog 实现引入 zerolog 依赖，仅在使用时导入。

#### Scenario: Zerolog implementation satisfies Logger interface
- **WHEN** 使用 `log.NewZerolog(zeroLogInstance)` 创建 Logger
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

#### Scenario: Zerolog With preserves fields
- **WHEN** 调用 `zerologLogger.With("key", "value")`
- **THEN** MUST 返回一个新的 Logger 实例，后续日志调用自动携带该字段

### Requirement: slog adapter implementation
框架 SHALL 使用 slog 作为 `core/log` 包的全局默认 Logger 实现。slog 实现使用 Go stdlib `log/slog`，零外部依赖。

#### Scenario: slog as package-level default
- **WHEN** 在 `quix.New()` 调用之前使用 `log.Info(ctx, "msg")`
- **THEN** SHALL 使用 slog 默认实现输出日志到 stderr

#### Scenario: slog implementation satisfies Logger interface
- **WHEN** 使用 `log.NewSlog()` 或 `log.NewSlog(sl)` 创建 Logger
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

#### Scenario: slog With returns new Logger
- **WHEN** 调用 `slogLogger.With("key", "value")`
- **THEN** MUST 返回一个新的 Logger 实例，后续日志调用自动携带该字段

### Requirement: Zap adapter implementation
框架 SHALL 提供可选的 Zap 适配器实现。Zap 实现引入 zap 依赖，仅在使用时导入。

#### Scenario: Zap implementation satisfies Logger interface
- **WHEN** 使用 `log.NewZap(zapLoggerInstance)` 创建 Logger
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

#### Scenario: Zap With preserves fields
- **WHEN** 调用 `zapLogger.With("key", "value")`
- **THEN** MUST 返回一个新的 Logger 实例，后续日志调用自动携带该字段

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
开发环境（`env == EnvDev`）的默认 zerolog logger MUST 输出 `caller` 字段。生产环境 MUST 不输出 `caller` 字段。`caller` MUST 通过 `CallerWithSkipFrameCount(4)` 配置，跳过框架内部帧，指向用户代码调用位置。

#### Scenario: Caller field in dev environment
- **WHEN** `quix.New()` 在开发环境创建 App 且未传入 `quix.WithLogger()`
- **THEN** 默认 logger MUST 在每条日志中输出 `caller` 字段，值为调用者文件名和行号

#### Scenario: No caller field in production
- **WHEN** `quix.New()` 在生产环境创建 App 且未传入 `quix.WithLogger()`
- **THEN** 默认 logger MUST 不输出 `caller` 字段

#### Scenario: Caller points to user code
- **WHEN** 用户通过 `log.Info(ctx, "msg")` 调用日志
- **THEN** `caller` 字段 MUST 指向用户代码位置，而非框架内部文件

### Requirement: WithLogger option function
框架 SHALL 提供 `quix.WithLogger(logger Logger)` Option 函数，允许用户在创建 App 时注入自定义 Logger 实现。`WithLogger` MUST 同时调用 `log.SetDefault()` 同步更新全局默认 Logger。

#### Scenario: Inject custom logger via option
- **WHEN** 用户调用 `quix.New(quix.WithLogger(myLogger))`
- **THEN** App 的 Logger MUST 等于用户传入的 `myLogger`，且全局默认 Logger MUST 也被设置为 `myLogger`

#### Scenario: Custom logger implements interface
- **WHEN** 用户传入一个自定义结构体作为 Logger
- **THEN** 自定义结构体 MUST 实现完整的 Logger 接口（编译期检查）

### Requirement: KV key-value logging format
Logger 接口的 `args ...any` 参数 MUST 采用 key-value 对格式（key string, value any 交替传入）。三个 adapter MUST 统一处理异常参数。

#### Scenario: Correct key-value pairs
- **WHEN** 调用 `log.Info(ctx, "msg", "method", "GET", "path", "/users")`
- **THEN** 日志输出 MUST 包含 `method=GET` 和 `path=/users` 字段

#### Scenario: Odd number of args handled gracefully
- **WHEN** 调用 `log.Info(ctx, "msg", "key1")`（奇数个 args）
- **THEN** 三个 adapter MUST 统一静默 drop 最后一个孤立的 key，不产生 panic

#### Scenario: Non-string key handling
- **WHEN** 调用 `log.Info(ctx, "msg", 123, "value")`（key 不是 string）
- **THEN** 三个 adapter MUST 统一将非字符串 key 转为 `"key"` 字面量作为字段名，不产生 panic

#### Scenario: Multiple non-string keys preserved
- **WHEN** 调用 `log.Info(ctx, "msg", 123, "a", 456, "b")`（多个非字符串 key）
- **THEN** MUST 使用序号后缀区分（如 `key_0=a`, `key_1=b`），不丢失数据

#### Scenario: normalizeArgs does not mutate caller slice
- **WHEN** 调用任意 adapter 的任何日志方法
- **THEN** MUST 不修改调用方传入的 `args` slice 原始数据

#### Scenario: Normalize args fast path
- **WHEN** 调用日志方法且所有 key 均为 string、args 数量为偶数
- **THEN** MUST 不创建新 slice，直接使用原始 args（零分配）

### Requirement: Logger usage examples
quix 框架 SHALL 在 `examples/log/` 目录下提供可运行的示例代码，覆盖 Logger 的主要使用场景。示例代码 MUST 可通过 `go run` 直接执行。后续每个组件 MUST 遵循此规范提供示例。

#### Scenario: Default logger example
- **WHEN** 开发者查看 `examples/log/slog/main.go`
- **THEN** SHALL 演示使用包级函数 `log.Info()` 进行日志记录

#### Scenario: Zerolog example
- **WHEN** 开发者查看 `examples/log/zerolog/main.go`
- **THEN** SHALL 演示创建 Zerolog Logger 并通过包级函数使用

#### Scenario: Zap example
- **WHEN** 开发者查看 `examples/log/zap/main.go`
- **THEN** SHALL 演示创建 Zap Logger 并通过包级函数使用

#### Scenario: Example code is runnable
- **WHEN** 开发者在项目根目录执行 `go run examples/log/slog/main.go`
- **THEN** MUST 编译通过并正常输出日志，无需额外配置

### Requirement: File naming convention
`core/log/` 包中定义 Logger 接口和全局函数的文件 MUST 命名为 `logger.go`，对应的测试文件 MUST 命名为 `logger_test.go`。

#### Scenario: Interface definition location
- **WHEN** 开发者查找 Logger 接口定义
- **THEN** MUST 在 `core/log/logger.go` 中找到

#### Scenario: Test file location
- **WHEN** 开发者查找 Logger 接口的单元测试
- **THEN** MUST 在 `core/log/logger_test.go` 中找到 SetDefault、全局函数的测试

### Requirement: Unified argument normalization
所有 adapter MUST 使用 `normalizeArgs` 统一函数处理参数。`normalizeArgs` MUST 提供快路径：当所有 key 均为 string 且 args 数量为偶数时，直接返回原 slice 不产生分配。

#### Scenario: All adapters use normalizeArgs
- **WHEN** 查看任意 adapter 的日志方法实现
- **THEN** MUST 调用 `normalizeArgs` 处理参数，不使用 adapter 私有的参数转换函数

#### Scenario: normalizeArgs signature
- **WHEN** 查看 `normalizeArgs` 函数签名
- **THEN** MUST 为 `func normalizeArgs(args []any) []any`

### Requirement: Test mock only in test files
`MockLogger` MUST 只在各包的 `_test.go` 文件中定义，不暴露到生产 API。各包在测试中使用真实 Logger 实现或局部定义的 mock。

#### Scenario: quix_test.go uses real Logger
- **WHEN** `quix_test.go` 需要验证 `WithLogger` 注入
- **THEN** MUST 使用真实 Logger 实现（如 `log.NewSlog()`），不依赖 `log.MockLogger`

#### Scenario: logging_test.go defines local mock
- **WHEN** `logging_test.go` 需要捕获日志调用
- **THEN** MUST 在测试文件内局部定义 mock 实现

### Requirement: zap adapter Close method safely handles Sync panic
zap adapter 的 `Close()` SHALL 调用 `Sync()` 并通过 `recover` 防止 panic 传播到调用方。当底层 writer 已关闭导致 `Sync` panic 时，`Close()` SHALL 返回 nil 而非 panic。

#### Scenario: Sync panic does not propagate
- **WHEN** 底层 writer 已关闭，调用 zap adapter 的 `Close()`
- **THEN** 不发生 panic，`Close()` 正常返回
