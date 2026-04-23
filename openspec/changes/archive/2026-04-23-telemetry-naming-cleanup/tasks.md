## 1. Remove duplicated ExtractTraceID/ExtractSpanID from middleware

- [x] 1.1 删除 `middleware/logging.go` 中的 `ExtractTraceID` / `ExtractSpanID` 导出变量声明
- [x] 1.2 在 `AccessLog()` 和 `WithRequestLogger()` 中将 `ExtractTraceID(ctx)` / `ExtractSpanID(ctx)` 替换为 `telemetry.ExtractTraceID(ctx)` / `telemetry.ExtractSpanID(ctx)`
- [x] 1.3 在 `middleware/logging.go` 中添加 `"github.com/fztcjjl/quix/core/telemetry"` import

## 2. Remove telemetry fields from server.go

- [x] 2.1 删除 `server.go` 中 `options` 结构体的 `telemetryServiceName` / `telemetryTracesEnabled` 字段
- [x] 2.2 删除 `WithTelemetryServiceName()` / `WithTelemetryTracesEnabled()` Option 函数
- [x] 2.3 删除 `NewServer()` 中 otelgin middleware 条件注入逻辑

## 3. Update quix.go to directly mount otelgin

- [x] 3.1 删除 `App` 结构体中的 `telemetryServiceName` / `telemetryTracesEnabled` 字段
- [x] 3.2 删除 `quix.New()` 中对 `middleware.ExtractTraceID` / `middleware.ExtractSpanID` 的赋值
- [x] 3.3 删除 `quix.New()` 中传递 `qhttp.WithTelemetryServiceName` / `qhttp.WithTelemetryTracesEnabled` 到 server options 的逻辑
- [x] 3.4 在 `quix.New()` 中 telemetry 初始化成功后，如果 `telCfg.TracesEnabled` 为 true，直接调用 `app.httpServer.Use(otelgin.Middleware(telCfg.ServiceName))` 挂载 otelgin
- [x] 3.5 添加 `"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"` import（如果尚未存在）

## 4. Update tests

- [x] 4.1 更新 middleware/logging 相关测试，删除对 `middleware.ExtractTraceID` / `middleware.ExtractSpanID` 变量的引用
- [x] 4.2 更新 server 测试，删除对 `WithTelemetryServiceName` / `WithTelemetryTracesEnabled` 的引用
- [x] 4.3 更新 quix 集成测试，验证 otelgin 由 quix.New() 自动挂载
- [x] 4.4 运行 `go test ./...` 确保所有测试通过

## 5. Cleanup

- [x] 5.1 运行 `go fmt ./...` 格式化代码
- [x] 5.2 运行 `golangci-lint run ./...` 确保 lint 通过
