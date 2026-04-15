## Context

`internal/protoc-gen-quix-gin/runtime` 包提供 Context 包装（`context.go`）和请求绑定（`bind.go`），是 `protoc-gen-quix-gin` 生成代码的运行时依赖。由于位于 `internal` 目录，Go 禁止外部模块导入，导致用户项目使用代码生成功能时编译失败。

runtime 包的内容（Context、ShouldBindUri/ShouldBindQuery/ShouldBindJSON、SetError/GetError）本质上是对 Gin 的薄封装，与 `core/transport/http/server/handler.go` 中 `Handler()` 函数的定位一致。

## Goals / Non-Goals

**Goals:**
- 将 runtime 包公开，解决外部项目无法导入的问题
- 保持生成代码的 API 不变（Context 类型和方法签名不变）

**Non-Goals:**
- 不修改 Context、Bind 等类型的行为或 API
- 不重构 handler.go 的实现
- 不修改 protoc 插件的代码生成逻辑（仅修改 import 路径）

## Decisions

### 1. 合并到 server 包而非新建包

**决策**: 将 `context.go`、`bind.go` 移动到 `core/transport/http/server/`，包名改为 `server`。

**替代方案**: 新建 `core/transport/http/server/ginruntime/` 或 `core/transport/http/server/runtime/`。

**选择理由**:
- runtime 内容与 handler.go 同属 Gin 薄封装，放在同一个包最自然
- 避免新增包的命名问题（runtime 遮蔽标准库、handler.Context 语义模糊等）
- 零新增包，结构更简洁

### 2. import 路径变更

**决策**: 生成代码的 import 从 `github.com/fztcjjl/quix/internal/protoc-gen-quix-gin/runtime` 变为 `github.com/fztcjjl/quix/core/transport/http/server`。

**影响**: 已有用户需重新生成代码（`buf generate`）。

**选择理由**: `core/transport/http/server` 已是公开包，生成代码使用该包的 Context 类型名变为 `server.Context`，语义清晰。

## Risks / Trade-offs

**[server 包膨胀]** → runtime 内容合并后 server 包文件增加。Mitigation: 文件按职责划分（handler.go、errors.go、context.go、bind.go），每个文件职责单一。
**[Breaking change]** → 生成代码 import 路径变更。Mitigation: 这是修复 internal 包无法导入的前提，用户重新生成即可。
**[Context 类型名冲突]** → 如果 server 包未来新增其他 Context 类型可能冲突。Mitigation: 当前 server 包无其他 Context 类型，短期内不是问题。
