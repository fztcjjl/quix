package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/fztcjjl/quix/core/log"
	"github.com/gin-gonic/gin"
)

// HideStackTraces controls whether Recovery omits full stack traces from logs.
// When true (production mode), only the panic value is logged.
// When false (default), the full stack trace is included.
var HideStackTraces bool

// Recovery returns a middleware that recovers from panics and logs the stack trace.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				reqID, _ := c.Get("X-Request-Id")
				fields := []any{
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"request_id", reqID,
				}
				if !HideStackTraces {
					fields = append(fields, "stack", string(debug.Stack()))
				}
				log.Error(c.Request.Context(), "panic recovered", fields...)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"code": "internal_error", "message": "Internal Server Error"},
				})
			}
		}()
		c.Next()
	}
}
