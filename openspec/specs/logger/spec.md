### Requirement: Unified Logger interface
quix 框架 SHALL 在 `core/log/` 包中定义最小化的 Logger 接口，包含 5 个方法：Info、Error、Warn、Debug、With。所有框架内部组件 MUST 通过此接口输出日志。

#### Scenario: Interface method signatures
- **WHEN** 开发者查看 Logger 接口定义
- **THEN** 接口 SHALL 包含以下方法签名：
  - `Info(ctx context.Context, msg string, args ...any)`
  - `Error(ctx context.Context, msg string, args ...any)`
  - `Warn(ctx context.Context, msg string, args ...any)`
  - `Debug(ctx context.Context, msg string, args ...any)`
  - `With(args ...any) Logger`

### Requirement: Global default Logger
`core/log/` 包 SHALL 提供全局默认 Logger 实例和包级日志函数，支持开箱即用。初始默认值 MUST 为 noopLogger（所有方法为空操作，不产生 panic）。

#### Scenario: Package-level logging functions
- **WHEN** 开发者在项目任意位置调用 `log.Info(ctx, "msg", "key", val)`
- **THEN** MUST 通过全局默认 Logger 输出日志，无需获取 App 实例或注入 Logger

#### Scenario: Global noopLogger before App creation
- **WHEN** 在 `quix.New()` 调用之前使用 `log.Info(ctx, "msg")`
- **THEN** SHALL 不产生 panic，不输出任何日志

#### Scenario: SetDefault replaces global logger
- **WHEN** 调用 `log.SetDefault(customLogger)`
- **THEN** 全局默认 Logger MUST 替换为 `customLogger`，后续包级函数调用 MUST 委托给 `customLogger`

### Requirement: App sets global default Logger automatically
`App.New()` SHALL 在创建 Logger 后自动调用 `log.SetDefault()` 设置全局默认。

#### Scenario: App.New sets global default
- **WHEN** 用户调用 `quix.New()` 且未传入 `WithLogger()`
- **THEN** MUST 自动调用 `log.SetDefault()` 将 zerolog Logger 设为全局默认

#### Scenario: WithLogger sets global default
- **WHEN** 用户调用 `quix.New(quix.WithLogger(customLogger))`
- **THEN** MUST 自动调用 `log.SetDefault()` 将自定义 Logger 设为全局默认

### Requirement: Zerolog default implementation
框架 SHALL 使用 Zerolog 作为默认 Logger 实现。当用户未指定 Logger 时，App MUST 使用 Zerolog 实现。

#### Scenario: Default logger without configuration
- **WHEN** 用户调用 `quix.New()` 且未传入 `quix.WithLogger()` 选项
- **THEN** 框架 MUST 使用 Zerolog 默认实现作为 App 的 Logger

#### Scenario: Zerolog implementation satisfies Logger interface
- **WHEN** 使用 `log.NewZerolog(zeroLogInstance)` 创建 Logger
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

### Requirement: slog adapter implementation
框架 SHALL 提供可选的 slog 适配器实现。slog 实现使用 Go stdlib `log/slog`，引入零外部依赖。

#### Scenario: slog implementation satisfies Logger interface
- **WHEN** 使用 `log.NewSlog()` 或 `log.NewSlog(sl)` 创建 Logger
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

#### Scenario: slog With returns new Logger
- **WHEN** 调用 `slogLogger.With("key", "value")`
- **THEN** MUST 返回一个新的 Logger 实例，后续日志调用自动携带该字段

### Requirement: Zerolog adapter implementation
框架 SHALL 提供可选的 Zerolog 适配器实现。Zerolog 实现引入 zerolog 依赖，仅在使用时导入。

#### Scenario: Zerolog implementation satisfies Logger interface
- **WHEN** 使用 `log.NewZerolog(zeroLogInstance)` 创建 Logger
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

#### Scenario: Zerolog With preserves fields
- **WHEN** 调用 `zerologLogger.With("key", "value")`
- **THEN** MUST 返回一个新的 Logger 实例，后续日志调用自动携带该字段

### Requirement: Zap adapter implementation
框架 SHALL 提供可选的 Zap 适配器实现。Zap 实现引入 zap 依赖，仅在使用时导入。

#### Scenario: Zap implementation satisfies Logger interface
- **WHEN** 使用 `log.NewZap(zapLoggerInstance)` 创建 Logger
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
- **WHEN** 调用 `log.Info(ctx, "msg", "method", "GET", "path", "/users")`
- **THEN** 日志输出 MUST 包含 `method=GET` 和 `path=/users` 字段

#### Scenario: Odd number of args handled gracefully
- **WHEN** 调用 `log.Info(ctx, "msg", "key1")`（奇数个 args）
- **THEN** SHALL 忽略多余的最后一个 key，不产生 panic

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
