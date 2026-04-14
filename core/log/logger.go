package log

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Level represents the log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger is the unified logging interface for quix framework.
// All framework components use this interface for log output.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
	Fatal(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
	SetLevel(level Level)
	Close() error
}

// defaultLogger is the global default Logger, protected by atomic.Pointer.
var defaultLogger atomic.Pointer[Logger]

func init() {
	l := NewSlog()
	defaultLogger.Store(&l)
}

// SetDefault sets the global default Logger.
func SetDefault(l Logger) {
	defaultLogger.Store(&l)
}

// Default returns the global default Logger.
func Default() Logger {
	if p := defaultLogger.Load(); p != nil {
		return *p
	}
	return NewSlog()
}

// Info logs an info message using the global default Logger.
func Info(ctx context.Context, msg string, args ...any) {
	Default().Info(ctx, msg, args...)
}

// Error logs an error message using the global default Logger.
func Error(ctx context.Context, msg string, args ...any) {
	Default().Error(ctx, msg, args...)
}

// Warn logs a warning message using the global default Logger.
func Warn(ctx context.Context, msg string, args ...any) {
	Default().Warn(ctx, msg, args...)
}

// Debug logs a debug message using the global default Logger.
func Debug(ctx context.Context, msg string, args ...any) {
	Default().Debug(ctx, msg, args...)
}

// Fatal logs a fatal message using the global default Logger and exits.
func Fatal(ctx context.Context, msg string, args ...any) {
	Default().Fatal(ctx, msg, args...)
}

// With creates a child Logger from the global default Logger with additional fields.
func With(args ...any) Logger {
	return Default().With(args...)
}

// SetLevel sets the log level of the global default Logger.
func SetLevel(level Level) {
	Default().SetLevel(level)
}

// normalizeArgs standardizes key-value pairs for all adapters.
// Non-string keys are converted to "key_0", "key_1", etc.
// Odd trailing args are silently dropped.
// Fast path: if all keys are strings and arg count is even, returns args without allocation.
func normalizeArgs(args []any) []any {
	if len(args) == 0 {
		return args
	}
	if len(args)%2 != 0 {
		args = args[:len(args)-1]
		if len(args) == 0 {
			return args
		}
	}
	// Fast path: check if all keys are strings
	needConvert := false
	for i := 0; i < len(args); i += 2 {
		if _, ok := args[i].(string); !ok {
			needConvert = true
			break
		}
	}
	if !needConvert {
		return args
	}
	// Slow path: copy and convert non-string keys
	out := make([]any, len(args))
	copy(out, args)
	nonStrIdx := 0
	for i := 0; i+1 < len(out); i += 2 {
		if _, ok := out[i].(string); !ok {
			out[i] = fmt.Sprintf("key_%d", nonStrIdx)
			nonStrIdx++
		}
	}
	return out
}
