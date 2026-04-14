package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fztcjjl/quix/core/log"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

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
	// Use a captureLogger to absorb panic output during test
	log.SetDefault(&captureLogger{})

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
