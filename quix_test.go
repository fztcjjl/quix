package quix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fztcjjl/quix/core/logger"
	qhttp "github.com/fztcjjl/quix/core/transport/http/server"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewDefaultLogger(t *testing.T) {
	app := New()
	if app == nil {
		t.Fatal("New() returned nil")
	}
	if app.Logger() == nil {
		t.Fatal("default logger is nil")
	}
}

func TestWithLoggerInject(t *testing.T) {
	custom := &mockLogger{}
	app := New(WithLogger(custom))
	if app.Logger() != custom {
		t.Fatal("WithLogger did not inject the custom logger")
	}
}

func TestWithLoggerCompileCheck(t *testing.T) {
	var _ Option = WithLogger(&mockLogger{})
}

func TestWithConfigInject(t *testing.T) {
	custom := &mockConfig{}
	app := New(WithConfig(custom))
	if app.Config() != custom {
		t.Fatal("WithConfig did not inject the custom config")
	}
}

func TestWithConfigCompileCheck(t *testing.T) {
	var _ Option = WithConfig(&mockConfig{})
}

func TestWithHttpServerInject(t *testing.T) {
	s := qhttp.NewServer(qhttp.WithAddr(":9999"))
	app := New(WithHttpServer(s))
	if app.httpServer != s {
		t.Fatal("WithHttpServer did not inject the custom server")
	}
}

func TestWithHttpServerCompileCheck(t *testing.T) {
	var _ Option = WithHttpServer(qhttp.NewServer())
}

func TestWithRpcServerInject(t *testing.T) {
	rpc := &mockTransportServer{}
	app := New(WithRpcServer(rpc))
	if app.rpcServer != rpc {
		t.Fatal("WithRpcServer did not inject the rpc server")
	}
}

func TestNewDefaultHttpServer(t *testing.T) {
	app := New()
	if app.httpServer == nil {
		t.Fatal("default http server is nil")
	}
}

func TestAppRouteProxy(t *testing.T) {
	app := New()
	app.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	app.httpServer.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAppRunAndShutdown(t *testing.T) {
	s := qhttp.NewServer(qhttp.WithAddr("127.0.0.1:0"))
	app := New(WithHttpServer(s))
	app.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	done := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = app.Shutdown(ctx)
		close(done)
	}()

	go app.Run()

	<-done
}

type mockConfig struct{}

func (m *mockConfig) Get(key string) any                { return nil }
func (m *mockConfig) String(key string) string          { return "" }
func (m *mockConfig) Int(key string) int                { return 0 }
func (m *mockConfig) Bool(key string) bool              { return false }
func (m *mockConfig) Bind(key string, target any) error { return nil }

type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (m *mockLogger) Error(ctx context.Context, msg string, args ...any) {}
func (m *mockLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (m *mockLogger) Debug(ctx context.Context, msg string, args ...any) {}
func (m *mockLogger) With(args ...any) logger.Logger                     { return m }

type mockTransportServer struct{}

func (m *mockTransportServer) Start() error                   { return nil }
func (m *mockTransportServer) Stop(ctx context.Context) error { return nil }
