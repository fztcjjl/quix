## Context

`core/config/koanf.go` 使用 koanf V1 API（`env v1.1.0`）加载环境变量。`normalizeEnvKey` 将所有 `_` 替换为 `.`，导致 snake_case key（如 `access_key_id`）被错误拆分为多层嵌套。同时未设置 prefix，系统环境变量全部灌入 koanf。

koanf V2 API（`env/v2`）提供更清晰的 `env.Opt` 结构体，支持 `Prefix`、`TransformFunc`、`EnvironFunc`。koanf 作者在其项目 listmonk 中使用双下划线 `__` 做嵌套分隔，社区 GitHub Issues（#188、#295、#108）也一致推荐此方案。

## Goals / Non-Goals

**Goals:**
- 环境变量嵌套分隔无歧义：`SMS__ACCESS_KEY_ID` → `sms.access_key_id`
- 过滤系统环境变量，只加载指定前缀的变量
- 升级到 koanf env provider V2 API
- `.env` 文件与真实环境变量使用相同的分隔规则

**Non-Goals:**
- 不做 Viper 式反向查找改造（反向查找对 `Bind` 支持有限，改动面大）
- 不支持无前缀模式（必须设 prefix，这是安全最佳实践）
- 不改变 Config 接口定义
- 不改变 YAML 文件中的 key 风格

## Decisions

### D1: 双下划线 `__` 做嵌套分隔符

**选择**: `__` 替换为 `.`，单 `_` 保持原样。

**备选方案**:
- ~~单 `_` 全量替换（现状）~~ — 歧义问题无解
- ~~koanf prefix 分段加载~~ — 需要额外的 section 发现机制，复杂度高
- ~~Viper 反向查找~~ — `Bind` 需要额外处理，改动面大

**理由**: koanf 作者和社区共识方案，跨生态标准（ASP.NET、Pydantic、Docker Compose），改动量最小（核心逻辑改一行）。

### D2: WithEnvPrefix 为必需选项

**选择**: 新增 `WithEnvPrefix(prefix string)` Option，prefix 包含尾部 `_`（如 `"QUIX_"`）。尾部的 `_` 作为 prefix 与 config path 的边界，`__` 只出现在 config path 内部做嵌套分隔。

**理由**: 遵循 ASP.NET Core 约定（业界最广泛采用的 `__` 嵌套方案）。`QUIX_SMS__ACCESS_KEY_ID` → strip `QUIX_` → `SMS__ACCESS_KEY_ID` → `sms.access_key_id`。prefix 的 `_` 和嵌套的 `__` 各司其职，不会冲突。

无 prefix 会将 `PATH`、`HOME` 等系统变量灌入 config map，产生难以排查的噪音。

**向后兼容**: 默认 prefix 为空字符串（不设 prefix），不设置时打印 warning 日志。这样不会破坏现有代码，但引导用户迁移。

### D3: 升级到 koanf env provider V2

**选择**: 从 `github.com/knadh/koanf/providers/env v1.1.0` 升级到 `github.com/knadh/koanf/providers/env/v2`。

**理由**: V2 API（`env.Opt`）结构更清晰，`TransformFunc func(k, v string) (string, any)` 可同时处理 key/value，`EnvironFunc` 方便测试时注入 mock 环境。

### D4: `.env` 文件与 env vars 使用相同的分隔规则

**选择**: `.env` 文件中的 key 也使用 `__` 分隔符（如 `SMS__ACCESS_KEY_ID=xxx`），复用 `normalizeEnvKey`。

**理由**: 保持一致性，用户不需要记两套规则。

## Risks / Trade-offs

- **[BREAKING] 环境变量命名变更** → 现有使用 `SERVER_PORT` 等单下划线命名的部署需要改为 `SERVER__PORT`（无嵌套时）或保持不变（单层 key 无需 `__`）。实际上 `SERVER_PORT` 是单层 key（`server.port`），改为 `SERVER__PORT`。迁移成本较低但需要文档说明。
- **koanf v2 API 依赖** → 需确认 `env/v2` 与当前 koanf 核心版本兼容。可通过 `go get github.com/knadh/koanf/providers/env/v2` 验证。
