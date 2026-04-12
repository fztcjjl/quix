package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fztcjjl/quix/core/log"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockLogger struct{}

func (m *mockLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Error(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Warn(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Debug(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) With(_ ...any) log.Logger                    { return m }

func TestRecoveryNoPanic(t *testing.T) {
	r := gin.New()
	r.Use(Recovery())
	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRecoveryCatchesPanic(t *testing.T) {
	log.SetDefault(&mockLogger{})

	r := gin.New()
	r.Use(Recovery())
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
