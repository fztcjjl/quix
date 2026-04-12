## ADDED Requirements

### Requirement: Unified Logger interface
quix 框架 SHALL 定义一个最小化的 Logger 接口，包含 5 个方法：Info、Error、Warn、Debug、With。所有框架内部组件 MUST 通过此接口输出日志。

#### Scenario: Interface method signatures
- **WHEN** 开发者查看 Logger 接口定义
- **THEN** 接口 SHALL 包含以下方法签名：
  - `Info(ctx context.Context, msg string, args ...any)`
  - `Error(ctx context.Context, msg string, args ...any)`
  - `Warn(ctx context.Context, msg string, args ...any)`
  - `Debug(ctx context.Context, msg string, args ...any)`
  - `With(args ...any) Logger`

### Requirement: slog default implementation
框架 SHALL 提供基于 Go stdlib `slog` 的默认 Logger 实现。当用户未指定 Logger 时，App MUST 使用 slog 实现。slog 实现引入零外部依赖。

#### Scenario: Default logger without configuration
- **WHEN** 用户调用 `quix.New()` 且未传入 `quix.WithLogger()` 选项
- **THEN** 框架 MUST 使用 slog 默认实现作为 App 的 Logger

#### Scenario: slog implementation satisfies Logger interface
- **WHEN** 使用 `logger.NewSlog()` 创建 Logger
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

#### Scenario: slog With returns new Logger
- **WHEN** 调用 `slogLogger.With("key", "value")`
- **THEN** MUST 返回一个新的 Logger 实例，后续日志调用自动携带该字段

### Requirement: Zerolog adapter implementation
框架 SHALL 提供可选的 Zerolog 适配器实现。Zerolog 实现引入 zerolog 依赖，仅在使用时导入。

#### Scenario: Zerolog implementation satisfies Logger interface
- **WHEN** 使用 `logger.NewZerolog(zeroLogInstance)` 创建 Logger
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

#### Scenario: Zerolog With preserves fields
- **WHEN** 调用 `zerologLogger.With("key", "value")`
- **THEN** MUST 返回一个新的 Logger 实例，后续日志调用自动携带该字段

### Requirement: Zap adapter implementation
框架 SHALL 提供可选的 Zap 适配器实现。Zap 实现引入 zap 依赖，仅在使用时导入。

#### Scenario: Zap implementation satisfies Logger interface
- **WHEN** 使用 `logger.NewZap(zapLoggerInstance)` 创建 Logger
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

#### Scenario: Zap With preserves fields
- **WHEN** 调用 `zapLogger.With("key", "value")`
- **THEN** MUST 返回一个新的 Logger 实例，后续日志调用自动携带该字段

### Requirement: WithLogger option function
框架 SHALL 提供 `quix.WithLogger(logger Logger)` Option 函数，允许用户在创建 App 时注入自定义 Logger 实现。

#### Scenario: Inject custom logger via option
- **WHEN** 用户调用 `quix.New(quix.WithLogger(myLogger))`
- **THEN** App 的 Logger MUST 等于用户传入的 `myLogger`

#### Scenario: Custom logger implements interface
- **WHEN** 用户传入一个自定义结构体作为 Logger
- **THEN** 自定义结构体 MUST 实现完整的 Logger 接口（编译期检查）

### Requirement: KV key-value logging format
Logger 接口的 `args ...any` 参数 MUST 采用 key-value 对格式（key string, value any 交替传入），与 slog 的键值对风格一致。

#### Scenario: Correct key-value pairs
- **WHEN** 调用 `logger.Info(ctx, "msg", "method", "GET", "path", "/users")`
- **THEN** 日志输出 MUST 包含 `method=GET` 和 `path=/users` 字段

#### Scenario: Odd number of args handled gracefully
- **WHEN** 调用 `logger.Info(ctx, "msg", "key1")`（奇数个 args）
- **THEN** SHALL 忽略多余的最后一个 key，不产生 panic

### Requirement: Logger usage examples
quix 框架 SHALL 在 `examples/logger/` 目录下提供可运行的示例代码，覆盖 Logger 的主要使用场景。示例代码 MUST 可通过 `go run` 直接执行。后续每个组件 MUST 遵循此规范提供示例。

#### Scenario: Default slog example
- **WHEN** 开发者查看 `examples/logger/slog_example.go`
- **THEN** SHALL 演示使用默认 slog 实现（NewSlog）进行日志记录，包括各日志级别和 With 字段追加

#### Scenario: Zerolog example
- **WHEN** 开发者查看 `examples/logger/zerolog_example.go`
- **THEN** SHALL 演示创建 Zerolog Logger 并注入到框架，包括 Zerolog 特有的配置方式

#### Scenario: Zap example
- **WHEN** 开发者查看 `examples/logger/zap_example.go`
- **THEN** SHALL 演示创建 Zap Logger 并注入到框架，包括 Zap 特有的配置方式

#### Scenario: Example code is runnable
- **WHEN** 开发者在项目根目录执行 `go run examples/logger/slog_example.go`
- **THEN** MUST 编译通过并正常输出日志，无需额外配置
