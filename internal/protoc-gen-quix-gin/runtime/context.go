package runtime

import (
	apperrors "github.com/fztcjjl/quix/core/errors"
	"github.com/fztcjjl/quix/core/transport/http/server"
	"github.com/gin-gonic/gin"
)

// Context wraps *gin.Context with additional methods for error handling
// and proto-aware request binding.
type Context struct {
	*gin.Context
}

// SetError stores an error in the gin context and aborts the request.
// It delegates to server.SetAppError for consistent error handling.
func (c *Context) SetError(err error) {
	server.SetAppError(c.Context, err)
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
