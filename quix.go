package quix

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fztcjjl/quix/core/config"
	"github.com/fztcjjl/quix/core/log"
	"github.com/fztcjjl/quix/core/telemetry"
	"github.com/fztcjjl/quix/core/transport"
	qhttp "github.com/fztcjjl/quix/core/transport/http/server"
	"github.com/fztcjjl/quix/core/transport/http/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// shutdownTimeout is the maximum duration for graceful shutdown.
const shutdownTimeout = 5 * time.Second

// Environment represents the application deployment environment.
type Environment string

const (
	// EnvDev is the local development environment.
	EnvDev Environment = "dev"
	// EnvTest is the CI/testing environment.
	EnvTest Environment = "test"
	// EnvStaging is the pre-production environment.
	EnvStaging Environment = "staging"
	// EnvProd is the production environment.
	EnvProd Environment = "prod"
)

// resolveEnv reads QUIX_ENV from the environment.
// Defaults to EnvDev. Unknown values are treated as EnvProd (safe default).
func resolveEnv() Environment {
	env := os.Getenv("QUIX_ENV")
	if env == "" {
		return EnvDev
	}
	switch Environment(env) {
	case EnvDev, EnvTest, EnvStaging, EnvProd:
		return Environment(env)
	default:
		return EnvProd
	}
}

// ginModeForEnv returns the Gin mode for the given environment.
func ginModeForEnv(env Environment) string {
	switch env {
	case EnvDev:
		return gin.DebugMode
	case EnvTest:
		return gin.TestMode
	default:
		return gin.ReleaseMode
	}
}

// App is the core framework application.
type App struct {
	httpServer             *qhttp.Server
	rpcServer              transport.Server
	logger                 log.Logger
	config                 config.Config
	env                    Environment
	defaultMiddleware      bool
	telemetryOpts          []telemetry.Option
	telemetryShutdown      func(context.Context) error
	telemetryServiceName   string
	telemetryTracesEnabled bool
	setupFuncs             []func(*App) error
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
// If no logger is provided, zerolog is used by default with format driven by QUIX_ENV.
// If no config is provided, koanf with environment variables is used by default.
func New(opts ...Option) *App {
	env := resolveEnv()
	// Default logger format driven by environment
	var logOutput io.Writer = os.Stderr
	if env == EnvDev {
		logOutput = zerolog.ConsoleWriter{Out: os.Stderr}
	}
	defaultLog := log.NewZerolog(zerolog.New(logOutput).With().Timestamp().Logger())
	defaultCfg, _ := config.NewKoanf()
	app := &App{
		logger: defaultLog,
		config: defaultCfg,
		env:    env,
	}
	log.SetDefault(defaultLog)
	// Auto-set Gin mode from environment; user options can override via WithGinMode
	gin.SetMode(ginModeForEnv(app.env))
	// Apply user options
	for _, opt := range opts {
		opt(app)
	}
	// Initialize telemetry if WithTelemetry was provided
	if len(app.telemetryOpts) > 0 {
		shutdown, err := telemetry.Init(context.Background(), app.telemetryOpts...)
		if err != nil {
			app.logger.Warn(context.Background(), "telemetry init failed, running without telemetry", "error", err)
		} else {
			app.telemetryShutdown = shutdown
			// Enable trace_id in logging middleware
			middleware.ExtractTraceID = telemetry.ExtractTraceID
		}
		// Resolve telemetry config for Server middleware injection
		cfg := &telemetry.Config{
			ServiceName:   "unknown_service",
			TracesEnabled: err == nil,
		}
		for _, opt := range app.telemetryOpts {
			opt(cfg)
		}
		app.telemetryServiceName = cfg.ServiceName
		app.telemetryTracesEnabled = cfg.TracesEnabled
	}
	// Config-driven server creation:
	// - If http is configured (http.addr or http.port), start HTTP server
	// - If rpc is configured (rpc.addr), start RPC server
	// - If neither is configured, default to HTTP server on :8080
	hasHttpConfig := app.config.String("http.addr") != "" || app.config.Int("http.port") != 0
	hasRpcConfig := app.config.String("rpc.addr") != ""

	if app.httpServer == nil && (hasHttpConfig || !hasRpcConfig) {
		var serverOpts []qhttp.Option
		serverOpts = append(serverOpts, qhttp.WithAddr(resolveHttpAddr(app.config)))
		if len(app.telemetryOpts) > 0 {
			serverOpts = append(serverOpts,
				qhttp.WithTelemetryServiceName(app.telemetryServiceName),
				qhttp.WithTelemetryTracesEnabled(app.telemetryTracesEnabled),
			)
		}
		app.httpServer = qhttp.NewServer(serverOpts...)
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
	// Output startup info log
	telemetryStatus := "disabled"
	if a.telemetryShutdown != nil {
		telemetryStatus = "enabled"
	}
	a.logger.Info(context.Background(), "starting server",
		"addr", resolveHttpAddr(a.config),
		"env", string(a.env),
		"gin_mode", gin.Mode(),
		"telemetry", telemetryStatus)

	// Execute WithSetup callbacks
	for _, fn := range a.setupFuncs {
		if err := fn(a); err != nil {
			a.logger.Error(context.Background(), "setup callback failed",
				"error", err)
			os.Exit(1)
		}
	}

	// Start servers
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

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info(context.Background(), "shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.Shutdown(ctx); err != nil {
		a.logger.Error(context.Background(), "shutdown failed",
			"error", fmt.Sprintf("%v", err))
	}

	a.logger.Info(context.Background(), "server exited")
}

// Shutdown gracefully shuts down all servers.
func (a *App) Shutdown(ctx context.Context) error {
	var errs []error

	if a.rpcServer != nil {
		a.logger.Info(ctx, "stopping rpc server...")
		if err := a.rpcServer.Stop(ctx); err != nil {
			a.logger.Error(ctx, "rpc server failed to stop", "error", err)
			errs = append(errs, err)
		}
	}
	if a.httpServer != nil {
		a.logger.Info(ctx, "stopping http server...")
		if err := a.httpServer.Stop(ctx); err != nil {
			a.logger.Error(ctx, "http server failed to stop", "error", err)
			errs = append(errs, err)
		}
	}
	// Flush telemetry providers after servers stop
	if a.telemetryShutdown != nil {
		a.logger.Info(ctx, "flushing telemetry...")
		if err := a.telemetryShutdown(ctx); err != nil {
			a.logger.Warn(ctx, "telemetry flush failed", "error", err)
			errs = append(errs, err)
		}
	}
	// Close logger last
	if a.logger != nil {
		a.logger.Info(ctx, "closing logger...")
		if err := a.logger.Close(); err != nil {
			a.logger.Error(ctx, "logger close failed", "error", err)
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
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
