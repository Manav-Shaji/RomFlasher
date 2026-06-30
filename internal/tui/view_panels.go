package tui

import (
	"fmt"
	"flashtool/internal/domain"
	"flashtool/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

var (
	confirmTitleStyle  = theme.TitleStyle.Copy().Background(theme.CurrentTheme.Warning).Foreground(theme.CurrentTheme.Background).Align(lipgloss.Center)
	confirmMsgStyle    = theme.TitleStyle.Copy().Foreground(theme.CurrentTheme.Highlight).Align(lipgloss.Center)
	confirmWarnStyle   = theme.DimStyle.Copy().Align(lipgloss.Center)
	confirmOptsStyle   = theme.BaseStyle.Copy().Bold(true)
	confirmBoxStyle    = theme.BorderStyle.Copy().BorderForeground(theme.CurrentTheme.Warning).Padding(1, 1)

	unauthHeaderStyle  = lipgloss.NewStyle().Background(theme.CurrentTheme.Warning).Foreground(theme.CurrentTheme.Background).Align(lipgloss.Center).Bold(true)
	unauthMsgStyle     = theme.TitleStyle.Copy().Foreground(theme.CurrentTheme.Warning).Align(lipgloss.Center)
	unauthHelpStyle    = theme.BaseStyle.Copy().Align(lipgloss.Center)

	hudStatusHeader    = lipgloss.NewStyle().Background(theme.CurrentTheme.Highlight).Foreground(theme.CurrentTheme.Background).Bold(true).Align(lipgloss.Center)
	hudInfoBox         = theme.SeparatorStyle.Copy().BorderStyle(lipgloss.NormalBorder()).Padding(0, 1)
	hudDescHeader      = theme.TitleStyle.Copy().Background(theme.CurrentTheme.Dim).Align(lipgloss.Center)
	hudDesc            = theme.BaseStyle.Copy()

	infoHeader         = lipgloss.NewStyle().Background(theme.CurrentTheme.StatusBg).Foreground(theme.CurrentTheme.StatusFg).Align(lipgloss.Center).Bold(true)
	infoDesc           = theme.BaseStyle.Copy().Padding(1, 1)
)

func renderConfirmView(m AppModel, width int) string {
	title := confirmTitleStyle.Copy().Width(width - 2).Render("⚠️  EXECUTION CONFIRM")
	msg := confirmMsgStyle.Copy().Width(width - 4).Render(m.Modal.ConfirmMsg)
	warn := confirmWarnStyle.Copy().Width(width - 4).Render("This action cannot be undone.")
	opts := confirmOptsStyle.Render("  [Y] PROCEED    [N] CANCEL")
	
	box := confirmBoxStyle.Copy().Width(width - 4).Render(lipgloss.JoinVertical(lipgloss.Center, msg, warn, "", opts))

	return lipgloss.JoinVertical(lipgloss.Left, title, box)
}

func renderBusyView(_ AppModel, _ int) string {
	return "" 
}

func renderUnauthorizedView(width int) string {
	header := unauthHeaderStyle.Copy().Width(width - 2).Render("⚠️  ACTION REQUIRED")
	msg := unauthMsgStyle.Copy().Width(width - 4).Render("DEVICE UNAUTHORIZED")
	help := unauthHelpStyle.Copy().Width(width - 4).Render("Please check your device screen and\naccept the USB debugging prompt.")
	
	return lipgloss.JoinVertical(lipgloss.Left, header, "", msg, "", help)
}

func renderDeviceHUD(m AppModel, width int) string {
	lblStyle, valStyle := theme.HUDLabelStyle, theme.HUDValueStyle

	statusHeader := hudStatusHeader.Copy().Width(width - 2).Render("DEVICE STATUS")

	row := lipgloss.JoinHorizontal(lipgloss.Left,
		lblStyle.Render(" MDL "), valStyle.Render(fmt.Sprintf("%-12s", m.Device.Model)),
		"   ",
		lblStyle.Render(" BAT "), valStyle.Render(fmt.Sprintf("%-9s", m.Device.Battery)),
	)
	if m.Device.Mode == domain.ModeFastboot {
		row += "   " + lblStyle.Render(" SEC ") + valStyle.Render(m.Device.Secure)
	}

	infoBox := hudInfoBox.Copy().Width(width - 4).Render(row)

	descHeader := hudDescHeader.Copy().Width(width - 2).Render(" COMMAND INFO ")
	desc := hudDesc.Copy().Width(width - 4).Render(m.Menu[m.Selection].Desc)

	return lipgloss.JoinVertical(lipgloss.Left, statusHeader, infoBox, "", descHeader, desc)
}

func renderInfoView(m AppModel, width int) string {
	header := infoHeader.Copy().Width(width - 2).Render("INFORMATION")
	desc := infoDesc.Copy().Width(width - 4).Render(m.Menu[m.Selection].Desc)
	return lipgloss.JoinVertical(lipgloss.Left, header, desc)
}
