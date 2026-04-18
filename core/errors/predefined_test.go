package errors_test

import (
	"net/http"
	"testing"

	qerrors "github.com/fztcjjl/quix/core/errors"
)

func TestPredefinedErrorsDefaultStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(string, string) *qerrors.Error
		wantStatus int
	}{
		{"BadRequest", qerrors.BadRequest, http.StatusBadRequest},
		{"Unauthorized", qerrors.Unauthorized, http.StatusUnauthorized},
		{"Forbidden", qerrors.Forbidden, http.StatusForbidden},
		{"NotFound", qerrors.NotFound, http.StatusNotFound},
		{"Internal", qerrors.Internal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn("test_code", "test message")
			if err.StatusCode != tt.wantStatus {
				t.Errorf("%s() StatusCode = %d, want %d", tt.name, err.StatusCode, tt.wantStatus)
			}
			if err.Code != "test_code" {
				t.Errorf("%s() Code = %q, want %q", tt.name, err.Code, "test_code")
			}
			if err.Message != "test message" {
				t.Errorf("%s() Message = %q, want %q", tt.name, err.Message, "test message")
			}
		})
	}
}

func TestPredefinedErrorsOverrideStatusCode(t *testing.T) {
	err := qerrors.NotFound("user_not_found", "用户不存在")
	err.StatusCode = http.StatusGone

	if err.StatusCode != http.StatusGone {
		t.Errorf("StatusCode after override = %d, want %d", err.StatusCode, http.StatusGone)
	}
}
