## Context

quix 的 protoc-gen-quix-gin 插件生成 Gin handler，handler 的请求处理流程为：`ShouldBind*` → `svc.Method()`。绑定成功后直接进入业务逻辑，无字段校验环节。

项目需要一种与 proto IDL 集成的校验方案，满足：校验规则定义在 proto 中（单一数据源）、代码生成自动处理（消除手写校验）、与现有错误体系兼容。

## Goals / Non-Goals

**Goals:**
- 提供 `ValidateRequest` 运行时函数，支持 protoc-gen-validate 生成的 `Validate()` 方法
- 生成的 handler 在绑定后自动调用校验
- 无校验规则的消息不受影响（no-op）
- 校验错误翻译为 `*apperrors.Error{Code: "validation_error", StatusCode: 400}`

**Non-Goals:**
- 不自建校验规则 DSL（使用行业标准的 `protoc-gen-validate`）
- 不修改 protoc-gen-validate 或 protovalidate 本身
- quix 本身不依赖 `protoc-gen-validate`（用户项目按需引入）

## Decisions

### 1. 使用 protoc-gen-validate（buf-validate）作为校验方案

**选择**: 用户在 proto 中使用 `[(buf.validate.field).string.min_len = 1]` 注解，`protoc-gen-validate` 插件生成 `Validate() error` 方法。

**替代方案**:
- go-playground/validator struct tags — rejected: proto 生成的 struct 无 validate tag
- 自定义 proto options + 新插件 — rejected: 重复造轮子，维护成本高
- 纯运行时 — rejected: 退化到 service 层手动校验

### 2. 接口检查实现零侵入集成

**选择**: 定义 `Validator` 接口（`Validate() error`），`ValidateRequest` 通过 type assert 检查。未实现接口的消息直接返回 nil，实现了则调用并翻译错误。

**理由**: 生成的 handler 无条件调用 `runtime.ValidateRequest(req)`，无需判断消息是否有校验规则。校验是运行时关注点，与代码生成解耦。

### 3. 通过接口解耦 protoc-gen-validate

**选择**: 错误解析通过 `fieldViolation` 接口（`Field() string` + `Reason() string`）和 `multiError` 接口（`Unwrap() []error`）提取字段违规信息，不导入 validate 包。

**理由**: quix 自身不依赖 `protoc-gen-validate`，校验库仅存在于用户项目中。

### 4. body "*" 路径变量绑定按冲突风险分路

**选择**: `RouteData` 新增 `PathVarConflict bool`，编译期检测路径变量名是否与 proto field 同名。模板据此二选一：同名 → `ShouldBindUriConflictCheck`（反射冲突检测），不同名 → `ShouldBindUri`（零开销普通绑定）。

**替代方案**: body "*" 有路径变量时始终调用 `ShouldBindUriConflictCheck`。 rejected — 不同名时无冲突可能，反射开销无意义。

**理由**: 大多数正常场景路径变量名不与 body 字段重名，走零开销的 `ShouldBindUri`；仅编译期警告的边界情况走反射冲突检测。

## Risks / Trade-offs

- [protoc-gen-validate 依赖] → 用户项目需自行添加 buf 依赖和插件配置，quix 只提供集成点
- [接口匹配] → protoc-gen-validate 生成的签名必须与 `Validator` 接口一致（当前一致：`Validate() error`）
- [无编译期校验] → proto 中写了校验规则但忘记配置 protoc-gen-validate 插件，运行时静默跳过。可通过 lint 或 CI 检查弥补
