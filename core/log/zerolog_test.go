package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func newTestZerologLogger() (zerolog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	l := zerolog.New(&buf).With().Timestamp().Logger()
	return l, &buf
}

func TestZerologLogLevels(t *testing.T) {
	l, buf := newTestZerologLogger()
	zl := NewZerolog(l)
	ctx := context.Background()

	tests := []struct {
		name  string
		level string
		call  func()
	}{
		{"info", "info", func() { zl.Info(ctx, "info msg") }},
		{"error", "error", func() { zl.Error(ctx, "error msg") }},
		{"warn", "warn", func() { zl.Warn(ctx, "warn msg") }},
		{"debug", "debug", func() { zl.Debug(ctx, "debug msg") }},
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

func TestZerologWithFields(t *testing.T) {
	l, buf := newTestZerologLogger()
	zl := NewZerolog(l)
	ctx := context.Background()

	zl.Info(ctx, "test", "method", "GET", "path", "/users")
	output := buf.String()
	if !strings.Contains(output, "method") || !strings.Contains(output, "GET") {
		t.Errorf("expected fields in output, got: %s", output)
	}
}

func TestZerologWithReturnsNewLogger(t *testing.T) {
	l, buf := newTestZerologLogger()
	zl := NewZerolog(l)
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

func TestZerologOddArgsDropped(t *testing.T) {
	l, _ := newTestZerologLogger()
	zl := NewZerolog(l)
	ctx := context.Background()

	// 3 args: odd trailing should be dropped, no panic
	zl.Info(ctx, "odd args", "key1")
}
