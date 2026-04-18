package quix

import (
	"time"

	"github.com/fztcjjl/quix/core/config"
	"github.com/fztcjjl/quix/core/log"
	"github.com/fztcjjl/quix/core/telemetry"
	"github.com/fztcjjl/quix/core/transport"
	qhttp "github.com/fztcjjl/quix/core/transport/http/server"
	"github.com/gin-gonic/gin"
)

// Option configures the App during creation.
type Option func(*App)

// WithLogger sets a custom Logger as the global default.
func WithLogger(l log.Logger) Option {
	return func(a *App) {
		log.SetDefault(l)
	}
}

// WithConfig sets a custom Config implementation for the App.
func WithConfig(c config.Config) Option {
	return func(a *App) {
		a.config = c
	}
}

// WithHttpServer sets a custom HTTP server for the App.
func WithHttpServer(s *qhttp.Server) Option {
	return func(a *App) {
		a.httpServer = s
	}
}

// WithRpcServer sets a custom RPC server for the App.
func WithRpcServer(s transport.Server) Option {
	return func(a *App) {
		a.rpcServer = s
	}
}

// WithGinMode sets the Gin mode (debug, release, test).
// This overrides the automatic Gin mode derived from QUIX_ENV.
func WithGinMode(mode string) Option {
	return func(a *App) {
		gin.SetMode(mode)
	}
}

// WithEnv sets the application environment, overriding QUIX_ENV.
func WithEnv(env Environment) Option {
	return func(a *App) {
		a.env = env
	}
}

// WithDefaultMiddleware controls whether default middleware (Recovery, RequestID) is mounted.
func WithDefaultMiddleware(enabled bool) Option {
	return func(a *App) {
		a.defaultMiddleware = enabled
	}
}

// WithTelemetry enables OpenTelemetry instrumentation.
// It initializes TracerProvider and MeterProvider, and configures the App
// to flush telemetry on shutdown.
func WithTelemetry(opts ...telemetry.Option) Option {
	return func(a *App) {
		a.telemetryOpts = opts
	}
}

// WithSetup registers startup callbacks that run before the HTTP server starts.
// Callbacks execute in registration order. If a callback returns an error,
// the app logs the error and exits with code 1.
func WithSetup(funcs ...func(*App) error) Option {
	return func(a *App) {
		a.setupFuncs = append(a.setupFuncs, funcs...)
	}
}

// WithShutdownTimeout sets the maximum duration for graceful shutdown.
// Defaults to 5 seconds if not set.
func WithShutdownTimeout(d time.Duration) Option {
	return func(a *App) {
		a.shutdownTimeout = d
	}
}
