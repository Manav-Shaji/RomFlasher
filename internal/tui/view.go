package tui

import (
	"fmt"
	"strings"
	"flashtool/internal/domain"
	"github.com/charmbracelet/lipgloss"
	"flashtool/internal/tui/theme"
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
	if mainH < 10 { mainH = 10 }

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
		lipgloss.WithWhitespaceForeground(theme.CurrentTheme.Background),
	)

	// Overlay Modal
	if m.ActiveModal != ModalNone && m.ActiveModal != ModalConfirm {
		modal := renderModal(m)
		return lipgloss.Place(m.Width, m.Height,
			lipgloss.Center, lipgloss.Center,
			modal,
			lipgloss.WithWhitespaceForeground(theme.CurrentTheme.Dim),
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

	t1 := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Title).Bold(true).Render(line1)
	t2 := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Accent).Bold(true).Render(line2)
	t3 := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Highlight).Bold(true).Render(line3)

	// Badges
	badge := theme.BadgeStyle.Copy()
	
	metadata := lipgloss.JoinHorizontal(lipgloss.Center,
		badge.Copy().Background(theme.CurrentTheme.Accent).Foreground(theme.CurrentTheme.Background).Render(" VOID "),
		badge.Copy().Background(theme.CurrentTheme.Highlight).Foreground(theme.CurrentTheme.Background).Render(" PRIME v1.2 "),
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
	sepStyle1 := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Title).Render(sepLine)
	sepStyle2 := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Highlight).Render(sepLine)
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
		Background(theme.CurrentTheme.Accent).Foreground(theme.CurrentTheme.Background).
		Width(width - 2).Align(lipgloss.Center).Bold(true).
		Render(" ❯❯ COMMANDS ")
	
	b.WriteString(header)
	b.WriteByte('\n')
	b.WriteString(theme.SeparatorStyle.Copy().Render(strings.Repeat("─", width-2)))
	b.WriteByte('\n')

	// Menu Items
	itemStyle := theme.BaseStyle.Copy().Width(width - 2)
	selStyle := theme.SelectedStyle.Copy().Width(width - 2).Foreground(theme.CurrentTheme.Background).Background(theme.CurrentTheme.Highlight)

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

	footer := theme.DimStyle.Copy().Width(width-2).Align(lipgloss.Center).Render("↑/↓ Nav • ↵ Run")
	b.WriteString(footer)

	return theme.BorderStyle.Copy().BorderForeground(theme.CurrentTheme.Accent).Width(width - 2).Height(height).Render(b.String())
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
	case m.Device.Mode == domain.ModeUnauthorized:
		body = renderUnauthorizedView(width)
	case m.Device.Mode != domain.ModeDisconnected && m.Device.Mode != domain.ModeOffline:
		body = renderDeviceHUD(m, width)
	default:
		body = renderInfoView(m, width)
	}

	// Live Logs Section
	logTitle := lipgloss.NewStyle().Background(theme.CurrentTheme.Dim).Foreground(theme.CurrentTheme.Foreground).
		Width(width - 2).Align(lipgloss.Center).Bold(true).Render(" ⚡ LIVE LOGSTREAM ")
	
	m.UI.Viewport.Width = detailW - 2
	m.UI.Viewport.Height = logH
	
	content := lipgloss.JoinVertical(lipgloss.Left, body, "", logTitle, m.UI.Viewport.View())
	
	return theme.BorderStyle.Copy().BorderForeground(theme.CurrentTheme.Highlight).Width(width - 2).Height(height).Render(content)
}

/* ───────────────────────────────
   STATUS BAR
 ─────────────────────────────── */

