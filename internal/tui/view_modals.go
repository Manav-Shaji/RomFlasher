package tui

import (
	"fmt"
	"strings"
	"flashtool/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func renderModal(m AppModel) string {
	w := int(float64(m.Width) * 0.8)
	if w < 60 { w = 60 }
	if w > 120 { w = 120 }
	
	h := 20
	if m.ActiveModal == ModalHelp { h = 26 }
	if m.ActiveModal == ModalCustom { h = 24 }
	if m.ActiveModal == ModalSettings { h = 22 }
	
	style := theme.BorderStyle.Copy().Width(w).Height(h).Border(lipgloss.RoundedBorder()).BorderForeground(theme.CurrentTheme.Highlight)
	
	switch m.ActiveModal {
	case ModalFile:
		return style.Render(renderFileModal(m, w))
	case ModalCustom:
		return style.Render(renderCustomModal(m, w, h))
	case ModalSettings:
		return style.Render(renderSettingsModal(m, w, h))
	case ModalHelp:
		return style.Render(renderHelpModal(m, w))
	case ModalConfirm:
		return "" 
	}
	return ""
}

func renderFileModal(m AppModel, w int) string {
	title := theme.TitleStyle.Copy().Width(w - 4).Align(lipgloss.Center).Render(m.Modal.FileTitle)
	var b strings.Builder
	
	// Path Box
	pathLabel := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Background).Background(theme.CurrentTheme.Highlight).Padding(0, 1).Bold(true).Render(" PATH ")
	pathText := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Highlight).Italic(true).Render(" " + m.Modal.FileDir)
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, pathLabel, pathText))
	b.WriteString("\n\n")

	// Search Box
	searchLabel := lipgloss.NewStyle().Padding(0, 1).Background(theme.CurrentTheme.Accent).Foreground(theme.CurrentTheme.Background).Bold(true).Render(" 🔍 SEARCH ")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, searchLabel, " ", m.UI.TextInput.View()))
	b.WriteString("\n\n")

	start := 0
	if m.Modal.FileCursor > 8 { start = m.Modal.FileCursor - 8 }
	end := start + 10
	if end > len(m.Modal.FileList) { end = len(m.Modal.FileList) }
	for i := start; i < end; i++ {
		f := m.Modal.FileList[i]
		icon := "📄"
		if f.IsDir { icon = "📁" }
		if f.Name == "[ SELECT THIS FOLDER ]" { icon = "✅" }
		name := f.Name
		if len(name) > 40 { name = name[:37] + "..." }
		if i == m.Modal.FileCursor {
			barWidth := w - 8
			b.WriteString(theme.SelectedStyle.Copy().Width(barWidth).Background(theme.CurrentTheme.Highlight).Foreground(theme.CurrentTheme.Background).Render(fmt.Sprintf(" ❯ %s  %s", icon, name)))
			b.WriteByte('\n')
		} else {
			s := theme.BaseStyle.Copy()
			if f.Name == "[ SELECT THIS FOLDER ]" { s = s.Foreground(theme.CurrentTheme.Highlight).Bold(true) }
			b.WriteString(fmt.Sprintf("   %s  %s\n", icon, s.Render(name)))
		}
	}
	for i := 0; i < (10 - (end - start)); i++ { b.WriteByte('\n') }
	fLeft := theme.DimStyle.Copy().Render(fmt.Sprintf(" %d items match", len(m.Modal.FileList)))
	fRight := theme.DimStyle.Copy().Render("↑/↓ Nav • ↵ Open • Esc Back ")
	space := w - lipgloss.Width(fLeft) - lipgloss.Width(fRight) - 4
	b.WriteString(fmt.Sprintf("\n%s%s%s", fLeft, strings.Repeat(" ", space), fRight))
	return lipgloss.JoinVertical(lipgloss.Center, title, "", lipgloss.NewStyle().Width(w - 4).Padding(0, 2).Render(b.String()))
}

