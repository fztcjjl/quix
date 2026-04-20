package middleware_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	qerrors "github.com/fztcjjl/quix/core/errors"
	"github.com/fztcjjl/quix/core/transport/http/server"
	"github.com/fztcjjl/quix/core/transport/http/server/middleware"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestResponseMiddlewareFormatsError(t *testing.T) {
	r := gin.New()
	r.Use(middleware.ResponseMiddleware())
	r.GET("/test", server.Handler(func(c *gin.Context) error {
		return qerrors.NotFound("user_not_found", "用户不存在")
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("response should contain 'error' object")
	}
	if errObj["code"] != "user_not_found" {
		t.Errorf("error.code = %v, want %q", errObj["code"], "user_not_found")
	}
	if errObj["message"] != "用户不存在" {
		t.Errorf("error.message = %v, want %q", errObj["message"], "用户不存在")
	}
	if _, exists := errObj["StatusCode"]; exists {
		t.Error("StatusCode should not be in response body")
	}
}

func TestResponseMiddlewareSkipsOnSuccess(t *testing.T) {
	r := gin.New()
	r.Use(middleware.ResponseMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if body["data"] != "ok" {
		t.Errorf("response data = %v, want %q", body["data"], "ok")
	}
	if _, exists := body["error"]; exists {
		t.Error("success response should not contain 'error'")
	}
}

func TestResponseMiddlewareStandardError(t *testing.T) {
	r := gin.New()
	r.Use(middleware.ResponseMiddleware())
	r.GET("/test", server.Handler(func(c *gin.Context) error {
		return fmt.Errorf("database connection failed")
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("response should contain 'error' object")
	}
	if errObj["code"] != "internal_error" {
		t.Errorf("error.code = %v, want %q", errObj["code"], "internal_error")
	}
	if errObj["message"] != "database connection failed" {
		t.Errorf("error.message = %v, want %q", errObj["message"], "database connection failed")
	}
}

func TestResponseMiddlewareErrorWithDetails(t *testing.T) {
	r := gin.New()
	r.Use(middleware.ResponseMiddleware())
	r.GET("/test", server.Handler(func(c *gin.Context) error {
		return &qerrors.Error{
			Code:       "param_invalid",
			Message:    "参数验证失败",
			Details:    map[string]string{"field": "email", "reason": "格式无效"},
			StatusCode: http.StatusBadRequest,
		}
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	errObj := body["error"].(map[string]any)
	details, ok := errObj["details"].(map[string]any)
	if !ok {
		t.Fatal("error.details should be present")
	}
	if details["field"] != "email" {
		t.Errorf("details.field = %v, want %q", details["field"], "email")
	}
}

func TestResponseMiddlewareHideInternalErrors(t *testing.T) {
	// Verify that HideInternalErrors replaces raw error messages with a generic status text.
	middleware.HideInternalErrors = true
	defer func() { middleware.HideInternalErrors = false }()

	r := gin.New()
	r.Use(middleware.ResponseMiddleware())
	r.GET("/test", server.Handler(func(c *gin.Context) error {
		return fmt.Errorf("database connection failed")
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	errObj := body["error"].(map[string]any)
	if errObj["message"] != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("error.message = %v, want %q (raw message should be hidden)", errObj["message"], http.StatusText(http.StatusInternalServerError))
	}
}
