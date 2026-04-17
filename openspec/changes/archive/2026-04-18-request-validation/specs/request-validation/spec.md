## ADDED Requirements

### Requirement: Validator 接口
runtime SHALL 定义 `Validator` 接口，包含 `Validate() error` 方法，与 protoc-gen-validate 生成的签名一致。

#### Scenario: 接口定义
- **WHEN** protoc-gen-validate 为消息生成了 `Validate() error` 方法
- **THEN** 该消息 MUST 自动满足 `Validator` 接口

### Requirement: ValidateRequest 函数
runtime SHALL 提供 `ValidateRequest(req any) error` 函数，检查 req 是否实现 `Validator` 接口并执行校验。

#### Scenario: req 实现了 Validator 且校验通过
- **WHEN** 调用 `ValidateRequest(req)` 且 `req.Validate()` 返回 nil
- **THEN** MUST 返回 nil

#### Scenario: req 实现了 Validator 且校验失败（单字段）
- **WHEN** 调用 `ValidateRequest(req)` 且 `req.Validate()` 返回单个字段违规错误
- **THEN** MUST 返回 `*apperrors.Error{Code: "validation_error", Message: "请求参数验证失败", StatusCode: 400, Details: [{Field, Message}]}`

#### Scenario: req 实现了 Validator 且校验失败（多字段）
- **WHEN** 调用 `ValidateRequest(req)` 且 `req.Validate()` 返回多字段违规错误
- **THEN** MUST 返回 `*apperrors.Error`，Details 包含所有违规字段的 FieldViolation 列表

#### Scenario: req 未实现 Validator
- **WHEN** 调用 `ValidateRequest(req)` 且 req 不实现 `Validator` 接口
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
- **WHEN** 请求消息无校验规则（未使用 protoc-gen-validate）
- **THEN** ValidateRequest 返回 nil，handler 正常继续
