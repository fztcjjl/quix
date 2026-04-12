package quix

import "github.com/fztcjjl/quix/core/logger"

// Option configures the App during creation.
type Option func(*App)

// WithLogger sets a custom Logger implementation for the App.
func WithLogger(l logger.Logger) Option {
	return func(a *App) {
		a.logger = l
	}
}
