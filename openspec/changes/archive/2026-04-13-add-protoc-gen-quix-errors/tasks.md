## 1. 框架 Proto 定义

- [x] 1.1 创建 `proto/errdesc/errdesc.proto` — 定义 `http_status` 和 `error_message` 自定义 option
- [x] 1.2 编译 proto 生成 `proto/errdesc/errdesc.pb.go`

## 2. 插件骨架

- [x] 2.1 创建 `cmd/protoc-gen-quix-errors/main.go` — 插件入口（复用 protoc-gen-quix-gin 模式）
- [x] 2.2 创建 `cmd/protoc-gen-quix-errors/template.go` — 数据结构 + toPascalCase/toSnakeCase 辅助函数
- [x] 2.3 创建 `cmd/protoc-gen-quix-errors/errors.tpl` — Go 模板（常量 + 构造函数 + WithDetails 变体）

## 3. 核心生成逻辑

- [x] 3.1 创建 `cmd/protoc-gen-quix-errors/generator.go` — 遍历 enum、读取 http_status extension、前缀剥离逻辑
- [x] 3.2 实现文件命名（`snake_case(enum_name)_errors.go`）
- [x] 3.3 实现无注解 enum 跳过逻辑

## 4. 测试

- [x] 4.1 创建 `testdata/test.proto` — 带注解 enum + 无注解 enum
- [x] 4.2 创建 `testdata/golden.golden` — 期望生成输出
- [x] 4.3 创建 `generator_test.go` — Golden file 测试 + 辅助函数单元测试
- [x] 4.4 验证测试通过

## 5. 示例

- [x] 5.1 创建 `examples/proto-errors/` 目录结构（buf.yaml、buf.gen.yaml、proto、main.go）
- [x] 5.2 端到端验证：`buf generate && go run main.go`

## 6. 收尾

- [x] 6.1 更新 `CLAUDE.md` 添加 protoc-gen-quix-errors 使用说明
- [x] 6.2 运行全部测试、构建、lint
