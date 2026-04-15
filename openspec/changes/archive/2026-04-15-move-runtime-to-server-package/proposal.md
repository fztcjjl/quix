## Why

`internal/protoc-gen-quix-gin/runtime` 是 `protoc-gen-quix-gin` 生成代码的运行时依赖，当前放在 `internal` 下导致外部项目无法导入。runtime 包（Context 包装、请求绑定、错误处理）本质上是对 Gin 的薄封装，与 `core/transport/http/server/handler.go` 的定位一致，应合并到公开的 `server` 包中。

## What Changes

- 将 `internal/protoc-gen-quix-gin/runtime/` 下的 `context.go`、`bind.go` 移动到 `core/transport/http/server/`，包名改为 `server`
- 删除 `internal/protoc-gen-quix-gin/runtime/` 目录
- 更新 `cmd/protoc-gen-quix-gin/template.go` 中 `runtimeImportPath()` 返回的 import 路径
- 更新生成代码中的 import 路径（`examples/proto-demo/gen/task/v1/task_gin.go`）
- 更新 golden test 文件（`cmd/protoc-gen-quix-gin/testdata/golden.golden`）
- runtime 测试移动到 `core/transport/http/server/` 包下

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

（无——纯包路径调整，行为不变）

## Impact

- **移动代码**: `internal/protoc-gen-quix-gin/runtime/context.go`、`bind.go` → `core/transport/http/server/`
- **删除目录**: `internal/protoc-gen-quix-gin/runtime/`
- **修改代码**: `cmd/protoc-gen-quix-gin/template.go`（import 路径）
- **修改生成代码**: `examples/proto-demo/gen/task/v1/task_gin.go`（import 路径）
- **修改测试**: golden test 文件、runtime 测试移动
- **BREAKING**: 生成代码的 import 路径从 `internal/.../runtime` 变为 `core/transport/http/server`，用户需重新生成代码
