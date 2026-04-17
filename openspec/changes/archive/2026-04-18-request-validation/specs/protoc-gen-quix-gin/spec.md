## MODIFIED Requirements

### Requirement: 请求绑定后的校验调用
生成的 handler SHALL 在 `shouldBind(req)` 成功后、调用 service 方法前，调用 `runtime.ValidateRequest(req)` 进行字段校验。

#### Scenario: 正常请求
- **WHEN** shouldBind 成功，ValidateRequest 返回 nil
- **THEN** MUST 继续调用 service 方法

#### Scenario: 校验失败
- **WHEN** shouldBind 成功，ValidateRequest 返回错误
- **THEN** MUST 调用 `c.SetError(err)` 并 return，不调用 service 方法

### Requirement: body 星号路径变量绑定按冲突风险分路
当 `body: "*"` 且有路径变量时，插件 SHALL 根据编译期同名检测结果选择不同的 URI 绑定方式，避免无冲突场景的不必要反射开销。

#### Scenario: 路径变量名与 body 字段同名
- **WHEN** `body: "*"` 且路径变量名与 input message 字段同名
- **THEN** 生成的 handler MUST 调用 `ShouldBindUriConflictCheck(req, pathVars)` 进行反射冲突检测

#### Scenario: 路径变量名与 body 字段不同名
- **WHEN** `body: "*"` 且路径变量名与 input message 所有字段不同名
- **THEN** 生成的 handler MUST 调用 `ShouldBindUri(req)` 进行零开销普通绑定
