# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

quix 是基于 Gin 的 Golang 快速开发框架，定位为薄封装。内置集成 Config/Log/Metrics/Tracing/Auth/Middleware 等基础设施组件，通过 Option 模式定制，提供合理的零配置默认值。

## 开发命令

```bash
go test ./...                              # 运行所有测试
go test ./core/logger/...                  # 运行指定组件的测试
go build ./...                             # 构建所有包
go fmt ./...                               # 格式化代码
go run examples/logger/slog/main.go        # 运行示例
```

无 Makefile 或构建系统，直接使用 Go 标准工具链。

## 项目架构

```
quix/
├── quix.go              # App 结构体、New()、Run()、Shutdown()
├── context.go           # Gin Context 薄封装（如使用）
├── option.go            # Option 类型和 WithXxx() 函数
├── core/                # 所有框架组件
│   ├── logger/          # 日志（默认 slog，可选 Zap/Zerolog 适配器）
│   ├── config/          # 配置加载（koanf）
│   ├── metrics/         # 指标收集（OpenTelemetry）
│   ├── tracing/         # 链路追踪（OpenTelemetry）
│   └── auth/            # 认证（默认 JWT，可插拔）
├── middleware/           # 内置 Gin 中间件（recovery、cors、ratelimit 等）
├── examples/            # 各组件的可运行示例
│   └── <component>/     # go run examples/<component>/xxx_example.go
└── internal/            # 框架内部工具，不对外暴露
```

## 关键约定

- **组件位置**: 所有框架组件放在 `core/<component>/` 下，不放在根包
- **接口优先**: 每个能力定义最小化 Go 接口，第三方库通过适配器模式封装
- **Option 模式**: App 配置使用 `func(*App)` 选项函数 — `quix.New(quix.WithLogger(...))`。默认实现零配置
- **Key-value 日志**: Logger 使用 `Info(ctx, "msg", "key", val)` 键值对交替风格（与 slog 一致）
- **示例必须**: 每个组件必须在 `examples/<component>/` 下提供可运行的示例代码

## 技术栈

- HTTP: Gin
- 默认日志: Go stdlib `slog`（零外部依赖）
- 可选日志: Zerolog、Zap（薄适配器）
- 配置: koanf
- 指标/追踪: OpenTelemetry（可选）
- Go 版本: 1.25

## 工作流

项目使用 OpenSpec 进行变更管理：
- `/opsx:propose <name>` — 创建变更提案，一次性生成所有制品（proposal、design、specs、tasks）
- `/opsx:explore` — 在变更前后探索想法
- `/opsx:apply` — 按任务清单实现变更
- `/opsx:archive` — 归档已完成的变更

变更制品存放在 `openspec/changes/<name>/` 下。规格文档使用 SHALL/MUST 规范语言和场景化测试标准。
