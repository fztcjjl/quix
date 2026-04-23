package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	qerrors "github.com/fztcjjl/quix/core/errors"
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

func (l *captureLogger) Trace(_ context.Context, msg string, args ...any) {
	l.add(slog.Level(-8), msg, args)
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

	r := setupRouter(AccessLog())

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

			r := setupRouter(AccessLog())

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

	r := setupRouter(AccessLog(WithSkipPaths("/healthz")))

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
	}, AccessLog())
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
	r.Use(AccessLog())
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

func (l *slogLoggerAdapter) Trace(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.Level(-8), msg, args...)
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
	r.Use(AccessLog(WithSkipPaths("/metrics/")))
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
	r.Use(AccessLog(WithHook(func(c *gin.Context, fields map[string]any) {
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

func TestLoggingLatencyMsField(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	latencyMs, ok := result["latency_ms"].(float64)
	if !ok {
		t.Fatal("latency_ms field missing or not a float64")
	}
	if latencyMs <= 0 {
		t.Errorf("latency_ms = %v, want > 0", latencyMs)
	}
}

func TestLoggingQueryField(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/ok?page=2&size=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := cap.toMap()
	if fields["query"] != "page=2&size=10" {
		t.Errorf("query = %v, want page=2&size=10", fields["query"])
	}
}

func TestLoggingNoQueryField(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := cap.toMap()
	if _, ok := fields["query"]; ok {
		t.Error("query field should not be present when URL has no query string")
	}
}

func TestLoggingRouteField(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.GET("/users/:id", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := cap.toMap()
	if fields["path"] != "/users/42" {
		t.Errorf("path = %v, want /users/42", fields["path"])
	}
	if fields["route"] != "/users/:id" {
		t.Errorf("route = %v, want /users/:id", fields["route"])
	}
}

func TestLoggingNoRouteFor404(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := cap.toMap()
	if _, ok := fields["route"]; ok {
		t.Error("route field should not be present for unmatched paths")
	}
}

func TestLoggingErrorCodeField(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.GET("/error", func(c *gin.Context) {
		c.Set("app_error", &qerrors.Error{Code: "not_found", Message: "user not found", StatusCode: 404})
		c.String(http.StatusNotFound, "not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := cap.toMap()
	if fields["error_code"] != "not_found" {
		t.Errorf("error_code = %v, want not_found", fields["error_code"])
	}
}

func TestLoggingErrorCodeForPlainError(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.GET("/error", func(c *gin.Context) {
		c.Set("app_error", errors.New("something went wrong"))
		c.String(http.StatusInternalServerError, "error")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := cap.toMap()
	if fields["error_code"] != "internal_error" {
		t.Errorf("error_code = %v, want internal_error", fields["error_code"])
	}
	if fields["error_message"] != "something went wrong" {
		t.Errorf("error_message = %v, want %q", fields["error_message"], "something went wrong")
	}
}

func TestLoggingErrorMessageField(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.GET("/error", func(c *gin.Context) {
		c.Set("app_error", &qerrors.Error{Code: "not_found", Message: "user not found", StatusCode: 404})
		c.String(http.StatusNotFound, "not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := cap.toMap()
	if fields["error_code"] != "not_found" {
		t.Errorf("error_code = %v, want not_found", fields["error_code"])
	}
	if fields["error_message"] != "user not found" {
		t.Errorf("error_message = %v, want %q", fields["error_message"], "user not found")
	}
}

func TestLoggingSlowRequest(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog(WithSlowThreshold(10 * time.Millisecond)))
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.String(http.StatusOK, "slow")
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cap.mu.Lock()
	hasSlowRequest := false
	for _, rec := range cap.records {
		if rec.msg == "slow request" {
			hasSlowRequest = true
		}
	}
	cap.mu.Unlock()

	if !hasSlowRequest {
		t.Error("expected slow request log entry")
	}
}

func TestLoggingNoSlowRequestWhenFast(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog(WithSlowThreshold(1 * time.Second)))
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cap.mu.Lock()
	hasSlowRequest := false
	for _, rec := range cap.records {
		if rec.msg == "slow request" {
			hasSlowRequest = true
		}
	}
	cap.mu.Unlock()

	if hasSlowRequest {
		t.Error("should not have slow request log for fast requests")
	}
}

func TestLoggingNoSlowRequestWithoutThreshold(t *testing.T) {
	cap := &captureLogger{}
	log.SetDefault(cap)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog()) // no slow threshold configured
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.String(http.StatusOK, "slow")
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cap.mu.Lock()
	hasSlowRequest := false
	for _, rec := range cap.records {
		if rec.msg == "slow request" {
			hasSlowRequest = true
		}
	}
	cap.mu.Unlock()

	if hasSlowRequest {
		t.Error("should not have slow request log when threshold is not configured")
	}
}

func TestWithRequestLoggerInjectsContextLogger(t *testing.T) {
	orig := log.Default()
	defer log.SetDefault(orig)

	// Use slog adapter to verify fields are preserved through With().
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := log.NewSlog(slog.New(handler))
	log.SetDefault(l)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("X-Request-Id", "req-abc")
		c.Next()
	}, WithRequestLogger())
	r.GET("/ok", func(c *gin.Context) {
		ctxLogger := log.FromContext(c.Request.Context())
		// Verify the context logger is different from default (has With fields)
		ctxLogger.Info(c.Request.Context(), "handler log", "key", "val")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v\nraw: %s", err, buf.String())
	}

	if result["request_id"] != "req-abc" {
		t.Errorf("request_id = %v, want req-abc", result["request_id"])
	}
}

func TestWithRequestLoggerWithoutRequestID(t *testing.T) {
	cap := &captureLogger{}
	orig := log.Default()
	log.SetDefault(cap)
	defer log.SetDefault(orig)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	// No requestid middleware
	r.Use(WithRequestLogger())
	r.GET("/ok", func(c *gin.Context) {
		l := log.FromContext(c.Request.Context())
		l.Info(c.Request.Context(), "handler log")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cap.mu.Lock()
	defer cap.mu.Unlock()
	if len(cap.records) == 0 {
		t.Fatal("expected log from handler, got none")
	}

	// Should not contain request_id
	rec := cap.records[0]
	for i := 0; i+1 < len(rec.args); i += 2 {
		if rec.args[i].(string) == "request_id" {
			t.Error("should not contain request_id when no requestid middleware")
		}
	}
}

func TestWithRequestLoggerFallsBackToDefault(t *testing.T) {
	orig := log.Default()
	defer log.SetDefault(orig)

	custom := &captureLogger{}
	log.SetDefault(custom)

	// WithRequestLogger should use FromContext which returns Default() when no logger in context
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(WithRequestLogger())
	r.GET("/ok", func(c *gin.Context) {
		l := log.FromContext(c.Request.Context())
		if l != custom {
			t.Error("FromContext should return default logger when no logger in context")
		}
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
}

func TestAccessLogBytesInField(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.POST("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/ok", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Length", "15")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	if result["request_size"] != float64(15) {
		t.Errorf("request_size = %v, want 15", result["request_size"])
	}
}

func TestAccessLogContentTypeField(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.POST("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/ok", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	if result["content_type"] != "application/json" {
		t.Errorf("content_type = %v, want application/json", result["content_type"])
	}
}

func TestAccessLogNoContentTypeWhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	if _, ok := result["content_type"]; ok {
		t.Error("content_type should not be present for GET without Content-Type header")
	}
}

func TestBodyLogJsonContentType(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog(WithBodyLog(1024)))
	r.POST("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	body := `{"user":"alice","email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/ok", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	if result["request_body"] != body {
		t.Errorf("request_body = %v, want %v", result["request_body"], body)
	}
}

func TestBodyLogTruncation(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog(WithBodyLog(10)))
	r.POST("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	body := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	req := httptest.NewRequest(http.MethodPost, "/ok", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	got, _ := result["request_body"].(string)
	if got != "0123456789" {
		t.Errorf("request_body = %q, want %q", got, "0123456789")
	}
	if result["body_truncated"] != true {
		t.Error("body_truncated should be true when body exceeds maxBytes")
	}
}

func TestBodyLogNotTruncated(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog(WithBodyLog(1024)))
	r.POST("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	body := "hello"
	req := httptest.NewRequest(http.MethodPost, "/ok", strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	if _, ok := result["body_truncated"]; ok {
		t.Error("body_truncated should not be present when body fits in maxBytes")
	}
}

func TestBodyLogSkipsMultipart(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog(WithBodyLog(1024)))
	r.POST("/upload", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("field1=value1"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=abc123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	if _, ok := result["request_body"]; ok {
		t.Error("request_body should not be present for multipart/form-data")
	}
}

func TestBodyLogSkipsOctetStream(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog(WithBodyLog(1024)))
	r.POST("/download", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/download", strings.NewReader("binary-data"))
	req.Header.Set("Content-Type", "application/octet-stream")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	if _, ok := result["request_body"]; ok {
		t.Error("request_body should not be present for application/octet-stream")
	}
}

func TestBodyLogSkipsProtobuf(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog(WithBodyLog(1024)))
	r.POST("/proto", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/proto", strings.NewReader("protobuf-data"))
	req.Header.Set("Content-Type", "application/x-protobuf")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	if _, ok := result["request_body"]; ok {
		t.Error("request_body should not be present for application/x-protobuf")
	}
}

func TestBodyLogDisabledByDefault(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog()) // no WithBodyLog
	r.POST("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/ok", strings.NewReader(`{"key":"val"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	if _, ok := result["request_body"]; ok {
		t.Error("request_body should not be present when WithBodyLog is not configured")
	}
}

func TestBodyLogCapturesTextPlain(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log.SetDefault(&slogLoggerAdapter{handler: handler})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog(WithBodyLog(1024)))
	r.POST("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	body := "plain text body content"
	req := httptest.NewRequest(http.MethodPost, "/ok", strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	if result["request_body"] != body {
		t.Errorf("request_body = %v, want %v", result["request_body"], body)
	}
}

func TestIsLoggableContentType(t *testing.T) {
	tests := []struct {
		ct   string
		want bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"application/x-www-form-urlencoded", true},
		{"application/xml", true},
		{"text/plain", true},
		{"text/html", true},
		{"multipart/form-data; boundary=abc", false},
		{"application/octet-stream", false},
		{"application/grpc+proto", false},
		{"application/x-protobuf", false},
		{"", false},
		{"application/msgpack", false},
	}
	for _, tt := range tests {
		t.Run(tt.ct, func(t *testing.T) {
			if got := isLoggableContentType(tt.ct); got != tt.want {
				t.Errorf("isLoggableContentType(%q) = %v, want %v", tt.ct, got, tt.want)
			}
		})
	}
}

func TestTruncateBody(t *testing.T) {
	// Within limit, not truncated.
	b, truncated := truncateBody([]byte("hello"), 10)
	if truncated {
		t.Error("should not be truncated")
	}
	if string(b) != "hello" {
		t.Errorf("got %q, want %q", string(b), "hello")
	}

	// Exceeds limit, truncated.
	b, truncated = truncateBody([]byte("0123456789ABCDEF"), 8)
	if !truncated {
		t.Error("should be truncated")
	}
	if string(b) != "01234567" {
		t.Errorf("got %q, want %q", string(b), "01234567")
	}

	// Zero max means no truncation.
	b, truncated = truncateBody([]byte("hello"), 0)
	if truncated {
		t.Error("should not be truncated when max is 0")
	}
	if string(b) != "hello" {
		t.Errorf("got %q, want %q", string(b), "hello")
	}
}
