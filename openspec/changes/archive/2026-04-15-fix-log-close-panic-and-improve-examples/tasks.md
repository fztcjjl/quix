## 1. zap Close panic 修复

- [x] 1.1 修改 `core/log/zap.go` Close() 方法：添加 `defer func() { _ = recover() }()`

## 2. level 并发安全 TODO 标记

- [x] 2.1 `core/log/slog.go` level 字段添加 TODO 注释
- [x] 2.2 `core/log/zerolog.go` level 字段添加 TODO 注释
- [x] 2.3 `core/log/zap.go` level 字段添加 TODO 注释
- [x] 2.4 `core/log/writer.go` level 字段添加 TODO 注释

## 3. 重写 log 示例

- [x] 3.1 重写 `examples/log/slog/main.go`：覆盖完整 Logger API
- [x] 3.2 重写 `examples/log/zerolog/main.go`：覆盖完整 Logger API
- [x] 3.3 重写 `examples/log/zap/main.go`：覆盖完整 Logger API

## 4. 验证

- [x] 4.1 执行 `go fmt ./...`、`go build ./...`、`go test ./core/log/...`、`golangci-lint run ./core/log/...`
