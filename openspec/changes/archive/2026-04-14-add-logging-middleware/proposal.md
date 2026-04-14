## Why

quix 的 HTTP Server 目前缺少请求日志中间件（access log）。Recovery 只记录 panic，日常运维无法追踪请求的 method、path、status、latency 等关键信息。对于一个 HTTP API 框架，请求日志是最基础的可观测性能力。

## What Changes

- 新增 `Logging()` 中间件，记录每个 HTTP 请求的关键信息（method、path、status、latency、request_id、client_ip、response_size）
- 支持按状态码分级日志：2xx/3xx=Info、4xx=Warn、5xx=Error
- 支持配置跳过路径（如健康检查 `/healthz`）
- 使用框架 `core/log.Logger` 接口，与现有日志体系一致
- 将 `Logging()` 加入默认中间件链（在 Recovery 之后、Response 之前）
- 新增使用示例 `examples/middleware/logging/`

## Capabilities

### New Capabilities
（无）

### Modified Capabilities
- `middleware`: 新增 Logging 中间件、更新默认中间件链

## Impact

- 新增 `core/transport/http/server/middleware/logging.go`
- 修改 `core/transport/http/server/server.go` — 默认中间件链加入 Logging
- 新增 `examples/middleware/logging/main.go`
- `openspec/specs/middleware/spec.md` 更新 Logging 相关需求
- 无新增外部依赖
