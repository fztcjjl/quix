package server

import (
	"context"
	"net/http"
	"time"

	"github.com/fztcjjl/quix/core/transport"
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
	addr string
}

// WithAddr sets the server listen address.
func WithAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

// NewServer creates a new HTTP Server with Gin engine.
func NewServer(opts ...Option) *Server {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	engine := gin.New()

	return &Server{
		Engine: engine,
		addr:   o.addr,
		server: &http.Server{Addr: o.addr, Handler: engine, ReadHeaderTimeout: 5 * time.Second},
	}
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
