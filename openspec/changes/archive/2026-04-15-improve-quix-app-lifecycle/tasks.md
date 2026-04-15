## 1. telemetry 初始化降级

- [x] 1.1 修改 `quix.go` telemetry 初始化：失败时 warn 日志 + 跳过，不再 panic

## 2. 统一环境概念 QUIX_ENV

- [x] 2.1 `quix.go` 新增 `Environment` 类型和 `EnvDev`/`EnvTest`/`EnvStaging`/`EnvProd` 常量
- [x] 2.2 `quix.go` `New()` 中读取 `QUIX_ENV` 环境变量，未设置默认 `dev`，未知值按 `prod` 处理
- [x] 2.3 `quix.go` `New()` 中根据环境自动设置 Gin mode（dev→debug, test→test, staging/prod→release），用户 `WithGinMode` 优先
- [x] 2.4 `quix.go` `New()` 中根据环境自动选择默认日志格式（dev→ConsoleWriter, test/staging/prod→JSON），用户 `WithLogger` 优先
- [x] 2.5 `option.go` 新增 `WithEnv(Environment)` Option

## 3. Run/Shutdown 统一

- [x] 3.1 修改 `quix.go` `Run()`：信号处理中调用 `Shutdown(ctx)` 而非重复实现关闭逻辑
- [x] 3.2 修改 `quix.go` `Shutdown()`：为每个组件关闭添加日志输出

## 4. 启动日志

- [x] 4.1 修改 `quix.go` `Run()`：启动 HTTP server 前输出 info 日志（addr、env、gin_mode、telemetry 状态）

## 5. WithSetup 启动前回调

- [x] 5.1 `quix.go` App 结构体添加 `setupFuncs []func(*App) error` 字段
- [x] 5.2 `option.go` 新增 `WithSetup(funcs ...func(*App) error)` Option
- [x] 5.3 `quix.go` `Run()` 中启动日志输出后、HTTP server 启动前执行 WithSetup 回调（按注册顺序，返回 error 时输出 error 日志并 `os.Exit(1)`）

## 6. 验证

- [x] 6.1 执行 `go fmt ./...`、`go build ./...`、`go test ./...`、`golangci-lint run ./...`
