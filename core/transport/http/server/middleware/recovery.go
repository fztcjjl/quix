package middleware

import (
	"context"
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
				log.Error(context.Background(), "panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
