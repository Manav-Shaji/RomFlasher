package app

import (
	"fmt"

	"flashtool/internal/config"
	"flashtool/internal/engine"
	"flashtool/internal/logger"
)

// Initialize executes the core startup sequence of the application.
// It loads configuration, initializes the logger, sets up the engine,
// and returns the fully prepared runtime application container.
func Initialize() (*App, error) {
	// 1. Load Configuration (Viper)
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("bootstrap config failed: %w", err)
	}

	// 2. Initialize Logger (Zap)
	log, err := logger.Initialize(cfg)
	if err != nil {
		return nil, fmt.Errorf("bootstrap logger failed: %w", err)
	}

	log.Info("Starting NexForge initialization sequence")

	// 3. Initialize Core Engine
	engineInstance := engine.NewEngine(cfg, log.Named("engine"))
	
	// 4. Create Application Container
	applicationContainer := New(cfg, log, engineInstance)

	log.Info("NexForge initialization sequence complete")
	return applicationContainer, nil
}
