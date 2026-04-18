### Requirement: protovalidate-go 运行时校验
runtime SHALL 使用 `protovalidate-go` 库对 proto message 进行运行时字段校验，不依赖代码生成。

#### Scenario: req 是 proto.Message 且校验通过
- **WHEN** 调用 `ValidateRequest(req)` 且 `req` 实现了 `proto.Message` 接口且校验通过
- **THEN** MUST 返回 nil

#### Scenario: req 是 proto.Message 且校验失败（单字段）
- **WHEN** 调用 `ValidateRequest(req)` 且 `req` 校验返回单字段违规
- **THEN** MUST 返回 `*qerrors.Error{Code: "validation_error", Message: "请求参数验证失败", StatusCode: 400, Details: [{Field, Message}]}`
- Field MUST 使用 `protovalidate.FieldPathString()` 格式化字段路径

#### Scenario: req 是 proto.Message 且校验失败（多字段）
- **WHEN** 调用 `ValidateRequest(req)` 且 `req` 校验返回多字段违规
- **THEN** MUST 返回 `*qerrors.Error`，Details 包含所有违规字段的 FieldViolation 列表

#### Scenario: req 未实现 proto.Message
- **WHEN** 调用 `ValidateRequest(req)` 且 req 不实现 `proto.Message` 接口
- **THEN** MUST 返回 nil（不执行校验）

#### Scenario: req 为 nil
- **WHEN** 调用 `ValidateRequest(nil)`
- **THEN** MUST 返回 nil

### Requirement: 生成的 handler 调用 ValidateRequest
protoc-gen-quix-gin 生成的 handler SHALL 在 `shouldBind(req)` 成功后、调用 service 方法前，无条件调用 `runtime.ValidateRequest(req)`。

#### Scenario: 正常请求流程
- **WHEN** 生成的 handler 处理请求，shouldBind 成功，ValidateRequest 返回 nil
- **THEN** MUST 继续调用 `svc.Method(c.Request.Context(), req)`

#### Scenario: 校验失败
- **WHEN** 生成的 handler 处理请求，shouldBind 成功，ValidateRequest 返回错误
- **THEN** MUST 调用 `c.SetError(err)` 并 return，不调用 service 方法

#### Scenario: 无校验规则
- **WHEN** 请求消息无校验规则（proto 未定义 `buf.validate.field` 注解）
- **THEN** protovalidate 返回 nil，handler 正常继续
