package internal

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/lipgloss"
	"flashtool/internal/ui"
)

/* ───────────────────────────────
   MAIN VIEW
─────────────────────────────── */
func (m AppModel) View() string {
	if m.Width == 0 {
		return "Initializing..."
	}

	// 1. Derived Layout dimensions
	header := renderHeader(m)
	status := renderStatusBar(m, m.Width)
	
	menuW, detailW, mainH, _, _ := m.GetLayoutDimensions()
	if mainH < 10 { mainH = 10 }

	menu := renderMenu(m, menuW, mainH)
	details := renderDetails(m, detailW, mainH)

	// 3. Assemble Primary Layout
	main := lipgloss.JoinHorizontal(lipgloss.Top, menu, details)
	content := lipgloss.JoinVertical(lipgloss.Left, header, main, status)
	
	// Apply top padding for Windows Terminal clipping
	paddedContent := lipgloss.NewStyle().PaddingTop(2).Render(content)

	finalView := lipgloss.Place(
		m.Width, m.Height,
		lipgloss.Left, lipgloss.Top,
		paddedContent,
		lipgloss.WithWhitespaceForeground(ui.CurrentTheme.Background),
	)

	// 4. Overlay Modal (Skip if handled inline in details)
	if m.ActiveModal != ModalNone && m.ActiveModal != ModalConfirm {
		modal := renderModal(m)
		return lipgloss.Place(m.Width, m.Height,
			lipgloss.Center, lipgloss.Center,
			modal,
			lipgloss.WithWhitespaceForeground(ui.CurrentTheme.Dim),
		)
	}

	return finalView
}

/* ───────────────────────────────
   HEADER
─────────────────────────────── */

func renderHeader(m AppModel) string {
	// 1. Large Block ASCII Title with Color Accents
	line1 := "█ █ █▀█ █ ▀▀█▄   █▀▀ █   █▀█ █▀▀ █  █ █▀▀ █▀█"
	line2 := "▀▄▀ █▄█ █ █  █   █▀  █   █▀█ ▀▀█ █▀▀█ █▀  █▀▄"
	line3 := " ▀  ▀▀▀ ▀ ▀▀▀    ▀   ▀▀▀ ▀ ▀ ▀▀▀ ▀  ▀ ▀▀▀ ▀ ▀"

	t1 := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Title).Bold(true).Render(line1)
	t2 := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Accent).Bold(true).Render(line2)
	t3 := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Highlight).Bold(true).Render(line3)

	// 2. HUD Badges
	badge := ui.GetBadgeStyle()
	
	metadata := lipgloss.JoinHorizontal(lipgloss.Center,
		badge.Copy().Background(ui.CurrentTheme.Accent).Foreground(ui.CurrentTheme.Background).Render(" VOID "),
		badge.Copy().Background(ui.CurrentTheme.Highlight).Foreground(ui.CurrentTheme.Background).Render(" PRIME v1.2 "),
		badge.Render(" ENGINE: ONLINE "),
	)

	// 3. Compose HUD
	banner := lipgloss.JoinVertical(lipgloss.Center,
		t1, t2, t3,
		"",
		metadata,
	)

	centered := lipgloss.NewStyle().Width(m.Width).Align(lipgloss.Center).Padding(1, 0).Render(banner)
	
	// Gradient Separator (simulated)
	sepLine := strings.Repeat("━", m.Width/2)
	sepStyle1 := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Title).Render(sepLine)
	sepStyle2 := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Highlight).Render(sepLine)
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
		Background(ui.CurrentTheme.Accent).Foreground(ui.CurrentTheme.Background).
		Width(width - 2).Align(lipgloss.Center).Bold(true).
		Render(" ❯❯ COMMANDS ")
	
	b.WriteString(header + "\n")
	b.WriteString(ui.GetSeparatorStyle().Render(strings.Repeat("─", width-2)) + "\n")

	// Menu Items
	itemStyle := ui.GetBaseStyle().Width(width - 2)
	selStyle := ui.GetSelectedStyle().Width(width - 2).Foreground(ui.CurrentTheme.Background).Background(ui.CurrentTheme.Highlight)

	for i := 0; i < len(m.Menu); i++ {
		it := m.Menu[i]
		prefix, style := "  ", itemStyle
		if i == m.Selection {
			prefix, style = " ❯ ", selStyle
		}
		b.WriteString(style.Render(fmt.Sprintf("%s%s %s", prefix, it.Icon, it.Label)) + "\n")
	}

	// Vertical Fill
	innerH := height - 2
	filledLines := 2 + len(m.Menu)
	
	fillCount := innerH - filledLines - 1
	if fillCount > 0 {
		b.WriteString(strings.Repeat("\n", fillCount))
	}

	footer := ui.GetDimStyle().Width(width-2).Align(lipgloss.Center).Render("↑/↓ Nav • ↵ Run")
	b.WriteString(footer)

	return ui.GetBorderStyle().BorderForeground(ui.CurrentTheme.Accent).Width(width - 2).Height(height).Render(b.String())
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
	case m.Device.Mode == ModeUnauthorized:
		body = renderUnauthorizedView(width)
	case m.Device.Mode != ModeDisconnected && m.Device.Mode != ModeOffline:
		body = renderDeviceHUD(m, width)
	default:
		body = renderInfoView(m, width)
	}

	// Live Logs Section
	logTitle := lipgloss.NewStyle().Background(ui.CurrentTheme.Dim).Foreground(ui.CurrentTheme.Foreground).
		Width(width - 2).Align(lipgloss.Center).Bold(true).Render(" ⚡ LIVE LOGSTREAM ")
	
	m.UI.Viewport.Width = detailW - 2
	m.UI.Viewport.Height = logH
	
	content := lipgloss.JoinVertical(lipgloss.Left, body, "", logTitle, m.UI.Viewport.View())
	
	return ui.GetBorderStyle().BorderForeground(ui.CurrentTheme.Highlight).Width(width - 2).Height(height).Render(content)
}

