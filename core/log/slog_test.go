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

	// l2 should carry the "service" field
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

	// 3 args: odd trailing "key1" should be dropped, no panic
	l.Info(ctx, "odd args", "key1")
	_ = buf.String() // just verify no panic
}

func TestSlogOddArgsPairPreserved(t *testing.T) {
	sl, buf := newTestSlogLogger()
	l := NewSlog(sl)
	ctx := context.Background()

	// 3 args: first pair preserved, last dropped
	l.Info(ctx, "mixed", "key1", "val1", "key2")
	output := buf.String()
	if !strings.Contains(output, "key1") || !strings.Contains(output, "val1") {
		t.Errorf("expected key1=val1 in output, got: %s", output)
	}
	if strings.Contains(output, "key2") {
		t.Errorf("trailing odd arg key2 should be dropped, got: %s", output)
	}
}
