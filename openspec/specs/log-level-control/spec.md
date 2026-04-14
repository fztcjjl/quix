### Requirement: Fatal log level
Logger 接口 SHALL 提供 `Fatal(ctx context.Context, msg string, args ...any)` 方法。Fatal MUST 记录 Error 级别日志后调用 `os.Exit(1)`。

#### Scenario: Fatal logs and exits
- **WHEN** 调用 `logger.Fatal(ctx, "critical error", "err", err)`
- **THEN** MUST 先记录 Error 级别日志，然后调用 `os.Exit(1)` 终止进程

#### Scenario: Fatal available via global function
- **WHEN** 调用 `log.Fatal(ctx, "msg")`
- **THEN** MUST 通过全局默认 Logger 执行 Fatal 行为

### Requirement: Log level type
`core/log/` 包 SHALL 定义 `Level` 类型，包含以下级别常量：`LevelDebug`、`LevelInfo`、`LevelWarn`、`LevelError`。

#### Scenario: Level constants
- **WHEN** 开发者查看 `core/log` 包导出的 Level 类型
- **THEN** MUST 包含 `LevelDebug`、`LevelInfo`、`LevelWarn`、`LevelError` 常量

### Requirement: SetLevel method
Logger 接口 SHALL 提供 `SetLevel(level Level)` 方法，允许运行时动态调整日志输出级别。低于设定级别的日志 MUST 被静默丢弃。

#### Scenario: SetLevel to Warn suppresses Info and Debug
- **WHEN** 调用 `logger.SetLevel(LevelWarn)` 后调用 `logger.Info(ctx, "msg")`
- **THEN** Info 级别日志 MUST 被丢弃，不产生输出

#### Scenario: SetLevel to Debug allows all levels
- **WHEN** 调用 `logger.SetLevel(LevelDebug)` 后调用 `logger.Debug(ctx, "msg")`
- **THEN** Debug 级别日志 MUST 正常输出

#### Scenario: SetLevel available via global function
- **WHEN** 调用 `log.SetLevel(log.LevelWarn)`
- **THEN** 全局默认 Logger MUST 调用其 `SetLevel` 方法
