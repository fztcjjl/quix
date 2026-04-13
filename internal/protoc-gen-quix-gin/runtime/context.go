package runtime

import (
	"errors"
	"net/http"

	apperrors "github.com/fztcjjl/quix/core/errors"
	"github.com/gin-gonic/gin"
)

// Context wraps *gin.Context with additional methods for error handling
// and proto-aware request binding.
type Context struct {
	*gin.Context
}

// SetError stores an error in the gin context and aborts the request.
// If err is *apperrors.Error, its StatusCode is used.
// Otherwise, the error is wrapped as an internal error with status 500.
func (c *Context) SetError(err error) {
	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		c.Context.Set("app_error", appErr)
		c.Context.AbortWithStatus(appErr.StatusCode)
	} else {
		wrapped := &apperrors.Error{
			Code:       "internal_error",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
		c.Context.Set("app_error", wrapped)
		c.Context.AbortWithStatus(http.StatusInternalServerError)
	}
}

// GetError retrieves the stored error from the gin context.
// Returns nil if no error was stored.
func (c *Context) GetError() *apperrors.Error {
	if raw, exists := c.Get("app_error"); exists {
		if appErr, ok := raw.(*apperrors.Error); ok {
			return appErr
		}
	}
	return nil
}
