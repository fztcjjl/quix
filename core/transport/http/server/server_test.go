package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fztcjjl/quix/core/errors"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewServer(t *testing.T) {
	s := NewServer(WithAddr(":8080"))
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.Engine == nil {
		t.Fatal("Engine() returned nil")
	}
	if s.Addr() != ":8080" {
		t.Errorf("expected :8080, got %s", s.Addr())
	}
}

func TestServerImplementsTransportServer(t *testing.T) {
	var s interface{} = NewServer()
	if _, ok := s.(interface {
		Start() error
		Stop(context.Context) error
	}); !ok {
		t.Fatal("Server does not implement transport.Server interface")
	}
}

func TestServerStartAndStop(t *testing.T) {
	s := NewServer(WithAddr("127.0.0.1:0"))

	go func() {
		time.Sleep(10 * time.Millisecond)
		_ = s.Stop(context.TODO())
	}()

	if err := s.Start(); err != http.ErrServerClosed {
		t.Fatalf("expected ErrServerClosed, got %v", err)
	}
}

func TestServerRouting(t *testing.T) {
	s := NewServer()
	s.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	s.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestServerMiddleware(t *testing.T) {
	s := NewServer()
	var called bool
	s.Use(func(c *gin.Context) {
		called = true
		c.Next()
	})
	s.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	s.Engine.ServeHTTP(w, req)

	if !called {
		t.Error("middleware was not called")
	}
}

func TestServerGroup(t *testing.T) {
	s := NewServer()
	api := s.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	s.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestServerAllMethods(t *testing.T) {
	s := NewServer()

	s.GET("/r", func(c *gin.Context) { c.Status(http.StatusOK) })
	s.POST("/r", func(c *gin.Context) { c.Status(http.StatusOK) })
	s.PUT("/r", func(c *gin.Context) { c.Status(http.StatusOK) })
	s.DELETE("/r", func(c *gin.Context) { c.Status(http.StatusOK) })
	s.PATCH("/r", func(c *gin.Context) { c.Status(http.StatusOK) })

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, m := range methods {
		req, _ := http.NewRequest(m, "/r", nil)
		w := httptest.NewRecorder()
		s.Engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("method %s: expected 200, got %d", m, w.Code)
		}
	}
}

func TestServerDefaultMiddleware(t *testing.T) {
	s := NewServer()

	// Default middleware should recover from panic
	s.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	s.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 (recovered), got %d", w.Code)
	}
}

func TestServerDisableDefaultMiddleware(t *testing.T) {
	s := NewServer(WithDefaultMiddleware(false))

	s.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	s.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestServerDefaultMiddlewareErrorFormat(t *testing.T) {
	s := NewServer()

	s.GET("/notfound", Handler(func(c *gin.Context) error {
		return errors.NotFound("user_not_found", "用户不存在")
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notfound", nil)
	s.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
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
}

func TestServerDefaultMiddlewareStandardError(t *testing.T) {
	s := NewServer()

	s.GET("/internal", Handler(func(c *gin.Context) error {
		return fmt.Errorf("unexpected error")
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/internal", nil)
	s.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
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
}
