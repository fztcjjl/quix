## Why

zap adapter 的 `Close()` 调用 `Sync()` 时，如果底层 writer 已关闭会 panic，用户调用 `Close()` 不应崩溃。同时各 adapter 的 `level` 字段存在并发读写 data race 风险，需要标记后续修复。此外 `examples/log/` 下的三个示例代码没有覆盖完整 Logger API，且结构不一致。

## What Changes

- 修复 zap adapter `Close()` 方法：添加 `recover` 防止 `Sync()` panic
- 在四个 adapter（slog/zerolog/zap/writer）的 `level` 字段添加 TODO 注释，标记后续改为 `atomic.Int32`
- 重写 `examples/log/slog/main.go`、`examples/log/zerolog/main.go`、`examples/log/zap/main.go`，统一结构并覆盖完整 API（各级别日志、key-value 字段、非字符串 key、奇数参数丢弃、With 子 logger、SetLevel 级别过滤、defer Close）

## Capabilities

### New Capabilities

（无新增能力）

### Modified Capabilities

- `logger`: zap adapter Close() 添加 panic 保护；level 字段添加 TODO 标记并发安全改进

## Impact

- **修改代码**: `core/log/zap.go`（Close recover）、`core/log/slog.go`、`core/log/zerolog.go`、`core/log/writer.go`（TODO 注释）
- **修改代码**: `examples/log/slog/main.go`、`examples/log/zerolog/main.go`、`examples/log/zap/main.go`
- **不改变 API**: Logger interface 不变，行为不变
