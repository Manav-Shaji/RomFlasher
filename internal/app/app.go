package app

import (
	"fmt"

	"flashtool/internal/config"
	"flashtool/internal/core"
	"go.uber.org/zap"
)

// App is the central dependency injection container for the CLI and TUI.
// It holds all active, initialized infrastructure required for the program to execute.
type App struct {
	Config *config.AppConfig
	Logger *zap.Logger
	Engine *core.Engine
}

// New creates a new Application container.
func New(cfg *config.AppConfig, logger *zap.Logger, engineInstance *core.Engine) *App {
	return &App{
		Config: cfg,
		Logger: logger,
		Engine: engineInstance,
	}
}

// Initialize executes the core startup sequence of the application.
// It loads configuration, initializes the logger, sets up the engine,
// and returns the fully prepared runtime application container.
func Initialize() (*App, error) {
	// Load Configuration (Viper)
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("bootstrap config failed: %w", err)
	}

	// Initialize Logger (Zap)
	log, err := InitLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("bootstrap logger failed: %w", err)
	}

	log.Info("Starting NexForge initialization sequence")

	// Initialize Core Engine
	engineInstance := core.NewEngine(cfg, log.Named("engine"))

	// Create Application Container
	applicationContainer := New(cfg, log, engineInstance)

	log.Info("NexForge initialization sequence complete")
	return applicationContainer, nil
}
