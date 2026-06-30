package theme

import "github.com/charmbracelet/lipgloss"

var (
	ColorNeonCyan = lipgloss.Color("#00f3ff")
	ColorPurple   = lipgloss.Color("#bd00ff")
	ColorDarkBg   = lipgloss.Color("#0a0a0a")
	ColorText     = lipgloss.Color("#ffffff")
	ColorError    = lipgloss.Color("#ff003c")
	ColorSuccess  = lipgloss.Color("#00ff66")

	StyleHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorNeonCyan).
		Padding(0, 1).
		Border(lipgloss.DoubleBorder(), true).
		BorderForeground(ColorPurple)

	StylePanel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorNeonCyan).
		Padding(0, 1)

	StyleStatus = lipgloss.NewStyle().
		Foreground(ColorText).
		Background(ColorPurple).
		Padding(0, 1)
		
	StyleError = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)
		
	StyleSuccess = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)
)
