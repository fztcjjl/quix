# Quix

基于 Gin 的 Golang 快速开发框架，为 HTTP API 服务提供开箱即用的基础设施集成。

## 特性

- **薄封装** — 基于 Gin 的轻量封装，不隐藏 Gin 的能力
- **零配置启动** — 合理默认值，`quix.New()` 即可运行
- **Option 模式** — 按需定制，灵活组合
- **统一接口** — 每个能力定义最小化接口，第三方库通过适配器接入
- **可插拔组件** — Logger、Config、Metrics、Tracing、Auth 均可替换

## 技术栈

| 模块 | 选型 |
|------|------|
| HTTP | [Gin](https://github.com/gin-gonic/gin) |
| 默认日志 | Go stdlib `slog` |
| 可选日志 | [Zerolog](https://github.com/rs/zerolog) / [Zap](https://github.com/uber-go/zap) |
| 配置 | [koanf](https://github.com/knadh/koanf) |
| 指标/追踪 | [OpenTelemetry](https://opentelemetry.io/) |

## 快速开始

```bash
go get github.com/fztcjjl/quix
```

```go
package main

import "github.com/fztcjjl/quix"

func main() {
    app := quix.New()
    app.GET("/hello", func(c *quix.Context) {
        c.JSON(200, gin.H{"message": "hello"})
    })
    app.Run(":8080")
}
```

## 项目结构

```
quix/
├── quix.go           # App 结构体、New()、Run()
├── option.go         # Option 类型和 WithXxx() 函数
├── core/             # 框架组件
│   ├── logger/       # 日志
│   ├── config/       # 配置
│   ├── metrics/      # 指标
│   ├── tracing/      # 链路追踪
│   └── auth/         # 认证
├── middleware/        # 内置中间件
├── examples/          # 使用示例
└── internal/          # 内部工具
```

## 开发

```bash
# 测试
go test ./...

# 代码检查
golangci-lint run

# 运行示例
go run examples/logger/slog/main.go
```

## License

MIT
