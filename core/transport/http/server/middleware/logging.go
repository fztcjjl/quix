package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/fztcjjl/quix/core/log"
	"github.com/gin-gonic/gin"
)

// LoggingHookFunc is called after each request with the collected log fields.
// It can be used to add custom fields or perform side effects.
type LoggingHookFunc func(c *gin.Context, fields map[string]any)

// ExtractTraceID extracts trace_id from context for logging middleware.
// Set this variable to enable trace_id output in access logs.
// When nil (default), no trace_id is included in log output.
var ExtractTraceID func(ctx context.Context) string

// loggingConfig holds configuration for the logging middleware.
type loggingConfig struct {
	skipPaths []string
	hook      LoggingHookFunc
}

// LoggingOption configures the logging middleware.
type LoggingOption func(*loggingConfig)

// WithSkipPaths sets paths to skip logging. Paths ending with "/" use prefix matching.
func WithSkipPaths(paths ...string) LoggingOption {
	return func(cfg *loggingConfig) {
		cfg.skipPaths = paths
	}
}

// WithHook sets a custom hook function called after each request.
func WithHook(fn LoggingHookFunc) LoggingOption {
	return func(cfg *loggingConfig) {
		cfg.hook = fn
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

// Logging returns a middleware that logs each HTTP request with structured fields.
// SkipPaths specifies exact paths to skip logging (e.g., "/healthz").
func Logging(skipPaths ...string) gin.HandlerFunc {
	return LoggingWith(WithSkipPaths(skipPaths...))
}

// LoggingWith returns a middleware that logs each HTTP request with structured fields.
// It supports functional options for customizing behavior.
func LoggingWith(opts ...LoggingOption) gin.HandlerFunc {
	var cfg loggingConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()

		if isSkipped(path, cfg.skipPaths) {
			return
		}

		latency := time.Since(start)
		status := c.Writer.Status()
		reqID, _ := c.Get("X-Request-Id")
		clientIP := c.ClientIP()

		fields := map[string]any{
			"method":        c.Request.Method,
			"path":          path,
			"status":        status,
			"latency":       latency.String(),
			"client_ip":     clientIP,
			"response_size": c.Writer.Size(),
		}
		if reqID != nil {
			fields["request_id"] = reqID
		}

		ctx := c.Request.Context()
		if ExtractTraceID != nil {
			if traceID := ExtractTraceID(ctx); traceID != "" {
				fields["trace_id"] = traceID
			}
		}

		if cfg.hook != nil {
			cfg.hook(c, fields)
		}

		args := mapToSlice(fields)

		switch {
		case status >= http.StatusInternalServerError:
			log.Error(ctx, "request completed", args...)
		case status >= http.StatusBadRequest:
			log.Warn(ctx, "request completed", args...)
		default:
			log.Info(ctx, "request completed", args...)
		}
	}
}

// mapToSlice converts a map to a flat key-value slice.
func mapToSlice(m map[string]any) []any {
	args := make([]any, 0, len(m)*2)
	for k, v := range m {
		args = append(args, k, v)
	}
	return args
}
