package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fztcjjl/quix/core/errors"
	"github.com/fztcjjl/quix/core/log"
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
	s.ServeHTTP(w, req)

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
	s.ServeHTTP(w, req)

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
	s.ServeHTTP(w, req)

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
		s.ServeHTTP(w, req)
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
	s.ServeHTTP(w, req)

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
	s.ServeHTTP(w, req)

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
	s.ServeHTTP(w, req)

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
	s.ServeHTTP(w, req)

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

func TestServerBodyLogOption(t *testing.T) {
	cap := &testCaptureLogger{}
	log.SetDefault(cap)

	s := NewServer(WithBodyLog(1024))
	s.POST("/echo", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	body := `{"user":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	fields := cap.toMap()
	if fields["request_body"] != body {
		t.Errorf("request_body = %v, want %v", fields["request_body"], body)
	}
}

func TestServerBodyLogDisabledByDefault(t *testing.T) {
	cap := &testCaptureLogger{}
	log.SetDefault(cap)

	s := NewServer()
	s.POST("/echo", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	body := `{"user":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	fields := cap.toMap()
	if _, ok := fields["request_body"]; ok {
		t.Error("request_body should not be present when WithBodyLog is not configured")
	}
}

// testCaptureLogger captures log records for testing.
type testCaptureLogger struct {
	records []testCaptureRecord
	mu      sync.Mutex
}

type testCaptureRecord struct {
	level slog.Level
	msg   string
	args  []any
}

func (l *testCaptureLogger) Trace(_ context.Context, msg string, args ...any) {
	l.add(slog.Level(-8), msg, args)
}
func (l *testCaptureLogger) Info(_ context.Context, msg string, args ...any) {
	l.add(slog.LevelInfo, msg, args)
}
func (l *testCaptureLogger) Error(_ context.Context, msg string, args ...any) {
	l.add(slog.LevelError, msg, args)
}
func (l *testCaptureLogger) Warn(_ context.Context, msg string, args ...any) {
	l.add(slog.LevelWarn, msg, args)
}
func (l *testCaptureLogger) Debug(_ context.Context, msg string, args ...any) {
	l.add(slog.LevelDebug, msg, args)
}
func (l *testCaptureLogger) Fatal(_ context.Context, _ string, _ ...any) {}
func (l *testCaptureLogger) With(_ ...any) log.Logger                    { return l }
func (l *testCaptureLogger) SetLevel(_ log.Level)                        {}
func (l *testCaptureLogger) Close() error                                { return nil }

func (l *testCaptureLogger) add(level slog.Level, msg string, args []any) {
	l.mu.Lock()
	l.records = append(l.records, testCaptureRecord{level, msg, args})
	l.mu.Unlock()
}

func (l *testCaptureLogger) toMap() map[string]any {
	m := make(map[string]any)
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, r := range l.records {
		for i := 0; i+1 < len(r.args); i += 2 {
			key, _ := r.args[i].(string)
			m[key] = r.args[i+1]
		}
	}
	return m
}
