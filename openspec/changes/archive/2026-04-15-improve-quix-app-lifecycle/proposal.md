## Why

quix.go 的 App 生命周期存在多个问题：telemetry 初始化失败直接 panic 导致生产环境崩溃、缺少统一环境概念导致 Gin mode 和日志格式无法按部署环境自动调整、Run/Shutdown 行为不一致、缺少启动和关闭日志。这些问题影响生产可用性和用户体验。

## What Changes

- telemetry.Init 失败时降级为 warn 日志，不再 panic
- 引入 `QUIX_ENV` 环境变量（dev/test/staging/prod），自动驱动 Gin mode（dev→debug, test→test, staging/prod→release）和默认日志格式（dev→console, test/staging/prod→json）
- `Run()` 内部复用 `Shutdown()` 逻辑，统一 shutdown 行为
- 启动时输出关键信息日志（监听地址、环境、Gin mode、telemetry 状态）
- `Shutdown()` 增加 shutdown 过程日志
- 新增 `WithSetup(funcs ...func(*App) error)` Option，支持注册多个启动前回调（路由注入、中间件注入、数据库连接等），在 `Run()` 启动服务前按注册顺序执行；回调返回 error 时输出 error 日志并 `os.Exit(1)` 退出程序

## Capabilities

### New Capabilities

- `app-lifecycle`: App 启动/关闭日志、telemetry 降级、QUIX_ENV 统一环境（驱动 Gin mode + 日志格式）、Run/Shutdown 统一、WithSetup 启动前回调（error 致命退出）

### Modified Capabilities

（无）

## Impact

- **修改代码**: `quix.go`（Run、Shutdown、New、环境感知、telemetry 初始化逻辑、WithSetup 回调执行）
- **修改代码**: `option.go`（新增 WithSetup、WithEnv Option）
- **新增公开类型**: `Environment`（dev/test/staging/prod）
- **新增公开 API**: `WithSetup(funcs ...func(*App) error)`、`WithEnv(Environment)`
- **新增公开常量**: `EnvDev`、`EnvTest`、`EnvStaging`、`EnvProd`
