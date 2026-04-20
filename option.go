package quix

import (
	"time"

	"github.com/fztcjjl/quix/core/config"
	"github.com/fztcjjl/quix/core/log"
	"github.com/fztcjjl/quix/core/telemetry"
	"github.com/fztcjjl/quix/core/transport"
	qhttp "github.com/fztcjjl/quix/core/transport/http/server"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// options holds all user-configurable settings for the App.
type options struct {
	config           config.Config
	env              Environment
	telemetryOpts    []telemetry.Option
	corsEnabled      bool
	corsConfig       *cors.Config
	loggingSkipPaths []string
	setupFuncs       []func(*App) error
	shutdownTimeout  time.Duration
}

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

// WithCORS controls whether CORS middleware is mounted in the default middleware chain.
// When set to false, CORS middleware is not mounted even if default middleware is enabled.
func WithCORS(enabled bool) Option {
	return func(a *App) {
		a.corsEnabled = enabled
	}
}

// WithCORSConfig sets a custom CORS configuration for the default middleware chain.
// When set, CORS middleware is mounted with this config instead of cors.Default().
func WithCORSConfig(cfg cors.Config) Option {
	return func(a *App) {
		a.corsConfig = &cfg
	}
}

// WithLoggingSkipPaths sets paths to skip logging.
// Paths ending with "/" use prefix matching.
func WithLoggingSkipPaths(paths ...string) Option {
	return func(a *App) {
		a.loggingSkipPaths = paths
	}
}
