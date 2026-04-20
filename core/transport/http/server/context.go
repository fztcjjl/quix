package server

import (
	"github.com/gin-gonic/gin"
)

// Context wraps *gin.Context with additional methods for error handling
// and proto-aware request binding.
type Context struct {
	*gin.Context
}

// SetError stores an error in the gin context.
// The error will be picked up by ResponseMiddleware to format the response.
func (c *Context) SetError(err error) {
	c.Set("app_error", err)
}

// GetError retrieves the stored error from the gin context.
// Returns nil if no error was stored.
func (c *Context) GetError() error {
	if val, exists := c.Get("app_error"); exists {
		if err, ok := val.(error); ok {
			return err
		}
	}
	return nil
}