func renderCustomModal(m AppModel, w, _ int) string {
	title := theme.TitleStyle.Copy().
		Background(theme.CurrentTheme.Highlight).
		Foreground(theme.CurrentTheme.Background).
		Width(w - 4).
		Align(lipgloss.Center).
		Render(" COMMAND CONSOLE ")
	var b strings.Builder
	innerW := w - 6
	// 1. Output Panel
	outputTitle := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Highlight).Bold(true).Render("🖥️  OUTPUT")
	b.WriteString(outputTitle)
	b.WriteByte('\n')
	
	m.Modal.CustomViewport.Width = innerW
	m.Modal.CustomViewport.Height = 11
	
	outputContent := m.Modal.CustomViewport.View()
	if m.Modal.CustomLogs.Len() == 0 && !m.Busy {
		outputContent = theme.DimStyle.Copy().Italic(true).Render("\n\n  Terminal initialized. Enter command below...")
	}
	
	outputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.CurrentTheme.Dim).
		Width(innerW).
		Height(11).
		Padding(0, 1).
		Render(outputContent)
	b.WriteString(outputBox)
	b.WriteByte('\n')

	// 2. Input Panel
	inputLabel := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Background).Background(theme.CurrentTheme.Accent).Bold(true).Padding(0, 1).Render(" COMMAND ")
	m.UI.TextInput.Prompt = lipgloss.NewStyle().Foreground(theme.CurrentTheme.Accent).Render(" ❯ ")
	inputField := m.UI.TextInput.View()
	if m.Busy { inputField = theme.DimStyle.Copy().Render("Executing command... ⚡") }
	
	inputLine := lipgloss.JoinHorizontal(lipgloss.Left, inputLabel, " ", inputField)
	inputBox := lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e")).Width(innerW).Padding(0, 1).Render(inputLine)
	b.WriteString(inputBox)
	b.WriteString("\n\n")

	
	// 3. Footer
	examples := theme.DimStyle.Copy().Render(" Try: 'adb shell getprop' or 'fastboot getvar all'")
	if m.Modal.CustomLogs.Len() > 0 || m.Busy {
		examples = theme.DimStyle.Copy().Render(fmt.Sprintf(" History: %d lines", m.Modal.CustomLogs.Len()))
	}

	fLeft, fRight := examples, theme.DimStyle.Copy().Render("↵ EXECUTE  •  ESC EXIT ")
	spaceCount := w - lipgloss.Width(fLeft) - lipgloss.Width(fRight) - 8
	if spaceCount < 0 { spaceCount = 0 }
	
	footer := lipgloss.JoinHorizontal(lipgloss.Bottom, fLeft, strings.Repeat(" ", spaceCount), fRight)
	b.WriteString(footer)
	
	return lipgloss.JoinVertical(lipgloss.Left, title, "", lipgloss.NewStyle().Padding(0, 2).Render(b.String()))
}

