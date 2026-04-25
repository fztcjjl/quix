## MODIFIED Requirements

### Requirement: koanf default implementation
框架 SHALL 提供基于 koanf 的默认 Config 实现。默认实现 MUST 支持从 YAML 文件、.env 文件和环境变量加载配置。环境变量 MUST 优先于 .env 文件，.env 文件 MUST 优先于 YAML 文件。

环境变量 MUST 使用双下划线 `__` 作为嵌套层级分隔符，单下划线 `_` MUST 保持原样不转换为分隔符。框架 SHALL 提供 `WithEnvPrefix(prefix string)` Option 用于指定环境变量前缀，仅加载匹配前缀的环境变量。

#### Scenario: Load from YAML file
- **WHEN** 调用 `config.NewKoanf(config.WithFile("config.yaml"))` 加载一个包含 `server.port: 8080` 的 YAML 文件
- **THEN** `cfg.Int("server.port")` MUST 返回 `8080`

#### Scenario: Environment variable override with double underscore
- **WHEN** 设置环境变量 `SERVER__PORT=9090` 并加载包含 `server.port: 8080` 的 YAML 文件
- **THEN** `cfg.Int("server.port")` MUST 返回 `9090`（环境变量优先）

#### Scenario: Nested env var preserves snake_case
- **WHEN** YAML 文件包含 `sms.access_key_id: "old_key"` 并设置环境变量 `SMS__ACCESS_KEY_ID=new_key`
- **THEN** `cfg.String("sms.access_key_id")` MUST 返回 `"new_key"`

#### Scenario: Single underscore in env var stays within key
- **WHEN** 设置环境变量 `SMS__POOL__MAX_SIZE=10`
- **THEN** 对应的配置路径 MUST 为 `sms.pool.max_size`（`__` 分隔嵌套，`_` 保持 snake_case）

#### Scenario: Env prefix filters system variables
- **WHEN** 调用 `config.NewKoanf(config.WithEnvPrefix("APP_"))` 且系统中存在 `PATH=/usr/bin` 等非 `APP_` 前缀的环境变量
- **THEN** koanf 配置中 MUST NOT 包含来自非 `APP_` 前缀环境变量的值

#### Scenario: Non-existent key returns zero value
- **WHEN** 调用 `cfg.String("nonexistent.key")`
- **THEN** MUST 返回空字符串 `""`，不产生 panic

#### Scenario: Bind to struct
- **WHEN** 调用 `cfg.Bind("server", &ServerConfig{})` 且配置中存在 `server` 键
- **THEN** MUST 将对应配置值填充到结构体字段中

#### Scenario: Bind with env override
- **WHEN** YAML 文件包含 `sms: { provider: "dev", access_key_id: "old" }` 并设置环境变量 `SMS__ACCESS_KEY_ID=new`
- **THEN** `cfg.Bind("sms", &SmsConfig{})` MUST 返回 `access_key_id` 为 `"new"`

## ADDED Requirements

### Requirement: .env file uses same delimiter convention
.env 文件中的 key MUST 与环境变量使用相同的双下划线 `__` 嵌套分隔规则。

#### Scenario: .env file with nested key
- **WHEN** .env 文件内容为 `SMS__ACCESS_KEY_ID=envkey123`
- **THEN** `cfg.String("sms.access_key_id")` MUST 返回 `"envkey123"`

#### Scenario: .env file overrides YAML file
- **WHEN** YAML 文件包含 `server.port: 8080` 且 .env 文件包含 `SERVER__PORT=3000`
- **THEN** `cfg.Int("server.port")` MUST 返回 `3000`

### Requirement: WithEnvPrefix option
框架 SHALL 提供 `WithEnvPrefix(prefix string)` Option 函数，用于指定环境变量的前缀过滤规则。

#### Scenario: Set env prefix
- **WHEN** 调用 `config.NewKoanf(config.WithEnvPrefix("MYAPP_"))` 并设置环境变量 `MYAPP_SERVER__PORT=8080`
- **THEN** `cfg.Int("server.port")` MUST 返回 `8080`

#### Scenario: Prefix is stripped from config key
- **WHEN** 调用 `config.NewKoanf(config.WithEnvPrefix("QUIX_"))` 并设置环境变量 `QUIX_SERVER__PORT=8080`
- **THEN** 前缀 `QUIX_` MUST 被剥离，配置路径为 `server.port`
