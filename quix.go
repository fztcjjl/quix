package quix

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fztcjjl/quix/core/config"
	"github.com/fztcjjl/quix/core/log"
	"github.com/fztcjjl/quix/core/transport"
	qhttp "github.com/fztcjjl/quix/core/transport/http/server"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// shutdownTimeout is the maximum duration for graceful shutdown.
const shutdownTimeout = 5 * time.Second

// App is the core framework application.
type App struct {
	httpServer *qhttp.Server
	rpcServer  transport.Server
	logger     log.Logger
	config     config.Config
}

// resolveHttpAddr reads the HTTP server address from config.
// It checks "http.addr" first, then falls back to "http.port" (default 8080).
func resolveHttpAddr(c config.Config) string {
	if addr := c.String("http.addr"); addr != "" {
		return addr
	}
	port := c.Int("http.port")
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf(":%d", port)
}

// New creates a new App with the given options.
// If no logger is provided, zerolog is used by default.
// If no config is provided, koanf with environment variables is used by default.
func New(opts ...Option) *App {
	defaultLog := log.NewZerolog(zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger())
	defaultCfg, _ := config.NewKoanf()
	app := &App{
		logger: defaultLog,
		config: defaultCfg,
	}
	log.SetDefault(defaultLog)
	for _, opt := range opts {
		opt(app)
	}
	// Config-driven server creation:
	// - If http is configured (http.addr or http.port), start HTTP server
	// - If rpc is configured (rpc.addr), start RPC server
	// - If neither is configured, default to HTTP server on :8080
	hasHttpConfig := app.config.String("http.addr") != "" || app.config.Int("http.port") != 0
	hasRpcConfig := app.config.String("rpc.addr") != ""

	if app.httpServer == nil && (hasHttpConfig || !hasRpcConfig) {
		app.httpServer = qhttp.NewServer(qhttp.WithAddr(resolveHttpAddr(app.config)))
	}
	// TODO: create RPC server when RPC transport is implemented
	// if app.rpcServer == nil && hasRpcConfig { ... }
	return app
}

// Logger returns the App's logger.
func (a *App) Logger() log.Logger {
	return a.logger
}

// Config returns the App's config.
func (a *App) Config() config.Config {
	return a.config
}

// HttpServer returns the App's HTTP server.
func (a *App) HttpServer() *qhttp.Server {
	return a.httpServer
}

// Run starts all servers and blocks until shutdown.
func (a *App) Run() {
	if a.httpServer != nil {
		go func() {
			if err := a.httpServer.Start(); err != nil {
				a.logger.Error(context.Background(), "http server failed to start",
					"error", fmt.Sprintf("%v", err))
			}
		}()
	}
	if a.rpcServer != nil {
		go func() {
			if err := a.rpcServer.Start(); err != nil {
				a.logger.Error(context.Background(), "rpc server failed to start",
					"error", fmt.Sprintf("%v", err))
			}
		}()
	}

	a.logger.Info(context.Background(), "server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info(context.Background(), "shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Stop RPC server first, then HTTP server
	if a.rpcServer != nil {
		if err := a.rpcServer.Stop(ctx); err != nil {
			a.logger.Error(context.Background(), "rpc server failed to stop",
				"error", fmt.Sprintf("%v", err))
		}
	}
	if a.httpServer != nil {
		if err := a.httpServer.Stop(ctx); err != nil {
			a.logger.Error(context.Background(), "http server failed to stop",
				"error", fmt.Sprintf("%v", err))
		}
	}

	a.logger.Info(context.Background(), "server exited")
}

// Shutdown gracefully shuts down all servers.
func (a *App) Shutdown(ctx context.Context) error {
	if a.rpcServer != nil {
		if err := a.rpcServer.Stop(ctx); err != nil {
			return err
		}
	}
	if a.httpServer != nil {
		if err := a.httpServer.Stop(ctx); err != nil {
			return err
		}
	}
	return nil
}

// HTTP routing proxy methods

func (a *App) GET(path string, handlers ...gin.HandlerFunc) {
	a.httpServer.GET(path, handlers...)
}

func (a *App) POST(path string, handlers ...gin.HandlerFunc) {
	a.httpServer.POST(path, handlers...)
}

func (a *App) PUT(path string, handlers ...gin.HandlerFunc) {
	a.httpServer.PUT(path, handlers...)
}

func (a *App) DELETE(path string, handlers ...gin.HandlerFunc) {
	a.httpServer.DELETE(path, handlers...)
}

func (a *App) PATCH(path string, handlers ...gin.HandlerFunc) {
	a.httpServer.PATCH(path, handlers...)
}

func (a *App) Group(path string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return a.httpServer.Group(path, handlers...)
}

func (a *App) Use(middleware ...gin.HandlerFunc) {
	a.httpServer.Use(middleware...)
}
