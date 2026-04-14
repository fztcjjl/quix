package middleware

import (
	"net/http"
	"time"

	"github.com/fztcjjl/quix/core/log"
	"github.com/gin-gonic/gin"
)

// Logging returns a middleware that logs each HTTP request with structured fields.
// SkipPaths specifies exact paths to skip logging (e.g., "/healthz").
func Logging(skipPaths ...string) gin.HandlerFunc {
	skip := make(map[string]struct{}, len(skipPaths))
	for _, p := range skipPaths {
		skip[p] = struct{}{}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()

		if _, ok := skip[path]; ok {
			return
		}

		latency := time.Since(start)
		status := c.Writer.Status()
		reqID, _ := c.Get("X-Request-Id")
		clientIP := c.ClientIP()

		fields := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency", latency.String(),
			"client_ip", clientIP,
			"response_size", c.Writer.Size(),
		}
		if reqID != nil {
			fields = append(fields, "request_id", reqID)
		}

		ctx := c.Request.Context()

		switch {
		case status >= http.StatusInternalServerError:
			log.Error(ctx, "request completed", fields...)
		case status >= http.StatusBadRequest:
			log.Warn(ctx, "request completed", fields...)
		default:
			log.Info(ctx, "request completed", fields...)
		}
	}
}
