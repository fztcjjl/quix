## ADDED Requirements

### Requirement: App telemetry init degrades on failure
当 `WithTelemetry` 启用但 `telemetry.Init` 失败时，App SHALL 输出 warn 日志说明失败原因，并正常启动（不启用 telemetry）。App SHALL NOT panic。

#### Scenario: Collector unreachable
- **WHEN** `WithTelemetry` 启用且 OTLP endpoint 不可达
- **THEN** 输出 warn 日志 "telemetry init failed: <error>, running without telemetry"，App 正常启动，无 telemetry 功能

#### Scenario: telemetry init succeeds
- **WHEN** `WithTelemetry` 启用且 Collector 可达
- **THEN** telemetry 正常初始化，与当前行为一致

### Requirement: App auto sets Gin mode from QUIX_ENV
App SHALL 在用户未通过 `WithGinMode` 设置时，根据 `QUIX_ENV` 环境变量自动设置 Gin mode：`dev` → debug，`test` → test，`staging`/`prod` → release。未设置 `QUIX_ENV` 时默认 `dev`（debug）。未知的自定义环境按 `prod`（release）处理。App SHALL 提供 `Environment` 类型和 `EnvDev`/`EnvTest`/`EnvStaging`/`EnvProd` 常量。

#### Scenario: QUIX_ENV=prod
- **WHEN** 环境变量 `QUIX_ENV=prod` 且用户未传 `WithGinMode`
- **THEN** `gin.SetMode(gin.ReleaseMode)` 被调用

#### Scenario: QUIX_ENV=test
- **WHEN** 环境变量 `QUIX_ENV=test` 且用户未传 `WithGinMode`
- **THEN** `gin.SetMode(gin.TestMode)` 被调用

#### Scenario: QUIX_ENV=staging
- **WHEN** 环境变量 `QUIX_ENV=staging` 且用户未传 `WithGinMode`
- **THEN** `gin.SetMode(gin.ReleaseMode)` 被调用

#### Scenario: User explicit WithGinMode takes precedence
- **WHEN** 用户传入 `WithGinMode("test")`
- **THEN** 使用用户指定的 mode，不读取环境变量

#### Scenario: No QUIX_ENV set
- **WHEN** `QUIX_ENV` 环境变量未设置且用户未传 `WithGinMode`
- **THEN** 默认 `dev`，`gin.SetMode(gin.DebugMode)` 被调用

#### Scenario: Unknown environment value
- **WHEN** `QUIX_ENV=uat`（未知值）且用户未传 `WithGinMode`
- **THEN** 按 `prod` 处理，`gin.SetMode(gin.ReleaseMode)` 被调用

### Requirement: App auto sets default log format from QUIX_ENV
App SHALL 在用户未通过 `WithLogger` 设置自定义 logger 时，根据 `QUIX_ENV` 环境变量选择默认日志格式：`dev` → ConsoleWriter（人眼可读），`test`/`staging`/`prod` → JSON（结构化）。未设置 `QUIX_ENV` 时默认 `dev`（console）。

#### Scenario: QUIX_ENV=prod with default logger
- **WHEN** `QUIX_ENV=prod` 且用户未传 `WithLogger`
- **THEN** 默认 zerolog 输出 JSON 格式到 stderr

#### Scenario: QUIX_ENV=dev with default logger
- **WHEN** `QUIX_ENV=dev` 且用户未传 `WithLogger`
- **THEN** 默认 zerolog 使用 ConsoleWriter 输出到 stderr

#### Scenario: User WithLogger takes precedence
- **WHEN** 用户传入 `WithLogger(customLogger)`
- **THEN** 使用用户指定的 logger，不根据环境自动选择格式

### Requirement: Run uses Shutdown internally
`Run()` SHALL 在收到终止信号后调用 `Shutdown(ctx)`，而非重复实现关闭逻辑。这确保 Run 和 Shutdown 的关闭行为一致（包括 telemetry flush 和 logger close）。

#### Scenario: Run triggers Shutdown
- **WHEN** `Run()` 收到 SIGINT/SIGTERM
- **THEN** 调用 `Shutdown(ctx)` 完成优雅关闭

#### Scenario: Run exits on server startup failure
- **WHEN** HTTP server 启动失败（如端口被占用）
- **THEN** MUST 输出 error 日志 "http server failed to start" 并调用 `os.Exit(1)` 退出程序

#### Scenario: Run exits on RPC server startup failure
- **WHEN** RPC server 启动失败
- **THEN** MUST 输出 error 日志 "rpc server failed to start" 并调用 `os.Exit(1)` 退出程序

### Requirement: App outputs startup info log
`Run()` SHALL 在启动 HTTP server 前输出一条 info 日志，包含 HTTP 监听地址、环境（QUIX_ENV）、Gin mode、telemetry 是否启用。

#### Scenario: Startup log with telemetry in prod
- **WHEN** `QUIX_ENV=prod` 且 `WithTelemetry` 启用
- **THEN** info 日志包含 addr、env="prod"、gin_mode="release"、telemetry="enabled"

#### Scenario: Startup log without telemetry in dev
- **WHEN** `QUIX_ENV=dev` 且 `WithTelemetry` 未启用
- **THEN** info 日志包含 addr、env="dev"、gin_mode="debug"、telemetry="disabled"

### Requirement: App Shutdown outputs process log
`Shutdown()` SHALL 在关闭每个组件时输出日志（info/warn）。

#### Scenario: Successful shutdown
- **WHEN** `Shutdown(ctx)` 正常完成所有组件关闭
- **THEN** 每个组件关闭前后有日志输出

#### Scenario: Shutdown error
- **WHEN** 某个组件关闭失败
- **THEN** 输出 error 日志并继续关闭后续组件

### Requirement: App supports WithShutdownTimeout option
App SHALL 提供 `WithShutdownTimeout(d time.Duration)` Option，允许自定义优雅关闭超时时间。默认值为 5 秒。

#### Scenario: Default shutdown timeout
- **WHEN** 用户未传入 `WithShutdownTimeout`
- **THEN** MUST 使用默认值 5 秒

#### Scenario: Custom shutdown timeout
- **WHEN** 用户传入 `quix.New(quix.WithShutdownTimeout(15 * time.Second))`
- **THEN** 优雅关闭超时 MUST 为 15 秒

### Requirement: App supports WithSetup startup callback
App SHALL 提供 `WithSetup(funcs ...func(*App) error)` Option，支持注册多个启动前回调。回调在 `Run()` 中启动日志输出后、HTTP server 启动前执行，按注册顺序执行。

#### Scenario: Single WithSetup callback
- **WHEN** 用户注册一个 WithSetup 回调
- **THEN** 回调在启动日志输出后、HTTP server 启动前执行

#### Scenario: Multiple WithSetup callbacks
- **WHEN** 用户注册多个 WithSetup 回调
- **THEN** 按注册顺序依次执行

#### Scenario: WithSetup callback fails
- **WHEN** 某个 WithSetup 回调返回 error
- **THEN** 输出 error 日志 "setup callback failed: <error>" 并调用 `os.Exit(1)` 退出程序

#### Scenario: No WithSetup callbacks
- **WHEN** 用户未注册任何 WithSetup 回调
- **THEN** 跳过回调执行，直接启动 HTTP server
