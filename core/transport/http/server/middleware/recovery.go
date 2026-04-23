package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/fztcjjl/quix/core/log"
	"github.com/gin-gonic/gin"
)

// recoveryConfig holds configuration for the Recovery middleware.
type recoveryConfig struct {
	hideStackTraces bool
}

// RecoveryOption configures the Recovery middleware.
type RecoveryOption func(*recoveryConfig)

// WithHideStackTraces controls whether Recovery omits full stack traces from logs.
// When true (production mode), only the panic value is logged.
// When false (default), the full stack trace is included.
func WithHideStackTraces(v bool) RecoveryOption {
	return func(cfg *recoveryConfig) {
		cfg.hideStackTraces = v
	}
}

// Recovery returns a middleware that recovers from panics and logs the stack trace.
func Recovery(opts ...RecoveryOption) gin.HandlerFunc {
	var cfg recoveryConfig
	for _, opt := range opts {
		opt(&cfg)
	}

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
				if !cfg.hideStackTraces {
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
