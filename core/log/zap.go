package log

import (
	"context"

	"go.uber.org/zap"
)

type zapLogger struct {
	sl *zap.SugaredLogger
}

// NewZap creates a Logger backed by zap.SugaredLogger.
func NewZap(sl *zap.SugaredLogger) Logger {
	return &zapLogger{sl: sl}
}

func (z *zapLogger) Info(ctx context.Context, msg string, args ...any) {
	z.sl.Infow(msg, toZapFields(ctx, args)...)
}

func (z *zapLogger) Error(ctx context.Context, msg string, args ...any) {
	z.sl.Errorw(msg, toZapFields(ctx, args)...)
}

func (z *zapLogger) Warn(ctx context.Context, msg string, args ...any) {
	z.sl.Warnw(msg, toZapFields(ctx, args)...)
}

func (z *zapLogger) Debug(ctx context.Context, msg string, args ...any) {
	z.sl.Debugw(msg, toZapFields(ctx, args)...)
}

func (z *zapLogger) With(args ...any) Logger {
	return &zapLogger{sl: z.sl.With(toZapFields(context.Background(), args)...)}
}

func toZapFields(_ context.Context, args []any) []any {
	fields := make([]any, 0, len(args))
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = "key"
		}
		fields = append(fields, key, args[i+1])
	}
	return fields
}

var _ Logger = (*zapLogger)(nil)
