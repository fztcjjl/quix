package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/fztcjjl/quix/core/errors"
	"github.com/gin-gonic/gin"
)

func TestHandlerNilReturn(t *testing.T) {
	r := gin.New()
	called := false
	r.GET("/test", Handler(func(c *gin.Context) error {
		called = true
		c.Status(http.StatusOK)
		return nil
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if !called {
		t.Error("handler should be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandlerReturnsAppError(t *testing.T) {
	r := gin.New()
	r.GET("/test", Handler(func(c *gin.Context) error {
		return apperrors.NotFound("user_not_found", "用户不存在")
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandlerReturnsAppErrorSetsContext(t *testing.T) {
	r := gin.New()
	r.GET("/test", Handler(func(c *gin.Context) error {
		return apperrors.BadRequest("param_invalid", "参数验证失败")
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlerReturnsStandardError(t *testing.T) {
	r := gin.New()
	r.GET("/test", Handler(func(c *gin.Context) error {
		return fmt.Errorf("db connection failed")
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandlerAbortsSubsequentHandlers(t *testing.T) {
	r := gin.New()
	secondCalled := false
	r.GET("/test", Handler(func(c *gin.Context) error {
		return apperrors.Forbidden("access_denied", "没有权限")
	}), func(c *gin.Context) {
		secondCalled = true
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if secondCalled {
		t.Error("second handler should not be called after error")
	}
}

func TestHandlerAppErrorInContext(t *testing.T) {
	r := gin.New()
	var capturedErr *apperrors.Error
	r.Use(func(c *gin.Context) {
		c.Next()
		if raw, ok := c.Get("app_error"); ok {
			capturedErr = raw.(*apperrors.Error)
		}
	})
	r.GET("/test", Handler(func(c *gin.Context) error {
		return apperrors.Unauthorized("token_expired", "令牌已过期")
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if capturedErr == nil {
		t.Fatal("app_error should be set in context")
	}
	if capturedErr.Code != "token_expired" {
		t.Errorf("app_error.Code = %q, want %q", capturedErr.Code, "token_expired")
	}
	if capturedErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("app_error.StatusCode = %d, want %d", capturedErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestSetAppErrorHideInternalErrors(t *testing.T) {
	tests := []struct {
		name        string
		hide        bool
		wantMessage string
		wantCode    string
	}{
		{
			name:        "dev mode exposes raw error",
			hide:        false,
			wantMessage: "db connection failed: host=db.example.com",
			wantCode:    "internal_error",
		},
		{
			name:        "prod mode hides raw error",
			hide:        true,
			wantMessage: "Internal Server Error",
			wantCode:    "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HideInternalErrors = tt.hide

			var capturedErr *apperrors.Error
			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Next()
				if raw, ok := c.Get("app_error"); ok {
					capturedErr = raw.(*apperrors.Error)
				}
			})
			r.GET("/test", Handler(func(c *gin.Context) error {
				return fmt.Errorf("db connection failed: host=db.example.com")
			}))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
			}
			if capturedErr == nil {
				t.Fatal("app_error should be set")
			}
			if capturedErr.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", capturedErr.Message, tt.wantMessage)
			}
			if capturedErr.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", capturedErr.Code, tt.wantCode)
			}
		})
	}

	HideInternalErrors = false
}
