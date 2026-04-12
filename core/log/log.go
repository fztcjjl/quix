package log

import "context"

// Logger is the unified logging interface for quix framework.
// All framework components use this interface for log output.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
}

// noopLogger is a no-operation Logger that does nothing.
type noopLogger struct{}

func (n *noopLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (n *noopLogger) Error(_ context.Context, _ string, _ ...any) {}
func (n *noopLogger) Warn(_ context.Context, _ string, _ ...any)  {}
func (n *noopLogger) Debug(_ context.Context, _ string, _ ...any) {}
func (n *noopLogger) With(_ ...any) Logger                        { return n }

var _ Logger = (*noopLogger)(nil)

// defaultLogger is the global default Logger.
var defaultLogger Logger = &noopLogger{}

// SetDefault sets the global default Logger.
func SetDefault(l Logger) {
	defaultLogger = l
}

// Info logs an info message using the global default Logger.
func Info(ctx context.Context, msg string, args ...any) {
	defaultLogger.Info(ctx, msg, args...)
}

// Error logs an error message using the global default Logger.
func Error(ctx context.Context, msg string, args ...any) {
	defaultLogger.Error(ctx, msg, args...)
}

// Warn logs a warning message using the global default Logger.
func Warn(ctx context.Context, msg string, args ...any) {
	defaultLogger.Warn(ctx, msg, args...)
}

// Debug logs a debug message using the global default Logger.
func Debug(ctx context.Context, msg string, args ...any) {
	defaultLogger.Debug(ctx, msg, args...)
}

// With creates a child Logger from the global default Logger with additional fields.
func With(args ...any) Logger {
	return defaultLogger.With(args...)
}
