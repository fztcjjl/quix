package log

import (
	"context"
	"os"

	"go.uber.org/zap"
)

type zapLogger struct {
	al *AtomicLevel
	sl *zap.SugaredLogger
}

// NewZap creates a Logger backed by zap.SugaredLogger.
func NewZap(sl *zap.SugaredLogger) Logger {
	return &zapLogger{al: NewAtomicLevel(LevelDebug), sl: sl}
}

func (z *zapLogger) Info(ctx context.Context, msg string, args ...any) {
	if !z.al.Enabled(LevelInfo) {
		return
	}
	z.sl.Infow(msg, normalizeArgs(args)...)
}

func (z *zapLogger) Error(ctx context.Context, msg string, args ...any) {
	if !z.al.Enabled(LevelError) {
		return
	}
	z.sl.Errorw(msg, normalizeArgs(args)...)
}

func (z *zapLogger) Warn(ctx context.Context, msg string, args ...any) {
	if !z.al.Enabled(LevelWarn) {
		return
	}
	z.sl.Warnw(msg, normalizeArgs(args)...)
}

func (z *zapLogger) Debug(ctx context.Context, msg string, args ...any) {
	if !z.al.Enabled(LevelDebug) {
		return
	}
	z.sl.Debugw(msg, normalizeArgs(args)...)
}

func (z *zapLogger) Trace(ctx context.Context, msg string, args ...any) {
	if !z.al.Enabled(LevelTrace) {
		return
	}
	z.sl.Debugw(msg, normalizeArgs(args)...)
}

func (z *zapLogger) Fatal(ctx context.Context, msg string, args ...any) {
	z.sl.Errorw(msg, normalizeArgs(args)...)
	os.Exit(1)
}

func (z *zapLogger) With(args ...any) Logger {
	return &zapLogger{al: z.al, sl: z.sl.With(normalizeArgs(args)...)}
}

func (z *zapLogger) SetLevel(level Level) {
	z.al.SetLevel(level)
}

func (z *zapLogger) Close() error {
	defer func() { _ = recover() }()
	return z.sl.Sync()
}

var _ Logger = (*zapLogger)(nil)
