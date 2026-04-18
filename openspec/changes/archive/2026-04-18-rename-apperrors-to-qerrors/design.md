## Context

`core/errors/` 包在框架源码和生成的代码中均使用 import 别名 `apperrors`。这一命名与 quix 框架其他组件风格不一致，且 `apperrors` 是过于通用的名称，容易与用户业务代码产生 import 冲突。当前共有 29 个文件引用了 `apperrors`。

## Goals / Non-Goals

**Goals:**
- 将 `apperrors` 别名统一重命名为 `qerrors`
- 更新所有框架源码、protoc 插件模板、golden file、示例代码中的引用
- 更新相关 spec 文档中的术语

**Non-Goals:**
- 不重命名 `core/errors/` 目录本身（目录名保持 `errors`，仅改 import 别名）
- 不修改 `Error` 类型定义或 API
- 不改动 protoc-gen-quix-errors 插件的功能逻辑

## Decisions

### 1. 仅改 import 别名，不改目录名

**选择**: `core/errors/` 目录保持不变，import 别名从 `apperrors` 改为 `qerrors`

**理由**: 目录名 `errors` 与 Go 标准库一致且语义清晰，问题仅出在 import 别名上。改别名即可解决命名冲突和风格不一致，无需引入目录重命名带来的额外复杂度。

### 2. 模板参数名同步更新

**选择**: 更新 `errors.tpl` 中的别名和 `template.go` 中的 `apperrorsImportPath` 函数名/变量名

**理由**: 保持代码可读性，避免函数名与实际行为不一致。

## Risks / Trade-offs

- **[Breaking Change]** 用户已生成的代码中 import 别名需要更新 → 生成新的代码即可自动修复，无需手动改动
- **[文档更新]** spec 文档中的 `apperrors` 引用需同步更新，否则文档与实现不一致 → 本次变更一并更新
