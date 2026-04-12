## Context

quix HTTP Server 使用 `gin.New()` 创建 Gin Engine，不包含任何默认中间件。handler panic 会直接 crash。框架需要内置常用中间件，提供合理的默认行为。

quix 定位为薄封装，应尽量复用 Gin 生态成熟的中间件库，避免重复造轮子。仅在需要与框架深度集成时（如 Recovery 需要 `log.Error()`）才自行实现。

## Goals / Non-Goals

**Goals:**
- 实现 Recovery 中间件（集成框架 Logger）
- 封装 `gin-contrib/requestid`，提供便捷函数
- 封装 `gin-contrib/cors`，提供便捷函数
- HTTP Server 默认挂载 Recovery + RequestID
- 提供机制让用户控制是否使用默认中间件

**Non-Goals:**
- 实现限流、认证等高级中间件（后续独立 change）
- 自行实现 RequestID 或 CORS（直接用 gin-contrib）

## Decisions

### 1. Recovery — 自行实现

Gin 内置 `gin.Recovery()` 但使用 gin 默认日志，无法集成 quix 的 Logger。自行实现 Recovery：
- 捕获 panic，通过 `log.Error()` 输出堆栈
- 返回 HTTP 500
- 使用 `runtime.Stack()` 获取堆栈信息
- 约 30 行代码，薄封装

### 2. RequestID — 封装 gin-contrib/requestid

```go
import "github.com/gin-contrib/requestid"

func RequestID() gin.HandlerFunc {
    return requestid.New()
}
```

薄封装，提供统一的 `middleware.RequestID()` 调用风格。用户需要自定义时可直接使用 `gin-contrib/requestid`。

### 3. CORS — 封装 gin-contrib/cors

```go
import "github.com/gin-contrib/cors"

func CORS() gin.HandlerFunc {
    return cors.Default()
}

func WithCORSConfig(cfg cors.Config) gin.HandlerFunc {
    return cors.New(cfg)
}
```

薄封装。`CORS()` 使用 gin-contrib 的默认配置（允许所有 Origin），`WithCORSConfig()` 允许自定义。

### 4. 默认中间件挂载

`App.New()` 创建 HTTP Server 时，默认挂载 Recovery + RequestID。通过 Option 控制：

```go
// 默认挂载 Recovery + RequestID
app := quix.New()

// 关闭默认中间件
app := quix.New(quix.WithDefaultMiddleware(false))
```

实现方式：
- `quix.option.go` 新增 `defaultMiddleware bool` 字段到 App
- `server.option.go` 新增 `WithDefaultMiddleware(bool)` 到 server Option
- `NewServer()` 中根据配置决定是否挂载

### 5. CORS 不作为默认中间件

CORS 策略因项目而异，用户需要时显式挂载：
```go
app.Use(middleware.CORS())
```

## Risks / Trade-offs

- [外部依赖] 引入两个 gin-contrib 依赖 → 均为 Gin 官方组织维护，广泛使用，风险低
- [Recovery 自实现复杂度] 捕获 panic 需要使用 defer + recover → 标准模式，复杂度可控
