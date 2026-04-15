## Context

core/log 组件代码审查发现两个改进点：(1) zap adapter `Close()` 调用 `Sync()` 可能 panic；(2) level 字段无并发保护。同时 examples/log 示例未覆盖完整 Logger API。

## Goals / Non-Goals

**Goals:**
- zap Close() 添加 recover 防止 panic
- 标记 level 字段并发安全问题（TODO，不立即修复）
- 重写三个 log 示例，统一结构并覆盖完整 API

**Non-Goals:**
- 不修复 level 字段的并发安全问题（仅标记 TODO，后续单独处理）
- 不改变 Logger interface

## Decisions

### 1. zap Close() 使用 defer recover 而非返回 error wrapper

**决策**: 在 `Close()` 中用 `defer func() { _ = recover() }()` 静默吞掉 panic。

**替代方案**: 包装返回 error，如 `return fmt.Errorf("sync panic: %v", r)`。

**选择理由**: `Sync` panic 是 zap 的已知行为（底层 writer 已关闭），用户无法从中恢复。静默吞掉比返回 error 更实用，因为 `Close()` 通常在 `defer` 中调用，返回值经常被忽略。

### 2. level 并发安全仅标记 TODO 不修复

**决策**: 在四个 adapter 的 `level` 字段添加 `// TODO: use atomic.Int32 for concurrent SetLevel safety` 注释。

**替代方案**: 立即改为 `atomic.Int32`。

**选择理由**: 实际场景中 `SetLevel` 通常只在启动时调用一次，data race 风险极低。改为 `atomic.Int32` 涉及所有 adapter + 测试的改动，属于独立变更，不适合混入此次修复。

## Risks / Trade-offs

**[recover 静默吞掉 panic]** → 如果 Sync panic 的原因不是 writer 关闭而是其他 bug，会被隐藏。Mitigation: Sync panic 几乎只在 writer 关闭时发生，且用户通常在程序退出时调用 Close，影响有限。
