package log

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func newTestSlogLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	sl := slog.New(h)
	return sl, &buf
}

func TestNewSlogDefault(t *testing.T) {
	l := NewSlog()
	if l == nil {
		t.Fatal("NewSlog() returned nil")
	}
}

func TestNewSlogWithLogger(t *testing.T) {
	sl, _ := newTestSlogLogger()
	l := NewSlog(sl)
	if l == nil {
		t.Fatal("NewSlog(sl) returned nil")
	}
}

func TestSlogLogLevels(t *testing.T) {
	sl, buf := newTestSlogLogger()
	l := NewSlog(sl)
	ctx := context.Background()

	tests := []struct {
		name  string
		level string
		call  func()
	}{
		{"info", "INFO", func() { l.Info(ctx, "info msg") }},
		{"error", "ERROR", func() { l.Error(ctx, "error msg") }},
		{"warn", "WARN", func() { l.Warn(ctx, "warn msg") }},
		{"debug", "DEBUG", func() { l.Debug(ctx, "debug msg") }},
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

func TestSlogWithFields(t *testing.T) {
	sl, buf := newTestSlogLogger()
	l := NewSlog(sl)
	ctx := context.Background()

	l.Info(ctx, "test", "method", "GET", "path", "/users")
	output := buf.String()
	if !strings.Contains(output, "method") || !strings.Contains(output, "GET") {
		t.Errorf("expected fields in output, got: %s", output)
	}
	if !strings.Contains(output, "path") || !strings.Contains(output, "/users") {
		t.Errorf("expected fields in output, got: %s", output)
	}
}

func TestSlogWithReturnsNewLogger(t *testing.T) {
	sl, buf := newTestSlogLogger()
	l := NewSlog(sl)
	ctx := context.Background()

	l2 := l.With("service", "quix")
	if l2 == nil {
		t.Fatal("With() returned nil")
	}

	l2.Info(ctx, "test")
	output := buf.String()
	if !strings.Contains(output, "service") || !strings.Contains(output, "quix") {
		t.Errorf("expected 'service=quix' in output, got: %s", output)
	}
}

func TestSlogOddArgsDropped(t *testing.T) {
	sl, buf := newTestSlogLogger()
	l := NewSlog(sl)
	ctx := context.Background()

	buf.Reset()
	l.Info(ctx, "odd args", "key1")
	// Odd trailing arg should be dropped, no panic
	output := buf.String()
	if strings.Contains(output, "key1") {
		t.Errorf("trailing odd arg key1 should be dropped, got: %s", output)
	}
}

func TestSlogOddArgsPairPreserved(t *testing.T) {
	sl, buf := newTestSlogLogger()
	l := NewSlog(sl)
	ctx := context.Background()

	l.Info(ctx, "mixed", "key1", "val1", "key2")
	output := buf.String()
	if !strings.Contains(output, "key1") || !strings.Contains(output, "val1") {
		t.Errorf("expected key1=val1 in output, got: %s", output)
	}
	if strings.Contains(output, "key2") {
		t.Errorf("trailing odd arg key2 should be dropped, got: %s", output)
	}
}

func TestSlogNonStringKey(t *testing.T) {
	sl, buf := newTestSlogLogger()
	l := NewSlog(sl)
	ctx := context.Background()

	l.Info(ctx, "test", 123, "value")
	output := buf.String()
	if !strings.Contains(output, "key_0") || !strings.Contains(output, "value") {
		t.Errorf("expected key_0=value in output, got: %s", output)
	}
}

func TestSlogMultipleNonStringKeys(t *testing.T) {
	sl, buf := newTestSlogLogger()
	l := NewSlog(sl)
	ctx := context.Background()

	l.Info(ctx, "test", 123, "a", 456, "b")
	output := buf.String()
	if !strings.Contains(output, "key_0") || !strings.Contains(output, "a") {
		t.Errorf("expected key_0=a in output, got: %s", output)
	}
	if !strings.Contains(output, "key_1") || !strings.Contains(output, "b") {
		t.Errorf("expected key_1=b in output, got: %s", output)
	}
}

func TestNormalizeArgsNoMutation(t *testing.T) {
	orig := []any{"key1", "val1", "key2", "val2"}
	backup := make([]any, len(orig))
	copy(backup, orig)

	_ = normalizeArgs(orig)

	for i, v := range orig {
		if v != backup[i] {
			t.Errorf("normalizeArgs mutated args[%d]: got %v, want %v", i, orig[i], backup[i])
		}
	}
}

func TestNormalizeArgsFastPath(t *testing.T) {
	args := []any{"key1", "val1", "key2", "val2"}
	result := normalizeArgs(args)
	if &result[0] != &args[0] {
		t.Error("expected fast path to return same slice, no allocation")
	}
}

func TestNormalizeArgsNonStringKey(t *testing.T) {
	result := normalizeArgs([]any{123, "a", 456, "b"})
	if result[0].(string) != "key_0" {
		t.Errorf("expected key_0, got %v", result[0])
	}
	if result[2].(string) != "key_1" {
		t.Errorf("expected key_1, got %v", result[2])
	}
}

func TestNormalizeArgsOddDropped(t *testing.T) {
	result := normalizeArgs([]any{"key1", "val1", "key2"})
	if len(result) != 2 {
		t.Errorf("expected 2 args, got %d", len(result))
	}
}

func TestNormalizeArgsEmpty(t *testing.T) {
	result := normalizeArgs([]any{})
	if len(result) != 0 {
		t.Errorf("expected 0 args, got %d", len(result))
	}
}

func TestNormalizeArgsSingleOddDropped(t *testing.T) {
	result := normalizeArgs([]any{"key1"})
	if len(result) != 0 {
		t.Errorf("expected 0 args (odd single dropped), got %d", len(result))
	}
}

func TestSlogSetLevel(t *testing.T) {
	sl, buf := newTestSlogLogger()
	l := NewSlog(sl)
	ctx := context.Background()

	l.SetLevel(LevelError)

	buf.Reset()
	l.Debug(ctx, "debug msg")
	if buf.String() != "" {
		t.Errorf("debug should be suppressed at LevelError, got: %s", buf.String())
	}

	buf.Reset()
	l.Info(ctx, "info msg")
	if buf.String() != "" {
		t.Errorf("info should be suppressed at LevelError, got: %s", buf.String())
	}

	buf.Reset()
	l.Error(ctx, "error msg")
	if !strings.Contains(buf.String(), "error msg") {
		t.Errorf("error should be emitted at LevelError, got: %s", buf.String())
	}
}

func TestSlogClose(t *testing.T) {
	sl, _ := newTestSlogLogger()
	l := NewSlog(sl)
	if err := l.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
}
