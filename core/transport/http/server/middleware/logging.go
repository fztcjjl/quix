package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fztcjjl/quix/core/errors"
	"github.com/fztcjjl/quix/core/log"
	"github.com/fztcjjl/quix/core/telemetry"
	"github.com/gin-gonic/gin"
)

// AccessLogHookFunc is called after each request with the collected log fields.
// It can be used to add custom fields or perform side effects.
type AccessLogHookFunc func(c *gin.Context, fields map[string]any)

// accessLogConfig holds configuration for the access log middleware.
type accessLogConfig struct {
	skipPaths     []string
	hook          AccessLogHookFunc
	slowThreshold time.Duration
	bodyLogMax    int
}

// AccessLogOption configures the AccessLog middleware.
type AccessLogOption func(*accessLogConfig)

// WithSkipPaths sets paths to skip logging. Paths ending with "/" use prefix matching.
func WithSkipPaths(paths ...string) AccessLogOption {
	return func(cfg *accessLogConfig) {
		cfg.skipPaths = paths
	}
}

// WithHook sets a custom hook function called after each request.
func WithHook(fn AccessLogHookFunc) AccessLogOption {
	return func(cfg *accessLogConfig) {
		cfg.hook = fn
	}
}

// WithSlowThreshold sets the slow request detection threshold.
// Requests exceeding this duration will generate an additional WARN log.
func WithSlowThreshold(d time.Duration) AccessLogOption {
	return func(cfg *accessLogConfig) {
		cfg.slowThreshold = d
	}
}

// WithBodyLog enables request body logging in access logs.
// maxBytes controls the maximum number of bytes to log per field;
// body exceeding this limit is truncated and a "body_truncated": true field is appended.
// Only text-like content types are captured; binary, multipart, and gRPC types are skipped.
func WithBodyLog(maxBytes int) AccessLogOption {
	return func(cfg *accessLogConfig) {
		cfg.bodyLogMax = maxBytes
	}
}

// isSkipped checks if a path should be skipped.
// Paths ending with "/" match any path with that prefix.
func isSkipped(path string, skipPaths []string) bool {
	for _, p := range skipPaths {
		if strings.HasSuffix(p, "/") {
			if strings.HasPrefix(path, p) {
				return true
			}
		} else {
			if path == p {
				return true
			}
		}
	}
	return false
}

// isLoggableContentType returns true for text-like content types safe to log.
func isLoggableContentType(ct string) bool {
	if ct == "" {
		return false
	}
	switch {
	case strings.HasPrefix(ct, "text/"):
		return true
	case strings.HasPrefix(ct, "application/json"):
		return true
	case strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		return true
	case strings.HasPrefix(ct, "application/xml"):
		return true
	}
	return false
}

// truncateBody truncates a byte slice to max bytes and returns whether it was truncated.
func truncateBody(b []byte, max int) ([]byte, bool) {
	if max <= 0 || len(b) <= max {
		return b, false
	}
	return b[:max], true
}

