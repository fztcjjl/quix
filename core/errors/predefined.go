package errors

import "net/http"

// BadRequest creates an Error with HTTP 400 status code.
func BadRequest(code, message string) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusBadRequest}
}

// Unauthorized creates an Error with HTTP 401 status code.
func Unauthorized(code, message string) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusUnauthorized}
}

// Forbidden creates an Error with HTTP 403 status code.
func Forbidden(code, message string) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusForbidden}
}

// NotFound creates an Error with HTTP 404 status code.
func NotFound(code, message string) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusNotFound}
}

// Internal creates an Error with HTTP 500 status code.
func Internal(code, message string) *Error {
	return &Error{Code: code, Message: message, StatusCode: http.StatusInternalServerError}
}
