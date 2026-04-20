## Why

框架当前缺少 CORS（跨域资源共享）支持。现有 `openspec/specs/middleware/spec.md` 已定义了 CORS 中间件的规格要求（`middleware.CORS()` 和 `middleware.WithCORSConfig(cfg)`），但尚未实现。大多数 Web API 服务都需要 CORS 支持，作为默认中间件链的一部分提供可以简化开发者的初始配置。

## What Changes

- 新增 `core/transport/http/server/middleware/cors.go`，实现 `CORS()` 和 `WithCORSConfig(cfg)` 函数
- 使用 `github.com/gin-contrib/cors` 作为底层实现
- 将 CORS 中间件加入默认中间件链（RequestID → otelgin → CORS → Recovery → Logging → Response）
- 在 App 级别增加 `WithCORS(enabled bool)` Option，支持禁用默认 CORS（保持向后兼容，默认启用）
- 在 Server 级别增加 `WithCORSConfig(cfg cors.Config)` Option，支持自定义 CORS 配置
- 更新 `openspec/specs/middleware/spec.md` 中默认中间件挂载需求，包含 CORS
- 在 `examples/middleware/` 下增加 CORS 示例

## Capabilities

### New Capabilities

_(无新增能力，CORS 已在 middleware spec 中定义)_

### Modified Capabilities

- `middleware`: 实现 CORS 中间件，并更新默认中间件挂载顺序（增加 CORS），增加 App/Server 级 CORS 配置 Option

## Impact

- **新增依赖**: `github.com/gin-contrib/cors`
- **影响文件**: `core/transport/http/server/server.go`（默认中间件链）、`quix/option.go`（新增 Option）、`core/transport/http/server/middleware/`（新增 cors.go）
- **示例代码**: `examples/middleware/`
- **向后兼容**: 默认行为变化——启用默认中间件时会自动挂载 CORS（允许所有 Origin）。如需禁用，使用 `WithCORS(false)`
