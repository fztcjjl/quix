## Context

quix.go 是框架入口，负责 App 创建、Server 启动、优雅关闭。当前代码在错误处理、环境感知、日志输出方面存在不足，影响生产可用性。

## Goals / Non-Goals

**Goals:**
- telemetry 初始化失败时降级而非崩溃
- 引入统一环境概念（QUIX_ENV），驱动 Gin mode 和日志格式等默认行为
- 统一 Run 和 Shutdown 的关闭逻辑
- 提供清晰的启动和关闭日志

**Non-Goals:**
- 不添加健康检查端点，后续单独处理
- 不根据环境自动设置日志级别，由用户通过 WithLogger 控制
- 不自动加载 config.yaml，配置文件加载由用户通过 WithConfig 控制
- 不扩展更多环境类型（uat、sandbox 等），业务层按需自行处理

## Decisions

### 1. telemetry 初始化失败降级而非 panic

**决策**: `telemetry.Init` 失败时输出 warn 日志并跳过，App 正常启动（无 telemetry 功能）。

**替代方案**: 返回 error 让用户决定。

**选择理由**: 可观测性是辅助功能，不应该因为 Collector 不可达就阻止应用启动。生产环境中 Collector 可能临时不可用，应用应该能正常运行。

### 2. 统一环境概念 QUIX_ENV

**决策**: 引入 `QUIX_ENV` 环境变量，内置四种环境类型：`dev`、`test`、`staging`、`prod`。环境驱动以下默认行为：

| | dev | test | staging | prod |
|---|---|---|---|---|
| Gin mode | debug | test | release | release |
| 默认日志格式 | console | json | json | json |

- 未设置 `QUIX_ENV` 时默认 `dev`
- 未知的自定义环境（如 `uat`、`sandbox`）按 `prod` 行为处理（release + json），安全保守
- 用户显式 `WithGinMode` / `WithLogger` 时以用户设置为准，优先级高于环境驱动

**环境类型定义**:
```go
type Environment string

const (
    EnvDev     Environment = "dev"
    EnvTest    Environment = "test"
    EnvStaging Environment = "staging"
    EnvProd    Environment = "prod"
)
```

**替代方案 A**: 继续读取 `GIN_MODE` 环境变量。
**替代方案 B**: 从 config 读取 `app.env`。

**选择理由**:
- `QUIX_ENV` 是框架级的环境概念，不绑定任何特定组件（Gin、日志等），未来可扩展驱动更多默认行为
- `GIN_MODE` 只控制 Gin，无法驱动日志格式等其他行为
- 环境变量是部署时最自然的配置方式，不需要改代码
- 不使用 `APP_ENV` 是因为太通用，容易与其他框架/工具冲突
- 保留 `WithGinMode` 和 `WithLogger` 让用户在需要时覆盖默认行为

**默认日志格式实现**:
```go
// New() 中，创建默认 logger 前根据 env 决定输出格式
output := os.Stderr
if env == EnvDev {
    output = zerolog.ConsoleWriter{Out: os.Stderr}
}
defaultLog := log.NewZerolog(zerolog.New(output).With().Timestamp().Logger())
```

### 3. Run() 内部调用 Shutdown()

**决策**: `Run()` 的关闭逻辑不再重复实现，改为调用 `Shutdown(ctx)`，确保 telemetry flush 和 logger close 两个路径一致。

**替代方案**: 在两处分别维护关闭逻辑。

**选择理由**: 消除重复代码，避免 Run 和 Shutdown 行为漂移。

### 4. 启动日志输出关键信息

**决策**: `Run()` 启动服务前输出一条 info 日志，包含：HTTP 监听地址、Gin mode、telemetry 是否启用。

**替代方案**: 提供 OnStart 回调让用户自己打日志。

**选择理由**: 框架自动输出基本信息是合理默认行为，用户不需要为看监听端口而写额外代码。

### 5. Shutdown 日志输出

**决策**: `Shutdown()` 在关闭每个组件前后输出 info/warn 日志。

### 6. WithSetup 启动前回调

**决策**: 新增 `WithSetup(funcs ...func(*App) error)` Option，支持注册多个启动前回调。回调在 `Run()` 中 HTTP server 启动前、启动日志输出后按注册顺序执行。

**替代方案**: 独立 setup 函数在 `main()` 中手动调用。

**选择理由**:
- 多模块场景下更解耦：每个模块通过 `WithSetup(myModule.Setup)` 注册初始化逻辑，不需要修改 `main.go`
- 回调接收 `*App` 参数，可以访问 `app.HttpServer().Engine`、`app.Logger()`、`app.Config()` 等
- 回调返回 `error`，符合 Go 惯例，可显式报告失败（DB 连接、缓存预热等）
- 回调返回 error 时输出 error 日志并调用 `os.Exit(1)` 退出程序。理由：WithSetup 语义是"启动前必须完成的初始化"，初始化失败时 app 处于损坏状态，启动无意义。非关键逻辑不返回 error 即可

**执行时序**:
```
Run():
  1. 输出启动日志
  2. 执行 WithSetup 回调 ← 新增（路由注入、中间件注入、DB 连接、缓存预热）
     - 按注册顺序依次执行
     - 某个回调返回 error → 输出 error 日志 → os.Exit(1)
  3. 启动 HTTP server (goroutine)
  4. 等待信号
```

## Risks / Trade-offs

**[telemetry 降级]** → 用户可能不知道 telemetry 没生效。Mitigation: warn 日志明确说明原因和影响。
**[QUIX_ENV vs GIN_MODE]** → 已有项目可能依赖 `GIN_MODE` 环境变量。Mitigation: `QUIX_ENV` 作为框架标准推荐，但 `WithGinMode` 始终可覆盖，不影响已有使用方式。
**[WithSetup 失败]** → 某个回调失败导致程序退出，是否过于激进。Mitigation: WithSetup 语义是启动前初始化，失败意味着 app 处于损坏状态，退出比带病启动更安全。非关键逻辑由用户自行决定不返回 error。
