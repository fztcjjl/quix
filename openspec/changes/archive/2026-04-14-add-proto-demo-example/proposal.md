## Why

当前 `examples/proto-api` 和 `examples/proto-errors` 分别独立演示 protoc-gen-quix-gin 和 protoc-gen-quix-errors，无法展示两者如何在实际项目中配合使用。需要一个综合性示例，演示从 proto 定义到 HTTP API + 错误处理的完整工作流。

## What Changes

- 新增 `examples/proto-demo/` 综合示例，包含任务管理 API（Task CRUD）
- proto 文件统一放在 `proto-demo/proto/task/v1/` 下，分离 `task.proto`（service + messages）和 `errors.proto`（error enum）
- `buf.yaml` 通过 BSR 依赖引用 `buf.build/fztcjjl/quix`（errdesc）和 `buf.build/googleapis/googleapis`
- `buf.gen.yaml` 同时启用 `go`、`protoc-gen-quix-gin`、`protoc-gen-quix-errors` 三个插件
- 生成的代码保存至 `proto-demo/gen/`
- 服务实现使用生成的错误码构造函数（如 `TaskNotFound()`、`TaskTitleRequired()`）
- 使用内存 map 模拟数据存储，无需数据库依赖

## Capabilities

### New Capabilities
（无，这是示例代码，不涉及框架能力变更）

### Modified Capabilities
（无）

## Impact

- 新增 `examples/proto-demo/` 目录（proto、gen、service、main.go、buf 配置）
- 删除 `examples/proto-api/`（已被 proto-demo 取代）
- 删除 `examples/proto-errors/`（已被 proto-demo 取代）
- 不修改任何框架代码
- 不新增外部依赖
