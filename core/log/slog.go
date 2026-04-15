package log

import (
	"context"
	"log/slog"
	"os"
)

type slogLogger struct {
	sl    *slog.Logger
	level Level // TODO: use atomic.Int32 for concurrent SetLevel safety
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
	if l.level > LevelInfo {
		return
	}
	l.sl.InfoContext(ctx, msg, normalizeArgs(args)...)
}

func (l *slogLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.level > LevelError {
		return
	}
	l.sl.ErrorContext(ctx, msg, normalizeArgs(args)...)
}

func (l *slogLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.level > LevelWarn {
		return
	}
	l.sl.WarnContext(ctx, msg, normalizeArgs(args)...)
}

func (l *slogLogger) Debug(ctx context.Context, msg string, args ...any) {
	if l.level > LevelDebug {
		return
	}
	l.sl.DebugContext(ctx, msg, normalizeArgs(args)...)
}

func (l *slogLogger) Fatal(ctx context.Context, msg string, args ...any) {
	l.sl.ErrorContext(ctx, msg, normalizeArgs(args)...)
	os.Exit(1)
}

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{sl: l.sl.With(normalizeArgs(args)...), level: l.level}
}

func (l *slogLogger) SetLevel(level Level) {
	l.level = level
}

func (l *slogLogger) Close() error {
	return nil
}

var _ Logger = (*slogLogger)(nil)
