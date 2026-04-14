package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestZapLogger() (*zap.SugaredLogger, *bytes.Buffer) {
	var buf bytes.Buffer
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		TimeKey:     "ts",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), zapcore.AddSync(&buf), zapcore.DebugLevel)
	sl := zap.New(core).Sugar()
	return sl, &buf
}

func TestZapLogLevels(t *testing.T) {
	sl, buf := newTestZapLogger()
	zl := NewZap(sl)
	ctx := context.Background()

	tests := []struct {
		name  string
		level string
		call  func()
	}{
		{"info", "INFO", func() { zl.Info(ctx, "info msg") }},
		{"error", "ERROR", func() { zl.Error(ctx, "error msg") }},
		{"warn", "WARN", func() { zl.Warn(ctx, "warn msg") }},
		{"debug", "DEBUG", func() { zl.Debug(ctx, "debug msg") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.call()
			if !strings.Contains(buf.String(), tt.level) {
				t.Errorf("expected level %q in output, got: %s", tt.level, buf.String())
			}
		})
	}
}

func TestZapWithFields(t *testing.T) {
	sl, buf := newTestZapLogger()
	zl := NewZap(sl)
	ctx := context.Background()

	zl.Info(ctx, "test", "method", "GET", "path", "/users")
	output := buf.String()
	if !strings.Contains(output, "method") || !strings.Contains(output, "GET") {
		t.Errorf("expected fields in output, got: %s", output)
	}
}

func TestZapWithReturnsNewLogger(t *testing.T) {
	sl, buf := newTestZapLogger()
	zl := NewZap(sl)
	ctx := context.Background()

	zl2 := zl.With("service", "quix")
	if zl2 == nil {
		t.Fatal("With() returned nil")
	}

	zl2.Info(ctx, "test")
	output := buf.String()
	if !strings.Contains(output, "service") || !strings.Contains(output, "quix") {
		t.Errorf("expected 'service=quix' in output, got: %s", output)
	}
}

func TestZapOddArgsDropped(t *testing.T) {
	sl, _ := newTestZapLogger()
	zl := NewZap(sl)
	ctx := context.Background()

	// 3 args: odd trailing should be dropped, no panic
	zl.Info(ctx, "odd args", "key1")
}

func TestZapNonStringKey(t *testing.T) {
	sl, buf := newTestZapLogger()
	zl := NewZap(sl)
	ctx := context.Background()

	zl.Info(ctx, "test", 123, "value")
	output := buf.String()
	if !strings.Contains(output, "key_0") || !strings.Contains(output, "value") {
		t.Errorf("expected key_0=value in output, got: %s", output)
	}
}

func TestZapMultipleNonStringKeys(t *testing.T) {
	sl, buf := newTestZapLogger()
	zl := NewZap(sl)
	ctx := context.Background()

	zl.Info(ctx, "test", 123, "a", 456, "b")
	output := buf.String()
	if !strings.Contains(output, "key_0") || !strings.Contains(output, "a") {
		t.Errorf("expected key_0=a in output, got: %s", output)
	}
	if !strings.Contains(output, "key_1") || !strings.Contains(output, "b") {
		t.Errorf("expected key_1=b in output, got: %s", output)
	}
}

func TestZapSetLevel(t *testing.T) {
	sl, buf := newTestZapLogger()
	zl := NewZap(sl)
	ctx := context.Background()

	zl.SetLevel(LevelError)

	buf.Reset()
	zl.Info(ctx, "info msg")
	if buf.String() != "" {
		t.Errorf("info should be suppressed at LevelError, got: %s", buf.String())
	}

	buf.Reset()
	zl.Error(ctx, "error msg")
	if !strings.Contains(buf.String(), "error msg") {
		t.Errorf("error should be emitted at LevelError, got: %s", buf.String())
	}
}

func TestZapClose(t *testing.T) {
	sl, _ := newTestZapLogger()
	zl := NewZap(sl)
	// Sync on a closed core may error, but Close should not panic
	_ = zl.Close()
}
