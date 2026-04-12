package quix

import (
	"github.com/fztcjjl/quix/core/config"
	"github.com/fztcjjl/quix/core/log"
	"github.com/fztcjjl/quix/core/transport"
	qhttp "github.com/fztcjjl/quix/core/transport/http/server"
	"github.com/gin-gonic/gin"
)

// Option configures the App during creation.
type Option func(*App)

// WithLogger sets a custom Logger implementation for the App.
func WithLogger(l log.Logger) Option {
	return func(a *App) {
		a.logger = l
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
func WithGinMode(mode string) Option {
	return func(a *App) {
		gin.SetMode(mode)
	}
}

// WithDefaultMiddleware controls whether default middleware (Recovery, RequestID) is mounted.
func WithDefaultMiddleware(enabled bool) Option {
	return func(a *App) {
		a.defaultMiddleware = enabled
	}
}
