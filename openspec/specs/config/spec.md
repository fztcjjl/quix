## ADDED Requirements

### Requirement: Unified Config interface
quix 框架 SHALL 定义一个最小化的 Config 接口，提供按 key 读取配置值的能力。所有框架组件 MUST 通过此接口获取配置。

#### Scenario: Interface method signatures
- **WHEN** 开发者查看 Config 接口定义
- **THEN** 接口 SHALL 包含以下方法签名：
  - `Get(key string) any`
  - `String(key string) string`
  - `Int(key string) int`
  - `Bool(key string) bool`
  - `Bind(key string, target any) error`

### Requirement: koanf default implementation
框架 SHALL 提供基于 koanf 的默认 Config 实现。默认实现 MUST 支持从 YAML 文件和环境变量加载配置，环境变量 MUST 优先于文件配置。

#### Scenario: Load from YAML file
- **WHEN** 调用 `config.NewKoanf(config.WithFile("config.yaml"))` 加载一个包含 `server.port: 8080` 的 YAML 文件
- **THEN** `cfg.Int("server.port")` MUST 返回 `8080`

#### Scenario: Environment variable override
- **WHEN** 设置环境变量 `SERVER_PORT=9090` 并加载包含 `server.port: 8080` 的 YAML 文件
- **THEN** `cfg.Int("server.port")` MUST 返回 `9090`（环境变量优先）

#### Scenario: Non-existent key returns zero value
- **WHEN** 调用 `cfg.String("nonexistent.key")`
- **THEN** MUST 返回空字符串 `""`，不产生 panic

#### Scenario: Bind to struct
- **WHEN** 调用 `cfg.Bind("server", &ServerConfig{})` 且配置中存在 `server` 键
- **THEN** MUST 将对应配置值填充到结构体字段中

### Requirement: WithConfig option function
框架 SHALL 提供 `quix.WithConfig(cfg Config)` Option 函数，允许用户在创建 App 时注入自定义 Config 实现。

#### Scenario: Inject custom config via option
- **WHEN** 用户调用 `quix.New(quix.WithConfig(myConfig))`
- **THEN** App 的 Config MUST 等于用户传入的 `myConfig`

#### Scenario: Custom config implements interface
- **WHEN** 用户传入一个自定义结构体作为 Config
- **THEN** 自定义结构体 MUST 实现完整的 Config 接口（编译期检查）

### Requirement: Nested key access
Config 接口 MUST 支持通过点号分隔的路径访问嵌套配置值。

#### Scenario: Access nested value
- **WHEN** YAML 文件内容为 `server: { host: "localhost", port: 8080 }`
- **THEN** `cfg.String("server.host")` MUST 返回 `"localhost"`

### Requirement: Config usage examples
quix 框架 SHALL 在 `examples/config/` 目录下提供可运行的示例代码，演示 Config 的主要使用场景。

#### Scenario: YAML file loading example
- **WHEN** 开发者查看 `examples/config/yaml_example.go`
- **THEN** SHALL 演示从 YAML 文件加载配置并读取各种类型的值

#### Scenario: Environment variable example
- **WHEN** 开发者查看 `examples/config/env_example.go`
- **THEN** SHALL 演示环境变量覆盖文件配置的用法

#### Scenario: Example code is runnable
- **WHEN** 开发者在项目根目录执行 `go run examples/config/yaml_example.go`
- **THEN** MUST 编译通过并正常输出，无需额外配置
