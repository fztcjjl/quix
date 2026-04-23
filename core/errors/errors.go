package errors

// Error represents a structured application error with code, message, optional details, and HTTP status code.
type Error struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
	StatusCode int    `json:"-"`
	cause      error  `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap returns the underlying cause of this error, supporting error chain traversal
// with errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.cause
}

// ResolveAppError converts a raw value (typically from gin.Context) into *Error.
// It is the single source of truth for app_error normalization used by both
// ResponseMiddleware and AccessLog middleware.
//
//   - *Error: returned as-is
//   - error: wrapped as {Code: "internal_error", StatusCode: 500, Message: err.Error()}
//   - nil or non-error: returns (nil, false)
func ResolveAppError(raw any) (*Error, bool) {
	if raw == nil {
		return nil, false
	}
	switch v := raw.(type) {
	case *Error:
		return v, true
	case error:
		return &Error{
			Code:       "internal_error",
			StatusCode: 500,
			Message:    v.Error(),
		}, true
	default:
		return nil, false
	}
}