func renderStatusBar(m AppModel, width int) string {
	dot, dotStyle := "●", lipgloss.NewStyle()
	switch m.Tick % 6 {
	case 0, 5: dot, dotStyle = "○", lipgloss.NewStyle().Foreground(theme.CurrentTheme.Border)
	case 1, 4: dotStyle = lipgloss.NewStyle().Foreground(theme.CurrentTheme.Highlight).Bold(true)
	case 2, 3: dotStyle = lipgloss.NewStyle().Foreground(theme.CurrentTheme.Title).Bold(true)
	}
	
	heartbeat := lipgloss.NewStyle().Padding(0, 1).Background(theme.CurrentTheme.Dim).Render(dotStyle.Render(dot))
	accent := theme.CurrentTheme.Accent

	modeBg, modeFg, modeIcon := theme.CurrentTheme.Dim, theme.CurrentTheme.Foreground, "⚠"
	switch m.Device.Mode {
	case domain.ModeFastboot: 
		modeBg, modeFg, modeIcon = theme.CurrentTheme.Highlight, theme.CurrentTheme.Background, "⚡"
	case domain.ModeRecovery, domain.ModeDevice, domain.ModeSideload: 
		modeBg, modeFg, modeIcon = theme.CurrentTheme.Success, theme.CurrentTheme.Background, "📱"
	case domain.ModeUnauthorized:
		modeBg, modeFg, modeIcon = theme.CurrentTheme.Error, theme.CurrentTheme.Background, "📵"
	case domain.ModeOffline:
		modeBg, modeFg, modeIcon = theme.CurrentTheme.Dim, theme.CurrentTheme.Foreground, "💤"
	}

	tri := ""
	statusStyle := lipgloss.NewStyle().Padding(0, 1)

	left := lipgloss.JoinHorizontal(lipgloss.Left,
		heartbeat,
		lipgloss.NewStyle().Foreground(theme.CurrentTheme.Dim).Background(accent).Render(tri),
		lipgloss.NewStyle().Background(accent).Foreground(theme.CurrentTheme.Background).Bold(true).Render(" VOID "),
		lipgloss.NewStyle().Foreground(accent).Background(theme.CurrentTheme.StatusBg).Render(tri),
		statusStyle.Background(theme.CurrentTheme.StatusBg).Foreground(theme.CurrentTheme.StatusFg).Bold(true).Render("SYSTEM"),
		lipgloss.NewStyle().Foreground(theme.CurrentTheme.StatusBg).Background(modeBg).Render(tri),
		statusStyle.Background(modeBg).Foreground(modeFg).Bold(true).Render(fmt.Sprintf("%s %s", modeIcon, m.Device.Mode)),
		lipgloss.NewStyle().Foreground(modeBg).Background(theme.CurrentTheme.Dim).Render(tri),
		statusStyle.Background(theme.CurrentTheme.Dim).Foreground(theme.CurrentTheme.Foreground).Render(m.Device.Serial),
		lipgloss.NewStyle().Foreground(theme.CurrentTheme.Dim).Background(theme.CurrentTheme.Background).Render(tri),
	)


	right := ""
	if m.ActiveToast != nil { right = renderToast(m.ActiveToast) }

	space := width - lipgloss.Width(left) - lipgloss.Width(right)
	if space < 0 { space = 0 }

	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", space), right)
}

func renderToast(t *Toast) string {
	bg, fg := theme.CurrentTheme.Accent, theme.CurrentTheme.Background
	accent := theme.CurrentTheme.Highlight 
	icon := "💠"
	label := "INFO"

	switch t.Type {
	case domain.LogError:
		bg, icon, label = theme.CurrentTheme.Error, "❌", "ERROR"
	case domain.LogSuccess:
		bg, icon, label = theme.CurrentTheme.Success, "✅", "SUCCESS"
	}

	tri := ""
	style := lipgloss.NewStyle().Padding(0, 1)

	// Label Segment (Main Color)
	left := style.Background(bg).Foreground(fg).Bold(true).Render(icon + " " + label)
	
	// Accent Separator
	mid := lipgloss.NewStyle().Foreground(bg).Background(accent).Render(tri)
	bar := lipgloss.NewStyle().Background(accent).Foreground(theme.CurrentTheme.Background).Bold(true).Render("⚡")
	sep := lipgloss.NewStyle().Foreground(accent).Background(theme.CurrentTheme.StatusBg).Render(tri)

	// Message Segment
	right := style.Background(theme.CurrentTheme.StatusBg).Foreground(theme.CurrentTheme.StatusFg).Render(t.Message)

	return lipgloss.JoinHorizontal(lipgloss.Left, left, mid, bar, sep, right)
}

