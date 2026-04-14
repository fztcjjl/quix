package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/fztcjjl/quix/core/log"
	"github.com/gin-gonic/gin"
)

// captureLogger captures log records for testing.
type captureLogger struct {
	records []captureRecord
	mu      sync.Mutex
}

type captureRecord struct {
	level slog.Level
	msg   string
	args  []any
}

func (l *captureLogger) Info(_ context.Context, msg string, args ...any) {
	l.add(slog.LevelInfo, msg, args)
}
func (l *captureLogger) Error(_ context.Context, msg string, args ...any) {
	l.add(slog.LevelError, msg, args)
}
func (l *captureLogger) Warn(_ context.Context, msg string, args ...any) {
	l.add(slog.LevelWarn, msg, args)
}
func (l *captureLogger) Debug(_ context.Context, msg string, args ...any) {
	l.add(slog.LevelDebug, msg, args)
}
func (l *captureLogger) Fatal(_ context.Context, _ string, _ ...any) {}
func (l *captureLogger) With(_ ...any) log.Logger                    { return l }
func (l *captureLogger) SetLevel(_ log.Level)                        {}
func (l *captureLogger) Close() error                                { return nil }

func (l *captureLogger) add(level slog.Level, msg string, args []any) {
	l.mu.Lock()
	l.records = append(l.records, captureRecord{level, msg, args})
	l.mu.Unlock()
}

func (l *captureLogger) toMap() map[string]any {
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

func (l *captureLogger) level() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, r := range l.records {
		switch r.level {
		case slog.LevelError:
			return "ERROR"
		case slog.LevelWarn:
			return "WARN"
		case slog.LevelInfo:
			return "INFO"
		}
	}
	return ""
}

func setupRouter(mw gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw)
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/notfound", func(c *gin.Context) { c.String(http.StatusNotFound, "not found") })
	r.GET("/server-error", func(c *gin.Context) { c.String(http.StatusInternalServerError, "oops") })
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "healthy") })
	return r
}

func TestLoggingFields(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	r := setupRouter(Logging())

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := cap.toMap()

	for _, key := range []string{"method", "path", "status", "latency", "client_ip", "response_size"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing field %q in log output", key)
		}
	}

	if fields["method"] != "GET" {
		t.Errorf("method = %v, want GET", fields["method"])
	}
	if fields["path"] != "/ok" {
		t.Errorf("path = %v, want /ok", fields["path"])
	}
	if fields["status"] != 200 {
		t.Errorf("status = %v, want 200", fields["status"])
	}
}

func TestLoggingLevelByStatusCode(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantLevel string
	}{
		{"2xx uses Info", "/ok", "INFO"},
		{"4xx uses Warn", "/notfound", "WARN"},
		{"5xx uses Error", "/server-error", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cap := &captureLogger{}
			log.SetDefault(cap)

			r := setupRouter(Logging())

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			got := cap.level()
			if got != tt.wantLevel {
				t.Errorf("log level = %v, want %v", got, tt.wantLevel)
			}
		})
	}
}

func TestLoggingSkipPaths(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	r := setupRouter(Logging("/healthz"))

	// Request to skipped path
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cap.mu.Lock()
	totalCalls := len(cap.records)
	cap.mu.Unlock()

	if totalCalls != 0 {
		t.Errorf("expected no log for skipped path /healthz, got %d records", totalCalls)
	}

	// Request to non-skipped path
	req = httptest.NewRequest(http.MethodGet, "/ok", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cap.mu.Lock()
	totalCalls = len(cap.records)
	cap.mu.Unlock()

	if totalCalls == 0 {
		t.Error("expected log for non-skipped path /ok, got 0 records")
	}
}

// TestLoggingRequestID verifies that request_id is included when available.
func TestLoggingRequestID(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("X-Request-Id", "test-123")
		c.Next()
	}, Logging())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := cap.toMap()
	if fields["request_id"] != "test-123" {
		t.Errorf("request_id = %v, want test-123", fields["request_id"])
	}
}

