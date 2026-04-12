## MODIFIED Requirements

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

## ADDED Requirements

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

## REMOVED Requirements

### Requirement: slog default implementation
**Reason**: 默认 Logger 已从 slog 改为 Zerolog，slog 保留为可选适配器
**Migration**: 使用 `quix.New()` 不需改动；显式使用 slog 时用 `log.NewSlog()` 创建

### Requirement: Logger usage examples (directory path)
**Reason**: 示例目录从 `examples/logger/` 迁移到 `examples/log/`
**Migration**: 运行示例时使用 `go run examples/log/slog/main.go` 等新路径
