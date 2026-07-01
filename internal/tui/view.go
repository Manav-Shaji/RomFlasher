package tui

import (
	"flashtool/internal/core"
	"flashtool/internal/platform"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

/* ───────────────────────────────
   MAIN VIEW
─────────────────────────────── */
func (m AppModel) View() string {
	if m.Width == 0 {
		return "Initializing..."
	}

	// Layout dimensions
	header := renderHeader(m)
	status := renderStatusBar(m, m.Width)

	menuW, detailW, mainH, _, _ := m.GetLayoutDimensions()
	if mainH < 10 {
		mainH = 10
	}

	menu := renderMenu(m, menuW, mainH)
	details := renderDetails(m, detailW, mainH)

	// Assemble Layout
	main := lipgloss.JoinHorizontal(lipgloss.Top, menu, details)
	content := lipgloss.JoinVertical(lipgloss.Left, header, main, status)

	// Apply top padding for Windows Terminal clipping
	paddedContent := lipgloss.NewStyle().PaddingTop(2).Render(content)

	finalView := lipgloss.Place(
		m.Width, m.Height,
		lipgloss.Left, lipgloss.Top,
		paddedContent,
		lipgloss.WithWhitespaceForeground(CurrentTheme.Background),
	)

	// Overlay Modal
	if m.ActiveModal != ModalNone && m.ActiveModal != ModalConfirm {
		modal := renderModal(m)
		return lipgloss.Place(m.Width, m.Height,
			lipgloss.Center, lipgloss.Center,
			modal,
			lipgloss.WithWhitespaceForeground(CurrentTheme.Dim),
		)
	}

	return finalView
}

/* ───────────────────────────────
   HEADER
─────────────────────────────── */

func renderHeader(m AppModel) string {
	// Title
	line1 := "█ █ █▀█ █ ▀▀█▄   █▀▀ █   █▀█ █▀▀ █  █ █▀▀ █▀█"
	line2 := "▀▄▀ █▄█ █ █  █   █▀  █   █▀█ ▀▀█ █▀▀█ █▀  █▀▄"
	line3 := " ▀  ▀▀▀ ▀ ▀▀▀    ▀   ▀▀▀ ▀ ▀ ▀▀▀ ▀  ▀ ▀▀▀ ▀ ▀"

	t1 := lipgloss.NewStyle().Foreground(CurrentTheme.Title).Bold(true).Render(line1)
	t2 := lipgloss.NewStyle().Foreground(CurrentTheme.Accent).Bold(true).Render(line2)
	t3 := lipgloss.NewStyle().Foreground(CurrentTheme.Highlight).Bold(true).Render(line3)

	// Badges
	badge := BadgeStyle.Copy()

	metadata := lipgloss.JoinHorizontal(lipgloss.Center,
		badge.Copy().Background(CurrentTheme.Accent).Foreground(CurrentTheme.Background).Render(" VOID "),
		badge.Copy().Background(CurrentTheme.Highlight).Foreground(CurrentTheme.Background).Render(" PRIME v1.2 "),
		badge.Render(" ENGINE: ONLINE "),
	)

	// HUD
	banner := lipgloss.JoinVertical(lipgloss.Center,
		t1, t2, t3,
		"",
		metadata,
	)

	centered := lipgloss.NewStyle().Width(m.Width).Align(lipgloss.Center).Padding(1, 0).Render(banner)

	// Separator
	sepLine := strings.Repeat("━", m.Width/2)
	sepStyle1 := lipgloss.NewStyle().Foreground(CurrentTheme.Title).Render(sepLine)
	sepStyle2 := lipgloss.NewStyle().Foreground(CurrentTheme.Highlight).Render(sepLine)
	sep := lipgloss.JoinHorizontal(lipgloss.Top, sepStyle1, sepStyle2)

	return lipgloss.JoinVertical(lipgloss.Left, centered, sep)
}

/* ───────────────────────────────
   MENU PANEL
 ─────────────────────────────── */

func renderMenu(m AppModel, width, height int) string {
	var b strings.Builder

	// Menu Title
	header := lipgloss.NewStyle().
		Background(CurrentTheme.Accent).Foreground(CurrentTheme.Background).
		Width(width - 2).Align(lipgloss.Center).Bold(true).
		Render(" ❯❯ COMMANDS ")

	b.WriteString(header)
	b.WriteByte('\n')
	b.WriteString(SeparatorStyle.Copy().Render(strings.Repeat("─", width-2)))
	b.WriteByte('\n')

	// Menu Items
	itemStyle := BaseStyle.Copy().Width(width - 2)
	selStyle := SelectedStyle.Copy().Width(width - 2).Foreground(CurrentTheme.Background).Background(CurrentTheme.Highlight)

	for i := 0; i < len(m.Menu); i++ {
		it := m.Menu[i]
		prefix, style := "  ", itemStyle
		if i == m.Selection {
			prefix, style = " ❯ ", selStyle
		}
		b.WriteString(style.Render(fmt.Sprintf("%s%s %s", prefix, it.Icon, it.Label)))
		b.WriteByte('\n')
	}

	// Vertical Fill
	innerH := height - 2
	filledLines := 2 + len(m.Menu)

	fillCount := innerH - filledLines - 1
	if fillCount > 0 {
		b.WriteString(strings.Repeat("\n", fillCount))
	}

	footer := DimStyle.Copy().Width(width - 2).Align(lipgloss.Center).Render("↑/↓ Nav • ↵ Run")
	b.WriteString(footer)

	return BorderStyle.Copy().BorderForeground(CurrentTheme.Accent).Width(width - 2).Height(height).Render(b.String())
}

/* ───────────────────────────────
   DETAILS PANEL
 ─────────────────────────────── */

func renderDetails(m AppModel, width, height int) string {
	var body string
	menuW, detailW, mainH, _, logH := m.GetLayoutDimensions()
	_ = menuW
	_ = detailW
	_ = mainH

	switch {
	case m.ActiveModal == ModalConfirm:
		body = renderConfirmView(m, width)
	case m.Busy:
		body = renderBusyView(m, width)
	case m.Device.Mode == platform.ModeUnauthorized:
		body = renderUnauthorizedView(width)
	case m.Device.Mode != platform.ModeDisconnected && m.Device.Mode != platform.ModeOffline:
		body = renderDeviceHUD(m, width)
	default:
		body = renderInfoView(m, width)
	}

	// Live Logs Section
	logTitle := lipgloss.NewStyle().Background(CurrentTheme.Dim).Foreground(CurrentTheme.Foreground).
		Width(width - 2).Align(lipgloss.Center).Bold(true).Render(" ⚡ LIVE LOGSTREAM ")

	m.UI.Viewport.Width = detailW - 2
	m.UI.Viewport.Height = logH

	content := lipgloss.JoinVertical(lipgloss.Left, body, "", logTitle, m.UI.Viewport.View())

	return BorderStyle.Copy().BorderForeground(CurrentTheme.Highlight).Width(width - 2).Height(height).Render(content)
}

/* ───────────────────────────────
   STATUS BAR
 ─────────────────────────────── */

func renderStatusBar(m AppModel, width int) string {
	dot, dotStyle := "●", lipgloss.NewStyle()
	switch m.Tick % 6 {
	case 0, 5:
		dot, dotStyle = "○", lipgloss.NewStyle().Foreground(CurrentTheme.Border)
	case 1, 4:
		dotStyle = lipgloss.NewStyle().Foreground(CurrentTheme.Highlight).Bold(true)
	case 2, 3:
		dotStyle = lipgloss.NewStyle().Foreground(CurrentTheme.Title).Bold(true)
	}

	heartbeat := lipgloss.NewStyle().Padding(0, 1).Background(CurrentTheme.Dim).Render(dotStyle.Render(dot))
	accent := CurrentTheme.Accent

	modeBg, modeFg, modeIcon := CurrentTheme.Dim, CurrentTheme.Foreground, "⚠"
	switch m.Device.Mode {
	case platform.ModeFastboot:
		modeBg, modeFg, modeIcon = CurrentTheme.Highlight, CurrentTheme.Background, "⚡"
	case platform.ModeRecovery, platform.ModeDevice, platform.ModeSideload:
		modeBg, modeFg, modeIcon = CurrentTheme.Success, CurrentTheme.Background, "📱"
	case platform.ModeUnauthorized:
		modeBg, modeFg, modeIcon = CurrentTheme.Error, CurrentTheme.Background, "📵"
	case platform.ModeOffline:
		modeBg, modeFg, modeIcon = CurrentTheme.Dim, CurrentTheme.Foreground, "💤"
	}

	tri := ""
	statusStyle := lipgloss.NewStyle().Padding(0, 1)

	left := lipgloss.JoinHorizontal(lipgloss.Left,
		heartbeat,
		lipgloss.NewStyle().Foreground(CurrentTheme.Dim).Background(accent).Render(tri),
		lipgloss.NewStyle().Background(accent).Foreground(CurrentTheme.Background).Bold(true).Render(" VOID "),
		lipgloss.NewStyle().Foreground(accent).Background(CurrentTheme.StatusBg).Render(tri),
		statusStyle.Background(CurrentTheme.StatusBg).Foreground(CurrentTheme.StatusFg).Bold(true).Render("SYSTEM"),
		lipgloss.NewStyle().Foreground(CurrentTheme.StatusBg).Background(modeBg).Render(tri),
		statusStyle.Background(modeBg).Foreground(modeFg).Bold(true).Render(fmt.Sprintf("%s %s", modeIcon, m.Device.Mode)),
		lipgloss.NewStyle().Foreground(modeBg).Background(CurrentTheme.Dim).Render(tri),
		statusStyle.Background(CurrentTheme.Dim).Foreground(CurrentTheme.Foreground).Render(m.Device.Serial),
		lipgloss.NewStyle().Foreground(CurrentTheme.Dim).Background(CurrentTheme.Background).Render(tri),
	)

	right := ""
	if m.ActiveToast != nil {
		right = renderToast(m.ActiveToast)
	}

	space := width - lipgloss.Width(left) - lipgloss.Width(right)
	if space < 0 {
		space = 0
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", space), right)
}

func renderToast(t *Toast) string {
	bg, fg := CurrentTheme.Accent, CurrentTheme.Background
	accent := CurrentTheme.Highlight
	icon := "💠"
	label := "INFO"

	switch t.Type {
	case core.LogError:
		bg, icon, label = CurrentTheme.Error, "❌", "ERROR"
	case core.LogSuccess:
		bg, icon, label = CurrentTheme.Success, "✅", "SUCCESS"
	}

	tri := ""
	style := lipgloss.NewStyle().Padding(0, 1)

	// Label Segment (Main Color)
	left := style.Background(bg).Foreground(fg).Bold(true).Render(icon + " " + label)

	// Accent Separator
	mid := lipgloss.NewStyle().Foreground(bg).Background(accent).Render(tri)
	bar := lipgloss.NewStyle().Background(accent).Foreground(CurrentTheme.Background).Bold(true).Render("⚡")
	sep := lipgloss.NewStyle().Foreground(accent).Background(CurrentTheme.StatusBg).Render(tri)

	// Message Segment
	right := style.Background(CurrentTheme.StatusBg).Foreground(CurrentTheme.StatusFg).Render(t.Message)

	return lipgloss.JoinHorizontal(lipgloss.Left, left, mid, bar, sep, right)
}

var (
	confirmTitleStyle = TitleStyle.Copy().Background(CurrentTheme.Warning).Foreground(CurrentTheme.Background).Align(lipgloss.Center)
	confirmMsgStyle   = TitleStyle.Copy().Foreground(CurrentTheme.Highlight).Align(lipgloss.Center)
	confirmWarnStyle  = DimStyle.Copy().Align(lipgloss.Center)
	confirmOptsStyle  = BaseStyle.Copy().Bold(true)
	confirmBoxStyle   = BorderStyle.Copy().BorderForeground(CurrentTheme.Warning).Padding(1, 1)

	unauthHeaderStyle = lipgloss.NewStyle().Background(CurrentTheme.Warning).Foreground(CurrentTheme.Background).Align(lipgloss.Center).Bold(true)
	unauthMsgStyle    = TitleStyle.Copy().Foreground(CurrentTheme.Warning).Align(lipgloss.Center)
	unauthHelpStyle   = BaseStyle.Copy().Align(lipgloss.Center)

	hudStatusHeader = lipgloss.NewStyle().Background(CurrentTheme.Highlight).Foreground(CurrentTheme.Background).Bold(true).Align(lipgloss.Center)
	hudInfoBox      = SeparatorStyle.Copy().BorderStyle(lipgloss.NormalBorder()).Padding(0, 1)
	hudDescHeader   = TitleStyle.Copy().Background(CurrentTheme.Dim).Align(lipgloss.Center)
	hudDesc         = BaseStyle.Copy()

	infoHeader = lipgloss.NewStyle().Background(CurrentTheme.StatusBg).Foreground(CurrentTheme.StatusFg).Align(lipgloss.Center).Bold(true)
	infoDesc   = BaseStyle.Copy().Padding(1, 1)
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
	lblStyle, valStyle := HUDLabelStyle, HUDValueStyle

	statusHeader := hudStatusHeader.Copy().Width(width - 2).Render("DEVICE STATUS")

	row := lipgloss.JoinHorizontal(lipgloss.Left,
		lblStyle.Render(" MDL "), valStyle.Render(fmt.Sprintf("%-12s", m.Device.Model)),
		"   ",
		lblStyle.Render(" BAT "), valStyle.Render(fmt.Sprintf("%-9s", m.Device.Battery)),
	)
	if m.Device.Mode == platform.ModeFastboot {
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

func renderModal(m AppModel) string {
	w := int(float64(m.Width) * 0.8)
	if w < 60 {
		w = 60
	}
	if w > 120 {
		w = 120
	}

	h := 20
	if m.ActiveModal == ModalHelp {
		h = 26
	}
	if m.ActiveModal == ModalCustom {
		h = 24
	}
	if m.ActiveModal == ModalSettings {
		h = 22
	}

	style := BorderStyle.Copy().Width(w).Height(h).Border(lipgloss.RoundedBorder()).BorderForeground(CurrentTheme.Highlight)

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
	title := TitleStyle.Copy().Width(w - 4).Align(lipgloss.Center).Render(m.Modal.FileTitle)
	var b strings.Builder

	// Path Box
	pathLabel := lipgloss.NewStyle().Foreground(CurrentTheme.Background).Background(CurrentTheme.Highlight).Padding(0, 1).Bold(true).Render(" PATH ")
	pathText := lipgloss.NewStyle().Foreground(CurrentTheme.Highlight).Italic(true).Render(" " + m.Modal.FileDir)
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, pathLabel, pathText))
	b.WriteString("\n\n")

	// Search Box
	searchLabel := lipgloss.NewStyle().Padding(0, 1).Background(CurrentTheme.Accent).Foreground(CurrentTheme.Background).Bold(true).Render(" 🔍 SEARCH ")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, searchLabel, " ", m.UI.TextInput.View()))
	b.WriteString("\n\n")

	start := 0
	if m.Modal.FileCursor > 8 {
		start = m.Modal.FileCursor - 8
	}
	end := start + 10
	if end > len(m.Modal.FileList) {
		end = len(m.Modal.FileList)
	}
	for i := start; i < end; i++ {
		f := m.Modal.FileList[i]
		icon := "📄"
		if f.IsDir {
			icon = "📁"
		}
		if f.Name == "[ SELECT THIS FOLDER ]" {
			icon = "✅"
		}
		name := f.Name
		if len(name) > 40 {
			name = name[:37] + "..."
		}
		if i == m.Modal.FileCursor {
			barWidth := w - 8
			b.WriteString(SelectedStyle.Copy().Width(barWidth).Background(CurrentTheme.Highlight).Foreground(CurrentTheme.Background).Render(fmt.Sprintf(" ❯ %s  %s", icon, name)))
			b.WriteByte('\n')
		} else {
			s := BaseStyle.Copy()
			if f.Name == "[ SELECT THIS FOLDER ]" {
				s = s.Foreground(CurrentTheme.Highlight).Bold(true)
			}
			b.WriteString(fmt.Sprintf("   %s  %s\n", icon, s.Render(name)))
		}
	}
	for i := 0; i < (10 - (end - start)); i++ {
		b.WriteByte('\n')
	}
	fLeft := DimStyle.Copy().Render(fmt.Sprintf(" %d items match", len(m.Modal.FileList)))
	fRight := DimStyle.Copy().Render("↑/↓ Nav • ↵ Open • Esc Back ")
	space := w - lipgloss.Width(fLeft) - lipgloss.Width(fRight) - 4
	b.WriteString(fmt.Sprintf("\n%s%s%s", fLeft, strings.Repeat(" ", space), fRight))
	return lipgloss.JoinVertical(lipgloss.Center, title, "", lipgloss.NewStyle().Width(w-4).Padding(0, 2).Render(b.String()))
}

