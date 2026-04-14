## 1. Proto 文件

- [x] 1.1 创建 `examples/proto-demo/proto/task/v1/errors.proto` — TaskError enum（4 个错误码）
- [x] 1.2 创建 `examples/proto-demo/proto/task/v1/task.proto` — TaskService（4 个 RPC）+ messages

## 2. Buf 配置

- [x] 2.1 创建 `examples/proto-demo/buf.yaml` — 模块配置，deps 引用 BSR（quix + googleapis）
- [x] 2.2 创建 `examples/proto-demo/buf.gen.yaml` — 启用 go + quix-gin + quix-errors 三个插件

## 3. 代码生成与验证

- [x] 3.1 执行 `buf generate` 验证生成成功
- [x] 3.2 检查生成产物（路由注册、错误码常量、构造函数）

## 4. 服务实现

- [x] 4.1 创建 `examples/proto-demo/service/task.go` — TaskService 实现（内存存储 + 错误码使用）

## 5. 示例入口

- [x] 5.1 创建 `examples/proto-demo/main.go` — 注册路由 + 启动服务
- [x] 5.2 端到端验证：`buf generate && go run main.go` + curl 测试

## 6. 清理旧示例

- [x] 6.1 删除 `examples/proto-api/`
- [x] 6.2 删除 `examples/proto-errors/`

## 7. 收尾

- [x] 7.1 更新 `CLAUDE.md` — 移除 proto-api/proto-errors 相关说明，添加 proto-demo 说明
- [x] 7.2 运行全部测试、构建、lint

## 8. 错误码集中管理 + 插件 bug 修复

- [x] 8.1 修复 `cmd/protoc-gen-quix-errors/generator.go` — `proto.GetExtension` + `ok` 改为 `proto.HasExtension`
- [x] 8.2 修改常量值格式：新增 `toLowerSpace()`，`Code` 从 `UPPER_SNAKE_CASE` 改为 `lower space case`
- [x] 8.3 更新 `cmd/protoc-gen-quix-errors/testdata/golden.golden` — 匹配修复后的行为
- [x] 8.2 更新 `cmd/protoc-gen-quix-errors/testdata/test.proto` — 确认无注解 enum（RegularEnum）不被生成
- [x] 8.3 更新 `cmd/protoc-gen-quix-errors/testdata/golden.golden` — 匹配修复后的行为
- [x] 8.4 重构 proto-demo 错误定义：`proto/task/v1/errors.proto` → `proto/errors/errors.proto`，`TaskError` → `ErrorDef`，值名 `TASK_ERROR_*` → `ERROR_TASK_*`
- [x] 8.5 更新 `proto/task/v1/task.proto` 的 import 路径
- [x] 8.6 重新 `buf generate` 并验证生成产物（无注解 enum 不再生成文件）
- [x] 8.7 更新 `service/task.go` 的 import（错误码从 `errors` 包引用）
- [x] 8.8 更新 `CLAUDE.md` — proto 目录结构、命名约定说明

## 9. 错误码常量值格式调整

- [x] 9.1 `template.go` 新增 `toLowerSpace()` 函数（UPPER_SNAKE → lower space）
- [x] 9.2 `generator.go` 的 `Code` 字段从 `valueName` 改为 `toLowerSpace(valueName)`
- [x] 9.3 更新 golden file 和 proto-demo 生成产物
