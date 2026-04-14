## Context

quix 有两个独立的 protoc 插件示例：`examples/proto-api`（Gin 路由生成）和 `examples/proto-errors`（错误码生成）。两者各自独立，无法展示插件配合使用的场景。`errdesc.proto` 已推送到 BSR（`buf.build/fztcjjl/quix`），示例应通过 buf 生态标准方式引用，不再本地复制。

## Goals / Non-Goals

**Goals:**
- 用一个真实业务场景（任务管理 CRUD）同时演示两个 protoc 插件
- 展示 `errors.proto` 与 `task.proto` 分离的组织方式
- 展示 BSR 依赖引用（`buf.build/fztcjjl/quix`）
- 展示生成的错误码在 handler 中的实际使用
- 示例自包含，`buf generate && go run main.go` 即可运行

**Non-Goals:**
- 不实现持久化存储（内存 map 即可）
- 不实现认证/鉴权
- 删除旧示例（proto-api、proto-errors 已被 proto-demo 取代）

## Decisions

### 1. 业务场景：Task（任务管理）

**选择**: Task CRUD — 创建、查询、列表、删除

**原因**: 简单易懂，但能覆盖多种 HTTP 方法（POST/GET/DELETE）和多种错误码（404/400/409/500）。

### 2. Proto 文件组织

**选择**: 错误定义集中到 `proto/errors/`，service 定义在 `proto/task/v1/`

```
proto/
├── errors/
│   └── errors.proto    # ErrorDef enum — 全项目统一错误定义
└── task/
    └── v1/
        └── task.proto  # TaskService + messages
```

**原因**: 与真实项目一致——错误码集中管理、与 service 解耦，便于跨 service 共享。

### 3. BSR 依赖引用

**选择**: `buf.yaml` 的 `deps` 引用 `buf.build/fztcjjl/quix` 和 `buf.build/googleapis/googleapis`

**原因**: errdesc.proto 已推送到 BSR，不应再本地复制。这是 buf 生态的标准做法。

### 4. 三个插件同时启用

**选择**: `buf.gen.yaml` 配置三个插件

```yaml
plugins:
  - protoc_builtin: go
  - local: protoc-gen-quix-gin
  - local: protoc-gen-quix-errors
```

**原因**: 完整演示 quix 的 protoc 插件工具链。go 插件生成 message 类型，quix-gin 生成路由，quix-errors 生成错误码。

### 5. 错误码定义

**选择**: 4 个错误码覆盖常见场景，使用 `ERROR_<模块>_<错误>` 命名约定

| 枚举值 | http_status | error_message |
|---|---|---|
| ERROR_TASK_UNSPECIFIED | 500 | 未知错误 |
| ERROR_TASK_NOT_FOUND | 404 | 任务不存在 |
| ERROR_TASK_TITLE_REQUIRED | 400 | 任务标题不能为空 |
| ERROR_TASK_ALREADY_DONE | 409 | 任务已完成 |

**原因**: 覆盖 4xx（参数校验、资源不存在、冲突）和 5xx（内部错误）三类场景。`ERROR_` 前缀统一标识错误定义，模块名作为二级前缀避免冲突。

### 6. 内存存储

**选择**: 使用 `sync.Map` 存储任务，UUID 作为 ID

**原因**: 无需外部依赖，启动即可用，示例足够简单。

### 7. 错误码集中管理（ErrorDef）

**选择**: 将 `enum TaskError` 重命名为 `enum ErrorDef`，集中到 `proto/errors/` 目录，作为项目统一错误定义

**原因**:
- 真实项目中多个模块（task、user、order...）的错误码应集中管理，而非分散在各模块 proto 中
- `ErrorDef` 比 `TaskError` 语义更明确——它是"错误定义"，不属于某个具体 service
- `errors.proto` 独立在 `proto/errors/` 目录，不与任何 service proto 耦合

**命名约定**: 枚举值采用 `ERROR_<模块>_<具体错误>` 格式

```protobuf
enum ErrorDef {
  ERROR_TASK_UNSPECIFIED    = 0;
  ERROR_TASK_NOT_FOUND      = 1;
  ERROR_TASK_TITLE_REQUIRED = 2;
  ERROR_TASK_ALREADY_DONE   = 3;
  ERROR_USER_NOT_FOUND      = 10;
  ...
}
```

生成结果：常量 `ErrorTaskNotFoundCode = "error task not found"`，函数 `func TaskNotFound()`。
无需修改 `protoc-gen-quix-errors` 的命名逻辑——`ErrorDef` 不以 `Error` 结尾，自动跳过当前的前缀拼接分支。

### 8. 错误码常量值格式

**选择**: 常量值使用小写空格分隔格式，而非 UPPER_SNAKE_CASE

```go
// 之前
ErrorTaskNotFoundCode = "ERROR_TASK_NOT_FOUND"

// 之后
ErrorTaskNotFoundCode = "error task not found"
```

**原因**:
- 错误码是面向开发者和日志的可读标识，小写空格格式更自然（类似英语句子）
- 常量名（Go 标识符）已提供 UPPER_SNAKE_CASE 的引用形式，运行时值无需重复
- 前端/日志中展示 `"error task not found"` 比 `"ERROR_TASK_NOT_FOUND"` 更友好

**实现**: `generator.go` 中新增 `toLowerSpace()` 函数，将 `UPPER_SNAKE_CASE` 转为 `lower space case`。

### 9. 修复 protoc-gen-quix-errors 无注解 enum bug

**问题**: 当前 `proto.GetExtension` 对未设置的字段返回 `(default, true)`，导致没有 `errdesc.http_status` 注解的 enum（如 `TaskStatus`）也被生成无效的错误码文件（StatusCode=0、函数名冲突）。

**修复**: 在 `generator.go` 中将 `proto.GetExtension` + `ok` 判断替换为 `proto.HasExtension`，只有显式设置了 `http_status` 才处理。

```go
// 之前
httpStatus, ok := proto.GetExtension(opts, errdesc.E_HttpStatus).(int32)
if !ok { continue }

// 之后
if !proto.HasExtension(opts, errdesc.E_HttpStatus) { continue }
httpStatus := proto.GetExtension(opts, errdesc.E_HttpStatus).(int32)
```

## Risks / Trade-offs

- **errdesc BSR 可用性**: 如果 BSR 不可达，`buf generate` 会失败 → 用户可改用本地 path 覆盖（buf 支持）
- **示例复杂度**: 比单独示例复杂 → 但更贴近真实项目，值得
- **ErrorDef 单文件膨胀**: 所有错误集中在一个 enum 中，大型项目可能数百个值 → 可通过编号分段（如 task=1-99, user=100-199）组织，未来如需拆分再引入多 enum
