package server

import (
	"context"
	"net/http"
	"time"

	"github.com/fztcjjl/quix/core/transport"
	"github.com/fztcjjl/quix/core/transport/http/server/middleware"
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
	telemetryServiceName   string
	telemetryTracesEnabled bool
}

// WithAddr sets the server listen address.
func WithAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

// WithDefaultMiddleware controls whether default middleware (Recovery, RequestID, ResponseMiddleware) is mounted.
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

// NewServer creates a new HTTP Server with Gin engine.
func NewServer(opts ...Option) *Server {
	o := &options{
		defaultMiddleware: true,
		readHeaderTimeout: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(o)
	}

	engine := gin.New()

	s := &Server{
		Engine: engine,
		addr:   o.addr,
		server: &http.Server{Addr: o.addr, Handler: engine, ReadHeaderTimeout: o.readHeaderTimeout},
	}

	if o.defaultMiddleware {
		engine.Use(middleware.Recovery())
		if o.telemetryServiceName != "" && o.telemetryTracesEnabled {
			engine.Use(otelgin.Middleware(o.telemetryServiceName))
		}
		engine.Use(requestid.New(), middleware.Logging(), middleware.ResponseMiddleware())
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
