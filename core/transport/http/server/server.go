package server

import (
	"context"
	"net/http"
	"time"

	"github.com/fztcjjl/quix/core/transport"
	"github.com/fztcjjl/quix/core/transport/http/server/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Server implements transport.Server for HTTP using Gin.
type Server struct {
	*gin.Engine
	addr   string
	server *http.Server
}

// Option configures the HTTP Server.
type Option func(*options)

type options struct {
	addr                   string
	defaultMiddleware      bool
	readHeaderTimeout      time.Duration
	readTimeout            time.Duration
	writeTimeout           time.Duration
	idleTimeout            time.Duration
	telemetryServiceName   string
	telemetryTracesEnabled bool
	corsEnabled            bool
	corsConfig             *cors.Config
	loggingSkipPaths       []string
}

// WithAddr sets the server listen address.
func WithAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

// WithDefaultMiddleware controls whether default middleware (RequestID, Recovery, CORS, Logging, Response) is mounted. Order: RequestID → [otelgin] → Recovery → CORS → Logging → Response.
func WithDefaultMiddleware(enabled bool) Option {
	return func(o *options) {
		o.defaultMiddleware = enabled
	}
}

// WithReadHeaderTimeout sets the server ReadHeaderTimeout.
func WithReadHeaderTimeout(d time.Duration) Option {
	return func(o *options) {
		o.readHeaderTimeout = d
	}
}

// WithReadTimeout sets the server ReadTimeout.
func WithReadTimeout(d time.Duration) Option {
	return func(o *options) {
		o.readTimeout = d
	}
}

// WithWriteTimeout sets the server WriteTimeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(o *options) {
		o.writeTimeout = d
	}
}

// WithIdleTimeout sets the server IdleTimeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(o *options) {
		o.idleTimeout = d
	}
}

// WithTelemetryServiceName sets the service name for otelgin middleware.
func WithTelemetryServiceName(name string) Option {
	return func(o *options) {
		o.telemetryServiceName = name
	}
}

// WithTelemetryTracesEnabled controls whether otelgin middleware is injected.
func WithTelemetryTracesEnabled(enabled bool) Option {
	return func(o *options) {
		o.telemetryTracesEnabled = enabled
	}
}

// WithCORSConfig sets a custom CORS configuration for the default middleware chain.
// When set, CORS middleware is mounted with this config instead of cors.Default().
func WithCORSConfig(cfg cors.Config) Option {
	return func(o *options) {
		o.corsConfig = &cfg
	}
}

// WithCORS controls whether CORS middleware is mounted in the default middleware chain.
// When set to false, CORS middleware is not mounted even if default middleware is enabled.
func WithCORS(enabled bool) Option {
	return func(o *options) {
		o.corsEnabled = enabled
	}
}

// WithLoggingSkipPaths sets paths to skip logging.
// Paths ending with "/" use prefix matching.
func WithLoggingSkipPaths(paths ...string) Option {
	return func(o *options) {
		o.loggingSkipPaths = paths
	}
}

// NewServer creates a new HTTP Server with Gin engine.
func NewServer(opts ...Option) *Server {
	o := &options{
		defaultMiddleware: true,
		corsEnabled:       true,
		readHeaderTimeout: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(o)
	}

	engine := gin.New()

	s := &Server{
		Engine: engine,
		addr:   o.addr,
		server: &http.Server{
			Addr:              o.addr,
			Handler:           engine,
			ReadHeaderTimeout: o.readHeaderTimeout,
			ReadTimeout:       o.readTimeout,
			WriteTimeout:      o.writeTimeout,
			IdleTimeout:       o.idleTimeout,
		},
	}

	if o.defaultMiddleware {
		engine.Use(requestid.New())
		if o.telemetryServiceName != "" && o.telemetryTracesEnabled {
			engine.Use(otelgin.Middleware(o.telemetryServiceName))
		}
		engine.Use(middleware.WithRequestLogger())
		engine.Use(middleware.Recovery())
		if o.corsEnabled {
			if o.corsConfig != nil {
				engine.Use(middleware.WithCORSConfig(*o.corsConfig))
			} else {
				engine.Use(middleware.CORS())
			}
		}
		engine.Use(
			middleware.AccessLog(middleware.WithSkipPaths(o.loggingSkipPaths...)),
			middleware.ResponseMiddleware(),
		)
	}

	return s
}

// Addr returns the server listen address.
func (s *Server) Addr() string {
	return s.addr
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Compile-time check
var _ transport.Server = (*Server)(nil)
