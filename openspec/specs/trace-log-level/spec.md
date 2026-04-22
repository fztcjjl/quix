### Requirement: Trace log level constant
`core/log/` 包 SHALL 定义 `LevelTrace Level = -1` 常量，位于 LevelDebug 之下。级别排序 MUST 为：LevelTrace < LevelDebug < LevelInfo < LevelWarn < LevelError。

#### Scenario: Trace level value
- **WHEN** 开发者引用 `log.LevelTrace`
- **THEN** 值 MUST 为 `-1`，且 `LevelTrace < LevelDebug`

### Requirement: Trace method on Logger interface
Logger 接口 SHALL 提供 `Trace(ctx context.Context, msg string, args ...any)` 方法。Trace MUST 仅在当前级别设置允许时输出日志（即 `level <= LevelTrace`）。

#### Scenario: Trace logs when level is LevelTrace
- **WHEN** 设置 `logger.SetLevel(LevelTrace)` 后调用 `logger.Trace(ctx, "msg")`
- **THEN** Trace 级别日志 MUST 正常输出

#### Scenario: Trace suppressed when level is LevelDebug
- **WHEN** 设置 `logger.SetLevel(LevelDebug)` 后调用 `logger.Trace(ctx, "msg")`
- **THEN** Trace 级别日志 MUST 被静默丢弃

### Requirement: Package-level Trace function
`core/log/` 包 SHALL 提供包级 `Trace(ctx context.Context, msg string, args ...any)` 函数，委托给全局默认 Logger。

#### Scenario: Global Trace function
- **WHEN** 调用 `log.Trace(ctx, "msg", "key", val)`
- **THEN** MUST 通过全局默认 Logger 输出 Trace 级别日志

### Requirement: slog adapter maps LevelTrace to slog.Level(-8)
slog adapter 实现 `Trace()` 时，MUST 将日志记录为 `slog.Level(-8)`，与 slog 生态的 Trace 级别约定一致。

#### Scenario: slog adapter Trace level mapping
- **WHEN** 使用 slog adapter 调用 `logger.Trace(ctx, "msg")`
- **THEN** 内部 MUST 使用 `slog.Level(-8)` 记录日志

### Requirement: zerolog adapter uses zerolog.TraceLevel
zerolog adapter 实现 `Trace()` 时，MUST 使用 `zerolog.TraceLevel`。

#### Scenario: zerolog adapter Trace level
- **WHEN** 使用 zerolog adapter 调用 `logger.Trace(ctx, "msg")`
- **THEN** 内部 MUST 使用 zerolog 的 Trace 级别输出日志

### Requirement: zap adapter uses appropriate debug level
zap adapter 实现 `Trace()` 时，MUST 使用低于 Debug 的级别输出日志。

#### Scenario: zap adapter Trace level
- **WHEN** 使用 zap adapter 调用 `logger.Trace(ctx, "msg")`
- **THEN** 日志 MUST 以 Trace 级别输出
