package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Background lipgloss.Color
	Foreground lipgloss.Color
	Dim        lipgloss.Color
	Highlight  lipgloss.Color
	Accent     lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Success    lipgloss.Color
	StatusBg   lipgloss.Color
	StatusFg   lipgloss.Color
	Title      lipgloss.Color
	Border     lipgloss.Color
}

var CurrentTheme = Theme{
	Background: lipgloss.Color("#0a0a0a"), // ColorDarkBg
	Foreground: lipgloss.Color("#ffffff"), // ColorText
	Dim:        lipgloss.Color("#555555"),
	Highlight:  lipgloss.Color("#00f3ff"), // ColorNeonCyan
	Accent:     lipgloss.Color("#bd00ff"), // ColorPurple
	Warning:    lipgloss.Color("#ffaa00"),
	Error:      lipgloss.Color("#ff003c"), // ColorError
	Success:    lipgloss.Color("#00ff66"), // ColorSuccess
	StatusBg:   lipgloss.Color("#bd00ff"), // ColorPurple
	StatusFg:   lipgloss.Color("#ffffff"),
	Title:      lipgloss.Color("#bd00ff"), // ColorPurple
	Border:     lipgloss.Color("#555555"),
}

func GetTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func GetDimStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(CurrentTheme.Dim)
}

func GetBaseStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(CurrentTheme.Foreground)
}

func GetBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(CurrentTheme.Border)
}

func GetHUDLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(CurrentTheme.Dim).Bold(true)
}

func GetHUDValueStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(CurrentTheme.Highlight).Bold(true)
}

func GetSeparatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().BorderForeground(CurrentTheme.Dim)
}

func GetSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func GetBadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1).Bold(true)
}
