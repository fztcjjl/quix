package config

import (
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

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
// It loads configuration from the given options (file, env).
// Environment variables take precedence over file values.
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

	// Load env (higher priority)
	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}), nil); err != nil {
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
