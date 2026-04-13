## 1. Bug 修复

- [x] 1.1 ResponseMiddleware 添加 comma-ok 安全类型断言（`response.go`）
- [x] 1.2 Content-Type 协商改为检查 Accept 头（`http.tpl`）
- [x] 1.3 ExtraImports 添加 `slices.Sort` 确保生成代码幂等（`generator.go`）

## 2. 健壮性改进

- [x] 2.1 Recovery 使用 `c.Request.Context()` 替代 `context.Background()`，添加 `request_id` 日志（`recovery.go`）
- [x] 2.2 Recovery panic 返回统一 JSON 错误格式（`recovery.go`）
- [x] 2.3 删除 `var _ = emptypb.Empty{}` 哑引用和未使用的 import（`generator.go`）
- [x] 2.4 生成阶段校验 body field 是否存在于输入 message（`generator.go`）

## 3. 代码质量

- [x] 3.1 抽取 `server.SetAppError()` 公共函数（新建 `errors.go`）
- [x] 3.2 `handler.go` 的 Handler() 委托调用 SetAppError
- [x] 3.3 `runtime/context.go` 的 SetError() 委托调用 server.SetAppError
- [x] 3.4 移除模板中冗余的 `g.Group("")`（`http.tpl`）
- [x] 3.5 main.go 双重参数解析添加说明注释（`main.go`）
- [x] 3.6 新增 `WithReadHeaderTimeout` Option（`server.go`）

## 4. 验证

- [x] 4.1 更新 golden file（`go test -update`）
- [x] 4.2 重新生成 proto-api 示例（`buf generate`）
- [x] 4.3 运行全部测试、构建、lint 确认通过