func renderCustomModal(m AppModel, w, _ int) string {
	title := TitleStyle.Copy().
		Background(CurrentTheme.Highlight).
		Foreground(CurrentTheme.Background).
		Width(w - 4).
		Align(lipgloss.Center).
		Render(" COMMAND CONSOLE ")
	var b strings.Builder
	innerW := w - 6
	// Output Panel
	outputTitle := lipgloss.NewStyle().Foreground(CurrentTheme.Highlight).Bold(true).Render("🖥️  OUTPUT")
	b.WriteString(outputTitle)
	b.WriteByte('\n')

	m.Modal.CustomViewport.Width = innerW
	m.Modal.CustomViewport.Height = 11

	outputContent := m.Modal.CustomViewport.View()
	if m.Modal.CustomLogs.Len() == 0 && !m.Busy {
		outputContent = DimStyle.Copy().Italic(true).Render("\n\n  Terminal initialized. Enter command below...")
	}

	outputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Dim).
		Width(innerW).
		Height(11).
		Padding(0, 1).
		Render(outputContent)
	b.WriteString(outputBox)
	b.WriteByte('\n')

	// Input Panel
	inputLabel := lipgloss.NewStyle().Foreground(CurrentTheme.Background).Background(CurrentTheme.Accent).Bold(true).Padding(0, 1).Render(" COMMAND ")
	m.UI.TextInput.Prompt = lipgloss.NewStyle().Foreground(CurrentTheme.Accent).Render(" ❯ ")
	inputField := m.UI.TextInput.View()
	if m.Busy {
		inputField = DimStyle.Copy().Render("Executing command... ⚡")
	}

	inputLine := lipgloss.JoinHorizontal(lipgloss.Left, inputLabel, " ", inputField)
	inputBox := lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e")).Width(innerW).Padding(0, 1).Render(inputLine)
	b.WriteString(inputBox)
	b.WriteString("\n\n")

	// Footer
	examples := DimStyle.Copy().Render(" Try: 'adb shell getprop' or 'fastboot getvar all'")
	if m.Modal.CustomLogs.Len() > 0 || m.Busy {
		examples = DimStyle.Copy().Render(fmt.Sprintf(" History: %d lines", m.Modal.CustomLogs.Len()))
	}

	fLeft, fRight := examples, DimStyle.Copy().Render("↵ EXECUTE  •  ESC EXIT ")
	spaceCount := w - lipgloss.Width(fLeft) - lipgloss.Width(fRight) - 8
	if spaceCount < 0 {
		spaceCount = 0
	}

	footer := lipgloss.JoinHorizontal(lipgloss.Bottom, fLeft, strings.Repeat(" ", spaceCount), fRight)
	b.WriteString(footer)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", lipgloss.NewStyle().Padding(0, 2).Render(b.String()))
}

