## 1. 框架源码

- [x] 1.1 `core/transport/http/server/errors.go` — import 别名 `apperrors` → `qerrors`
- [x] 1.2 `core/transport/http/server/context.go` — import 别名 `apperrors` → `qerrors`
- [x] 1.3 `core/transport/http/server/validate.go` — import 别名 `apperrors` → `qerrors`
- [x] 1.4 `core/transport/http/server/middleware/response.go` — import 别名 `apperrors` → `qerrors`
- [x] 1.5 `core/transport/http/server/handler_test.go` — import 别名 `apperrors` → `qerrors`
- [x] 1.6 `core/transport/http/server/runtime_test.go` — import 别名 `apperrors` → `qerrors`
- [x] 1.7 `core/transport/http/server/middleware/response_test.go` — import 别名 `apperrors` → `qerrors`
- [x] 1.8 `core/transport/http/server/validate_test.go` — import 别名 `apperrors` → `qerrors`
- [x] 1.9 `core/errors/predefined_test.go` — 如有引用则更新
- [x] 1.10 `core/errors/errors_test.go` — 如有引用则更新

## 2. protoc-gen-quix-errors 插件

- [x] 2.1 `cmd/protoc-gen-quix-errors/errors.tpl` — 模板中 `apperrors` → `qerrors`
- [x] 2.2 `cmd/protoc-gen-quix-errors/template.go` — `apperrorsImportPath` 变量/函数名 → `qerrorsImportPath`
- [x] 2.3 `cmd/protoc-gen-quix-errors/testdata/golden.golden` — import 别名 `apperrors` → `qerrors`
- [x] 2.4 运行 `go test ./cmd/protoc-gen-quix-errors/... -run TestGenerate -update` 验证 golden file

## 3. protoc-gen-quix-gin runtime

- [x] 3.1 `internal/protoc-gen-quix-gin/runtime/context.go` — 已移至 server 包，在 1.2 中已处理

## 4. 示例代码

- [x] 4.1 `examples/proto-demo/gen/errors/error_def_errors.go` — import 别名 `apperrors` → `qerrors`

## 5. 文档更新

- [x] 5.1 `CLAUDE.md` — 更新 `apperrors` 相关描述为 `qerrors`
- [x] 5.2 更新 `openspec/specs/protoc-gen-quix-errors/spec.md` 中的 `apperrors` → `qerrors`
- [x] 5.3 更新 `openspec/specs/protoc-gen-quix-gin-runtime/spec.md` 中的 `apperrors` → `qerrors`
- [x] 5.4 更新 `openspec/specs/gin-wrapper/spec.md` 中的 `apperrors` → `qerrors`

## 6. 验证

- [x] 6.1 `go build ./...` 确认编译通过
- [x] 6.2 `go test ./...` 确认所有测试通过
- [x] 6.3 `golangci-lint run ./...` 确认无 lint 警告
