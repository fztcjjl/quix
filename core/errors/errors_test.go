package errors_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	qerrors "github.com/fztcjjl/quix/core/errors"
)

func TestErrorImplementsErrorInterface(t *testing.T) {
	e := &qerrors.Error{Code: "test", Message: "test message"}
	var _ error = e

	if e.Error() != "test message" {
		t.Errorf("Error() = %q, want %q", e.Error(), "test message")
	}
}

func TestErrorJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		err      *qerrors.Error
		wantJSON string
	}{
		{
			name:     "full error with details",
			err:      &qerrors.Error{Code: "param_invalid", Message: "参数验证失败", Details: "field: name", StatusCode: 400},
			wantJSON: `{"code":"param_invalid","message":"参数验证失败","details":"field: name"}`,
		},
		{
			name:     "error without details",
			err:      &qerrors.Error{Code: "not_found", Message: "不存在", StatusCode: 404},
			wantJSON: `{"code":"not_found","message":"不存在"}`,
		},
		{
			name:     "error with map details",
			err:      &qerrors.Error{Code: "validation", Message: "验证失败", Details: map[string]string{"field": "email", "reason": "invalid format"}, StatusCode: 400},
			wantJSON: `{"code":"validation","message":"验证失败","details":{"field":"email","reason":"invalid format"}}`,
		},
		{
			name:     "error with slice details",
			err:      &qerrors.Error{Code: "batch_error", Message: "批量操作部分失败", Details: []map[string]any{{"id": 1, "error": "not found"}, {"id": 2, "error": "forbidden"}}, StatusCode: 207},
			wantJSON: `{"code":"batch_error","message":"批量操作部分失败","details":[{"error":"not found","id":1},{"error":"forbidden","id":2}]}`,
		},
		{
			name:     "status code not in JSON",
			err:      &qerrors.Error{Code: "internal", Message: "内部错误", StatusCode: 500},
			wantJSON: `{"code":"internal","message":"内部错误"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.err)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if string(data) != tt.wantJSON {
				t.Errorf("json.Marshal() = %s, want %s", string(data), tt.wantJSON)
			}
		})
	}
}

func TestErrorDetailsIsOptional(t *testing.T) {
	e := &qerrors.Error{Code: "test", Message: "test", StatusCode: 400}
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if string(data) != `{"code":"test","message":"test"}` {
		t.Errorf("json.Marshal() with nil details = %s, want details omitted", string(data))
	}
}

func TestErrorWithCustomStructDetails(t *testing.T) {
	type FieldError struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}

	e := &qerrors.Error{
		Code:       "validation",
		Message:    "验证失败",
		Details:    []FieldError{{Field: "email", Message: "格式无效"}, {Field: "age", Message: "必须大于0"}},
		StatusCode: 400,
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	want := `{"code":"validation","message":"验证失败","details":[{"field":"email","message":"格式无效"},{"field":"age","message":"必须大于0"}]}`
	if string(data) != want {
		t.Errorf("json.Marshal() = %s, want %s", string(data), want)
	}
}

func TestErrorIsComparableWithErrorsIs(t *testing.T) {
	e := &qerrors.Error{Code: "test", Message: "test"}
	var _ error = e

	if !errors.Is(e, e) {
		t.Error("errors.Is should return true for same pointer")
	}
}

func TestResolveAppError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got, ok := qerrors.ResolveAppError(nil)
		if ok {
			t.Fatal("expected ok=false for nil")
		}
		if got != nil {
			t.Fatal("expected nil Error for nil input")
		}
	})

	t.Run("structured error", func(t *testing.T) {
		err := &qerrors.Error{Code: "not_found", Message: "gone", StatusCode: 404}
		got, ok := qerrors.ResolveAppError(err)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if got != err {
			t.Error("should return the same pointer for *Error")
		}
	})

	t.Run("native error", func(t *testing.T) {
		native := fmt.Errorf("db connection failed")
		got, ok := qerrors.ResolveAppError(native)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if got.Code != "internal_error" {
			t.Errorf("Code = %q, want %q", got.Code, "internal_error")
		}
		if got.Message != "db connection failed" {
			t.Errorf("Message = %q, want %q", got.Message, "db connection failed")
		}
		if got.StatusCode != 500 {
			t.Errorf("StatusCode = %d, want 500", got.StatusCode)
		}
	})

	t.Run("non-error value", func(t *testing.T) {
		got, ok := qerrors.ResolveAppError("just a string")
		if ok {
			t.Fatal("expected ok=false for non-error value")
		}
		if got != nil {
			t.Fatal("expected nil Error for non-error value")
		}
	})
}
