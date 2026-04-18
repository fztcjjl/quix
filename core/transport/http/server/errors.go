package server

import (
	"errors"
	"net/http"

	apperrors "github.com/fztcjjl/quix/core/errors"
	"github.com/gin-gonic/gin"
)

// HideInternalErrors controls whether SetAppError hides raw error messages
// from non-apperrors.Error in HTTP responses. When true (production mode),
// such errors are logged but replaced with a generic message in the response.
// When false (default), raw error messages are included.
var HideInternalErrors bool

// SetAppError stores an error in the gin context and aborts the request.
// If err is *apperrors.Error, its StatusCode is used.
// Otherwise, the error is wrapped as an internal error with status 500.
func SetAppError(c *gin.Context, err error) {
	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		c.Set("app_error", appErr)
		c.AbortWithStatus(appErr.StatusCode)
	} else {
		wrapped := &apperrors.Error{
			Code:       "internal_error",
			StatusCode: http.StatusInternalServerError,
		}
		if HideInternalErrors {
			wrapped.Message = http.StatusText(http.StatusInternalServerError)
		} else {
			wrapped.Message = err.Error()
		}
		c.Set("app_error", wrapped)
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
