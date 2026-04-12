package logger

import (
	"context"
	"testing"
)

// mockLogger is a test implementation that verifies the Logger interface
// is correctly satisfied at compile time.
type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (m *mockLogger) Error(ctx context.Context, msg string, args ...any) {}
func (m *mockLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (m *mockLogger) Debug(ctx context.Context, msg string, args ...any) {}
func (m *mockLogger) With(args ...any) Logger                            { return m }

// Compile-time check: mockLogger implements Logger
var _ Logger = (*mockLogger)(nil)

func TestMockLoggerSatisfiesInterface(t *testing.T) {
	var l Logger = &mockLogger{}
	ctx := context.Background()

	// Verify all methods are callable without panic
	l.Info(ctx, "info")
	l.Error(ctx, "error")
	l.Warn(ctx, "warn")
	l.Debug(ctx, "debug")

	child := l.With("key", "value")
	if child == nil {
		t.Fatal("With() returned nil")
	}
}
