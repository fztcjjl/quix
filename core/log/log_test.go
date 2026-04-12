package log

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

func TestNoopLoggerDoesNotPanic(t *testing.T) {
	n := &noopLogger{}
	ctx := context.Background()

	n.Info(ctx, "info")
	n.Error(ctx, "error")
	n.Warn(ctx, "warn")
	n.Debug(ctx, "debug")

	child := n.With("key", "value")
	if child == nil {
		t.Fatal("noopLogger.With() returned nil")
	}
}

func TestSetDefault(t *testing.T) {
	// Save and restore original
	orig := defaultLogger
	defer func() { defaultLogger = orig }()

	custom := &mockLogger{}
	SetDefault(custom)

	if defaultLogger != custom {
		t.Fatal("SetDefault did not update defaultLogger")
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	// Save and restore original
	orig := defaultLogger
	defer func() { defaultLogger = orig }()

	// Use noopLogger — should not panic
	SetDefault(&noopLogger{})
	ctx := context.Background()

	Info(ctx, "info msg")
	Error(ctx, "error msg")
	Warn(ctx, "warn msg")
	Debug(ctx, "debug msg")

	child := With("key", "value")
	if child == nil {
		t.Fatal("global With() returned nil")
	}
}
