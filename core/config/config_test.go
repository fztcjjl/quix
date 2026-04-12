package config

import "testing"

// mockConfig is a test implementation that verifies the Config interface
// is correctly satisfied at compile time.
type mockConfig struct{}

func (m *mockConfig) Get(key string) any                { return nil }
func (m *mockConfig) String(key string) string          { return "" }
func (m *mockConfig) Int(key string) int                { return 0 }
func (m *mockConfig) Bool(key string) bool              { return false }
func (m *mockConfig) Bind(key string, target any) error { return nil }

// Compile-time check
var _ Config = (*mockConfig)(nil)

func TestMockConfigSatisfiesInterface(t *testing.T) {
	var c Config = &mockConfig{}

	if c.Get("key") != nil {
		t.Fatal("expected nil")
	}
	if c.String("key") != "" {
		t.Fatal("expected empty string")
	}
	if c.Int("key") != 0 {
		t.Fatal("expected 0")
	}
	if c.Bool("key") != false {
		t.Fatal("expected false")
	}
	if c.Bind("key", &struct{}{}) != nil {
		t.Fatal("expected nil error")
	}
}
