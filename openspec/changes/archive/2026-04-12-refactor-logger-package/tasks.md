## 1. 包重命名

- [x] 1.1 将 `core/logger/` 目录重命名为 `core/log/`
- [x] 1.2 所有文件中 `package logger` → `package log`
- [x] 1.3 所有文件中 import 路径 `core/logger` → `core/log`

## 2. 全局默认 Logger

- [x] 2.1 在 `core/log/log.go` 中新增 noopLogger 实现
- [x] 2.2 新增全局变量 `defaultLogger`，初始为 noopLogger
- [x] 2.3 新增 `SetDefault(l Logger)` 函数
- [x] 2.4 新增包级函数 `Info/Error/Warn/Debug/With`，委托给 defaultLogger

## 3. App 集成

- [x] 3.1 `App.New()` 创建 Logger 后调用 `log.SetDefault(app.logger)`
- [x] 3.2 更新 `quix.go` 和 `option.go` 的 import 路径

## 4. 示例迁移

- [x] 4.1 将 `examples/logger/` 目录重命名为 `examples/log/`
- [x] 4.2 更新示例代码中的 import 路径和包名

## 5. 测试更新

- [x] 5.1 更新 `quix_test.go` 中 mock 的 import 路径
- [x] 5.2 新增全局默认 Logger 的测试（包级函数、SetDefault、noopLogger）

## 6. 文档与规范

- [x] 6.1 更新 `CLAUDE.md` 中 `logger/` 相关描述
- [x] 6.2 更新 `openspec/specs/logger/spec.md`（sync delta spec）
- [x] 6.3 运行 `go fmt ./...` 和 `golangci-lint run ./...` 验证
