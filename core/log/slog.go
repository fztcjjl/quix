package log

import (
	"context"
	"log/slog"
	"os"
)

type slogLogger struct {
	al *AtomicLevel
	sl *slog.Logger
}

// NewSlog creates a Logger backed by Go stdlib slog.
// If sl is nil, slog.Default() is used.
func NewSlog(sl ...*slog.Logger) Logger {
	if len(sl) > 0 && sl[0] != nil {
		return &slogLogger{al: NewAtomicLevel(LevelDebug), sl: sl[0]}
	}
	return &slogLogger{al: NewAtomicLevel(LevelDebug), sl: slog.Default()}
}

func (l *slogLogger) Info(ctx context.Context, msg string, args ...any) {
	if !l.al.Enabled(LevelInfo) {
		return
	}
	l.sl.InfoContext(ctx, msg, normalizeArgs(args)...)
}

func (l *slogLogger) Error(ctx context.Context, msg string, args ...any) {
	if !l.al.Enabled(LevelError) {
		return
	}
	l.sl.ErrorContext(ctx, msg, normalizeArgs(args)...)
}

func (l *slogLogger) Warn(ctx context.Context, msg string, args ...any) {
	if !l.al.Enabled(LevelWarn) {
		return
	}
	l.sl.WarnContext(ctx, msg, normalizeArgs(args)...)
}

func (l *slogLogger) Debug(ctx context.Context, msg string, args ...any) {
	if !l.al.Enabled(LevelDebug) {
		return
	}
	l.sl.DebugContext(ctx, msg, normalizeArgs(args)...)
}

func (l *slogLogger) Trace(ctx context.Context, msg string, args ...any) {
	if !l.al.Enabled(LevelTrace) {
		return
	}
	l.sl.Log(ctx, slog.Level(-8), msg, normalizeArgs(args)...)
}

func (l *slogLogger) Fatal(ctx context.Context, msg string, args ...any) {
	l.sl.ErrorContext(ctx, msg, normalizeArgs(args)...)
	os.Exit(1)
}

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{al: l.al, sl: l.sl.With(normalizeArgs(args)...)}
}

func (l *slogLogger) SetLevel(level Level) {
	l.al.SetLevel(level)
}

func (l *slogLogger) Close() error {
	return nil
}

var _ Logger = (*slogLogger)(nil)
