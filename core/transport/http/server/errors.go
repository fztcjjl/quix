package server

import (
	"errors"
	"net/http"

	apperrors "github.com/fztcjjl/quix/core/errors"
	"github.com/gin-gonic/gin"
)

// SetAppError stores an error in the gin context and aborts the request.
// If err is *apperrors.Error, its StatusCode is used.
// Otherwise, the error is wrapped as an internal error with status 500.
func SetAppError(c *gin.Context, err error) {
	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		c.Set("app_error", appErr)
		c.AbortWithStatus(appErr.StatusCode)
	} else {
		c.Set("app_error", &apperrors.Error{
			Code:       "internal_error",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
