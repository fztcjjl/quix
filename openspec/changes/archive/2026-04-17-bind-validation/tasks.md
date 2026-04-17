## 1. Runtime 冲突检测实现

- [x] 1.1 在 `core/transport/http/server/bind.go` 中实现 `ShouldBindUriConflictCheck(req any, pathVars []string) error`：反射查找 json tag 匹配的字段，检测非零值冲突，无冲突时绑定 URI
- [x] 1.2 为 `ShouldBindUriConflictCheck` 编写单元测试：覆盖无冲突、值一致、值不一致、非 struct 指针四个场景

## 2. Generator 编译期校验

- [x] 2.1 在 `cmd/protoc-gen-quix-gin/` 新增 `bind_validation.go`，定义 `bindErrorf(plugin, format, args)` 和 `bindWarnf(format, args)` 辅助函数（分别通过 plugin.Error 红色输出 + 黄色 stderr 输出）
- [x] 2.2 在 `cmd/protoc-gen-quix-gin/generator.go` 中新增 GET/DELETE body 校验：调用 `bindErrorf` 终止生成
- [x] 2.3 在 `cmd/protoc-gen-quix-gin/generator.go` 中新增 body `*` 路径变量同名警告：遍历 PathVars，检查是否与 input message 字段同名，同名时调用 `bindWarnf` 输出黄色警告并继续生成
- [x] 2.3 为以上校验和警告补充 generator 测试

## 3. 模板更新

- [x] 3.1 修改 `cmd/protoc-gen-quix-gin/http.tpl` 中 body `*` 分支：当有路径变量时生成 `ShouldBindJSON(req)` + `ShouldBindUriConflictCheck(req, pathVars)`；无路径变量时保持原有 `ShouldBindJSON(req)`

## 4. 测试更新

- [x] 4.1 在 `cmd/protoc-gen-quix-gin/testdata/test.proto` 中新增 body `*` + 路径变量（不同名）的 RPC 定义
- [x] 4.2 更新 `cmd/protoc-gen-quix-gin/testdata/golden.golden`，运行 `go test ... -run TestGenerate -update`
- [x] 4.3 运行全量测试 `go test ./...` 验证通过
- [x] 4.4 运行 `golangci-lint run ./...` 验证通过
