package server

import (
	"context"
	"net/http"
	"time"

	"github.com/fztcjjl/quix/core/transport"
	"github.com/fztcjjl/quix/core/transport/http/server/middleware"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
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
	addr              string
	defaultMiddleware bool
}

// WithAddr sets the server listen address.
func WithAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

// WithDefaultMiddleware controls whether default middleware (Recovery, RequestID) is mounted.
func WithDefaultMiddleware(enabled bool) Option {
	return func(o *options) {
		o.defaultMiddleware = enabled
	}
}

// NewServer creates a new HTTP Server with Gin engine.
func NewServer(opts ...Option) *Server {
	o := &options{defaultMiddleware: true}
	for _, opt := range opts {
		opt(o)
	}

	engine := gin.New()

	s := &Server{
		Engine: engine,
		addr:   o.addr,
		server: &http.Server{Addr: o.addr, Handler: engine, ReadHeaderTimeout: 5 * time.Second},
	}

	if o.defaultMiddleware {
		engine.Use(middleware.Recovery(), requestid.New())
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
