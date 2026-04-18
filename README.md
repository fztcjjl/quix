# Quix

基于 Gin 的 Golang 快速开发框架，为 HTTP API 服务提供开箱即用的基础设施集成。

## 特性

- **薄封装** — 基于 Gin 的轻量封装，不隐藏 Gin 的能力
- **零配置启动** — 合理默认值，`quix.New()` 即可运行
- **Option 模式** — 按需定制，灵活组合
- **统一接口** — 每个能力定义最小化接口，第三方库通过适配器接入
- **可插拔组件** — Logger、Config、Metrics、Tracing、Auth 均可替换
- **IDL 驱动** — 通过 `protoc-gen-quix-gin` 插件从 protobuf 自动生成 Gin 路由代码
- **请求校验** — 集成 protovalidate-go，通过 proto 注解定义校验规则，自动执行字段验证

## 技术栈

| 模块 | 选型 |
|------|------|
| HTTP | [Gin](https://github.com/gin-gonic/gin) |
| 日志 | Go stdlib `slog`（默认）、[Zerolog](https://github.com/rs/zerolog)、[Zap](https://github.com/uber-go/zap) |
| 配置 | [koanf](https://github.com/knadh/koanf) |
| 指标/追踪 | [OpenTelemetry](https://opentelemetry.io/) |
| IDL | [protobuf](https://protobuf.dev/) + [google.api.http](https://aip.dev/4320) |
| 校验 | [protovalidate-go](https://github.com/bufbuild/protovalidate-go) |

## 快速开始

### 基础用法

```go
package main

import "github.com/fztcjjl/quix"

func main() {
    app := quix.New()
    app.GET("/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "hello"})
    })
    app.Run(":8080")
}
```

### IDL 驱动开发

定义 proto 文件，通过注解映射 HTTP 路由：

```protobuf
service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply) {
    option (google.api.http) = {
      get: "/hello/{name}"
    };
  }
  rpc CreateUser (CreateUserRequest) returns (UserResponse) {
    option (google.api.http) = {
      post: "/users"
      body: "*"
    };
  }
}
```

使用 buf 或 protoc 生成代码：

```bash
cd examples/proto-api && buf generate
```

生成的 `greeter_gin.go` 包含服务接口、路由注册和 handler，实现接口即可运行：

```go
package service

type GreeterService struct{}

func (s *GreeterService) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
    return &pb.HelloReply{Message: "Hello, " + req.Name}, nil
}
```

```go
func main() {
    app := quix.New()
    pb.RegisterGreeterHTTPService(app.Group("/api"), service.NewGreeterService())
    app.Run(":8080")
}
```

## 项目结构

```
quix/
├── quix.go           # App 结构体、New()、Run()、Shutdown()
├── option.go         # Option 类型和 WithXxx() 函数
├── core/             # 框架组件
│   ├── errors/       # 结构化错误类型
│   ├── config/       # 配置加载
│   ├── log/          # 日志适配器
│   └── transport/
│       └── http/server/  # HTTP Server、Handler 包装、默认中间件
├── middleware/        # 内置中间件
├── cmd/protoc-gen-quix-gin/  # protoc 代码生成插件
├── internal/         # 插件运行时
├── examples/          # 使用示例
└── openspec/         # 变更管理
```

## 组件

### 错误处理

统一的错误处理模式：Handler 返回 error，框架自动包装为结构化响应。

```go
app.GET("/user/:id", qhttp.Handler(func(c *gin.Context) error {
    return errors.NotFound("user_not_found", "用户不存在")
}))
// 响应: {"error": {"code": "user_not_found", "message": "用户不存在"}} HTTP 404
```

预定义错误：`errors.BadRequest()`、`errors.NotFound()`、`errors.Unauthorized()`、`errors.Forbidden()`、`errors.Internal()`

### HTTP Server

内置 Recovery、RequestID、ResponseMiddleware 中间件，支持优雅关闭和信号处理。通过 `qhttp.Handler()` 包装支持 `func(c *gin.Context) error` 签名。

### 请求校验

集成 [protovalidate-go](https://github.com/bufbuild/protovalidate-go)，通过 proto 注解定义校验规则，生成的 handler 自动执行字段验证：

```protobuf
message CreateTaskRequest {
  string title = 1 [(buf.validate.field).string = {
    min_len: 1
    max_len: 200
  }];
}
```

校验失败返回结构化 400 响应：`{"error": {"code": "validation_error", "message": "请求参数验证失败", "details": [{"field": "title", "message": "..."}]}}`

## 开发

```bash
go test ./...
go build ./...
go fmt ./...
golangci-lint run ./...
```

## License

MIT
