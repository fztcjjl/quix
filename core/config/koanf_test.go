package config

import (
	"os"
	"path/filepath"
	"testing"
)

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

	t.Setenv("SERVER_PORT", "9090")

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
	t.Setenv("APP_NAME", "test")

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
