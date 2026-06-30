package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"

	"flashtool/internal/app"
	"flashtool/internal/tui"
)

var (
	applicationContainer *app.App
)

var rootCmd = &cobra.Command{
	Use:   "nexforge",
	Short: "NexForge - Android ROM Flashing TUI/CLI",
	Long:  `NexForge is a powerful TUI and CLI application for flashing Android ROMs via Fastboot and ADB.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		applicationContainer, err = app.Initialize()
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		m := tui.NewModel(applicationContainer)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error starting TUI: %w", err)
		}
		return nil
	},
}

func Execute() {
	cobra.MousetrapHelpText = ""
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
