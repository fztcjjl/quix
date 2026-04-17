package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "github.com/fztcjjl/quix/core/errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockValidator implements Validator with a configurable error.
type mockValidator struct {
	err error
}

func (m *mockValidator) Validate() error {
	return m.err
}

// mockMultiError implements multiError for testing multiple violations.
type mockMultiError struct {
	errs []error
}

func (m *mockMultiError) Error() string {
	return "multiple errors"
}

func (m *mockMultiError) Unwrap() []error {
	return m.errs
}

// mockFieldViolation implements fieldViolation for testing.
type mockFieldViolation struct {
	field  string
	reason string
}

func (m *mockFieldViolation) Error() string {
	return m.field + ": " + m.reason
}

func (m *mockFieldViolation) Field() string {
	return m.field
}

func (m *mockFieldViolation) Reason() string {
	return m.reason
}

// mockNonValidator does NOT implement Validator.
type mockNonValidator struct {
	Name string `json:"name"`
}

func TestValidateRequest_ImplementsValidator_Error(t *testing.T) {
	req := &mockValidator{
		err: &mockFieldViolation{field: "name", reason: "must not be empty"},
	}

	err := ValidateRequest(req)
	assert.NotNil(t, err)

	var appErr *apperrors.Error
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "validation_error", appErr.Code)
	assert.Equal(t, http.StatusBadRequest, appErr.StatusCode)
	assert.Equal(t, "请求参数验证失败", appErr.Message)

	details, ok := appErr.Details.([]FieldViolation)
	assert.True(t, ok)
	assert.Len(t, details, 1)
	assert.Equal(t, "name", details[0].Field)
	assert.Equal(t, "must not be empty", details[0].Message)
}

func TestValidateRequest_ImplementsValidator_MultiError(t *testing.T) {
	req := &mockValidator{
		err: &mockMultiError{
			errs: []error{
				&mockFieldViolation{field: "name", reason: "must not be empty"},
				&mockFieldViolation{field: "email", reason: "invalid format"},
			},
		},
	}

	err := ValidateRequest(req)
	assert.NotNil(t, err)

	var appErr *apperrors.Error
	assert.True(t, errors.As(err, &appErr))

	details, ok := appErr.Details.([]FieldViolation)
	assert.True(t, ok)
	assert.Len(t, details, 2)
	assert.Equal(t, "name", details[0].Field)
	assert.Equal(t, "email", details[1].Field)
}

func TestValidateRequest_ImplementsValidator_NoError(t *testing.T) {
	req := &mockValidator{err: nil}

	err := ValidateRequest(req)
	assert.Nil(t, err)
}

func TestValidateRequest_NotImplementValidator(t *testing.T) {
	req := &mockNonValidator{Name: "alice"}

	err := ValidateRequest(req)
	assert.Nil(t, err)
}

func TestValidateRequest_Nil(t *testing.T) {
	err := ValidateRequest(nil)
	assert.Nil(t, err)
}

func TestValidateRequest_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		ctx := &Context{Context: c}
		var req mockNonValidator

		// Simulate generated handler flow: bind -> validate -> service
		if err := c.ShouldBindJSON(&req); err != nil {
			ctx.SetError(err)
			return
		}
		if err := ValidateRequest(&req); err != nil {
			ctx.SetError(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": req.Name})
	})

	// mockNonValidator doesn't implement Validator, so ValidateRequest is no-op
	w := httptest.NewRecorder()
	body, _ := json.Marshal(mockNonValidator{Name: "alice"})
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateRequest_IntegrationWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		ctx := &Context{Context: c}
		req := &mockValidator{err: &mockFieldViolation{field: "name", reason: "required"}}

		if err := ValidateRequest(req); err != nil {
			ctx.SetError(err)
			return
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var appErr *apperrors.Error
	assert.True(t, errors.As(ValidateRequest(&mockValidator{err: &mockFieldViolation{field: "name", reason: "required"}}), &appErr))
	assert.Equal(t, "validation_error", appErr.Code)
}
