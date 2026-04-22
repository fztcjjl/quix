package log

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
)

// AtomicLevel provides concurrent-safe log level management.
// Multiple Logger instances can share the same AtomicLevel to synchronize level changes.
type AtomicLevel struct {
	level atomic.Int32
}

// NewAtomicLevel creates an AtomicLevel with the given initial level.
func NewAtomicLevel(l Level) *AtomicLevel {
	al := &AtomicLevel{}
	//nolint:gosec // Level values are small bounded integers (-1 to 3)
	al.level.Store(int32(l))
	return al
}

// Level returns the current log level.
func (al *AtomicLevel) Level() Level {
	return Level(al.level.Load())
}

// SetLevel atomically sets the log level.
func (al *AtomicLevel) SetLevel(l Level) {
	//nolint:gosec // Level values are small bounded integers (-1 to 3)
	al.level.Store(int32(l))
}

func (al *AtomicLevel) Enabled(l Level) bool {
	return Level(al.level.Load()) <= l
}

// Level represents the log level.
type Level int

const (
	LevelTrace Level = -1
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLevel parses a case-insensitive level string.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "trace":
		return LevelTrace, nil
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return Level(0), fmt.Errorf("unknown log level: %q", s)
	}
}

// Logger is the unified logging interface for quix framework.
// All framework components use this interface for log output.
type Logger interface {
	Trace(ctx context.Context, msg string, args ...any)
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

// Trace logs a trace message using the global default Logger.
func Trace(ctx context.Context, msg string, args ...any) {
	Default().Trace(ctx, msg, args...)
}

// With creates a child Logger from the global default Logger with additional fields.
func With(args ...any) Logger {
	return Default().With(args...)
}

// SetLevel sets the log level of the global default Logger.
func SetLevel(level Level) {
	Default().SetLevel(level)
}

// Close closes the global default Logger, flushing any buffered output.
func Close() error {
	return Default().Close()
}

// contextKey is the unexported type used for context keys to avoid collisions.
type contextKey struct{}

// NewContext stores a Logger in the context.
func NewContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext extracts a Logger from the context.
// Returns the global default Logger if no Logger is stored.
func FromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(contextKey{}).(Logger); ok {
		return l
	}
	return Default()
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
