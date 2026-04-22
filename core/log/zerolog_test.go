package log

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestZerologNonStringKey(t *testing.T) {
	l, buf := newTestZerologLogger()
	zl := NewZerolog(l)
	ctx := context.Background()

	zl.Info(ctx, "test", 123, "value")
	output := buf.String()
	if !strings.Contains(output, "key_0") || !strings.Contains(output, "value") {
		t.Errorf("expected key_0=value in output, got: %s", output)
	}
}

func TestZerologMultipleNonStringKeys(t *testing.T) {
	l, buf := newTestZerologLogger()
	zl := NewZerolog(l)
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

func TestZerologSetLevel(t *testing.T) {
	l, buf := newTestZerologLogger()
	zl := NewZerolog(l)
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

func TestZerologClose(t *testing.T) {
	l, _ := newTestZerologLogger()
	zl := NewZerolog(l)
	if err := zl.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
}

func TestZerologTimestampField(t *testing.T) {
	var buf bytes.Buffer
	l := zerolog.New(&buf).With().Timestamp().Logger()
	zl := NewZerolog(l)

	zl.Info(context.Background(), "ts test")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if _, ok := m["time"]; !ok {
		t.Errorf("expected 'time' field in output, got: %s", buf.String())
	}
}

func TestZerologCallerField(t *testing.T) {
	var buf bytes.Buffer
	l := zerolog.New(&buf).With().Timestamp().CallerWithSkipFrameCount(4).Logger()
	zl := NewZerolog(l)

	orig := Default()
	defer SetDefault(orig)
	SetDefault(zl)

	// Call through package-level Info() to match real usage pattern.
	Info(context.Background(), "caller test")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	caller, ok := m["caller"]
	if !ok {
		t.Errorf("expected 'caller' field in output, got: %s", buf.String())
		return
	}
	callerStr, ok := caller.(string)
	if !ok {
		t.Fatalf("expected caller to be string, got %T", caller)
	}
	// Caller should contain the test file name
	if !strings.Contains(callerStr, "zerolog_test.go") {
		t.Errorf("expected caller to contain 'zerolog_test.go', got: %s", callerStr)
	}
}
