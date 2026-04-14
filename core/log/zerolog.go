package log

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

type zerologLogger struct {
	l     zerolog.Logger
	level Level
}

// NewZerolog creates a Logger backed by zerolog.
func NewZerolog(l zerolog.Logger) Logger {
	return &zerologLogger{l: l}
}

func (z *zerologLogger) Info(ctx context.Context, msg string, args ...any) {
	if z.level > LevelInfo {
		return
	}
	z.l.Info().Ctx(ctx).Fields(argsToMap(normalizeArgs(args))).Msg(msg)
}

func (z *zerologLogger) Error(ctx context.Context, msg string, args ...any) {
	if z.level > LevelError {
		return
	}
	z.l.Error().Ctx(ctx).Fields(argsToMap(normalizeArgs(args))).Msg(msg)
}

func (z *zerologLogger) Warn(ctx context.Context, msg string, args ...any) {
	if z.level > LevelWarn {
		return
	}
	z.l.Warn().Ctx(ctx).Fields(argsToMap(normalizeArgs(args))).Msg(msg)
}

func (z *zerologLogger) Debug(ctx context.Context, msg string, args ...any) {
	if z.level > LevelDebug {
		return
	}
	z.l.Debug().Ctx(ctx).Fields(argsToMap(normalizeArgs(args))).Msg(msg)
}

func (z *zerologLogger) Fatal(ctx context.Context, msg string, args ...any) {
	z.l.Error().Ctx(ctx).Fields(argsToMap(normalizeArgs(args))).Msg(msg)
	os.Exit(1)
}

func (z *zerologLogger) With(args ...any) Logger {
	return &zerologLogger{l: z.l.With().Fields(argsToMap(normalizeArgs(args))).Logger(), level: z.level}
}

func (z *zerologLogger) SetLevel(level Level) {
	z.level = level
}

func (z *zerologLogger) Close() error {
	return nil
}

// argsToMap converts a flat key-value slice to a map.
// Precondition: args must be even length with string keys (from normalizeArgs).
func argsToMap(args []any) map[string]any {
	m := make(map[string]any, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		m[args[i].(string)] = args[i+1]
	}
	return m
}

var _ Logger = (*zerologLogger)(nil)
