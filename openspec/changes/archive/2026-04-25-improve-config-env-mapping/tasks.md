## 1. 依赖升级

- [x] 1.1 升级 koanf env provider 从 v1 到 v2：`go get github.com/knadh/koanf/providers/env/v2`，更新 `core/config/koanf.go` import 路径
- [x] 1.2 验证升级后现有测试是否通过（预期会失败，因为 env 命名规则变更）

## 2. 核心实现

- [x] 2.1 修改 `normalizeEnvKey` 函数：将 `strings.ReplaceAll(_, "_", ".")` 改为 `strings.ReplaceAll(_, "__", ".")`
- [x] 2.2 添加 `envPrefix string` 到 `options` 结构体，新增 `WithEnvPrefix(prefix string) Option`
- [x] 2.3 将 `env.Provider("", ".", normalizeEnvKey)` 替换为 V2 API：`env.Provider(".", env.Opt{Prefix: o.envPrefix, TransformFunc: ...})`，TransformFunc 中先 TrimPrefix 再调 normalizeEnvKey
- [x] 2.4 更新 `.env` 文件加载逻辑，使用与 env vars 相同的 `__` 分隔规则和 prefix 过滤

## 3. 测试更新

- [x] 3.1 更新 `koanf_test.go` 中所有环境变量命名为 `__` 格式（如 `SERVER_PORT` → `SERVER__PORT`，`APP_NAME` → `APP__NAME`）
- [x] 3.2 新增测试：嵌套 snake_case key 的环境变量覆盖（如 `SMS__ACCESS_KEY_ID` → `sms.access_key_id`）
- [x] 3.3 新增测试：`WithEnvPrefix` 过滤非前缀环境变量
- [x] 3.4 更新 `.env` 文件测试用例中的 key 格式

## 4. 示例更新

- [x] 4.1 更新 `examples/config/` 下的示例代码，环境变量命名改为 `__` 格式

## 5. 收尾

- [x] 5.1 运行 `go fmt ./...` 格式化代码
- [x] 5.2 运行 `golangci-lint run ./...` 确保无 lint 错误
- [x] 5.3 运行 `go test ./core/config/...` 确保所有测试通过
