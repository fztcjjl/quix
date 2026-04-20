package middleware

import (
	"net/http"

	qerrors "github.com/fztcjjl/quix/core/errors"
	"github.com/gin-gonic/gin"
)

// HideInternalErrors controls whether ResponseMiddleware hides raw error messages
// from non-qerrors.Error in HTTP responses. When true (production mode),
// raw error messages are replaced with a generic status text.
var HideInternalErrors bool

// ResponseMiddleware checks for app_error in the gin context after handler execution.
// It formats the error as {"error": {"code": ..., "message": ..., "details": ...}}.
// - *qerrors.Error: uses its Code, Message, StatusCode, Details directly.
// - Other error: wraps as {Code: "internal_error", StatusCode: 500, Message: err.Error()}.
// - HideInternalErrors: when true, replaces the raw message with status text.
// If no error exists, it does nothing (success responses are handled by handlers).
func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		raw, exists := c.Get("app_error")
		if !exists {
			return
		}

		var appErr *qerrors.Error
		switch v := raw.(type) {
		case *qerrors.Error:
			appErr = v
		case error:
			msg := v.Error()
			if HideInternalErrors {
				msg = http.StatusText(http.StatusInternalServerError)
			}
			appErr = &qerrors.Error{
				Code:       "internal_error",
				StatusCode: http.StatusInternalServerError,
				Message:    msg,
			}
		}
		c.JSON(appErr.StatusCode, gin.H{"error": appErr})
	}
}
