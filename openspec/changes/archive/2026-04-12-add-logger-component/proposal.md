## Why

quix 框架需要一个统一的日志能力作为基础设施。所有内置组件（中间件、Metrics、Auth 等）都依赖日志输出，日志也是应用开发最基本的需求。作为第一个组件，它的接口设计将为后续所有组件的集成方式定调。

## What Changes

- 新增 `quix/core/logger` 子包，提供统一的 Logger 接口
- 默认实现基于 Go stdlib `slog`（零外部依赖）
- 提供可选的 Zerolog 适配器实现
- 提供可选的 Zap 适配器实现
- 提供 `quix.WithLogger()` Option 函数，支持注入自定义 Logger 到 App
- 在 `examples/logger/` 中提供 Logger 使用示例，演示默认 slog、Zerolog、Zap 三种用法
- 确立 `examples/` 目录规范：每个组件 MUST 提供可运行的示例代码

## Capabilities

### New Capabilities
- `logger`: 统一日志接口、默认 slog 实现、可选 Zap/Zerolog 适配器、使用示例

### Modified Capabilities

## Impact

- 新增依赖：Zerolog 和 Zap 作为可选依赖（仅在用户选择时引入）
- 新增 `quix/core/logger/` 子包，对外公开
- 新增 `examples/logger/` 目录，包含可运行的示例代码
- 确立 `examples/` 目录规范，后续每个组件均需提供示例
- 后续所有框架组件将依赖此 Logger 接口
- 暂不影响现有代码（首个组件）
