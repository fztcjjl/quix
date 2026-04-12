## ADDED Requirements

### Requirement: HTTP Server default middleware
HTTP Server 创建时 SHALL 支持通过 Option 控制是否自动挂载 Recovery 和 RequestID 中间件。

#### Scenario: NewServer mounts default middleware when enabled
- **WHEN** 创建 HTTP Server 且 `server.WithDefaultMiddleware(true)` 或未配置
- **THEN** MUST 在 Engine 上挂载 Recovery 和 RequestID 中间件

#### Scenario: Disable default middleware
- **WHEN** 创建 HTTP Server 时传入 `server.WithDefaultMiddleware(false)`
- **THEN** MUST 不挂载任何默认中间件
