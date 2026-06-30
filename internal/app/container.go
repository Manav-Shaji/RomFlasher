package app

import (
	"flashtool/internal/config"
	"flashtool/internal/engine"
	"go.uber.org/zap"
)

// App is the central dependency injection container for the CLI and TUI.
// It holds all active, initialized infrastructure required for the program to execute.
type App struct {
	Config *config.AppConfig
	Logger *zap.Logger
	Engine *engine.Engine
}

// New creates a new Application container.
func New(cfg *config.AppConfig, logger *zap.Logger, engineInstance *engine.Engine) *App {
	return &App{
		Config: cfg,
		Logger: logger,
		Engine: engineInstance,
	}
}
