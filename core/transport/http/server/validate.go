package server

import (
	"errors"
	"net/http"

	apperrors "github.com/fztcjjl/quix/core/errors"
)

// Validator is implemented by proto messages that have validation rules.
// protoc-gen-validate generates Validate() methods with this signature.
type Validator interface {
	Validate() error
}

// FieldViolation represents a single field validation failure.
type FieldViolation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// fieldViolation is the interface used to extract field-level details
// from protoc-gen-validate errors without importing the validate package.
type fieldViolation interface {
	Field() string
	Reason() string
}

// multiError is the interface used to unwrap multiple validation errors
// from protoc-gen-validate without importing the validate package.
type multiError interface {
	Unwrap() []error
}

// ValidateRequest checks if req implements Validator and calls Validate().
// Returns nil if req does not implement Validator (no validation rules).
// Translates validation errors to *apperrors.Error with HTTP 400.
func ValidateRequest(req any) error {
	v, ok := req.(Validator)
	if !ok {
		return nil
	}
	if err := v.Validate(); err != nil {
		return toValidationError(err)
	}
	return nil
}

func toValidationError(err error) *apperrors.Error {
	var violations []FieldViolation

	// Handle multi-error (multiple field violations)
	var me multiError
	if errors.As(err, &me) {
		for _, e := range me.Unwrap() {
			if v := extractViolation(e); v != nil {
				violations = append(violations, *v)
			}
		}
	} else if v := extractViolation(err); v != nil {
		violations = append(violations, *v)
	}

	if len(violations) == 0 {
		violations = []FieldViolation{{Message: err.Error()}}
	}

	return &apperrors.Error{
		Code:       "validation_error",
		Message:    "请求参数验证失败",
		Details:    violations,
		StatusCode: http.StatusBadRequest,
	}
}

func extractViolation(err error) *FieldViolation {
	var fv fieldViolation
	if errors.As(err, &fv) {
		return &FieldViolation{Field: fv.Field(), Message: fv.Reason()}
	}
	return nil
}
