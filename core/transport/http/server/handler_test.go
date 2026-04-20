package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	qerrors "github.com/fztcjjl/quix/core/errors"
	"github.com/fztcjjl/quix/core/transport/http/server/middleware"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandlerNilReturn(t *testing.T) {
	r := gin.New()
	r.Use(middleware.ResponseMiddleware())
	called := false
	r.GET("/test", Handler(func(c *gin.Context) error {
		called = true
		c.Status(http.StatusOK)
		return nil
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
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
	r.Use(middleware.ResponseMiddleware())
	r.GET("/test", Handler(func(c *gin.Context) error {
		return qerrors.NotFound("user_not_found", "用户不存在")
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandlerReturnsAppErrorSetsContext(t *testing.T) {
	r := gin.New()
	r.Use(middleware.ResponseMiddleware())
	r.GET("/test", Handler(func(c *gin.Context) error {
		return qerrors.BadRequest("param_invalid", "参数验证失败")
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlerReturnsStandardError(t *testing.T) {
	r := gin.New()
	r.Use(middleware.ResponseMiddleware())
	r.GET("/test", Handler(func(c *gin.Context) error {
		return fmt.Errorf("db connection failed")
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