/* ───────────────────────────────
   STATUS BAR
 ─────────────────────────────── */

func renderStatusBar(m AppModel, width int) string {
	dot, dotStyle := "●", lipgloss.NewStyle()
	switch m.Tick % 6 {
	case 0, 5: dot, dotStyle = "○", lipgloss.NewStyle().Foreground(ui.CurrentTheme.Border)
	case 1, 4: dotStyle = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Highlight).Bold(true)
	case 2, 3: dotStyle = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Title).Bold(true)
	}
	
	heartbeat := lipgloss.NewStyle().Padding(0, 1).Background(ui.CurrentTheme.Dim).Render(dotStyle.Render(dot))
	accent := ui.CurrentTheme.Accent

	modeBg, modeFg, modeIcon := ui.CurrentTheme.Dim, ui.CurrentTheme.Foreground, "⚠"
	switch m.Device.Mode {
	case ModeFastboot: 
		modeBg, modeFg, modeIcon = ui.CurrentTheme.Highlight, ui.CurrentTheme.Background, "⚡"
	case ModeRecovery, ModeDevice, ModeSideload: 
		modeBg, modeFg, modeIcon = ui.CurrentTheme.Success, ui.CurrentTheme.Background, "📱"
	case ModeUnauthorized:
		modeBg, modeFg, modeIcon = ui.CurrentTheme.Error, ui.CurrentTheme.Background, "📵"
	case ModeOffline:
		modeBg, modeFg, modeIcon = ui.CurrentTheme.Dim, ui.CurrentTheme.Foreground, "💤"
	}

	tri := ""
	statusStyle := lipgloss.NewStyle().Padding(0, 1)

	left := lipgloss.JoinHorizontal(lipgloss.Left,
		heartbeat,
		lipgloss.NewStyle().Foreground(ui.CurrentTheme.Dim).Background(accent).Render(tri),
		lipgloss.NewStyle().Background(accent).Foreground(ui.CurrentTheme.Background).Bold(true).Render(" VOID "),
		lipgloss.NewStyle().Foreground(accent).Background(ui.CurrentTheme.StatusBg).Render(tri),
		statusStyle.Background(ui.CurrentTheme.StatusBg).Foreground(ui.CurrentTheme.StatusFg).Bold(true).Render("SYSTEM"),
		lipgloss.NewStyle().Foreground(ui.CurrentTheme.StatusBg).Background(modeBg).Render(tri),
		statusStyle.Background(modeBg).Foreground(modeFg).Bold(true).Render(fmt.Sprintf("%s %s", modeIcon, m.Device.Mode)),
		lipgloss.NewStyle().Foreground(modeBg).Background(ui.CurrentTheme.Dim).Render(tri),
		statusStyle.Background(ui.CurrentTheme.Dim).Foreground(ui.CurrentTheme.Foreground).Render(m.Device.Serial),
		lipgloss.NewStyle().Foreground(ui.CurrentTheme.Dim).Background(ui.CurrentTheme.Background).Render(tri),
	)


	right := ""
	if m.ActiveToast != nil { right = renderToast(m.ActiveToast) }

	space := width - lipgloss.Width(left) - lipgloss.Width(right)
	if space < 0 { space = 0 }

	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", space), right)
}

func renderToast(t *Toast) string {
	bg, fg := ui.CurrentTheme.Accent, ui.CurrentTheme.Background
	accent := ui.CurrentTheme.Highlight 
	icon := "💠"
	label := "INFO"

	switch t.Type {
	case LogError:
		bg, icon, label = ui.CurrentTheme.Error, "❌", "ERROR"
	case LogSuccess:
		bg, icon, label = ui.CurrentTheme.Success, "✅", "SUCCESS"
	}

	tri := ""
	style := lipgloss.NewStyle().Padding(0, 1)

	// Label Segment (Main Color)
	left := style.Background(bg).Foreground(fg).Bold(true).Render(icon + " " + label)
	
	// Accent Separator
	mid := lipgloss.NewStyle().Foreground(bg).Background(accent).Render(tri)
	bar := lipgloss.NewStyle().Background(accent).Foreground(ui.CurrentTheme.Background).Bold(true).Render("⚡")
	sep := lipgloss.NewStyle().Foreground(accent).Background(ui.CurrentTheme.StatusBg).Render(tri)

	// Message Segment
	right := style.Background(ui.CurrentTheme.StatusBg).Foreground(ui.CurrentTheme.StatusFg).Render(t.Message)

	return lipgloss.JoinHorizontal(lipgloss.Left, left, mid, bar, sep, right)
}

