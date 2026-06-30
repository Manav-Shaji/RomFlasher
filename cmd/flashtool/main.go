package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"flashtool/internal/tui"
)

func main() {
	// Start Bubble Tea TUI using the main internal AppModel
	m := tui.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting UI: %v\n", err)
		os.Exit(1)
	}
}
