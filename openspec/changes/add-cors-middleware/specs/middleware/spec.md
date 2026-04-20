## MODIFIED Requirements

### Requirement: Default middleware mounting
App SHALL 默认挂载 RequestID、otelgin（如启用）、CORS、Recovery、Logging 和 Response 中间件到 HTTP Server。

#### Scenario: Default middleware mounted automatically
- **WHEN** 用户调用 `quix.New()` 未传入 `WithDefaultMiddleware(false)`
- **THEN** HTTP Server MUST 自动挂载 RequestID、otelgin（如启用 Telemetry）、CORS、Recovery、Logging 和 Response 中间件，顺序为 `RequestID → [otelgin] → CORS → Recovery → Logging → Response`

#### Scenario: Disable default middleware
- **WHEN** 用户调用 `quix.New(quix.WithDefaultMiddleware(false))`
- **THEN** HTTP Server MUST 不挂载任何默认中间件

#### Scenario: Disable CORS only
- **WHEN** 用户调用 `quix.New(quix.WithCORS(false))`
- **THEN** HTTP Server MUST 不挂载 CORS 中间件，但其他默认中间件正常挂载

#### Scenario: Custom CORS at App level
- **WHEN** 用户调用 `quix.New(quix.WithCORSConfig(cfg))`
- **THEN** HTTP Server MUST 使用自定义 `cors.Config` 挂载 CORS 中间件
