## Why

当前可观测性命名存在两个问题：1) `middleware/logging.go` 导出的 `ExtractTraceID` / `ExtractSpanID` 函数变量与 `core/telemetry/` 包的同名导出函数职责重叠，造成概念混淆；2) `server.go` 的 `telemetryServiceName` / `telemetryTracesEnabled` 字段与 `telemetry.Config` 冗余，quix.go 需要在初始化后手动从 telemetry Config 复制到 server options，增加维护负担。

## What Changes

- 删除 `middleware/logging.go` 中导出的 `ExtractTraceID` / `ExtractSpanID` 函数变量，改为 middleware 包内直接调用 `telemetry.ExtractTraceID()` / `telemetry.ExtractSpanID()`，消除重复定义
- 删除 `server.go` 中 `telemetryServiceName` / `telemetryTracesEnabled` 字段及对应的 `WithTelemetryServiceName` / `WithTelemetryTracesEnabled` Option
- 修改 `quix.go` 中 server 创建逻辑，otelgin middleware 的注入改为由 quix.New() 在创建 server 之前直接挂载到 engine，不再通过 server options 传递遥测配置
- 更新相关测试

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `logging-middleware-trace-id`: 去除函数变量解耦模式，改为直接调用 telemetry 包
- `otel-gin-middleware`: otelgin 注入方式从 server options 改为 quix.New() 直接挂载
- `app-telemetry`: quix.New() 不再存储 telemetryServiceName/telemetryTracesEnabled 字段

## Impact

- **API 变更**: `middleware.ExtractTraceID` / `middleware.ExtractSpanID` 导出变量被删除（**BREAKING**），外部代码应改为调用 `telemetry.ExtractTraceID()`；`qhttp.WithTelemetryServiceName` / `qhttp.WithTelemetryTracesEnabled` 被删除（**BREAKING**）
- **依赖变更**: `middleware` 包新增对 `core/telemetry` 包的 import（之前刻意避免 OTel import，现改为直接依赖）
- **受影响文件**: `core/transport/http/server/middleware/logging.go`、`core/transport/http/server/server.go`、`quix.go`、相关测试文件