func renderHelpModal(_ AppModel, w int) string {
	title := TitleStyle.Copy().Width(w - 4).Align(lipgloss.Center).Render("SYSTEM DOCUMENTATION")
	hl := TitleStyle.Copy().Foreground(CurrentTheme.Highlight).Width(w - 8).Align(lipgloss.Center)
	row := func(k, v string) string {
		return fmt.Sprintf(" %s %s\n", lipgloss.NewStyle().Width(18).Foreground(CurrentTheme.Highlight).Render(k), v)
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

	b.WriteString(fmt.Sprintf("\n%s", DimStyle.Copy().Width(w-8).Align(lipgloss.Center).Render("Build: stable-1.2.0 • Pro Flasher Core")))

	return lipgloss.JoinVertical(lipgloss.Center, title, "", lipgloss.NewStyle().Width(w-4).Padding(0, 2).Render(b.String()))
}

func renderSettingsModal(m AppModel, w, _ int) string {
	title := TitleStyle.Copy().Width(w - 4).Align(lipgloss.Center).Render("APPLICATION CONFIGURATION")
	var b strings.Builder
	innerW := w - 6

	renderItem := func(index int, label, desc, pathVal string) {
		b.WriteString(fmt.Sprintf(" %s\n", TitleStyle.Copy().Foreground(CurrentTheme.Highlight).Render(label)))
		b.WriteString(fmt.Sprintf("  %s\n", DimStyle.Copy().Render(desc)))

		pad := "  "
		if m.Modal.SettingsCursor == index {
			pad = lipgloss.NewStyle().Foreground(CurrentTheme.Accent).Render("❯ ")
		}

		val := pathVal
		if val == "" {
			val = "(NOT SET)"
		}

		pathStr := lipgloss.NewStyle().Foreground(CurrentTheme.Foreground).Render(val)
		b.WriteString(fmt.Sprintf("%s%s\n\n", pad, pathStr))
	}

	renderItem(0, "📂 Base ROM Directory", "The root folder where your custom ROMs and images are stored.", m.App.Config.BaseDir)
	renderItem(1, "📱 Target Device Folder", "Default folder structure path for the current device.", m.App.Config.DevicePath)

	for i := 0; i < 1; i++ {
		b.WriteByte('\n')
	}

	saveLabel := "  [ SAVE AND APPLY CONFIGURATION ]  "
	saveStyle := lipgloss.NewStyle().Foreground(CurrentTheme.Dim)
	if m.Modal.SettingsCursor == 2 {
		saveStyle = lipgloss.NewStyle().Foreground(CurrentTheme.Background).Background(CurrentTheme.Accent).Bold(true)
	}
	saveBtn := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Center).Render(saveStyle.Render(saveLabel))
	b.WriteString(saveBtn)
	b.WriteString("\n\n")

	fLeft, fRight := DimStyle.Copy().Render(" TAB Nav • ↵ SELECT "), DimStyle.Copy().Render(" ESC CANCEL ")
	spaceCount := w - lipgloss.Width(fLeft) - lipgloss.Width(fRight) - 8
	if spaceCount < 0 {
		spaceCount = 0
	}
	footer := lipgloss.JoinHorizontal(lipgloss.Bottom, fLeft, strings.Repeat(" ", spaceCount), fRight)
	b.WriteString(footer)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", lipgloss.NewStyle().Padding(0, 2).Render(b.String()))
}
