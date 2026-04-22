## ADDED Requirements

### Requirement: NewContext stores Logger in context
`core/log/` 包 SHALL 提供 `NewContext(ctx context.Context, l Logger) context.Context` 函数，将 Logger 存入 context。后续调用 `FromContext` MUST 返回存入的 Logger。

#### Scenario: Store and retrieve Logger
- **WHEN** 调用 `ctx := log.NewContext(background, customLogger)` 后调用 `log.FromContext(ctx)`
- **THEN** MUST 返回 `customLogger`

#### Scenario: Nested NewContext overwrites
- **WHEN** 调用 `log.NewContext(ctx, logger1)` 后再调用 `log.NewContext(ctx, logger2)`
- **THEN** `log.FromContext(ctx)` MUST 返回 `logger2`

### Requirement: FromContext extracts Logger or returns default
`core/log/` 包 SHALL 提供 `FromContext(ctx context.Context) Logger` 函数。当 context 中无 Logger 时，MUST 返回全局默认 Logger（`Default()`），永不返回 nil。

#### Scenario: Context without Logger returns default
- **WHEN** 调用 `log.FromContext(context.Background())`
- **THEN** MUST 返回全局默认 Logger，不返回 nil

#### Scenario: Context with Logger returns stored Logger
- **WHEN** 通过 `log.IntoContext` 存入 customLogger 后调用 `log.FromContext`
- **THEN** MUST 返回 customLogger

### Requirement: contextKey uses unexported type
IntoContext/FromContext 的 context key MUST 使用未导出类型 `contextKey struct{}`，避免与用户代码的 key 冲突。

#### Scenario: Key isolation
- **WHEN** 用户代码使用其他 context key 存值
- **THEN** MUST 不影响 `log.FromContext` 的行为
