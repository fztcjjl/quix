## Context

quix 框架已完成 Logger 组件，Config 是第二个基础设施组件。后续的 Middleware、Metrics、Tracing、Auth 等组件都需要从配置中读取参数。Config 组件沿用 logger 建立的模式：最小接口 + koanf 默认实现 + Option 注入。

## Goals / Non-Goals

**Goals:**
- 定义一个最小、稳定的 Config 接口，作为框架所有组件的配置抽象
- 提供基于 koanf 的默认实现，支持 YAML 文件和环境变量
- 环境变量优先于文件配置（符合 12-Factor App 原则）
- 通过 Option 注入到 App

**Non-Goals:**
- 不支持热重载配置（后续按需添加）
- 不内置配置校验（如类型约束、必填检查，由用户在业务层处理）
- 不支持远程配置中心（Consul、etcd 等，后续按需通过 koanf provider 扩展）
- 不实现配置加密/解密

## Decisions

### D1: 接口设计 — 5 个方法 + Bind

**选择**: `Get/String/Int/Bool/Bind` 五个方法。

**替代方案**:
- 只暴露 `Get(key string) any`：太底层，用户每次都要类型断言
- 增加 `Float/StringSlice/IntSlice` 等：过度设计，按需再扩展

**理由**: 5 个方法覆盖 95% 的使用场景，`Bind` 用于将整个配置段映射到结构体（koanf 原生支持 `Unmarshal`）。

### D2: 环境变量覆盖文件配置

**选择**: 环境变量优先级高于文件配置。

**替代方案**:
- 文件优先，环境变量作为补充：不符合 12-Factor App
- 只用环境变量，不支持文件：开发体验差

**理由**: 文件用于开发环境默认值，环境变量用于生产环境覆盖，这是云原生应用的通用做法。

### D3: 键名映射策略

**选择**: 使用点号分隔路径（`server.port`），koanf 原生支持。环境变量使用下划线分隔 + 大写（`SERVER_PORT`），通过 koanf 的 env 映射自动转换。

**理由**: 文件配置用点号分隔直观，环境变量用大写下划线是 POSIX 惯例。koanf 内置支持这两种风格的映射。

### D4: WithFile 作为配置选项

**选择**: `config.NewKoanf(config.WithFile("config.yaml"))` 风格，而非在构造函数里硬编码文件路径。

**理由**: 与 quix 的 Option 模式一致，灵活可组合。用户可以添加多个配置源。

## Risks / Trade-offs

- **[键名冲突]** 不同配置源可能定义相同的键 → koanf 的加载顺序决定优先级，文档中明确说明
- **[环境变量映射]** 复杂嵌套结构的环境变量命名可能不直观 → 提供清晰的示例和文档
- **[性能]** koanf 启动时一次性加载，运行时只读 → 无性能问题
