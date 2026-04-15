## 1. 移动 runtime 代码到 server 包

- [x] 1.1 将 `internal/protoc-gen-quix-gin/runtime/context.go` 移动到 `core/transport/http/server/context.go`，包名改为 `server`
- [x] 1.2 将 `internal/protoc-gen-quix-gin/runtime/bind.go` 移动到 `core/transport/http/server/bind.go`，包名改为 `server`
- [x] 1.3 将 `internal/protoc-gen-quix-gin/runtime/runtime_test.go` 移动到 `core/transport/http/server/runtime_test.go`，包名改为 `server`

## 2. 更新 protoc 插件 import 路径与命名

- [x] 2.1 修改 `cmd/protoc-gen-quix-gin/template.go` 中 `runtimeImportPath()` 返回值为 `github.com/fztcjjl/quix/core/transport/http/server`
- [x] 2.2 更新 `cmd/protoc-gen-quix-gin/testdata/golden.golden` 中的 import 路径
- [x] 2.3 运行 `go test ./cmd/protoc-gen-quix-gin/... -run TestGenerate -update` 更新 golden file
- [x] 2.4 修改 `cmd/protoc-gen-quix-gin/template.go`：`runtimeImportPath` → `serverImportPath`，`"runtimePkg"` → `"serverPkg"`，更新注释
- [x] 2.5 修改 `cmd/protoc-gen-quix-gin/http.tpl`：`{{runtimePkg}}` → `{{serverPkg}}`
- [x] 2.6 运行 `go test ./cmd/protoc-gen-quix-gin/... -run TestGenerate -update` 验证 golden file 无变化

## 3. 重新生成 proto-demo 示例代码

- [x] 3.1 安装更新后的 protoc-gen-quix-gin 插件（`go install ./cmd/protoc-gen-quix-gin`）
- [x] 3.2 在 `examples/proto-demo` 目录执行 `buf generate` 重新生成代码
- [x] 3.3 验证生成代码的 import 路径已更新为 `runtime "github.com/fztcjjl/quix/core/transport/http/server"`

## 4. 清理

- [x] 4.1 删除 `internal/protoc-gen-quix-gin/runtime/` 目录

## 5. 验证

- [x] 5.1 执行 `go fmt ./...`、`go build ./...`、`go test ./...`、`golangci-lint run ./...`
