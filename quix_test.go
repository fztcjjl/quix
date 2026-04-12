package quix

import (
	"context"
	"testing"

	"github.com/fztcjjl/quix/core/logger"
)

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
	// Ensure any Logger implementation satisfies WithLogger
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
