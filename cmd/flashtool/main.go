package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"flashtool/internal"
	tea "github.com/charmbracelet/bubbletea"
)

//go:embed config.json
var embeddedConfig []byte

func main() {
	// 1. Configuration (Local overrides Embedded)
	var config internal.AppConfig
	
	// Try loading local config first
	data, err := os.ReadFile("config.json")
	if err != nil {
		// Fallback to embedded config
		data = embeddedConfig
	}
	
	json.Unmarshal(data, &config)

	// 2. Model Initialization
	model := internal.NewModel()
	model.Config = config
	model.BaseDir = config.BaseDir
	model.DevicePath = config.DevicePath
	model.Menu = internal.GetDefaultMenu()
	model.SetupUI()

	// 3. Execution
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal Error: %v\n", err)
		os.Exit(1)
	}
}
