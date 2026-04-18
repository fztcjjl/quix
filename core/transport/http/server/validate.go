package server

import (
	"errors"
	"net/http"

	protovalidate "buf.build/go/protovalidate"
	qerrors "github.com/fztcjjl/quix/core/errors"
	"google.golang.org/protobuf/proto"
)

// ValidationMessage is the message used when field validation fails with violations.
// Override this to customize or localize the message.
var ValidationMessage = "请求参数验证失败"

// FieldViolation represents a single field validation failure.
type FieldViolation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateRequest checks if req is a proto message and validates it using protovalidate.
// Returns nil if req is not a proto.Message (no validation rules).
// Translates validation errors to *qerrors.Error with HTTP 400.
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

func toValidationError(err error) *qerrors.Error {
	var valErr *protovalidate.ValidationError
	if !errors.As(err, &valErr) {
		return &qerrors.Error{
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
		return &qerrors.Error{
			Code:       "validation_error",
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return &qerrors.Error{
		Code:       "validation_error",
		Message:    ValidationMessage,
		Details:    violations,
		StatusCode: http.StatusBadRequest,
	}
}
