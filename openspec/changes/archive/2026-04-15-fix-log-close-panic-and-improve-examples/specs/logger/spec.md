## MODIFIED Requirements

### Requirement: zap adapter Close method safely handles Sync panic
zap adapter 的 `Close()` SHALL 调用 `Sync()` 并通过 `recover` 防止 panic 传播到调用方。当底层 writer 已关闭导致 `Sync` panic 时，`Close()` SHALL 返回 nil 而非 panic。

#### Scenario: Sync panic does not propagate
- **WHEN** 底层 writer 已关闭，调用 zap adapter 的 `Close()`
- **THEN** 不发生 panic，`Close()` 正常返回

### Requirement: level 字段标记并发安全 TODO
四个 adapter（slog/zerolog/zap/writer）的 `level` 字段 SHALL 添加 TODO 注释，标记后续需要改为 `atomic.Int32` 以保证并发安全。

#### Scenario: TODO 注释存在
- **WHEN** 检查四个 adapter 源码
- **THEN** `level` 字段旁包含 `TODO: use atomic.Int32 for concurrent SetLevel safety` 注释
