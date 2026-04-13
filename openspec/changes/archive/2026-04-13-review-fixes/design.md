## Context

代码审检发现 gin-wrapper 和 protoc-gen-quix-gin 存在 Bug 级别问题（panic 风险、响应格式错误、生成代码不稳定）和健壮性缺陷（上下文丢失、未校验输入、代码重复）。变更已实施完毕，此处记录技术决策。

## Goals / Non-Goals

**Goals:**
- 修复可能导致生产 panic 的缺陷
- 确保生成代码幂等（ExtraImports 排序）
- 消除错误处理逻辑重复（SetAppError 抽取）
- 提升可观测性（Recovery 保留请求上下文）

**Non-Goals:**
- 不新增功能或 API
- 不改变外部接口签名（生成代码的接口定义不变）
- 不处理 #10（formDecoder 全局变量）和 #13（exportField 替代方案），当前可接受

## Decisions

### 1. SetAppError 抽取到 server 包而非 errors 包

**选择**: 放在 `core/transport/http/server/errors.go`

**原因**: SetAppError 依赖 gin.Context，是 HTTP 层概念，不属于 core/errors（纯错误定义，无 gin 依赖）。runtime 包依赖 server 包是单向的（runtime → server），不产生循环依赖。

**替代方案**: 放在 core/errors — 被否决，因为 errors 包不应引入 gin 依赖。

### 2. runtime.Context.SetError 委托 server.SetAppError

**选择**: SetError 内部调用 server.SetAppError(c.Context, err)

**原因**: 消除 15 行重复代码，确保两处行为一致。生成的 handler 调用 runtime.SetError，手动写的 handler 调用 qhttp.Handler，两者最终走同一条路径。

### 3. Content-Type 协商使用 Accept 头而非请求 Content-Type

**选择**: `c.GetHeader("Accept") == "application/x-protobuf"`

**原因**: 响应格式由客户端期望决定（Accept），而非客户端发送的数据格式（Content-Type）。这与 proto-gen-go-grpc 的行为一致。

### 4. Recovery panic 返回统一 JSON 格式

**选择**: 使用 AbortWithStatusJSON 返回 `{"error": {"code": "internal_error", ...}}`

**原因**: 与框架其他错误响应格式一致，客户端可以统一解析。

## Risks / Trade-offs

- **runtime 依赖 server 包**: 引入 internal/protoc-gen-quix-gin/runtime → core/transport/http/server 依赖。评估为低风险——runtime 是生成代码的运行时库，依赖 HTTP server 包是合理的。
- **Recovery JSON body 在 ResponseMiddleware 之前**: Recovery 使用 AbortWithStatusJSON 直接写入响应，不经过 ResponseMiddleware。这是正确的——panic 不应产生正常错误格式，且 ResponseMiddleware 在 Recovery 之后执行，不会被触发。
