package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name          string
	Foreground    lipgloss.Color
	Background    lipgloss.Color
	Border        lipgloss.Color
	Title         lipgloss.Color
	Highlight     lipgloss.Color
	Selection     lipgloss.Color
	SelectionText lipgloss.Color
	Dim           lipgloss.Color
	StatusBg      lipgloss.Color
	StatusFg      lipgloss.Color
	Error         lipgloss.Color
	Success       lipgloss.Color
	Warning       lipgloss.Color
	Accent        lipgloss.Color
	Accent1       lipgloss.Color
}

var CurrentTheme = Theme{
	Name:          "Cyberpunk",
	Foreground:    "#e0e0e0",
	Background:    "#0a0a0f",
	Border:        "#4b0082", // Indigo
	Title:         "#ff00ff", // Neon Pink
	Highlight:     "#00ffff", // Neon Cyan
	Selection:     "#ff00ff", // Pink highlight
	SelectionText: "#000000",
	Dim:           "#4a4a4a",
	StatusBg:      "#1a1a2e",
	StatusFg:      "#ffffff",
	Error:         "#ff3131",
	Success:       "#39ff14",
	Warning:       "#faff00",
	Accent:        "#bc13fe", // Neon Purple
	Accent1:       "#00d2ff", // Sky Blue
}


/* STYLE HELPERS */

func GetBaseStyle() lipgloss.Style      { return lipgloss.NewStyle().Foreground(CurrentTheme.Foreground) }
func GetBorderStyle() lipgloss.Style    { return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(CurrentTheme.Border) }
func GetTitleStyle() lipgloss.Style     { return lipgloss.NewStyle().Foreground(CurrentTheme.Title).Bold(true) }
func GetSelectedStyle() lipgloss.Style  { return lipgloss.NewStyle().Foreground(CurrentTheme.SelectionText).Background(CurrentTheme.Selection).Bold(true) }
func GetToastStyle() lipgloss.Style     { return lipgloss.NewStyle().Padding(0, 1) }
func GetDimStyle() lipgloss.Style       { return lipgloss.NewStyle().Foreground(CurrentTheme.Dim) }
func GetSeparatorStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(CurrentTheme.Border) }
func GetHUDLabelStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(CurrentTheme.Dim).Bold(true) }
func GetHUDValueStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(CurrentTheme.Title) }

func GetBadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(CurrentTheme.StatusBg).
		Foreground(CurrentTheme.StatusFg).
		Padding(0, 1).
		MarginRight(1).
		Bold(true)
}
