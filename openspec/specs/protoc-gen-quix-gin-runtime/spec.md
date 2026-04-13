### Requirement: Context 包装器
Runtime SHALL 在 `internal/protoc-gen-quix-gin/runtime/` 中提供 `Context` 结构体，嵌入 `*gin.Context`。

#### Scenario: Context 创建
- **WHEN** 生成代码创建 `&runtime.Context{Context: ginCtx}`
- **THEN** MUST 返回嵌入 `*gin.Context` 的 `Context` 实例，可访问所有 gin.Context 方法

### Requirement: SetError 错误处理
Context SHALL 提供 `SetError(err error)` 方法，内部委托 `server.SetAppError` 实现，确保与 qhttp.Handler 的错误处理行为一致。

#### Scenario: apperrors.Error
- **WHEN** 调用 `SetError(&apperrors.Error{Code: "not_found", StatusCode: 404})`
- **THEN** MUST 将 error 存入 `c.Set("app_error", err)`，并调用 `c.AbortWithStatus(404)`

#### Scenario: 标准 error
- **WHEN** 调用 `SetError(fmt.Errorf("db failed"))`
- **THEN** MUST 包装为 `*Error{Code: "internal_error", StatusCode: 500}`，存入 `app_error`，并调用 `c.AbortWithStatus(500)`

### Requirement: GetError 获取错误
Context SHALL 提供 `GetError() *apperrors.Error` 方法。

#### Scenario: 有错误
- **WHEN** 之前调用了 `SetError(err)`
- **THEN** MUST 返回存储的 `*apperrors.Error`

#### Scenario: 无错误
- **WHEN** 未调用 `SetError`
- **THEN** MUST 返回 nil

### Requirement: ShouldBindUri 路径参数绑定
Context SHALL 提供 `ShouldBindUri(req any) error` 方法，使用 form decoder 将路径参数绑定到 req。

#### Scenario: 路径参数绑定
- **WHEN** 调用 `ShouldBindUri(&HelloRequest{})` 且路径中有 `:name=world`
- **THEN** MUST 将 `world` 值绑定到 `req.Name` 字段

#### Scenario: json tag 匹配
- **WHEN** proto 生成结构体的字段使用 `json:"user_id"` tag
- **THEN** MUST 通过 json tag 名称匹配路径参数（而非 uri tag）

### Requirement: ShouldBindQuery 查询参数绑定
Context SHALL 提供 `ShouldBindQuery(req any) error` 方法，使用 form decoder 将查询参数绑定到 req。

#### Scenario: 查询参数绑定
- **WHEN** 调用 `ShouldBindQuery(&SearchRequest{})` 且 URL 为 `?query=hello&page_size=10`
- **THEN** MUST 将 `hello` 绑定到 `Query`，`10` 绑定到 `PageSize`

### Requirement: ShouldBindJSON 请求体绑定
Context SHALL 提供 `ShouldBindJSON(req any) error` 方法，将 JSON 请求体绑定到 req。

#### Scenario: JSON 绑定
- **WHEN** 调用 `ShouldBindJSON(&CreateUserRequest{})` 且请求体为 `{"name":"alice"}`
- **THEN** MUST 将 `alice` 绑定到 `Name` 字段
