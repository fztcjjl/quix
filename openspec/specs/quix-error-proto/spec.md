## ADDED Requirements

### Requirement: 自定义 http_status 和 error_message option
quix 框架 SHALL 在 `proto/errdesc/errdesc.proto` 中定义两个自定义 field option，扩展 `google.protobuf.EnumValueOptions`：
- `int32 http_status = 84001` — HTTP 状态码
- `string error_message = 84002` — 错误消息

#### Scenario: option 定义
- **WHEN** 查看 `proto/errdesc/errdesc.proto`
- **THEN** MUST 包含 `extend google.protobuf.EnumValueOptions { int32 http_status = 84001; string error_message = 84002; }`

#### Scenario: Go package 路径
- **WHEN** proto 文件 option `go_package` 为 `github.com/fztcjjl/quix/proto/errdesc;errdesc`
- **THEN** 生成的 Go 代码 MUST 放在 `proto/errdesc/` 目录下，包名为 `errdesc`

#### Scenario: 生成的 Go 代码可用
- **WHEN** protoc 插件 import `errdesc "github.com/fztcjjl/quix/proto/errdesc"`
- **THEN** MUST 能通过 `errdesc.E_HttpStatus` 和 `errdesc.E_ErrorMessage` 访问扩展描述符

### Requirement: 生成的 pb.go 提交到仓库
`proto/errdesc/errdesc.pb.go` SHALL 由 protoc-gen-go 生成并提交到 git 仓库。

#### Scenario: 文件存在
- **WHEN** 检查仓库中 `proto/errdesc/errdesc.pb.go`
- **THEN** MUST 存在且包含 `E_HttpStatus` 和 `E_ErrorMessage` 变量定义
