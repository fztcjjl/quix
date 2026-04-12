package log

import (
	"context"

	"github.com/rs/zerolog"
)

type zerologLogger struct {
	l zerolog.Logger
}

// NewZerolog creates a Logger backed by zerolog.
func NewZerolog(l zerolog.Logger) Logger {
	return &zerologLogger{l: l}
}

func (z *zerologLogger) Info(ctx context.Context, msg string, args ...any) {
	z.l.Info().Ctx(ctx).Fields(toZerologFields(args)).Msg(msg)
}

func (z *zerologLogger) Error(ctx context.Context, msg string, args ...any) {
	z.l.Error().Ctx(ctx).Fields(toZerologFields(args)).Msg(msg)
}

func (z *zerologLogger) Warn(ctx context.Context, msg string, args ...any) {
	z.l.Warn().Ctx(ctx).Fields(toZerologFields(args)).Msg(msg)
}

func (z *zerologLogger) Debug(ctx context.Context, msg string, args ...any) {
	z.l.Debug().Ctx(ctx).Fields(toZerologFields(args)).Msg(msg)
}

func (z *zerologLogger) With(args ...any) Logger {
	return &zerologLogger{l: z.l.With().Fields(toZerologFields(args)).Logger()}
}

func toZerologFields(args []any) map[string]any {
	m := make(map[string]any, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = "key"
		}
		m[key] = args[i+1]
	}
	return m
}

var _ Logger = (*zerologLogger)(nil)
