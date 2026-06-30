package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"flashtool/internal/android"
	"flashtool/internal/device"
	"flashtool/internal/flasher"
	"flashtool/internal/logger"
	"flashtool/internal/ui"
	"flashtool/internal/updater"
	"flashtool/internal/version"
)

func main() {
	// 1. Initialization: Logger & Base Dependencies
	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(baseDir); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal Error: Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Log.Info("Starting VoidFlasher PRIME",
		"version", version.Version,
		"commit", version.Commit,
		"build_date", version.Date,
	)

	// 2. Setup System Commander
	sysCmd, err := android.NewSystemCommander()
	var baseCmd android.Commander
	if err != nil {
		logger.Log.Warn("Failed to initialize system commander, UI will display mock data", "error", err)
		baseCmd = &android.MockCommander{}
	} else {
		baseCmd = sysCmd
	}

	// Wrap with Reliability Layer
	reliableCmd := android.NewReliableCommander(baseCmd)

	// 3. Setup Device Detector and Flasher Engine
	detector := device.NewDetector(reliableCmd)
	_ = flasher.New(reliableCmd, detector)

	// 4. Check for Updates asynchronously
	go func() {
		info, err := updater.CheckForUpdates(context.Background(), version.Version)
		if err == nil && info.Available {
			logger.Log.Info("Update available!", "version", info.Version, "url", info.DownloadURL)
		}
	}()

	// 5. Start Bubble Tea TUI
	m := ui.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting UI: %v\n", err)
		os.Exit(1)
	}
}
