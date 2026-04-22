## MODIFIED Requirements

### Requirement: Log level type
`core/log/` 包 SHALL 定义 `Level` 类型，包含以下级别常量：`LevelTrace`、`LevelDebug`、`LevelInfo`、`LevelWarn`、`LevelError`。常量值 MUST 为递增整数序列：LevelTrace=-1, LevelDebug=0, LevelInfo=1, LevelWarn=2, LevelError=3。

#### Scenario: Level constants ordering
- **WHEN** 开发者比较日志级别
- **THEN** MUST 满足 `LevelTrace < LevelDebug < LevelInfo < LevelWarn < LevelError`

#### Scenario: Level constant values
- **WHEN** 开发者检查 LevelTrace 和 LevelDebug 的值
- **THEN** LevelTrace MUST 为 -1，LevelDebug MUST 为 0

### Requirement: SetLevel method with atomic safety
Logger 接口 SHALL 提供 `SetLevel(level Level)` 方法，允许运行时动态调整日志输出级别。低于设定级别的日志 MUST 被静默丢弃。`SetLevel` MUST 通过 `atomic.Int32` 实现并发安全。

#### Scenario: SetLevel to Warn suppresses Info, Debug, and Trace
- **WHEN** 调用 `logger.SetLevel(LevelWarn)` 后调用 `logger.Trace(ctx, "msg")`、`logger.Debug(ctx, "msg")`、`logger.Info(ctx, "msg")`
- **THEN** Trace、Debug、Info 级别日志 MUST 全部被丢弃

#### Scenario: SetLevel to Trace allows all levels
- **WHEN** 调用 `logger.SetLevel(LevelTrace)` 后调用 `logger.Trace(ctx, "msg")`
- **THEN** Trace 级别日志 MUST 正常输出
