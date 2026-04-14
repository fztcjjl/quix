## ADDED Requirements

### Requirement: Close method for resource cleanup
Logger 接口 SHALL 提供 `Close() error` 方法，用于刷新缓冲区和释放资源。`Close` MUST 是幂等的——多次调用 MUST 不产生错误。

#### Scenario: Zap Close flushes buffer
- **WHEN** 调用 zapLogger 的 `Close()`
- **THEN** MUST 调用底层 zap.Logger 的 `Sync()` 方法，返回可能的错误

#### Scenario: Slog Close is no-op
- **WHEN** 调用 slogLogger 的 `Close()`
- **THEN** MUST 返回 nil，不产生错误

#### Scenario: Zerolog Close is no-op
- **WHEN** 调用 zerologLogger 的 `Close()`
- **THEN** MUST 返回 nil，不产生错误

#### Scenario: Close is idempotent
- **WHEN** 对同一个 Logger 实例多次调用 `Close()`
- **THEN** 第二次及后续调用 MUST 返回 nil，不产生 panic

#### Scenario: Close on default slog Logger
- **WHEN** 调用默认 slogLogger 的 `Close()`
- **THEN** MUST 返回 nil
