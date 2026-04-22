package log

import (
	"context"
	"sync"
	"testing"
)

// captureLogger is a local test mock that records log calls.
type captureLogger struct {
	infos  []logCall
	errors []logCall
	warns  []logCall
	debugs []logCall
	traces []logCall
	mu     sync.Mutex
}

type logCall struct {
	msg  string
	args []any
}

func (l *captureLogger) Trace(_ context.Context, msg string, args ...any) {
	l.mu.Lock()
	l.traces = append(l.traces, logCall{msg, args})
	l.mu.Unlock()
}
func (l *captureLogger) Info(_ context.Context, msg string, args ...any) {
	l.mu.Lock()
	l.infos = append(l.infos, logCall{msg, args})
	l.mu.Unlock()
}
func (l *captureLogger) Error(_ context.Context, msg string, args ...any) {
	l.mu.Lock()
	l.errors = append(l.errors, logCall{msg, args})
	l.mu.Unlock()
}
func (l *captureLogger) Warn(_ context.Context, msg string, args ...any) {
	l.mu.Lock()
	l.warns = append(l.warns, logCall{msg, args})
	l.mu.Unlock()
}
func (l *captureLogger) Debug(_ context.Context, msg string, args ...any) {
	l.mu.Lock()
	l.debugs = append(l.debugs, logCall{msg, args})
	l.mu.Unlock()
}
func (l *captureLogger) Fatal(_ context.Context, _ string, _ ...any) {}
func (l *captureLogger) With(_ ...any) Logger                        { return l }
func (l *captureLogger) SetLevel(_ Level)                            {}
func (l *captureLogger) Close() error                                { return nil }

var _ Logger = (*captureLogger)(nil)

func TestCaptureLoggerSatisfiesInterface(t *testing.T) {
	var l Logger = &captureLogger{}
	ctx := context.Background()

	l.Info(ctx, "info")
	l.Error(ctx, "error")
	l.Warn(ctx, "warn")
	l.Debug(ctx, "debug")
	_ = l.Fatal // Fatal is defined but we can't call it (os.Exit)

	child := l.With("key", "value")
	if child == nil {
		t.Fatal("With() returned nil")
	}

	if err := l.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	l.SetLevel(LevelWarn)
}

func TestDefaultIsSlog(t *testing.T) {
	l := Default()
	if l == nil {
		t.Fatal("Default() returned nil")
	}
}

func TestSetDefault(t *testing.T) {
	orig := Default()
	defer SetDefault(orig)

	custom := &captureLogger{}
	SetDefault(custom)

	if Default() != custom {
		t.Fatal("SetDefault did not update defaultLogger")
	}
}

func TestDefault(t *testing.T) {
	orig := Default()
	defer SetDefault(orig)

	custom := &captureLogger{}
	SetDefault(custom)

	got := Default()
	if got != custom {
		t.Fatal("Default() did not return the custom logger")
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	orig := Default()
	defer SetDefault(orig)

	m := &captureLogger{}
	SetDefault(m)
	ctx := context.Background()

	Info(ctx, "info msg")
	Error(ctx, "error msg")
	Warn(ctx, "warn msg")
	Debug(ctx, "debug msg")
	With("key", "value")
	SetLevel(LevelWarn)

	if len(m.infos) != 1 || m.infos[0].msg != "info msg" {
		t.Errorf("expected 1 info log, got %v", m.infos)
	}
	if len(m.errors) != 1 || m.errors[0].msg != "error msg" {
		t.Errorf("expected 1 error log, got %v", m.errors)
	}
	if len(m.warns) != 1 || m.warns[0].msg != "warn msg" {
		t.Errorf("expected 1 warn log, got %v", m.warns)
	}
	if len(m.debugs) != 1 || m.debugs[0].msg != "debug msg" {
		t.Errorf("expected 1 debug log, got %v", m.debugs)
	}
}

func TestConcurrentAccess(t *testing.T) {
	orig := Default()
	defer SetDefault(orig)

	m := &captureLogger{}
	SetDefault(m)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			Info(ctx, "concurrent")
		})
		if i == 0 {
			continue // avoid lint warning
		}
	}
	wg.Wait()

	m.mu.Lock()
	count := len(m.infos)
	m.mu.Unlock()

	if count != 100 {
		t.Errorf("expected 100 info logs, got %d", count)
	}
}

func TestConcurrentSetDefaultAndLog(t *testing.T) {
	orig := Default()
	defer SetDefault(orig)

	m1 := &captureLogger{}
	m2 := &captureLogger{}

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				SetDefault(m1)
			} else {
				SetDefault(m2)
			}
		}(i)
		go func() {
			defer wg.Done()
			Info(context.Background(), "race test")
		}()
	}
	wg.Wait()
}

func TestLevelOrdering(t *testing.T) {
	if LevelTrace >= LevelDebug {
		t.Error("LevelTrace should be less than LevelDebug")
	}
	if LevelDebug >= LevelInfo {
		t.Error("LevelDebug should be less than LevelInfo")
	}
	if LevelInfo >= LevelWarn {
		t.Error("LevelInfo should be less than LevelWarn")
	}
	if LevelWarn >= LevelError {
		t.Error("LevelWarn should be less than LevelError")
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelTrace, "trace"},
		{LevelDebug, "debug"},
		{LevelInfo, "info"},
		{LevelWarn, "warn"},
		{LevelError, "error"},
		{Level(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  Level
		ok    bool
	}{
		{"trace", LevelTrace, true},
		{"debug", LevelDebug, true},
		{"info", LevelInfo, true},
		{"warn", LevelWarn, true},
		{"error", LevelError, true},
		{"INFO", LevelInfo, true},
		{"DEBUG", LevelDebug, true},
		{"invalid", Level(0), false},
	}
	for _, tt := range tests {
		got, err := ParseLevel(tt.input)
		if tt.ok && err != nil {
			t.Errorf("ParseLevel(%q) unexpected error: %v", tt.input, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("ParseLevel(%q) expected error, got nil", tt.input)
		}
		if err == nil && got != tt.want {
			t.Errorf("ParseLevel(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestPackageLevelTrace(t *testing.T) {
	orig := Default()
	defer SetDefault(orig)

	m := &captureLogger{}
	SetDefault(m)
	ctx := context.Background()

	Trace(ctx, "trace msg")

	if len(m.traces) != 1 || m.traces[0].msg != "trace msg" {
		t.Errorf("expected 1 trace log, got %v", m.traces)
	}
}

func TestNewContextAndFromContext(t *testing.T) {
	orig := Default()
	defer SetDefault(orig)

	custom := &captureLogger{}
	ctx := NewContext(context.Background(), custom)

	got := FromContext(ctx)
	if got != custom {
		t.Error("FromContext should return the stored logger")
	}
}

func TestNewContextOverwrites(t *testing.T) {
	l1 := &captureLogger{}
	l2 := &captureLogger{}

	ctx := NewContext(context.Background(), l1)
	ctx = NewContext(ctx, l2)

	if FromContext(ctx) != l2 {
		t.Error("second NewContext should overwrite the first")
	}
}

func TestFromContextReturnsDefaultWhenEmpty(t *testing.T) {
	ctx := context.Background()
	got := FromContext(ctx)
	if got != Default() {
		t.Error("FromContext should return Default() when no logger in context")
	}
}
