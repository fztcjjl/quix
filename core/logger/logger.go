package logger

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
