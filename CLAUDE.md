# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

quix 是基于 Gin 的 Golang 快速开发框架，定位为薄封装。内置集成 Config/Log/Errors/Middleware/Transport 等基础设施组件，通过 Option 模式定制，提供合理的零配置默认值。支持通过 `protoc-gen-quix-gin` 插件从 protobuf IDL 自动生成 Gin 路由注册代码，通过 `protoc-gen-quix-errors` 插件从 proto enum 自动生成错误码常量和构造函数。

## 开发命令

```bash
go test ./...                              # 运行所有测试
go test ./core/log/...                     # 运行指定组件的测试
go test ./cmd/protoc-gen-quix-gin/... -run TestGenerate -update  # 更新 gin 插件 golden file
go test ./cmd/protoc-gen-quix-errors/... -run TestGenerate -update  # 更新 errors 插件 golden file
go build ./...                             # 构建所有包
go fmt ./...                               # 格式化代码
golangci-lint run ./...                     # 代码检查
go install ./cmd/protoc-gen-quix-gin           # 安装 protoc-gin 插件到 $PATH
go install ./cmd/protoc-gen-quix-errors        # 安装 protoc-errors 插件到 $PATH
```

无 Makefile 或构建系统，直接使用 Go 标准工具链。

## 项目架构

```
quix/
├── quix.go              # App 结构体、New()、Run()、Shutdown()
├── option.go            # Option 类型和 WithXxx() 函数
├── core/
│   ├── errors/           # 结构化错误类型（Error + 预定义函数）
│   ├── config/           # 配置加载（koanf）
│   ├── log/              # 日志（默认 slog，可选 Zerolog/Zap 适配器）
│   └── transport/
│       └── http/server/  # HTTP Server（嵌入 gin.Engine）+ Handler 包装 + 默认中间件
│           └── middleware/  # Recovery、ResponseMiddleware
├── middleware/           # 内置 Gin 中间件（recovery）
├── cmd/protoc-gen-quix-gin/  # protoc 插件：从 proto 生成 Gin 路由代码
├── cmd/protoc-gen-quix-errors/  # protoc 插件：从 proto enum 生成错误码常量和构造函数
├── proto/errdesc/           # 框架 proto：自定义 EnumValueOptions（http_status + error_message）
├── internal/protoc-gen-quix-gin/runtime/  # 插件运行时：Context 包装器、请求绑定、错误处理
├── examples/            # 各组件的可运行示例
└── openspec/             # 变更管理（specs + changes/archive）
```

## 关键约定

- **组件位置**: 所有框架组件放在 `core/<component>/` 下，不放在根包
- **接口优先**: 每个能力定义最小化 Go 接口，第三方库通过适配器模式封装
- **Option 模式**: App 配置使用 `func(*App)` 选项函数 — `quix.New(quix.WithLogger(...))`。默认实现零配置
- **Key-value 日志**: Logger 使用 `log.Info(ctx, "msg", "key", val)` 键值对交替风格
- **示例必须**: 每个组件必须在 `examples/<component>/` 下提供可运行的示例代码
- **写完代码必须格式化**: 每次 Write/Edit Go 文件后，必须执行 `go fmt ./...`
- **完成一组任务后必须 lint**: 每完成一个任务组（如接口定义、实现、测试），必须执行 `golangci-lint run ./...`
- **错误处理模式**: Handler 返回 error，由 `qhttp.Handler()` 包装为 gin.HandlerFunc，`ResponseMiddleware` 统一格式化 `{"error": {...}}` 响应

## protoc-gen-quix-gin 插件

从 proto service + `google.api.http` 注解自动生成 Gin 路由注册代码。

```bash
# 生成代码（buf）
cd examples/proto-api && buf generate

# 生成代码（protoc）
protoc -I proto --go_out=gen --go_opt=paths=source_relative \
  --quix-gin_out=gen --quix-gin_opt=paths=source_relative \
  proto/greeter/greeter.proto
```

生成 `xxx_gin.go` 文件，包含：服务接口（`XxxHTTPService`）、路由注册函数（`RegisterXxxHTTPService`）、handler 函数（使用 `runtime.Context` 包装器）。

## protoc-gen-quix-errors 插件

从 proto enum + `errdesc.http_status` / `errdesc.error_message` 注解自动生成错误码常量和无参构造函数。

```bash
# 生成代码（buf）
cd examples/proto-errors && buf generate
```

生成 `<enum_name>_errors.go` 文件，包含：错误码常量（`XxxCode`）、构造函数（`func Xxx() *apperrors.Error`）、WithDetails 变体（`func XxxWithDetails(details any) *apperrors.Error`）。Code/Message/StatusCode 全部来自 proto 定义，构造函数无参数。

用户 proto 文件需 import `errdesc/errdesc.proto`。enum 值命名约定：`<ENUM_NAME>_<VALUE_NAME>`，如 `UserError::USER_ERROR_NOT_FOUND` 生成函数 `UserNotFound`。

## 技术栈

- HTTP: Gin
- 默认日志: slog
- 可选日志: Zerolog、Zap（薄适配器）
- 配置: koanf
- 指标/追踪: OpenTelemetry（可选）
- IDL: protobuf + `google.api.http` 注解
- 测试: testify（断言）、go-cmp（diff）、Golden File（插件测试）
- Go 版本: 1.25

## 工作流

项目使用 OpenSpec 进行变更管理：
- `/opsx:propose <name>` — 创建变更提案，一次性生成所有制品（proposal、design、specs、tasks）
- `/opsx:explore` — 在变更前后探索想法
- `/opsx:apply` — 按任务清单实现变更
- `/opsx:archive` — 归档已完成的变更

变更制品存放在 `openspec/changes/<name>/` 下。规格文档使用 SHALL/MUST 规范语言和场景化测试标准。

## golangci-lint 配置

启用 12 个 linter（errcheck、govet、staticcheck、unused、gosimple、ineffassign、gosec、gofmt、goimports、misspell、gocritic）。示例代码放宽 errcheck 和 gosec 检查。
