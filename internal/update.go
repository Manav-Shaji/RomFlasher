package internal

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"flashtool/internal/ui"
)

/* MESSAGES */

type ToastTimeoutMsg struct{}
type SetupMainDirMsg string
type SetupDeviceDirMsg string

type SetupConfirmMsg struct {
	Msg string
	Cmd tea.Cmd
}

type SettingsFolderSelectedMsg struct {
	Index int
	Path  string
}

/* INIT */

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		PollDeviceCmd(),
		WaitForLogs(LogChan),
		textinput.Blink,
	)
}

/* UPDATE */

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case HeartbeatMsg:
		m.Tick++
		return m, HeartbeatCmd()

	case PollMsg:
		return m.handlePollMsg(msg)

	case DeviceUpdateMsg:
		return m.handleDeviceUpdate(msg)

	case LogMsg:
		return m.handleLogMsg(msg)

	case TaskCompleteMsg:

		return m.handleTaskComplete(msg)

	case ToastTimeoutMsg:
		m.ActiveToast = nil
		return m, nil

	case SetupMainDirMsg, SetupDeviceDirMsg, SetupConfirmMsg, SettingsFolderSelectedMsg:
		return m.handleSetupMsg(msg)
	}

	return m, nil
}

func (m AppModel) GetLayoutDimensions() (menuW, detailW, mainH, bodyH, logH int) {
	if m.Width == 0 || m.Height == 0 { return }

	// 1. Header & Status heights
	// Verification: ASCII(3) + blank(1) + Meta(1) + Padding(1,0)=2 + Sep(1) = 8 lines
	headerH := 8 
	statusH := 1
	
	mainH = m.Height - headerH - statusH - 2
	if mainH < 10 { mainH = 10 }

	menuW = m.Width / 3
	if menuW < 25 { menuW = 25 } else if menuW > 40 { menuW = 40 }
	detailW = m.Width - menuW

	// 2. Body height (Active HUD/Info)
	bodyH = 6 // Default HUD height
	if m.Busy || m.ActiveModal == ModalConfirm {
		bodyH = 4 // Confirms/Busy are shorter
	}
	
	spacing := 3
	if m.Busy { spacing = 1 }

	logH = mainH - bodyH - spacing

	if logH < 0 { logH = 0 }

	return
}

func toastTimeoutCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return ToastTimeoutMsg{} })
}

func RenderLogsStr(logs []LogEntry, width int) string {
	var b strings.Builder
	for _, l := range logs {
		style := lipgloss.NewStyle().
			Foreground(ui.CurrentTheme.Foreground).
			Width(width).
			PaddingRight(1)
		
		text := l.Text
		
		// 1. Level-based styling
		if l.Level == LogError {
			style = style.Foreground(ui.CurrentTheme.Error).Bold(true)
		} else if l.Level == LogSuccess {
			style = style.Foreground(ui.CurrentTheme.Success).Bold(true)
		}

		// 2. Keyword Highlighting
		if strings.HasPrefix(text, ">") {
			// It's a command
			cmdPart := text
			if strings.Contains(text, "adb") {
				cmdPart = strings.Replace(text, "adb", lipgloss.NewStyle().Foreground(ui.CurrentTheme.Highlight).Bold(true).Render("adb"), 1)
			} else if strings.Contains(text, "fastboot") {
				cmdPart = strings.Replace(text, "fastboot", lipgloss.NewStyle().Foreground(ui.CurrentTheme.Accent).Bold(true).Render("fastboot"), 1)
			}
			
			// Highlight actions
			for _, action := range []string{"flash", "sideload", "wipe-super", "reboot"} {
				if strings.Contains(cmdPart, action) {
					coloredAction := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Warning).Bold(true).Render(action)
					cmdPart = strings.Replace(cmdPart, action, coloredAction, 1)
				}
			}
			text = cmdPart
		}

		// 3. Status Highlights
		if strings.Contains(text, "[ DONE ]") {
			text = strings.Replace(text, "[ DONE ]", lipgloss.NewStyle().Foreground(ui.CurrentTheme.Success).Bold(true).Render("[ DONE ]"), 1)
		} else if strings.Contains(text, "[ FAILED") {
			text = strings.Replace(text, "[ FAILED", lipgloss.NewStyle().Foreground(ui.CurrentTheme.Error).Bold(true).Render("[ FAILED"), 1)
		}

		b.WriteString(style.Render(text))
		b.WriteByte('\n')
	}
	return b.String()
}


func LoadFiles(dir, filter string) []FileItem {
	entries, _ := os.ReadDir(dir)
	items := make([]FileItem, 0, len(entries)+2)

	if filter == "" {
		items = append(items, FileItem{Name: "[ SELECT THIS FOLDER ]", Path: dir, IsDir: true})
	}

	if parent := filepath.Dir(dir); parent != dir {
		items = append(items, FileItem{Name: "..", Path: parent, IsDir: true})
	}

	filter = strings.ToLower(filter)
	for _, e := range entries {
		name := e.Name()
		if filter != "" && e.IsDir() { continue }
		if !e.IsDir() && filter != "" && !strings.HasSuffix(strings.ToLower(name), filter) { continue }
		
		items = append(items, FileItem{
			Name:  name,
			Path:  filepath.Join(dir, name),
			IsDir: e.IsDir(),
		})
	}
	return items
}
