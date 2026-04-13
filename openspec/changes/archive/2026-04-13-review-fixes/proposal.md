## Why

资深架构视角审检 gin-wrapper（`core/transport/http/server/`）和 protoc-gen-quix-gin 代码生成工具后，发现若干 Bug、健壮性缺陷和代码质量问题需要修复。这些问题会影响生产稳定性（如 panic 风险、错误响应格式）和开发体验（如生成代码不稳定）。

## What Changes

- **Bug 修复**: ResponseMiddleware 不安全类型断言、Content-Type 协商检查错请求头、ExtraImports 生成顺序不确定
- **健壮性改进**: Recovery 中间件丢失请求上下文、protoc 插件 emptypb 哑引用、body field 未校验
- **代码质量**: 错误处理逻辑重复抽取、冗余 `g.Group("")` 移除、参数解析注释补全
- **小优化**: ReadHeaderTimeout 可配置、Recovery 返回统一 JSON 格式

## Capabilities

### New Capabilities

（无新增能力）

### Modified Capabilities

- `gin-wrapper`: ResponseMiddleware 安全断言、Recovery 上下文保留与 JSON 响应、错误处理抽取为 SetAppError、ReadHeaderTimeout 可配置
- `protoc-gen-quix-gin`: Accept 头协商、ExtraImports 排序、emptypb 清理、body field 校验、g.Group("") 移除、参数解析注释

## Impact

- `core/transport/http/server/middleware/response.go` — 类型断言安全化
- `core/transport/http/server/middleware/recovery.go` — 上下文保留 + JSON 响应
- `core/transport/http/server/server.go` — 新增 WithReadHeaderTimeout Option
- `core/transport/http/server/handler.go` — 使用 SetAppError 消除重复
- `core/transport/http/server/errors.go` — 新文件，SetAppError 公共函数
- `cmd/protoc-gen-quix-gin/generator.go` — ExtraImports 排序、body field 校验、删除 emptypb
- `cmd/protoc-gen-quix-gin/http.tpl` — Accept 头、移除 g.Group("")
- `cmd/protoc-gen-quix-gin/main.go` — 参数解析注释
- `internal/protoc-gen-quix-gin/runtime/context.go` — 委托 SetAppError
- `cmd/protoc-gen-quix-gin/testdata/golden.golden` — 更新生成输出
- `examples/proto-api/gen/greeter/greeter_gin.go` — 重新生成
