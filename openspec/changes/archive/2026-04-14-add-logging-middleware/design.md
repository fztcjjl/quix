## Context

quix 的 HTTP Server 默认中间件链为 `Recovery → RequestID → Response`。Recovery 仅在 panic 时输出日志，日常请求没有任何日志记录。框架已有 `core/log.Logger` 接口和 `gin-contrib/requestid` 中间件（在 context 中存储 `X-Request-Id`），Logging 中间件需要与这些现有设施集成。

## Goals / Non-Goals

**Goals:**
- 为每个 HTTP 请求输出一行结构化 access log（method、path、status、latency、request_id、client_ip、response_size）
- 按状态码自动选择日志级别（2xx/3xx=Info、4xx=Warn、5xx=Error）
- 支持配置跳过路径（如 `/healthz`、`/readyz`）
- 加入默认中间件链，零配置即可使用

**Non-Goals:**
- 不记录请求/响应 body（敏感数据风险大，且对性能影响显著）
- 不实现 Metrics 或 Tracing（后续独立变更）
- 不支持自定义日志字段或动态字段注入
- 不支持异步/缓冲日志写入（由 Logger 实现层决定）

## Decisions

### 1. 中间件位置

**选择**: 放在 `core/transport/http/server/middleware/logging.go`，与 Recovery、Response 同目录

**原因**: Logging 与 Recovery、Response 同属 HTTP Server 的核心中间件，放在同一目录保持一致性。

**替代方案**: 放在 `middleware/` 根目录 → 被否决，Logging 是 HTTP 层概念，不属于通用中间件。

### 2. 默认中间件链顺序

**选择**: `Recovery → RequestID → Logging → Response`

**原因**: Recovery 最先确保 panic 不泄漏；RequestID 生成 ID；Logging 需要读取 request_id 和最终 status code，必须在 Response 之前（Response 只处理 error 格式化，不影响 status code）；Response 在最后格式化错误响应。

```
Request → Recovery → RequestID → Logging → Response → Handler
```

### 3. 使用全局 Logger

**选择**: 使用 `log.Info/Error/Warn/Debug()` 全局函数，不通过参数注入 Logger

**原因**: 与 Recovery 中间件保持一致（Recovery 也使用全局 log 函数）。框架 Logger 已通过 `quix.New(quix.WithLogger(...))` 设置全局默认值。

### 4. 状态码到日志级别的映射

**选择**: 固定映射，不可配置

| 状态码范围 | 日志级别 |
|---|---|
| 1xx, 2xx, 3xx | Info |
| 4xx | Warn |
| 5xx | Error |

**原因**: Info 表示正常流量，运维设为 Info 可看全量请求；Warn 用于客户端异常需关注；Error 用于服务端故障需告警。Debug 留给应用代码的详细调试日志。

### 5. 路径跳过配置

**选择**: `Logging()` 接受可变 `SkipPaths []string` 参数，`Logging("/healthz", "/readyz")` 跳过指定路径

**原因**: 健康检查等高频路径的日志通常无价值。可变参数比 Options 模式更简洁，符合 Go 惯例（参考 `gin.LoggerWithConfig`）。路径匹配使用精确匹配（前缀匹配容易误跳过）。

### 6. 日志字段固定

**选择**: 固定字段集合，不支持自定义

```
method, path, status, latency, request_id, client_ip, response_size
```

**原因**: 固定字段满足 90% 场景。自定义字段增加复杂度，且可通过在 handler 中直接调用 `log.Info()` 实现。quix 定位薄封装，不过度设计。

### 7. response_size 来源

**选择**: 从 `gin.ResponseWriter.Size()` 获取

**原因**: Gin 的 ResponseWriter 已经包装了写入字节数统计，无需额外计算。

## Risks / Trade-offs

- **高频日志性能**: 每个请求产生一条日志 → 可通过 SkipPaths 减少健康检查等路径的日志量；性能瓶颈在 Logger 实现层而非中间件
- **固定字段限制**: 无法添加自定义字段 → 可通过 handler 内直接调 `log.Info()` 补充；如需强需求可后续扩展
- **精确匹配限制**: SkipPaths 只支持精确路径 → 足够健康检查场景；如需通配符可后续改为前缀匹配
