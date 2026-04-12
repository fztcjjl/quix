## Why

quix 框架需要一个统一的配置加载能力。后续所有组件（Middleware、Metrics、Tracing、Auth 等）都需要从配置中读取参数（如监听地址、日志级别、数据库连接串等）。Config 组件是框架基础设施的第二块拼图，紧接在 Logger 之后。

## What Changes

- 新增 `quix/core/config/` 子包，提供统一的 Config 接口
- 默认实现基于 koanf，支持多数据源（YAML 文件、环境变量、命令行参数）
- 提供 `quix.WithConfig()` Option 函数，支持注入自定义 Config 到 App
- 在 `examples/config/` 中提供使用示例

## Capabilities

### New Capabilities
- `config`: 统一配置接口、koanf 默认实现、多数据源支持、使用示例

### Modified Capabilities

## Impact

- 新增依赖：koanf 及相关 provider（yaml、env、file）
- 新增 `quix/core/config/` 子包，对外公开
- 新增 `examples/config/` 目录
- App 结构体新增 Config 字段
