package tui

import (
	"fmt"
	"flashtool/internal/core"
	"flashtool/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func renderConfirmView(m AppModel, width int) string {
	title := theme.GetTitleStyle().Background(theme.CurrentTheme.Warning).Foreground(theme.CurrentTheme.Background).
		Width(width - 2).Align(lipgloss.Center).Render("⚠️  EXECUTION CONFIRM")
	
	msg := theme.GetTitleStyle().Foreground(theme.CurrentTheme.Highlight).Width(width - 4).Align(lipgloss.Center).Render(m.Modal.ConfirmMsg)
	warn := theme.GetDimStyle().Width(width - 4).Align(lipgloss.Center).Render("This action cannot be undone.")
	opts := theme.GetBaseStyle().Bold(true).Render("  [Y] PROCEED    [N] CANCEL")
	
	box := theme.GetBorderStyle().BorderForeground(theme.CurrentTheme.Warning).
		Width(width - 4).Padding(1, 1).
		Render(lipgloss.JoinVertical(lipgloss.Center, msg, warn, "", opts))

	return lipgloss.JoinVertical(lipgloss.Left, title, box)
}

func renderBusyView(m AppModel, width int) string {
	return "" 
}

func renderUnauthorizedView(width int) string {
	header := lipgloss.NewStyle().Background(theme.CurrentTheme.Warning).Foreground(theme.CurrentTheme.Background).
		Width(width - 2).Align(lipgloss.Center).Bold(true).Render("⚠️  ACTION REQUIRED")
	
	msg := theme.GetTitleStyle().Foreground(theme.CurrentTheme.Warning).Width(width - 4).Align(lipgloss.Center).Render("DEVICE UNAUTHORIZED")
	help := theme.GetBaseStyle().Width(width - 4).Align(lipgloss.Center).Render("Please check your device screen and\naccept the USB debugging prompt.")
	
	return lipgloss.JoinVertical(lipgloss.Left, header, "", msg, "", help)
}

func renderDeviceHUD(m AppModel, width int) string {
	lblStyle, valStyle := theme.GetHUDLabelStyle(), theme.GetHUDValueStyle()

	statusHeader := lipgloss.NewStyle().
		Background(theme.CurrentTheme.Highlight).Foreground(theme.CurrentTheme.Background).
		Width(width - 2).Bold(true).Align(lipgloss.Center).Render("DEVICE STATUS")

	row := lipgloss.JoinHorizontal(lipgloss.Left,
		lblStyle.Render(" MDL "), valStyle.Render(fmt.Sprintf("%-12s", m.Device.Model)),
		"   ",
		lblStyle.Render(" BAT "), valStyle.Render(fmt.Sprintf("%-9s", m.Device.Battery)),
	)
	if m.Device.Mode == core.ModeFastboot {
		row += "   " + lblStyle.Render(" SEC ") + valStyle.Render(m.Device.Secure)
	}

	infoBox := theme.GetSeparatorStyle().BorderStyle(lipgloss.NormalBorder()).
		Width(width - 4).Padding(0, 1).Render(row)

	descHeader := theme.GetTitleStyle().Width(width - 2).Background(theme.CurrentTheme.Dim).Align(lipgloss.Center).Render(" COMMAND INFO ")
	desc := theme.GetBaseStyle().Width(width - 4).Render(m.Menu[m.Selection].Desc)

	return lipgloss.JoinVertical(lipgloss.Left, statusHeader, infoBox, "", descHeader, desc)
}

func renderInfoView(m AppModel, width int) string {
	header := lipgloss.NewStyle().Background(theme.CurrentTheme.StatusBg).Foreground(theme.CurrentTheme.StatusFg).
		Width(width - 2).Align(lipgloss.Center).Bold(true).Render("INFORMATION")
	desc := theme.GetBaseStyle().Width(width - 4).Padding(1, 1).Render(m.Menu[m.Selection].Desc)
	return lipgloss.JoinVertical(lipgloss.Left, header, desc)
}
