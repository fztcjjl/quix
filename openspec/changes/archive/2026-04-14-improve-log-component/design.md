## Context

quix 框架的日志组件（`core/log/`）提供统一 Logger 接口和 slog/zap/zerolog 三个适配器。当前存在三类问题：

1. **正确性**: 参数预处理逻辑分散在三个 adapter 中重复实现，行为不一致；全局变量无并发保护
2. **一致性**: 三个 adapter 对非字符串 key、奇数参数处理行为不同；中间件用全局函数不感知 `WithLogger` 替换
3. **功能性**: 缺少级别控制、Fatal 级别、Close/Sync 生命周期管理

## Goals / Non-Goals

**Goals:**
- 统一三个 adapter 的参数处理行为，消除数据丢失风险
- 修复已知 bug（slice 副作用、并发安全）
- 解决中间件全局函数与 App 注入的不一致
- 分层默认 logger（`core/log` = slog，`quix` = zerolog）
- 扩展 Logger 接口（Fatal、SetLevel、Close）
- 增强 Logging 中间件（前缀匹配 skipPaths、自定义 hook）
- 提供零依赖的 `NewWriter` 适配器

**Non-Goals:**
- 不改变现有 adapter 的公开 API 签名（`NewSlog`、`NewZerolog`、`NewZap`）
- 不引入新的第三方日志库
- 不做日志采样、日志轮转等功能
- 不重构中间件的依赖注入模式（通过 `WithLogger` + `SetDefault` 解决即可）

## Decisions

### 1. 非字符串 key 统一策略：转为 `"key"` 字面量，值追加序号后缀避免覆盖

**选择**: 三个 adapter 统一为：非字符串 key 转为 `"key"` 字面量，多个非字符串 key 时追加序号后缀（`key_0`、`key_1`...）避免 map 覆盖。

**备选方案**:
- A) slog 的嵌套结构（`{"key": "k=v", "value": ...}`）— 输出结构复杂，不直观
- B) `fmt.Sprintf("%v", key)` 作为字段名 — 可能产生无效 JSON key
- C) panic — 过于严格

**理由**: 方案最简单，不丢数据，JSON 输出干净。

### 2. 奇数参数统一策略：静默 drop 尾部

**选择**: 三个 adapter 统一为静默 drop 最后一个孤立的 key，不 panic。

**备选方案**: panic（过于严格）、自动补 nil 值（语义模糊）

**理由**: 与 slog 现有行为一致，最安全。

### 3. 参数预处理统一：`normalizeArgs`

**选择**: 提取 `normalizeArgs` 统一函数，放在 `logger.go` 中。快路径（全字符串 key + 偶数参数）零分配直接返回原 slice；慢路径才 copy + 转换非字符串 key + drop 奇数尾部。

**备选方案**:
- A) 各 adapter 各自预处理（当前重构前方案）— 重复代码，行为不一致
- B) 在 adapter 构造时预处理 — 无法处理运行时动态参数

**理由**: 单一实现消除重复，快路径优化避免正常调用时的无用分配。

### 4. 并发安全：`atomic.Pointer[Logger]`

**选择**: `defaultLogger` 改用 `atomic.Pointer[Logger]` 存储。

**备选方案**: `atomic.Value`（存储不同 concrete type 会 panic）、`sync.RWMutex`（有锁开销）

**理由**: `atomic.Pointer[Logger]` 存储 `*Logger`（pointer-to-interface），与 slog 标准库一致。虽然 pointer-to-interface 略不常见，但是类型安全的并发方案。

### 5. 分层默认 logger

**选择**:
- `core/log/` 包级默认: `NewSlog()`（零外部依赖，开箱即用）
- `quix.New()`: 通过默认 zerolog Logger 初始化，并 `log.SetDefault()` 覆盖全局

**理由**: `core/log` 作为独立包使用 slog 默认，零外部依赖且产生可见输出；`quix` 应用层偏好 zerolog ConsoleWriter 的开发体验。

### 6. Logger 接口扩展

**选择**: 新增三个方法：
- `Fatal(ctx, msg, args...)` — 记录 Error 级别日志后 `os.Exit(1)`
- `SetLevel(level Level)` — 设置日志级别
- `Close() error` — 刷缓冲区释放资源

**备选方案**: 不加 Fatal（用户直接调 `log.Error` + `os.Exit`）；不加 SetLevel（依赖底层库）；不加 Close（依赖 defer）

**理由**: 框架层统一接口让用户不感知底层库差异。`Close` 对 zap 的 `Sync()` 尤其必要。

### 7. Logging 中间件增强

**选择**:
- skipPaths 支持前缀匹配（`/metrics` 匹配 `/metrics/health`）
- 新增 `LoggingWith(opts ...LoggingOption)` 函数式选项模式，支持自定义 hook

**理由**: 前缀匹配覆盖常见场景（健康检查、metrics）；hook 回调让用户灵活添加自定义字段。

### 8. `NewWriter` 适配器

**选择**: 新增 `NewWriter(w io.Writer) Logger`，内部用 `slog.NewJSONHandler(w)` 实现。

**理由**: 零额外依赖，满足"只想重定向输出"的简单需求。

### 9. MockLogger 不暴露到生产 API

**选择**: 删除 `mock.go`，各包测试在 `_test.go` 内局部定义 mock 或使用真实 Logger 实现。

**备选方案**: 保留 `mock.go` 作为共享测试辅助（但会暴露到生产 API）

**理由**: 生产包不应包含仅供测试使用的类型。`quix_test.go` 只需验证注入，用真实 Logger 即可；`logging_test.go` 需要捕获调用，局部定义 mock 足够。

## Risks / Trade-offs

- **[BREAKING] Logger 接口新增 3 个方法** → 所有现有实现和 mock 需同步更新。Mitigation: 框架内部代码量小，更新范围可控。
- **[BREAKING] 全局默认从 noopLogger 改为 slog** → 用户如果在 `quix.New()` 之前使用全局函数，现在会输出日志到 stderr 而非静默。Mitigation: 这是更合理的行为，且 `quix.New()` 会立即覆盖为 zerolog。
- **`SetLevel` 的粒度** → 当前设计只有全局级别控制，不支持 per-logger 级别。Mitigation: 满足大多数场景，per-logger 级别属于过度设计。
- **`Fatal` 不可恢复** → 与 `os.Exit` 绑定，无法被 recover。这是 Fatal 的标准语义。
