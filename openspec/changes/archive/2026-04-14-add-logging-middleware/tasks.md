## 1. Logging 中间件实现

- [x] 1.1 创建 `core/transport/http/server/middleware/logging.go` — 实现 `Logging(skipPaths ...string)` 函数，记录 method/path/status/latency/request_id/client_ip/response_size
- [x] 1.2 实现状态码到日志级别的映射（2xx/3xx=Info, 4xx=Warn, 5xx=Error）
- [x] 1.3 实现 SkipPaths 精确路径跳过逻辑

## 2. 默认中间件链更新

- [x] 2.1 修改 `core/transport/http/server/server.go` — 默认中间件链加入 Logging，顺序为 `Recovery → RequestID → Logging → Response`

## 3. 测试

- [x] 3.1 创建 `core/transport/http/server/middleware/logging_test.go` — 测试日志字段完整性、状态码级别映射、跳过路径逻辑
- [x] 3.2 验证测试通过

## 4. 示例

- [x] 4.1 创建 `examples/middleware/logging/main.go` — 演示默认日志和跳过路径
- [x] 4.2 端到端验证：`go run examples/middleware/logging/main.go` 并发送请求查看日志输出

## 5. 收尾

- [x] 5.1 更新 `CLAUDE.md` — 如有必要更新中间件相关说明
- [x] 5.2 运行全部测试、构建、lint
