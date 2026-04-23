package middleware

import (
	"net/http"

	qerrors "github.com/fztcjjl/quix/core/errors"
	"github.com/gin-gonic/gin"
)

// responseConfig holds configuration for the ResponseMiddleware.
type responseConfig struct {
	hideInternalErrors bool
}

// ResponseOption configures the ResponseMiddleware.
type ResponseOption func(*responseConfig)

// WithHideInternalErrors controls whether ResponseMiddleware hides raw error messages
// from non-qerrors.Error in HTTP responses. When true (production mode),
// raw error messages are replaced with a generic status text.
func WithHideInternalErrors(v bool) ResponseOption {
	return func(cfg *responseConfig) {
		cfg.hideInternalErrors = v
	}
}

// ResponseMiddleware checks for app_error in the gin context after handler execution.
// It formats the error as {"error": {"code": ..., "message": ..., "details": ...}}.
// - *qerrors.Error: uses its Code, Message, StatusCode, Details directly.
// - Other error: wraps as {Code: "internal_error", StatusCode: 500, Message: err.Error()}.
// - WithHideInternalErrors: when true, replaces the raw message with status text.
// If no error exists, it does nothing (success responses are handled by handlers).
func ResponseMiddleware(opts ...ResponseOption) gin.HandlerFunc {
	var cfg responseConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(c *gin.Context) {
		c.Next()
		raw, exists := c.Get("app_error")
		if !exists {
			return
		}

		appErr, ok := qerrors.ResolveAppError(raw)
		if !ok {
			return
		}

		if cfg.hideInternalErrors && appErr.Code == "internal_error" {
			appErr.Message = http.StatusText(http.StatusInternalServerError)
		}
		c.JSON(appErr.StatusCode, gin.H{"error": appErr})
	}
}
