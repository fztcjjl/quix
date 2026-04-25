## Why

当前 `core/config/koanf.go` 使用 koanf V1 API 加载环境变量，`normalizeEnvKey` 将所有 `_` 替换为 `.`（如 `SMS_ACCESS_KEY_ID` → `sms.access.key.id`），导致 snake_case key（如 `access_key_id`）被错误拆分为嵌套层级。此外未设置 Prefix，系统环境变量（`PATH`、`HOME` 等）全部灌入 koanf 产生噪音。koanf 社区（包括作者在 listmonk 中的实践）一致推荐双下划线 `__` 做嵌套分隔。

## What Changes

- 将 `env.Provider` 从 V1 API（`env.Provider(prefix, delim, cb)`）升级到 V2 API（`env.Provider(delim, env.Opt{...})`）
- 环境变量嵌套分隔符从单 `_` 改为双 `__`（如 `SMS__ACCESS_KEY_ID` → `sms.access_key_id`）
- 添加 `WithEnvPrefix(prefix string)` Option，过滤只加载指定前缀的环境变量，防止系统变量污染
- 同步更新 `.env` 文件加载逻辑，使用相同的 `__` 分隔规则
- **BREAKING**: 环境变量命名约定变更，现有使用单 `_` 的环境变量需迁移为 `__` 格式

## Capabilities

### New Capabilities

_无新能力_

### Modified Capabilities

- `config`: 环境变量映射规则从单 `_` 改为双 `__` 嵌套分隔，新增 `WithEnvPrefix` Option，升级 koanf env provider 到 V2 API

## Impact

- **代码**: `core/config/koanf.go`（主要改动）、`core/config/koanf_test.go`（测试更新）
- **示例**: `examples/config/` 下的环境变量示例需同步更新命名约定
- **依赖**: `github.com/knadh/koanf/providers/env/v2`（V2 API，需确认当前 go.mod 是否已有）
- **破坏性变更**: 用户现有的单 `_` 环境变量需改为 `__` 格式
