package log

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// ZerologOption configures the zerolog Logger adapter.
type ZerologOption func(*zerologLogger)

// WithCaller enables caller (file:line) output on each log entry.
// The caller field is computed by walking the call stack to find the first
// frame outside the runtime and core/log packages, so it works correctly
// for all calling patterns: log.Info(), log.FromContext(ctx).Info(), etc.
func WithCaller() ZerologOption {
	return func(z *zerologLogger) { z.callerEnabled = true }
}

type zerologLogger struct {
	al            *AtomicLevel
	l             zerolog.Logger
	callerEnabled bool
}

// NewZerolog creates a Logger backed by zerolog.
func NewZerolog(l zerolog.Logger, opts ...ZerologOption) Logger {
	z := &zerologLogger{al: NewAtomicLevel(LevelDebug), l: l}
	for _, opt := range opts {
		opt(z)
	}
	return z
}

func (z *zerologLogger) Trace(ctx context.Context, msg string, args ...any) {
	if !z.al.Enabled(LevelTrace) {
		return
	}
	e := z.l.Trace().Ctx(ctx)
	addFieldsToEvent(e, normalizeArgs(args))
	z.addCallerField(e)
	e.Msg(msg)
}

func (z *zerologLogger) Info(ctx context.Context, msg string, args ...any) {
	if !z.al.Enabled(LevelInfo) {
		return
	}
	e := z.l.Info().Ctx(ctx)
	addFieldsToEvent(e, normalizeArgs(args))
	z.addCallerField(e)
	e.Msg(msg)
}

func (z *zerologLogger) Error(ctx context.Context, msg string, args ...any) {
	if !z.al.Enabled(LevelError) {
		return
	}
	e := z.l.Error().Ctx(ctx)
	addFieldsToEvent(e, normalizeArgs(args))
	z.addCallerField(e)
	e.Msg(msg)
}

func (z *zerologLogger) Warn(ctx context.Context, msg string, args ...any) {
	if !z.al.Enabled(LevelWarn) {
		return
	}
	e := z.l.Warn().Ctx(ctx)
	addFieldsToEvent(e, normalizeArgs(args))
	z.addCallerField(e)
	e.Msg(msg)
}

func (z *zerologLogger) Debug(ctx context.Context, msg string, args ...any) {
	if !z.al.Enabled(LevelDebug) {
		return
	}
	e := z.l.Debug().Ctx(ctx)
	addFieldsToEvent(e, normalizeArgs(args))
	z.addCallerField(e)
	e.Msg(msg)
}

func (z *zerologLogger) Fatal(ctx context.Context, msg string, args ...any) {
	e := z.l.Error().Ctx(ctx)
	addFieldsToEvent(e, normalizeArgs(args))
	z.addCallerField(e)
	e.Msg(msg)
	os.Exit(1)
}

func (z *zerologLogger) With(args ...any) Logger {
	normalized := normalizeArgs(args)
	m := make(map[string]any, len(normalized)/2)
	for i := 0; i+1 < len(normalized); i += 2 {
		m[normalized[i].(string)] = normalized[i+1]
	}
	return &zerologLogger{al: z.al, l: z.l.With().Fields(m).Logger(), callerEnabled: z.callerEnabled}
}

func (z *zerologLogger) SetLevel(level Level) {
	z.al.SetLevel(level)
}

func (z *zerologLogger) Close() error {
	return nil
}

// addCallerField adds the caller file:line to the event if caller is enabled.
func (z *zerologLogger) addCallerField(e *zerolog.Event) {
	if !z.callerEnabled {
		return
	}
	// skip=2: addCallerField → zerologLogger.Info (core/log/, will be skipped by findCaller)
	if file, line, ok := findCaller(2); ok {
		e.Str(zerolog.CallerFieldName, zerolog.CallerMarshalFunc(0, file, line))
	}
}

// addFieldsToEvent adds key-value pairs to a zerolog.Event using type dispatch.
// This avoids allocating a map per log call.
func addFieldsToEvent(e *zerolog.Event, args []any) *zerolog.Event {
	for i := 0; i+1 < len(args); i += 2 {
		key := args[i].(string)
		switch v := args[i+1].(type) {
		case string:
			e.Str(key, v)
		case error:
			e.AnErr(key, v)
		case int:
			e.Int(key, v)
		case int64:
			e.Int64(key, v)
		case float64:
			e.Float64(key, v)
		case bool:
			e.Bool(key, v)
		case time.Duration:
			e.Dur(key, v)
		default:
			e.Interface(key, v)
		}
	}
	return e
}

var _ Logger = (*zerologLogger)(nil)
