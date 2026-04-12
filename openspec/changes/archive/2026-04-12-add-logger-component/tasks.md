## 1. 项目结构与依赖

- [x] 1.1 创建 `core/logger/` 目录结构
- [x] 1.2 添加 Gin 依赖到 go.mod

## 2. Logger 接口定义

- [x] 2.1 在 `core/logger/logger.go` 中定义 Logger 接口（Info/Error/Warn/Debug/With）
- [x] 2.2 编写接口单元测试，验证编译期类型检查

## 3. slog 默认实现

- [x] 3.1 在 `core/logger/slog.go` 中实现 slog 适配器（NewSlog 构造函数）
- [x] 3.2 实现 With 方法，返回携带附加字段的 slog Logger
- [x] 3.3 处理奇数个 args 的边界情况（忽略多余的最后一个 key）
- [x] 3.4 编写 slog 实现的单元测试

## 4. Zerolog 适配器

- [x] 4.1 在 `core/logger/zerolog.go` 中实现 Zerolog 适配器（NewZerolog 构造函数）
- [x] 4.2 实现 With 方法，返回携带附加字段的 Zerolog Logger
- [x] 4.3 编写 Zerolog 实现的单元测试

## 5. Zap 适配器

- [x] 5.1 在 `core/logger/zap.go` 中实现 Zap 适配器（NewZap 构造函数）
- [x] 5.2 实现 With 方法，返回携带附加字段的 Zap Logger（使用 zap.SugaredLogger）
- [x] 5.3 编写 Zap 实现的单元测试

## 6. Option 函数

- [x] 6.1 在根包创建 `option.go`，定义 Option 类型和 WithLogger 函数
- [x] 6.2 创建 App 结构体骨架（quix.go），包含 Logger 字段和 New() 构造函数
- [x] 6.3 验证 WithLogger 注入到 App 的行为

## 7. 使用示例

- [x] 7.1 创建 `examples/logger/` 目录
- [x] 7.2 编写 `examples/logger/slog/main.go`（默认 slog 用法）
- [x] 7.3 编写 `examples/logger/zerolog/main.go`（Zerolog 用法）
- [x] 7.4 编写 `examples/logger/zap/main.go`（Zap 用法）
- [x] 7.5 验证所有示例可通过 `go run` 执行
