## Why

当前 protoc-gen-quix-gin 生成的 handler 在 `ShouldBind*` 绑定请求参数后直接调用 service 方法，缺少字段级校验。非法请求（如空标题、无效 ID）会直接透传到 service 层，需要每个 service 方法手动校验，产生重复代码。

## What Changes

- 新增 `ValidateRequest(req any) error` 运行时函数：通过接口检查判断 req 是否实现了 `Validator`（`Validate() error`），实现了则调用并翻译错误，未实现则跳过（no-op）
- 生成的 handler 在 `shouldBind` 成功后无条件调用 `runtime.ValidateRequest(req)`
- 错误翻译：通过接口解耦 `protoc-gen-validate`，将其 `ValidationError` 转换为 `*apperrors.Error{Code: "validation_error", StatusCode: 400, Details: [...]}`

## Capabilities

### New Capabilities
- `request-validation`: 请求字段绑定校验集成，定义 Validator 接口和 ValidateRequest 函数

### Modified Capabilities
- `protoc-gen-quix-gin`: 生成的 handler 在 shouldBind 后插入 `runtime.ValidateRequest(req)` 调用

## Impact

- `core/transport/http/server/validate.go` — 新增
- `core/transport/http/server/validate_test.go` — 新增
- `cmd/protoc-gen-quix-gin/http.tpl` — 修改
- `cmd/protoc-gen-quix-gin/testdata/golden.golden` — 更新
- `examples/proto-demo/` — 更新 buf.yaml、buf.gen.yaml、task.proto
