package log

import (
	"context"
	"os"

	"go.uber.org/zap"
)

type zapLogger struct {
	sl    *zap.SugaredLogger
	level Level // TODO: use atomic.Int32 for concurrent SetLevel safety
}

// NewZap creates a Logger backed by zap.SugaredLogger.
func NewZap(sl *zap.SugaredLogger) Logger {
	return &zapLogger{sl: sl}
}

func (z *zapLogger) Info(ctx context.Context, msg string, args ...any) {
	if z.level > LevelInfo {
		return
	}
	z.sl.Infow(msg, normalizeArgs(args)...)
}

func (z *zapLogger) Error(ctx context.Context, msg string, args ...any) {
	if z.level > LevelError {
		return
	}
	z.sl.Errorw(msg, normalizeArgs(args)...)
}

func (z *zapLogger) Warn(ctx context.Context, msg string, args ...any) {
	if z.level > LevelWarn {
		return
	}
	z.sl.Warnw(msg, normalizeArgs(args)...)
}

func (z *zapLogger) Debug(ctx context.Context, msg string, args ...any) {
	if z.level > LevelDebug {
		return
	}
	z.sl.Debugw(msg, normalizeArgs(args)...)
}

func (z *zapLogger) Fatal(ctx context.Context, msg string, args ...any) {
	z.sl.Errorw(msg, normalizeArgs(args)...)
	os.Exit(1)
}

func (z *zapLogger) With(args ...any) Logger {
	return &zapLogger{sl: z.sl.With(normalizeArgs(args)...), level: z.level}
}

func (z *zapLogger) SetLevel(level Level) {
	z.level = level
}

func (z *zapLogger) Close() error {
	defer func() { _ = recover() }()
	return z.sl.Sync()
}

var _ Logger = (*zapLogger)(nil)
