package log

import (
	"context"
	"fmt"
	"log/slog"
)

type slogLogger struct {
	sl *slog.Logger
}

// NewSlog creates a Logger backed by Go stdlib slog.
// If sl is nil, slog.Default() is used.
func NewSlog(sl ...*slog.Logger) Logger {
	if len(sl) > 0 && sl[0] != nil {
		return &slogLogger{sl: sl[0]}
	}
	return &slogLogger{sl: slog.Default()}
}

func (l *slogLogger) Info(ctx context.Context, msg string, args ...any) {
	l.sl.InfoContext(ctx, msg, toSlogArgs(args)...)
}

func (l *slogLogger) Error(ctx context.Context, msg string, args ...any) {
	l.sl.ErrorContext(ctx, msg, toSlogArgs(args)...)
}

func (l *slogLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.sl.WarnContext(ctx, msg, toSlogArgs(args)...)
}

func (l *slogLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.sl.DebugContext(ctx, msg, toSlogArgs(args)...)
}

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{sl: l.sl.With(toSlogArgs(args)...)}
}

// toSlogArgs converts key-value pairs to slog arguments.
// Non-string keys are wrapped with a "key" prefix. Odd trailing args are dropped.
func toSlogArgs(args []any) []any {
	for i := 0; i+1 < len(args); i += 2 {
		if _, ok := args[i].(string); !ok {
			args[i] = slog.String("key", fmt.Sprintf("%v", args[i]))
			args[i+1] = slog.Any("value", args[i+1])
		}
	}
	if len(args)%2 != 0 {
		args = args[:len(args)-1]
	}
	return args
}

var _ Logger = (*slogLogger)(nil)
