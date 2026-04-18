package server

import (
	"errors"
	"net/http"
	"testing"

	validate "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	protovalidate "buf.build/go/protovalidate"
	apperrors "github.com/fztcjjl/quix/core/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func makeFieldPath(fieldName string) *validate.FieldPath {
	return &validate.FieldPath{
		Elements: []*validate.FieldPathElement{
			{FieldName: proto.String(fieldName)},
		},
	}
}

func makeViolation(fieldName, message string) *protovalidate.Violation {
	return &protovalidate.Violation{
		Proto: &validate.Violation{
			Field:   makeFieldPath(fieldName),
			Message: proto.String(message),
		},
	}
}

func TestValidateRequest_Nil(t *testing.T) {
	err := ValidateRequest(nil)
	assert.Nil(t, err)
}

func TestValidateRequest_NonProtoMessage(t *testing.T) {
	type plain struct {
		Name string
	}
	err := ValidateRequest(&plain{Name: "alice"})
	assert.Nil(t, err)
}

func TestValidateRequest_ProtoMessage_NoRules(t *testing.T) {
	// emptypb.Empty has no validation rules, should pass
	err := ValidateRequest(&emptypb.Empty{})
	assert.Nil(t, err)
}

func TestToValidationError_SingleViolation(t *testing.T) {
	valErr := &protovalidate.ValidationError{
		Violations: []*protovalidate.Violation{
			makeViolation("title", "value must be at least 1 character(s) long"),
		},
	}

	err := toValidationError(valErr)
	assert.NotNil(t, err)

	var appErr *apperrors.Error
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "validation_error", appErr.Code)
	assert.Equal(t, http.StatusBadRequest, appErr.StatusCode)
	assert.Equal(t, "请求参数验证失败", appErr.Message)

	details, ok := appErr.Details.([]FieldViolation)
	assert.True(t, ok)
	assert.Len(t, details, 1)
	assert.Equal(t, "title", details[0].Field)
	assert.Equal(t, "value must be at least 1 character(s) long", details[0].Message)
}

func TestToValidationError_MultipleViolations(t *testing.T) {
	valErr := &protovalidate.ValidationError{
		Violations: []*protovalidate.Violation{
			makeViolation("title", "value must be at least 1 character(s) long"),
			makeViolation("email", "value must be a valid email address"),
		},
	}

	err := toValidationError(valErr)

	var appErr *apperrors.Error
	assert.True(t, errors.As(err, &appErr))

	details, ok := appErr.Details.([]FieldViolation)
	assert.True(t, ok)
	assert.Len(t, details, 2)
	assert.Equal(t, "title", details[0].Field)
	assert.Equal(t, "email", details[1].Field)
}

func TestToValidationError_NonValidationError(t *testing.T) {
	err := toValidationError(assert.AnError)

	var appErr *apperrors.Error
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "validation_error", appErr.Code)
	assert.Equal(t, http.StatusBadRequest, appErr.StatusCode)
	assert.Equal(t, assert.AnError.Error(), appErr.Message)
	assert.Nil(t, appErr.Details)
}

func TestFieldPathString(t *testing.T) {
	path := makeFieldPath("user_id")
	result := protovalidate.FieldPathString(path)
	assert.Equal(t, "user_id", result)
}
