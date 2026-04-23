package config

import (
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// normalizeEnvKey converts SERVER_PORT to server.port, matching the env provider convention.
func normalizeEnvKey(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "_", ".")
}

type koanfConfig struct {
	k *koanf.Koanf
}

// Option is a functional option for configuring koanfConfig.
type Option func(*options)

type options struct {
	filePath string
}

// WithFile sets the YAML configuration file path.
func WithFile(path string) Option {
	return func(o *options) {
		o.filePath = path
	}
}

// NewKoanf creates a Config backed by koanf.
// It loads configuration from .env file and environment variables.
// If present, a .env file in the working directory is loaded automatically.
// Priority (low to high): YAML file < .env file < environment variables.
func NewKoanf(opts ...Option) (Config, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	k := koanf.New(".")

	// Load file first (lower priority)
	if o.filePath != "" {
		if err := k.Load(file.Provider(o.filePath), yaml.Parser()); err != nil {
			return nil, err
		}
	}

	// Load .env file (higher priority than file, lower than env vars)
	if _, err := os.Stat(".env"); err == nil {
		data, err := os.ReadFile(".env")
		if err != nil {
			return nil, err
		}
		parsed, err := dotenv.Parser().Unmarshal(data)
		if err != nil {
			return nil, err
		}
		normalized := make(map[string]any, len(parsed))
		for key, val := range parsed {
			normalized[normalizeEnvKey(key)] = val
		}
		if err := k.Load(confmap.Provider(normalized, "."), nil); err != nil {
			return nil, err
		}
	}

	// Load env (higher priority)
	if err := k.Load(env.Provider("", ".", normalizeEnvKey), nil); err != nil {
		return nil, err
	}

	return &koanfConfig{k: k}, nil
}

func (c *koanfConfig) Get(key string) any {
	return c.k.Get(key)
}

func (c *koanfConfig) String(key string) string {
	return c.k.String(key)
}

func (c *koanfConfig) Int(key string) int {
	return c.k.Int(key)
}

func (c *koanfConfig) Bool(key string) bool {
	return c.k.Bool(key)
}

func (c *koanfConfig) Bind(key string, target any) error {
	return c.k.Unmarshal(key, target)
}

var _ Config = (*koanfConfig)(nil)
