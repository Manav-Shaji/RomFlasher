package internal

import (
	"fmt"
	"flashtool/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func renderConfirmView(m AppModel, width int) string {
	title := ui.GetTitleStyle().Background(ui.CurrentTheme.Warning).Foreground(ui.CurrentTheme.Background).
		Width(width - 2).Align(lipgloss.Center).Render("⚠️  EXECUTION CONFIRM")
	
	msg := ui.GetTitleStyle().Foreground(ui.CurrentTheme.Highlight).Width(width - 4).Align(lipgloss.Center).Render(m.Modal.ConfirmMsg)
	warn := ui.GetDimStyle().Width(width - 4).Align(lipgloss.Center).Render("This action cannot be undone.")
	opts := ui.GetBaseStyle().Bold(true).Render("  [Y] PROCEED    [N] CANCEL")
	
	box := ui.GetBorderStyle().BorderForeground(ui.CurrentTheme.Warning).
		Width(width - 4).Padding(1, 1).
		Render(lipgloss.JoinVertical(lipgloss.Center, msg, warn, "", opts))

	return lipgloss.JoinVertical(lipgloss.Left, title, box)
}

func renderBusyView(m AppModel, width int) string {
	return "" 
}

func renderUnauthorizedView(width int) string {
	header := lipgloss.NewStyle().Background(ui.CurrentTheme.Warning).Foreground(ui.CurrentTheme.Background).
		Width(width - 2).Align(lipgloss.Center).Bold(true).Render("⚠️  ACTION REQUIRED")
	
	msg := ui.GetTitleStyle().Foreground(ui.CurrentTheme.Warning).Width(width - 4).Align(lipgloss.Center).Render("DEVICE UNAUTHORIZED")
	help := ui.GetBaseStyle().Width(width - 4).Align(lipgloss.Center).Render("Please check your device screen and\naccept the USB debugging prompt.")
	
	return lipgloss.JoinVertical(lipgloss.Left, header, "", msg, "", help)
}

func renderDeviceHUD(m AppModel, width int) string {
	lblStyle, valStyle := ui.GetHUDLabelStyle(), ui.GetHUDValueStyle()

	statusHeader := lipgloss.NewStyle().
		Background(ui.CurrentTheme.Highlight).Foreground(ui.CurrentTheme.Background).
		Width(width - 2).Bold(true).Align(lipgloss.Center).Render("DEVICE STATUS")

	row := lipgloss.JoinHorizontal(lipgloss.Left,
		lblStyle.Render(" MDL "), valStyle.Render(fmt.Sprintf("%-12s", m.Device.Model)),
		"   ",
		lblStyle.Render(" BAT "), valStyle.Render(fmt.Sprintf("%-9s", m.Device.Battery)),
	)
	if m.Device.Mode == ModeFastboot {
		row += "   " + lblStyle.Render(" SEC ") + valStyle.Render(m.Device.Secure)
	}

	infoBox := ui.GetSeparatorStyle().BorderStyle(lipgloss.NormalBorder()).
		Width(width - 4).Padding(0, 1).Render(row)

	descHeader := ui.GetTitleStyle().Width(width - 2).Background(ui.CurrentTheme.Dim).Align(lipgloss.Center).Render(" COMMAND INFO ")
	desc := ui.GetBaseStyle().Width(width - 4).Render(m.Menu[m.Selection].Desc)

	return lipgloss.JoinVertical(lipgloss.Left, statusHeader, infoBox, "", descHeader, desc)
}

func renderInfoView(m AppModel, width int) string {
	header := lipgloss.NewStyle().Background(ui.CurrentTheme.StatusBg).Foreground(ui.CurrentTheme.StatusFg).
		Width(width - 2).Align(lipgloss.Center).Bold(true).Render("INFORMATION")
	desc := ui.GetBaseStyle().Width(width - 4).Padding(1, 1).Render(m.Menu[m.Selection].Desc)
	return lipgloss.JoinVertical(lipgloss.Left, header, desc)
}