// AccessLog returns a middleware that logs each HTTP request with structured fields.
func AccessLog(opts ...AccessLogOption) gin.HandlerFunc {
	var cfg accessLogConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Read request body before c.Next() consumes it.
		var reqBody []byte
		if cfg.bodyLogMax > 0 {
			ct := c.ContentType()
			if isLoggableContentType(ct) {
				var buf bytes.Buffer
				reqBody, _ = io.ReadAll(io.TeeReader(c.Request.Body, &buf))
				c.Request.Body = io.NopCloser(&buf)
			}
		}

		c.Next()

		if isSkipped(path, cfg.skipPaths) {
			return
		}

		latency := time.Since(start)
		status := c.Writer.Status()
		reqID, _ := c.Get("X-Request-Id")
		clientIP := c.ClientIP()
		ct := c.ContentType()

		// Build log args directly as a slice (avoids intermediate map allocation).
		args := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency", latency.String(),
			"latency_ms", float64(latency) / float64(time.Millisecond),
			"client_ip", clientIP,
			"request_size", c.Request.ContentLength,
			"response_size", c.Writer.Size(),
		}
		if ct != "" {
			args = append(args, "content_type", ct)
		}
		if reqID != nil {
			args = append(args, "request_id", reqID)
		}

		ctx := c.Request.Context()
		if traceID := telemetry.ExtractTraceID(ctx); traceID != "" {
			args = append(args, "trace_id", traceID)
		}
		if spanID := telemetry.ExtractSpanID(ctx); spanID != "" {
			args = append(args, "span_id", spanID)
		}

		if query := c.Request.URL.RawQuery; query != "" {
			args = append(args, "query", query)
		}
		if ua := c.Request.UserAgent(); ua != "" {
			args = append(args, "user_agent", ua)
		}
		if route := c.FullPath(); route != "" {
			args = append(args, "route", route)
		}

		// Request body logging.
		if len(reqBody) > 0 {
			body, truncated := truncateBody(reqBody, cfg.bodyLogMax)
			args = append(args, "request_body", string(body))
			if truncated {
				args = append(args, "body_truncated", true)
			}
		}

		// Extract error_code from application error if present.
		if appErrVal, exists := c.Get("app_error"); exists {
			if appErr, ok := appErrVal.(*errors.Error); ok && appErr.Code != "" {
				args = append(args, "error_code", appErr.Code)
			}
		}

		if cfg.hook != nil {
			// Build a map from args for the hook, then rebuild args with any additions.
			fields := sliceToMap(args)
			cfg.hook(c, fields)
			args = mapToSlice(fields)
		}

		switch {
		case status >= http.StatusInternalServerError:
			log.Error(ctx, "request completed", args...)
		case status >= http.StatusBadRequest:
			log.Warn(ctx, "request completed", args...)
		default:
			log.Info(ctx, "request completed", args...)
		}

		// Slow request detection.
		if cfg.slowThreshold > 0 && latency > cfg.slowThreshold {
			slowArgs := []any{
				"path", path,
				"latency_ms", float64(latency) / float64(time.Millisecond),
				"threshold_ms", float64(cfg.slowThreshold) / float64(time.Millisecond),
			}
			log.Warn(ctx, "slow request", slowArgs...)
		}
	}
}

// sliceToMap converts a flat key-value slice to a map.
func sliceToMap(args []any) map[string]any {
	m := make(map[string]any, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		if k, ok := args[i].(string); ok {
			m[k] = args[i+1]
		}
	}
	return m
}

// mapToSlice converts a map to a flat key-value slice.
func mapToSlice(m map[string]any) []any {
	args := make([]any, 0, len(m)*2)
	for k, v := range m {
		args = append(args, k, v)
	}
	return args
}

// RequestLoggerOption configures the WithRequestLogger middleware.
type RequestLoggerOption func(*requestLoggerConfig)

type requestLoggerConfig struct{}

// WithRequestLogger returns a middleware that creates a request-scoped Logger
// enriched with trace_id, span_id, and request_id, and stores it in the context.
// Downstream handlers can retrieve it via log.FromContext(ctx).
func WithRequestLogger(opts ...RequestLoggerOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		base := log.FromContext(c.Request.Context())

		var fields []any

		// Extract request_id from gin context (set by requestid middleware).
		if reqID, exists := c.Get("X-Request-Id"); exists {
			fields = append(fields, "request_id", reqID)
		}

		ctx := c.Request.Context()

		// Extract trace_id.
		if traceID := telemetry.ExtractTraceID(ctx); traceID != "" {
			fields = append(fields, "trace_id", traceID)
		}

		// Extract span_id.
		if spanID := telemetry.ExtractSpanID(ctx); spanID != "" {
			fields = append(fields, "span_id", spanID)
		}

		child := base.With(fields...)
		newCtx := log.NewContext(ctx, child)
		c.Request = c.Request.WithContext(newCtx)

		c.Next()
	}
}
