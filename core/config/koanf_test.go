package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestEnv(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test .env: %v", err)
	}
	return path
}

func writeTestYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test yaml: %v", err)
	}
	return path
}

func TestNewKoanfFromFile(t *testing.T) {
	yamlContent := `
server:
  host: localhost
  port: 8080
debug: true`
	path := writeTestYAML(t, yamlContent)

	cfg, err := NewKoanf(WithFile(path))
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.String("server.host") != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.String("server.host"))
	}
	if cfg.Int("server.port") != 8080 {
		t.Errorf("expected 8080, got %d", cfg.Int("server.port"))
	}
	if !cfg.Bool("debug") {
		t.Error("expected debug to be true")
	}
}

func TestEnvOverrideFile(t *testing.T) {
	yamlContent := `
server:
  port: 8080`
	path := writeTestYAML(t, yamlContent)

	t.Setenv("SERVER__PORT", "9090")

	cfg, err := NewKoanf(WithFile(path))
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.Int("server.port") != 9090 {
		t.Errorf("expected 9090 (env override), got %d", cfg.Int("server.port"))
	}
}

func TestNonExistentKey(t *testing.T) {
	path := writeTestYAML(t, "key: value")

	cfg, err := NewKoanf(WithFile(path))
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.String("nonexistent") != "" {
		t.Errorf("expected empty string, got %s", cfg.String("nonexistent"))
	}
	if cfg.Int("nonexistent") != 0 {
		t.Errorf("expected 0, got %d", cfg.Int("nonexistent"))
	}
	if cfg.Get("nonexistent") != nil {
		t.Errorf("expected nil, got %v", cfg.Get("nonexistent"))
	}
}

func TestBindToStruct(t *testing.T) {
	yamlContent := `
server:
  host: 0.0.0.0
  port: 3000`
	path := writeTestYAML(t, yamlContent)

	type ServerConfig struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	cfg, err := NewKoanf(WithFile(path))
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	var sc ServerConfig
	if err := cfg.Bind("server", &sc); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	if sc.Host != "0.0.0.0" {
		t.Errorf("expected 0.0.0.0, got %s", sc.Host)
	}
	if sc.Port != 3000 {
		t.Errorf("expected 3000, got %d", sc.Port)
	}
}

func TestNestedKeyAccess(t *testing.T) {
	yamlContent := `
app:
  server:
    host: localhost
    port: 8080`
	path := writeTestYAML(t, yamlContent)

	cfg, err := NewKoanf(WithFile(path))
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.String("app.server.host") != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.String("app.server.host"))
	}
	if cfg.Int("app.server.port") != 8080 {
		t.Errorf("expected 8080, got %d", cfg.Int("app.server.port"))
	}
}

func TestNewKoanfNoFile(t *testing.T) {
	t.Setenv("APP__NAME", "test")

	cfg, err := NewKoanf()
	if err != nil {
		t.Fatalf("NewKoanf without file failed: %v", err)
	}

	if cfg.String("app.name") != "test" {
		t.Errorf("expected test, got %s", cfg.String("app.name"))
	}
}

func TestNewKoanfFileNotFound(t *testing.T) {
	_, err := NewKoanf(WithFile("/nonexistent/config.yaml"))
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestEnvFileLoaded(t *testing.T) {
	dir := t.TempDir()
	writeTestEnv(t, dir, "APP__NAME=fromenvfile")

	// .env is loaded from cwd, so we must chdir
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	defer os.Chdir(old)

	cfg, err := NewKoanf()
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.String("app.name") != "fromenvfile" {
		t.Errorf("expected fromenvfile, got %s", cfg.String("app.name"))
	}
}

func TestEnvFileSkippedWhenMissing(t *testing.T) {
	dir := t.TempDir()
	// No .env file in dir

	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	defer os.Chdir(old)

	_, err = NewKoanf()
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}
}

func TestEnvVarOverridesEnvFile(t *testing.T) {
	dir := t.TempDir()
	writeTestEnv(t, dir, "SERVER__PORT=3000")

	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	defer os.Chdir(old)

	t.Setenv("SERVER__PORT", "9090")

	cfg, err := NewKoanf()
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.Int("server.port") != 9090 {
		t.Errorf("expected 9090 (env override), got %d", cfg.Int("server.port"))
	}
}

func TestEnvFileOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	yamlContent := "server:\n  port: 8080"
	yamlPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0600); err != nil {
		t.Fatal(err)
	}
	writeTestEnv(t, dir, "SERVER__PORT=3000")

	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	defer os.Chdir(old)

	cfg, err := NewKoanf(WithFile(yamlPath))
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.Int("server.port") != 3000 {
		t.Errorf("expected 3000 (from .env), got %d", cfg.Int("server.port"))
	}
}

func TestNestedSnakeCaseEnvOverride(t *testing.T) {
	yamlContent := `
sms:
  access_key_id: old_key
  pool:
    max_size: 5`
	path := writeTestYAML(t, yamlContent)

	t.Setenv("SMS__ACCESS_KEY_ID", "new_key")

	cfg, err := NewKoanf(WithFile(path))
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.String("sms.access_key_id") != "new_key" {
		t.Errorf("expected new_key, got %s", cfg.String("sms.access_key_id"))
	}
	// Pool section without env override should keep YAML value
	if cfg.Int("sms.pool.max_size") != 5 {
		t.Errorf("expected 5, got %d", cfg.Int("sms.pool.max_size"))
	}
}

func TestDoubleUnderscoreNesting(t *testing.T) {
	t.Setenv("SMS__POOL__MAX_SIZE", "10")

	cfg, err := NewKoanf()
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.Int("sms.pool.max_size") != 10 {
		t.Errorf("expected 10, got %d", cfg.Int("sms.pool.max_size"))
	}
}

func TestWithEnvPrefixFiltersVars(t *testing.T) {
	yamlContent := `
server:
  port: 8080`
	path := writeTestYAML(t, yamlContent)

	t.Setenv("MYAPP_SERVER__PORT", "9090")
	// Non-prefixed var should be ignored
	t.Setenv("SERVER__PORT", "3000")

	cfg, err := NewKoanf(WithFile(path), WithEnvPrefix("MYAPP_"))
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.Int("server.port") != 9090 {
		t.Errorf("expected 9090 (from MYAPP_ prefix), got %d", cfg.Int("server.port"))
	}
}

func TestWithEnvPrefixStripsPrefix(t *testing.T) {
	t.Setenv("QUIX_SERVER__PORT", "8080")
	t.Setenv("QUIX_APP__NAME", "myapp")

	cfg, err := NewKoanf(WithEnvPrefix("QUIX_"))
	if err != nil {
		t.Fatalf("NewKoanf failed: %v", err)
	}

	if cfg.Int("server.port") != 8080 {
		t.Errorf("expected 8080, got %d", cfg.Int("server.port"))
	}
	if cfg.String("app.name") != "myapp" {
		t.Errorf("expected myapp, got %s", cfg.String("app.name"))
	}
}
