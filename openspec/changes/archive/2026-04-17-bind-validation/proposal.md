## Why

protoc-gen-quix-gin 在生成 HTTP handler 的 bind 代码时缺少关键校验：GET/DELETE 方法允许声明 body（违反 HTTP 语义，运行时字段全为零值），以及 body `*` 与路径变量存在歧义（JSON body 可覆盖路径变量值）。这些问题在编译期和运行时都静默通过，导致难以排查的 bug。

## What Changes

- 新增编译期校验：GET/DELETE 方法声明 body 时，plugin 报错并终止生成
- 新增编译期冲突警告：body `*` 且路径变量与 input message 字段同名时，输出警告提示潜在冲突（不终止生成）
- 新增运行时冲突检测：body `*` 有路径变量时（已通过编译期同名检查），绑定 URI 前检测 body 是否实际传了同名字段且值不一致，不一致则返回错误；无冲突则正常绑定 URI
- 模板更新：body `*` 有路径变量时，生成 `ShouldBindJSON(req)` + `ShouldBindUriConflictCheck(req, pathVars)`

## Capabilities

### New Capabilities

### Modified Capabilities
- `protoc-gen-quix-gin`: 新增 GET/DELETE body 编译期校验，body `*` 路径变量同名编译期警告，body `*` 有路径变量时模板增加冲突检测调用
- `protoc-gen-quix-gin-runtime`: 新增 ShouldBindUriConflictCheck 方法（运行时冲突检测 + URI 绑定）

## Impact

- `cmd/protoc-gen-quix-gin/generator.go` — 新增校验逻辑
- `cmd/protoc-gen-quix-gin/http.tpl` — body `*` 有路径变量时增加 ShouldBindUriConflictCheck 调用
- `cmd/protoc-gen-quix-gin/testdata/` — golden file 更新
- `core/transport/http/server/bind.go` — 新增 ShouldBindUriConflictCheck 方法
