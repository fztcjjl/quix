package errors

import "net/http"

// BadRequest creates an Error with HTTP 400 status code.
func BadRequest(code, message string) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusBadRequest}
}

// BadRequestWrap creates an Error with HTTP 400 status code, wrapping an underlying cause.
func BadRequestWrap(code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusBadRequest, cause: cause}
}

// Unauthorized creates an Error with HTTP 401 status code.
func Unauthorized(code, message string) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusUnauthorized}
}

// UnauthorizedWrap creates an Error with HTTP 401 status code, wrapping an underlying cause.
func UnauthorizedWrap(code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusUnauthorized, cause: cause}
}

// Forbidden creates an Error with HTTP 403 status code.
func Forbidden(code, message string) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusForbidden}
}

// ForbiddenWrap creates an Error with HTTP 403 status code, wrapping an underlying cause.
func ForbiddenWrap(code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusForbidden, cause: cause}
}

// NotFound creates an Error with HTTP 404 status code.
func NotFound(code, message string) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusNotFound}
}

// NotFoundWrap creates an Error with HTTP 404 status code, wrapping an underlying cause.
func NotFoundWrap(code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusNotFound, cause: cause}
}

// Internal creates an Error with HTTP 500 status code.
func Internal(code, message string) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusInternalServerError}
}

// InternalWrap creates an Error with HTTP 500 status code, wrapping an underlying cause.
func InternalWrap(code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusInternalServerError, cause: cause}
}
