package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/fztcjjl/quix/core/log"
	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics and logs the stack trace.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				reqID, _ := c.Get("X-Request-Id")
				log.Error(c.Request.Context(), "panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"request_id", reqID,
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"code": "internal_error", "message": "Internal Server Error"},
				})
			}
		}()
		c.Next()
	}
}
