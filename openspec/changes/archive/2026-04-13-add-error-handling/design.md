## Context

quix HTTP Server 目前没有统一的错误响应格式。开发者需要手动构建错误 JSON，格式各异。需要一个标准的错误类型和响应机制，确保 API 错误响应一致。应用层只需编写返回 `error` 的 handler 函数，框架自动处理错误响应格式化和 HTTP 状态码。request_id 由已有的 RequestID 中间件负责，通过响应头传递。

## Goals / Non-Goals

**Goals:**
- 定义 `Error` 类型，Code 为字符串、Details 为 `any`（最灵活）
- 提供预定义错误函数（BadRequest、NotFound、Unauthorized、Internal、Forbidden）
- 提供 `Handler()` 包装器，将 `func(c *gin.Context) error` 转为 gin.HandlerFunc，自动处理错误响应
- ResponseMiddleware 自动格式化错误响应
- ResponseMiddleware 默认挂载到 HTTP Server

**Non-Goals:**
- 不包装成功响应，用户直接用 `c.JSON(status, data)`
- 不实现验证器集成（后续独立 change）
- 不实现国际化（i18n 留给上层处理）

## Decisions

### 1. Error 类型放在 `core/errors/`

```go
type Error struct {
    Code        string `json:"code"`
    Message     string `json:"message"`
    Details     any    `json:"details,omitempty"`
    StatusCode  int    `json:"-"`
}

func (e *Error) Error() string { return e.Message }
```

- `Code` 使用字符串，如 `"param_invalid"`，比整型更具可读性
- `Details` 为 `any`，支持数组、map、自定义结构等任意类型
- `StatusCode` 为 HTTP 状态码，使用 `json:"-"` 不序列化到 JSON 响应体中，仅用于设置 HTTP 响应状态码

### 2. Handler 包装器放在 `core/transport/http/server/`

Handler 包装器属于 HTTP 传输层关注点（依赖 gin），放在 `core/transport/http/server/handler.go`。

```go
func Handler(fn func(c *gin.Context) error) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := fn(c); err != nil {
            var appErr *apperrors.Error
            if errors.As(err, &appErr) {
                c.Set("app_error", appErr)
                c.AbortWithStatus(appErr.StatusCode)
            } else {
                c.Set("app_error", &apperrors.Error{
                    Code:       "internal_error",
                    Message:    err.Error(),
                    StatusCode: http.StatusInternalServerError,
                })
                c.AbortWithStatus(http.StatusInternalServerError)
            }
        }
    }
}
```

- 用户 handler 签名为 `func(c *gin.Context) error`，返回 nil 表示成功
- 如果返回 `*Error`：使用其 StatusCode 作为 HTTP 状态码
- 如果返回普通 `error`（如 `fmt.Errorf`）：自动包装为 500 内部错误
- 内部使用 `errors.As` 判断错误类型（server 包不遮蔽标准库 errors）

### 3. 预定义错误函数

```go
func BadRequest(code, message string) *Error
func NotFound(code, message string) *Error
func Unauthorized(code, message string) *Error
func Internal(code, message string) *Error
func Forbidden(code, message string) *Error
```

预定义函数返回的 `*Error` 包含默认的 StatusCode：
- BadRequest → 400
- Unauthorized → 401
- Forbidden → 403
- NotFound → 404
- Internal → 500

用户可覆盖默认值：`err := errors.NotFound("user_not_found", "用户不存在"); err.StatusCode = 404`。

### 4. ResponseMiddleware 实现

```go
func ResponseMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        if raw, exists := c.Get("app_error"); exists {
            err := raw.(*Error)
            c.JSON(err.StatusCode, gin.H{
                "error": err,
            })
        }
    }
}
```

- 在 handler 执行后检查 `app_error`
- 如果存在：用 `c.JSON()` 写出统一错误格式，HTTP status 从 `err.StatusCode` 读取
- 如果不存在：不做任何处理（成功响应已由 handler 写入）
- request_id 由 RequestID 中间件通过响应头注入，ResponseMiddleware 不负责

### 5. 默认挂载

ResponseMiddleware 加入默认中间件列表，随 Recovery 和 RequestID 一起挂载。挂载顺序：Recovery → RequestID → ResponseMiddleware。

### 6. 包名 `errors` 遮蔽标准库

`core/errors/` 包名为 `errors`，使用时 `import "github.com/fztcjjl/quix/core/errors"`。遮蔽标准库 `errors` 包，项目内部需要 `errors.Is()` 时使用别名。由于框架内部几乎不需要 `errors.Is()`，影响可接受。

## Risks / Trade-offs

- [stdlib 遮蔽] → 项目内部使用 `errors.Is()` 时需别名 → 影响极小，可接受
- [Abort 依赖 c.Next()] → gin 的 Abort 只阻止 handler chain，不阻止 middleware chain → 符合需求，ResponseMiddleware 在 handler 之后检查
- [Handler 包装器增加一层调用] → 性能影响可忽略，换来更简洁的 handler 编写体验