func renderHelpModal(_ AppModel, w int) string {
	title := theme.TitleStyle.Copy().Width(w - 4).Align(lipgloss.Center).Render("SYSTEM DOCUMENTATION")
	hl := theme.TitleStyle.Copy().Foreground(theme.CurrentTheme.Highlight).Width(w - 8).Align(lipgloss.Center)
	row := func(k, v string) string {
		return fmt.Sprintf(" %s %s\n", lipgloss.NewStyle().Width(18).Foreground(theme.CurrentTheme.Highlight).Render(k), v)
	}

	var b strings.Builder
	b.WriteString(hl.Render("🎮 NAVIGATION & CONTROLS"))
	b.WriteByte('\n')
	b.WriteString(row("[↑/↓] / [K/J]", "Move Selection"))
	b.WriteString(row("[ENTER]", "Execute Action"))
	b.WriteString(row("[ESC/BACK]", "Close Window"))
	
	b.WriteString(fmt.Sprintf("\n%s\n", hl.Render("📱 DEVICE MODES GUIDE")))
	b.WriteString(" • FASTBOOT : Phone is in Bootloader mode (flashing images)\n")
	b.WriteString(" • SIDELOAD : Recovery mode -> Apply update from ADB\n")
	b.WriteString(" • UNAUTH   : Check phone screen for USB debug prompt\n")

	b.WriteString(fmt.Sprintf("\n%s\n", hl.Render("🛡️ FLASHING SAFETY TIPS")))
	b.WriteString(" • Use original USB 2.0 cables whenever possible\n")
	b.WriteString(" • Ensure battery is at least 30% before flashing\n")
	b.WriteString(" • A/B SLOTS: Tool targets the active slot automatically\n")

	b.WriteString(fmt.Sprintf("\n%s\n", hl.Render("⌨️ CUSTOM COMMANDS")))
	b.WriteString(" Runs directly in your flashtool directory. Use it for manual\n")
	b.WriteString(" adb shell commands or custom fastboot flags.\n")

	b.WriteString(fmt.Sprintf("\n%s", theme.DimStyle.Copy().Width(w - 8).Align(lipgloss.Center).Render("Build: stable-1.2.0 • Pro Flasher Core")))
	
	return lipgloss.JoinVertical(lipgloss.Center, title, "", lipgloss.NewStyle().Width(w - 4).Padding(0, 2).Render(b.String()))
}

func renderSettingsModal(m AppModel, w, _ int) string {
	title := theme.TitleStyle.Copy().Width(w - 4).Align(lipgloss.Center).Render("APPLICATION CONFIGURATION")
	var b strings.Builder
	innerW := w - 6

	renderItem := func(index int, label, desc, pathVal string) {
		b.WriteString(fmt.Sprintf(" %s\n", theme.TitleStyle.Copy().Foreground(theme.CurrentTheme.Highlight).Render(label)))
		b.WriteString(fmt.Sprintf("  %s\n", theme.DimStyle.Copy().Render(desc)))
		
		pad := "  "
		if m.Modal.SettingsCursor == index {
			pad = lipgloss.NewStyle().Foreground(theme.CurrentTheme.Accent).Render("❯ ")
		}
		
		val := pathVal
		if val == "" { val = "(NOT SET)" }
		
		pathStr := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Foreground).Render(val)
		b.WriteString(fmt.Sprintf("%s%s\n\n", pad, pathStr))
	}

	renderItem(0, "📂 Base ROM Directory", "The root folder where your custom ROMs and images are stored.", m.App.Config.BaseDir)
	renderItem(1, "📱 Target Device Folder", "Default folder structure path for the current device.", m.App.Config.DevicePath)
	
	for i := 0; i < 1; i++ { b.WriteByte('\n') }
	
	saveLabel := "  [ SAVE AND APPLY CONFIGURATION ]  "
	saveStyle := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Dim)
	if m.Modal.SettingsCursor == 2 {
		saveStyle = lipgloss.NewStyle().Foreground(theme.CurrentTheme.Background).Background(theme.CurrentTheme.Accent).Bold(true)
	}
	saveBtn := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Center).Render(saveStyle.Render(saveLabel))
	b.WriteString(saveBtn)
	b.WriteString("\n\n")

	fLeft, fRight := theme.DimStyle.Copy().Render(" TAB Nav • ↵ SELECT "), theme.DimStyle.Copy().Render(" ESC CANCEL ")
	spaceCount := w - lipgloss.Width(fLeft) - lipgloss.Width(fRight) - 8
	if spaceCount < 0 { spaceCount = 0 }
	footer := lipgloss.JoinHorizontal(lipgloss.Bottom, fLeft, strings.Repeat(" ", spaceCount), fRight)
	b.WriteString(footer)
	
	return lipgloss.JoinVertical(lipgloss.Left, title, "", lipgloss.NewStyle().Padding(0, 2).Render(b.String()))
}
