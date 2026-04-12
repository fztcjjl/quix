package middleware

import (
	apperrors "github.com/fztcjjl/quix/core/errors"
	"github.com/gin-gonic/gin"
)

// ResponseMiddleware checks for app_error in the gin context after handler execution.
// If an error exists, it formats the response as {"error": {...}}.
// If no error exists, it does nothing (success responses are handled by handlers).
func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if raw, exists := c.Get("app_error"); exists {
			err := raw.(*apperrors.Error)
			c.JSON(err.StatusCode, gin.H{
				"error": err,
			})
		}
	}
}
