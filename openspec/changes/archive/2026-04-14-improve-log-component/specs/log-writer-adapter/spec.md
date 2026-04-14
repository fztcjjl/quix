## ADDED Requirements

### Requirement: NewWriter adapter
`core/log/` 包 SHALL 提供 `NewWriter(w io.Writer) Logger` 函数，创建一个基于 `io.Writer` 的 Logger 实现。内部 MUST 使用 `slog.NewJSONHandler` 实现，零额外外部依赖。

#### Scenario: NewWriter creates valid Logger
- **WHEN** 调用 `log.NewWriter(os.Stdout)`
- **THEN** 返回值 MUST 实现 Logger 接口的所有方法

#### Scenario: NewWriter outputs JSON to writer
- **WHEN** 调用 `logger.Info(ctx, "msg", "key", "val")` 其中 logger 由 `NewWriter(buf)` 创建
- **THEN** `buf` 中 MUST 包含合法 JSON 格式的日志输出，包含 `msg` 和 `key` 字段

#### Scenario: NewWriter zero external dependencies
- **WHEN** 用户仅导入 `core/log` 包并使用 `NewWriter`
- **THEN** MUST 不引入 zerolog 或 zap 等外部依赖
