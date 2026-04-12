## 1. 项目结构与依赖

- [x] 1.1 创建 `core/config/` 目录结构
- [x] 1.2 添加 koanf 及相关依赖（koanf、yaml、env provider、file provider）

## 2. Config 接口定义

- [x] 2.1 在 `core/config/config.go` 中定义 Config 接口（Get/String/Int/Bool/Bind）
- [x] 2.2 编写接口单元测试，验证编译期类型检查

## 3. koanf 默认实现

- [x] 3.1 在 `core/config/koanf.go` 中实现 koanf 适配器（NewKoanf 构造函数 + Option 模式）
- [x] 3.2 实现 WithFile 配置选项（从 YAML 文件加载）
- [x] 3.3 实现环境变量加载，确保环境变量优先于文件配置
- [x] 3.4 实现 Get/String/Int/Bool/Bind 方法
- [x] 3.5 支持点号分隔的嵌套键名访问
- [x] 3.6 编写 koanf 实现的单元测试

## 4. 集成到 App

- [x] 4.1 在 `option.go` 中添加 WithConfig 函数
- [x] 4.2 在 `quix.go` 的 App 结构体中添加 Config 字段
- [x] 4.3 提供 Config() 方法返回 App 的 Config
- [x] 4.4 编写 WithConfig 注入测试

## 5. 使用示例

- [x] 5.1 创建 `examples/config/` 目录
- [x] 5.2 编写 `examples/config/yaml/main.go`（YAML 文件加载示例 + 示例配置文件）
- [x] 5.3 编写 `examples/config/env/main.go`（环境变量覆盖示例）
- [x] 5.4 验证所有示例可通过 `go run` 执行
