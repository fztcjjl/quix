## 1. Runtime 包

- [x] 1.1 创建 `internal/protoc-gen-quix-gin/runtime/context.go` — Context 包装器、SetError、GetError
- [x] 1.2 创建 `internal/protoc-gen-quix-gin/runtime/bind.go` — ShouldBindUri、ShouldBindQuery（form decoder + SetTagName("json")）、ShouldBindJSON
- [x] 1.3 创建 runtime 单元测试

## 2. 插件骨架

- [x] 2.1 创建 `cmd/protoc-gen-quix-gin/main.go` — protogen 入口，SupportedFeatures、paths=source_relative 参数解析
- [x] 2.2 创建 `cmd/protoc-gen-quix-gin/pathutil.go` — ConvertPath、ExtractPathVars
- [x] 2.3 创建 `cmd/protoc-gen-quix-gin/template.go` — 模板数据结构、exportField、runtimeImportPath
- [x] 2.4 创建 `cmd/protoc-gen-quix-gin/http.tpl` — 代码生成模板（interface、RegisterFunc、Handler）

## 3. 代码生成逻辑

- [x] 3.1 创建 `cmd/protoc-gen-quix-gin/generator.go` — generateFile（解析 http 注解、路由提取、额外 import 收集、void 检测）
- [x] 3.2 支持 GET/POST/PUT/DELETE/PATCH 全 HTTP 方法
- [x] 3.3 支持 `body: "*"` 和 `body: "field_name"` 两种绑定模式
- [x] 3.4 支持 additional_bindings（一个 RPC 多条路由）
- [x] 3.5 支持 google.protobuf.Empty（接口 void 返回，handler 204）
- [x] 3.6 支持 Content-Type 协商（protobuf vs JSON）
- [x] 3.7 支持 paths=source_relative 文件输出路径
- [x] 3.8 创建 `cmd/protoc-gen-quix-gin/testdata/test.proto` — 覆盖所有场景的测试 proto
- [x] 3.9 创建 `cmd/protoc-gen-quix-gin/generator_test.go` — Golden File 测试

## 4. 依赖与构建

- [x] 4.1 添加依赖：google.golang.org/protobuf、google.golang.org/genproto、go-playground/form/v4、github.com/google/go-cmp/cmp、github.com/stretchr/testify
- [x] 4.2 验证 `go build ./...` 通过
- [x] 4.3 验证 `go test ./...` 通过
- [x] 4.4 验证 `golangci-lint run ./...` 通过

## 5. 示例项目

- [x] 5.1 创建 `examples/proto-api/proto/greeter/greeter.proto`
- [x] 5.2 创建 `examples/proto-api/buf.yaml`、`buf.gen.yaml`
- [x] 5.3 创建 `examples/proto-api/service/greeter.go` — 接口实现
- [x] 5.4 创建 `examples/proto-api/main.go` — 服务启动
- [x] 5.5 验证 `buf generate` 正常工作
- [x] 5.6 验证 `go run examples/proto-api/main.go` + curl 端到端通过
