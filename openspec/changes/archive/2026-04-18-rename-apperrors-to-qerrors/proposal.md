## Why

`core/errors/` 包在生成的代码中使用别名 `apperrors`（如 `import apperrors "github.com/fztcjjl/quix/core/errors"`），与 quix 框架其他组件的命名风格不一致（`qhttp` 等），且 `apperrors` 名称过于通用，容易与用户业务代码中的同名包产生 import 冲突。

## What Changes

- **BREAKING** 将 `core/errors/` 包的 import 别名从 `apperrors` 重命名为 `qerrors`
- 更新 `protoc-gen-quix-errors` 插件的代码生成模板，将 `apperrors` 替换为 `qerrors`
- 更新 `protoc-gen-quix-errors` 插件的 golden file 测试
- 更新 `protoc-gen-quix-gin` runtime 中的 import 别名引用
- 更新 `proto-demo` 示例中的生成代码 import
- 更新框架源码中所有引用 `apperrors` 别名的位置
- 更新相关 spec 文档中的 `apperrors` 引用

## Capabilities

### New Capabilities

（无新能力）

### Modified Capabilities

- `errors`: Error 包 import 别名从 `apperrors` 变更为 `qerrors`
- `protoc-gen-quix-errors`: 代码生成模板和 golden file 中的 import 别名变更
- `protoc-gen-quix-gin-runtime`: runtime 中引用 `apperrors` 的位置变更
- `gin-wrapper`: ResponseMiddleware 等引用 `apperrors` 的位置变更

## Impact

- **代码**: `core/errors/`、`core/transport/http/server/` 下多个文件、`cmd/protoc-gen-quix-errors/` 模板和测试、`cmd/protoc-gen-quix-gin/` runtime 引用
- **生成代码**: 所有通过 `protoc-gen-quix-errors` 生成的错误码文件需要重新生成（import 别名变更）
- **文档**: CLAUDE.md、相关 spec 文档中提及 `apperrors` 的位置
- **兼容性**: **BREAKING** — 使用生成代码的项目需要重新运行代码生成并更新 import 别名
