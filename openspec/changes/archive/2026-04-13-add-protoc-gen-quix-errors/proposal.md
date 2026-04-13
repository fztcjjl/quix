## Why

quix 的 `core/errors.Error` 使用手写字符串作为错误码（如 `errors.NotFound("user_not_found", "...")`），缺乏集中管理，容易拼写错误和跨服务不一致。需要一个独立的 protoc 插件，从 proto enum 定义自动生成错误码常量和构造函数，实现错误码的 IDL 驱动管理。

## What Changes

- 新增框架 proto 文件 `proto/errdesc/errdesc.proto`，定义自定义 field option `http_status` 和 `error_message`（标注在 `EnumValueOptions` 上）
- 新增独立 protoc 插件 `protoc-gen-quix-errors`，从 proto enum 生成 Go 错误码常量和构造函数
- 生成的代码包含：string 常量（错误码）、构造函数（`Xxx(msg) *Error`）、带 Details 变体（`XxxWithDetails(msg, details) *Error`）
- 新增使用示例 `examples/proto-errors/`

## Capabilities

### New Capabilities
- `protoc-gen-quix-errors`: 独立 protoc 插件，从 proto enum 定义生成错误码常量和构造函数
- `quix-error-proto`: 框架 proto 定义文件，提供 `http_status` 自定义 option

### Modified Capabilities
（无现有 capability 行为变更）

## Impact

- 新增 `proto/errdesc/` 目录下的 `errdesc.proto` 和 `errdesc.pb.go`
- 新增 `cmd/protoc-gen-quix-errors/` 目录（插件代码 + 测试）
- 新增 `examples/proto-errors/` 目录（使用示例）
- `CLAUDE.md` 更新插件使用说明
- go.mod 无需变更（protobuf/protogen 依赖已存在）
