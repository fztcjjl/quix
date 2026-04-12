package config

// Config is the unified configuration interface for quix framework.
// All framework components use this interface to access configuration values.
type Config interface {
	Get(key string) any
	String(key string) string
	Int(key string) int
	Bool(key string) bool
	Bind(key string, target any) error
}
