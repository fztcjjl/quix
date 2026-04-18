package server

import (
	"errors"
	"net/http"

	protovalidate "buf.build/go/protovalidate"
	apperrors "github.com/fztcjjl/quix/core/errors"
	"google.golang.org/protobuf/proto"
)

// FieldViolation represents a single field validation failure.
type FieldViolation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateRequest checks if req is a proto message and validates it using protovalidate.
// Returns nil if req is not a proto.Message (no validation rules).
// Translates validation errors to *apperrors.Error with HTTP 400.
func ValidateRequest(req any) error {
	msg, ok := req.(proto.Message)
	if !ok {
		return nil
	}
	if err := protovalidate.Validate(msg); err != nil {
		return toValidationError(err)
	}
	return nil
}

func toValidationError(err error) *apperrors.Error {
	var valErr *protovalidate.ValidationError
	if !errors.As(err, &valErr) {
		return &apperrors.Error{
			Code:       "validation_error",
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	violations := make([]FieldViolation, 0, len(valErr.Violations))
	for _, v := range valErr.Violations {
		violations = append(violations, FieldViolation{
			Field:   protovalidate.FieldPathString(v.Proto.GetField()),
			Message: v.Proto.GetMessage(),
		})
	}

	if len(violations) == 0 {
		return &apperrors.Error{
			Code:       "validation_error",
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return &apperrors.Error{
		Code:       "validation_error",
		Message:    "请求参数验证失败",
		Details:    violations,
		StatusCode: http.StatusBadRequest,
	}
}
