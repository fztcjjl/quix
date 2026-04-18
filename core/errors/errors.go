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
