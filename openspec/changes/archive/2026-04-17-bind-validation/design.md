## Context

protoc-gen-quix-gin 根据 `google.api.http` 注解生成 Gin handler，包含请求参数绑定逻辑。当前 bind 规范存在两个缺陷：GET/DELETE 方法允许声明 body（违反 HTTP 语义），以及 body `*` 与路径变量可能冲突（JSON body 覆盖路径变量值）。两者都静默通过，不产生任何警告或错误。

## Goals / Non-Goals

**Goals:**
- GET/DELETE 带 body 时编译期报错，终止代码生成
- body `*` 路径变量与 proto field 同名时编译期警告（提示潜在冲突，不终止生成）
- body `*` 有路径变量时运行时冲突检测（兜底 json_name 别名等绕过编译期检查的情况）
- body `*` 有路径变量但无冲突时，URI 正常覆盖/补充绑定

**Non-Goals:**
- 不处理 query 参数与 body 的冲突
- 不处理 header 绑定
- 不修改现有 ShouldBindUri 方法的行为

## Decisions

### 1. GET/DELETE body 校验：编译期 plugin.Error

**选择**: generator.go 遍历 bindings 时，GET 或 DELETE 且 `body != ""` → `plugin.Error` 终止生成。

**理由**: 违反 HTTP 规范，没有合理使用场景。

### 2. body `*` 路径变量同名：编译期警告

**选择**: generator 遍历路径变量，检查是否与 input message 字段同名（proto field name），同名 → 通过 `fmt.Fprintf(os.Stderr, "\033[33mwarning: ...\033[0m\n")` 黄色输出到 stderr，继续生成代码。

**替代方案**: `plugin.Error` 终止生成。 rejected — 同名只是冲突的潜在可能，body 可能不传该字段，运行时检测才是真正的安全网。

**理由**: 编译期提醒用户注意，运行时检测实际拦截。

### 3. body `*` 有路径变量：运行时冲突检测 + URI 绑定

**选择**: 模板生成 `ShouldBindJSON(req)` + `ShouldBindUriConflictCheck(req, pathVars)`。runtime 方法用反射检测 body 是否传了同名字段且值不一致，不一致返回错误；一致或 body 没传则正常绑定 URI。

**替代方案**: 不做运行时检测，URI 静默覆盖。 rejected — 用户传了冲突值却不知道，难以排查。

**理由**: 编译期按 proto field name 匹配，但 proto 的 `json_name` 可以给字段起别名，运行时 body 用的是 json name。两层互补。

### 4. ShouldBindUriConflictCheck 作为独立方法

**选择**: 新增 `ShouldBindUriConflictCheck(req any, pathVars []string) error`，不修改现有 `ShouldBindUri`。

**理由**: `ShouldBindUri` 已被 no-body 和 body-field 场景使用，冲突检测是 body `*` 特有需求，保持独立避免影响已有行为。

### 5. 反射实现冲突检测

**选择**: 在 `ShouldBindUriConflictCheck` 中用 reflect 遍历 json tag 匹配字段、比较值。

**理由**: 仅 body `*` 有路径变量时触发，reflect 开销相比 JSON 反序列化可忽略。作为代码生成器也可以在模板中生成硬编码比较，但需要传递字段类型信息增加模板复杂度，不值得。

## Risks / Trade-offs

- [反射开销] → 仅 body `*` 有路径变量时触发，1-2 次反射调用，开销可忽略
- [路径变量名与 json tag 不匹配] → findFieldByJSONTag 基于 SetTagName("json") 同一约定，一致性有保证
- [字段类型多样] → compareFieldValue 对 string 直接比较，int/uint 解析后比较，其他类型不走路径绑定不受影响
