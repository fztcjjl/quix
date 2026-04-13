package server

import (
	"github.com/gin-gonic/gin"
)

// Handler wraps a handler function that returns error into a gin.HandlerFunc.
// If the handler returns an *errors.Error, it uses its StatusCode.
// If it returns any other error, it wraps it as a 500 internal error.
// If it returns nil, the request continues normally.
func Handler(fn func(c *gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := fn(c); err != nil {
			SetAppError(c, err)
		}
	}
}
