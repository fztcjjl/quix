package log

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewWriterCreatesLogger(t *testing.T) {
	var buf bytes.Buffer
	l := NewWriter(&buf)
	if l == nil {
		t.Fatal("NewWriter() returned nil")
	}
}

func TestNewWriterOutputsJSON(t *testing.T) {
	var buf bytes.Buffer
	l := NewWriter(&buf)
	ctx := context.Background()

	l.Info(ctx, "test msg", "key", "val")

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw: %s", err, buf.String())
	}

	if result["msg"] != "test msg" {
		t.Errorf("msg = %v, want 'test msg'", result["msg"])
	}
	if result["key"] != "val" {
		t.Errorf("key = %v, want 'val'", result["key"])
	}
}

func TestNewWriterLogLevels(t *testing.T) {
	var buf bytes.Buffer
	l := NewWriter(&buf)
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

func TestNewWriterSetLevel(t *testing.T) {
	var buf bytes.Buffer
	l := NewWriter(&buf)
	ctx := context.Background()

	l.SetLevel(LevelError)

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

func TestNewWriterClose(t *testing.T) {
	var buf bytes.Buffer
	l := NewWriter(&buf)
	if err := l.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
}
