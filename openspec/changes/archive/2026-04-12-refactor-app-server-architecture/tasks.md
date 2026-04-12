## 1. App 结构体重构

- [x] 1.1 修改 App 结构体：`server transport.Server` + `servers []transport.Server` → `httpServer *qhttp.Server` + `rpcServer transport.Server`
- [x] 1.2 新增 `HttpServer()` getter 方法

## 2. Option 函数重构

- [x] 2.1 `WithServer(transport.Server)` → `WithHttpServer(*qhttp.Server)` + `WithRpcServer(transport.Server)`
- [x] 2.2 移除 `AddServer()` 方法

## 3. 配置驱动服务创建

- [x] 3.1 配置 key 变更：`server.addr`/`server.port` → `http.addr`/`http.port`，新增 `rpc.addr`
- [x] 3.2 `New()` 中实现配置驱动逻辑：有 http 配置启动 HTTP，有 rpc 配置启动 RPC，都没有默认启动 HTTP
- [x] 3.3 地址在 `New()` 中传递给 `NewServer()`，不再依赖运行时设置

## 4. Server 侧清理

- [x] 4.1 移除 `SetAddr()` 方法

## 5. Run/Shutdown 调整

- [x] 5.1 `Run()` 移除 `addr` 参数，分别处理 `httpServer` 和 `rpcServer`
- [x] 5.2 `Shutdown()` 分别处理 `httpServer` 和 `rpcServer`
- [x] 5.3 路由代理方法直接使用 `a.httpServer`，移除 `defaultHTTPServer()` 类型断言

## 6. 测试与示例更新

- [x] 6.1 更新 `quix_test.go`：适配新 API（WithHttpServer、移除 AddServer 测试、Run 无参数）
- [x] 6.2 更新 `examples/http/main.go`：`app.Run()` 无参数
- [x] 6.3 运行 `go fmt ./...` 和 `golangci-lint run ./...` 验证