// TestLoggingJSONOutput verifies that log output is valid JSON when using slog JSON handler.
func TestLoggingJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Logging())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v\nraw: %s", err, buf.String())
	}

	for _, key := range []string{"method", "path", "status", "latency", "client_ip", "response_size"} {
		if _, ok := result[key]; !ok {
			t.Errorf("missing field %q in JSON log output", key)
		}
	}
}

// slogLoggerAdapter adapts slog.Handler to log.Logger interface.
type slogLoggerAdapter struct {
	handler slog.Handler
}

func (l *slogLoggerAdapter) Info(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, msg, args...)
}
func (l *slogLoggerAdapter) Error(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelError, msg, args...)
}
func (l *slogLoggerAdapter) Warn(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, msg, args...)
}
func (l *slogLoggerAdapter) Debug(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, msg, args...)
}

func (l *slogLoggerAdapter) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	var attrs []slog.Attr
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			attrs = append(attrs, slog.Any(args[i].(string), args[i+1]))
		}
	}
	r := slog.NewRecord(time.Time{}, level, msg, 0)
	r.AddAttrs(attrs...)
	_ = l.handler.Handle(ctx, r)
}

func (l *slogLoggerAdapter) Fatal(_ context.Context, _ string, _ ...any) {}
func (l *slogLoggerAdapter) With(_ ...any) log.Logger                    { return l }
func (l *slogLoggerAdapter) SetLevel(_ log.Level)                        {}
func (l *slogLoggerAdapter) Close() error                                { return nil }

func TestIsSkipped(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		skipPaths []string
		want      bool
	}{
		{"exact match", "/healthz", []string{"/healthz"}, true},
		{"exact no match", "/healthz/ready", []string{"/healthz"}, false},
		{"prefix match", "/metrics/health", []string{"/metrics/"}, true},
		{"prefix rejects parent", "/metrics", []string{"/metrics/"}, false},
		{"prefix match nested", "/metrics/cpu", []string{"/metrics/"}, true},
		{"empty skipPaths", "/healthz", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSkipped(tt.path, tt.skipPaths)
			if got != tt.want {
				t.Errorf("isSkipped(%q, %v) = %v, want %v", tt.path, tt.skipPaths, got, tt.want)
			}
		})
	}
}

func TestLoggingWithSkipPathsPrefixMatch(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoggingWith(WithSkipPaths("/metrics/")))
	r.GET("/metrics/health", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/metrics", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// Prefix match should skip
	req := httptest.NewRequest(http.MethodGet, "/metrics/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cap.mu.Lock()
	totalCalls := len(cap.records)
	cap.mu.Unlock()

	if totalCalls != 0 {
		t.Errorf("expected no log for /metrics/health, got %d records", totalCalls)
	}

	// Parent path should NOT be skipped
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cap.mu.Lock()
	totalCalls = len(cap.records)
	cap.mu.Unlock()

	if totalCalls == 0 {
		t.Error("expected log for /metrics, got 0 records")
	}

	// Unrelated path should NOT be skipped
	req = httptest.NewRequest(http.MethodGet, "/ok", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cap.mu.Lock()
	totalCalls = len(cap.records)
	cap.mu.Unlock()

	if totalCalls == 0 {
		t.Error("expected log for /ok, got 0 records")
	}
}

func TestLoggingWithHook(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	var hookCalled bool
	var hookFields map[string]any

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoggingWith(WithHook(func(c *gin.Context, fields map[string]any) {
		hookCalled = true
		hookFields = fields
		fields["custom_key"] = "custom_val"
	})))
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !hookCalled {
		t.Fatal("hook was not called")
	}
	if hookFields == nil {
		t.Fatal("hook received nil fields")
	}
	if hookFields["method"] != "GET" {
		t.Errorf("hook fields method = %v, want GET", hookFields["method"])
	}

	// Custom field should appear in the log output
	fields := cap.toMap()
	if fields["custom_key"] != "custom_val" {
		t.Errorf("expected custom_key=custom_val in log output, got %v", fields)
	}
}
