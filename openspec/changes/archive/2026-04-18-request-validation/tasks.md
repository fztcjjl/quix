## 1. Runtime 实现

- [x] 1.1 新增 `core/transport/http/server/validate.go`：定义 `Validator` 接口、`ValidateRequest` 函数、`FieldViolation` 类型、接口解耦的错误解析
- [x] 1.2 新增 `core/transport/http/server/validate_test.go`：覆盖单字段错误、多字段错误、无错误、未实现接口、nil 五个场景 + 集成测试

## 2. 模板更新

- [x] 2.1 修改 `cmd/protoc-gen-quix-gin/http.tpl`：在 shouldBind 成功后插入 `runtime.ValidateRequest(req)` 调用
- [x] 2.2 更新 `cmd/protoc-gen-quix-gin/testdata/golden.golden`
- [x] 2.3 `RouteData` 新增 `PathVarConflict bool`，编译期同名检测时设置该标记
- [x] 2.4 `http.tpl` 中 body `*` + 路径变量分支按 `PathVarConflict` 分路：同名 → `ShouldBindUriConflictCheck`，不同名 → `ShouldBindUri`
- [x] 2.5 更新 `cmd/protoc-gen-quix-gin/testdata/golden.golden` 反映分路优化

## 3. 示例更新

- [x] 3.1 更新 `examples/proto-demo/buf.yaml`：添加 `buf.build/bufbuild/protovalidate` 依赖
- [x] 3.2 更新 `examples/proto-demo/buf.gen.yaml`：添加 `buf.build/bufbuild/validate-go` 插件
- [x] 3.3 更新 `examples/proto-demo/proto/task/v1/task.proto`：添加 `buf.validate.field` 注解

## 4. 验证

- [x] 4.1 运行全量测试 `go test ./...` 通过
- [x] 4.2 运行 `golangci-lint run ./...` 无告警
