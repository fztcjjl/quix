package quix

import (
	"os"

	"github.com/fztcjjl/quix/core/logger"
	"github.com/rs/zerolog"
)

// App is the core framework application.
type App struct {
	logger logger.Logger
}

// New creates a new App with the given options.
// If no logger is provided, zerolog is used by default.
func New(opts ...Option) *App {
	defaultLog := logger.NewZerolog(zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger())
	app := &App{
		logger: defaultLog,
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// Logger returns the App's logger.
func (a *App) Logger() logger.Logger {
	return a.logger
}
