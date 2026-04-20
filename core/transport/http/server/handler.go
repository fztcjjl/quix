package server

import (
	"github.com/gin-gonic/gin"
)

// Handler wraps a handler function that returns error into a gin.HandlerFunc.
// If the handler returns an error, it stores the error in the gin context.
// ResponseMiddleware will detect the error and format the response.
func Handler(fn func(ctx *gin.Context) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c := &Context{Context: ctx}
		if err := fn(ctx); err != nil {
			c.SetError(err)
		}
	}
}
