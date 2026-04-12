## 1. Transport 接口定义

- [x] 1.1 创建 `core/transport/` 目录
- [x] 1.2 在 `core/transport/transport.go` 中定义 Server 接口（Start/Stop）
- [x] 1.3 编写接口编译期检查测试

## 2. HTTP Server 实现

- [x] 2.1 创建 `core/transport/http/` 目录
- [x] 2.2 在 `core/transport/http/server.go` 中实现 Server（封装 gin.Engine）
- [x] 2.3 实现 NewServer 构造函数（addr + Option 模式）
- [x] 2.4 实现 Engine() 方法返回底层 *gin.Engine
- [x] 2.5 代理 Gin 路由方法（GET/POST/PUT/DELETE/PATCH/GROUP/USE）
- [x] 2.6 实现 Start() — 创建 http.Server 并启动
- [x] 2.7 实现 Stop() — 优雅关闭 http.Server
- [x] 2.8 编写 HTTP Server 单元测试（httptest）

## 3. App 集成

- [x] 3.1 重构 `quix.go`：App 持有默认 Server + 额外 Server 列表
- [x] 3.2 修改 `New()` 创建默认 HTTP Server
- [x] 3.3 实现 `App.Run(addr)` — 设置地址、启动所有 Server、监听信号
- [x] 3.4 实现 `App.Shutdown(ctx)` — 关闭所有 Server
- [x] 3.5 实现 `App.AddServer(s transport.Server)` 添加额外 Server
- [x] 3.6 实现便捷方法代理（GET/POST/PUT/DELETE/PATCH/GROUP/USE）
- [x] 3.7 Config 集成：addr 为空时从 server.port 读取
- [x] 3.8 启动/关闭时输出日志
- [x] 3.9 添加 WithGinMode Option
- [x] 3.10 编写 App 集成测试

## 4. 使用示例

- [x] 4.1 编写 `examples/http/main.go`（完整 HTTP 服务示例）
- [x] 4.2 验证示例可通过 `go run` 执行并通过 curl 访问
