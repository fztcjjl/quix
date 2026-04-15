package log

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type writerLogger struct {
	sl    *slog.Logger
	level Level // TODO: use atomic.Int32 for concurrent SetLevel safety
}

// NewWriter creates a Logger backed by an io.Writer using slog.JSONHandler.
// This provides a zero-dependency Logger that outputs structured JSON.
func NewWriter(w io.Writer) Logger {
	sl := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return &writerLogger{sl: sl}
}

func (l *writerLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.level > LevelInfo {
		return
	}
	l.sl.InfoContext(ctx, msg, normalizeArgs(args)...)
}

func (l *writerLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.level > LevelError {
		return
	}
	l.sl.ErrorContext(ctx, msg, normalizeArgs(args)...)
}

func (l *writerLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.level > LevelWarn {
		return
	}
	l.sl.WarnContext(ctx, msg, normalizeArgs(args)...)
}

func (l *writerLogger) Debug(ctx context.Context, msg string, args ...any) {
	if l.level > LevelDebug {
		return
	}
	l.sl.DebugContext(ctx, msg, normalizeArgs(args)...)
}

func (l *writerLogger) Fatal(ctx context.Context, msg string, args ...any) {
	l.sl.ErrorContext(ctx, msg, normalizeArgs(args)...)
	os.Exit(1)
}

func (l *writerLogger) With(args ...any) Logger {
	return &writerLogger{sl: l.sl.With(normalizeArgs(args)...), level: l.level}
}

func (l *writerLogger) SetLevel(level Level) {
	l.level = level
}

func (l *writerLogger) Close() error {
	return nil
}

var _ Logger = (*writerLogger)(nil)
