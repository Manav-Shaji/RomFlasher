package tui

import (
	"flashtool/internal/domain"
	"flashtool/internal/engine"
	"flashtool/internal/platform/adb"

	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"flashtool/internal/tui/theme"
)

/* MESSAGES */

type ToastTimeoutMsg struct{}
type SetupMainDirMsg string
type SetupDeviceDirMsg string
type LogTickMsg time.Time

type SetupConfirmMsg struct {
	Msg string
	Cmd tea.Cmd
}

type SettingsFolderSelectedMsg struct {
	Index int
	Path  string
}

type SearchDebounceMsg struct {
	Query string
}

/* INIT */

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		adb.PollDeviceCmd(),
		m.App.Engine.WaitForLogs(),
		LogTickCmd(),
		textinput.Blink,
	)
}

func LogTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return LogTickMsg(t)
	})
}

/* UPDATE */

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case adb.HeartbeatMsg:
		m.Tick++
		return m, adb.HeartbeatCmd()

	case LogTickMsg:
		if m.LogsDirty {
			if m.ActiveModal == ModalCustom {
				innerW := m.Modal.CustomViewport.Width
				m.Modal.CustomViewport.SetContent(RenderLogsStr(m.Modal.CustomLogs, innerW))
				m.Modal.CustomViewport.GotoBottom()
			} else {
				_, detailW, _, _, logH := m.GetLayoutDimensions()
				m.UI.Viewport.Width = detailW - 2
				m.UI.Viewport.Height = logH
				m.UI.Viewport.SetContent(RenderLogsStr(m.Logs, m.UI.Viewport.Width))
				m.UI.Viewport.GotoBottom()
			}
			m.LogsDirty = false
		}
		return m, LogTickCmd()

	case adb.PollMsg:
		return m.handlePollMsg(msg)

	case adb.DeviceUpdateMsg:
		return m.handleDeviceUpdate(msg)

	case engine.LogMsg:
		return m.handleLogMsg(msg)

	case engine.TaskCompleteMsg:
		return m.handleTaskComplete(msg)

	case ToastTimeoutMsg:
		m.ActiveToast = nil
		return m, nil

	case SearchDebounceMsg:
		if m.ActiveModal == ModalFile && m.UI.TextInput.Value() == msg.Query {
			val := strings.ToLower(msg.Query)
			if val != "" {
				var filtered []FileItem
				for _, f := range m.Modal.FullFileList {
					if strings.Contains(strings.ToLower(f.Name), val) {
						filtered = append(filtered, f)
					}
				}
				m.Modal.FileList = filtered
			} else {
				m.Modal.FileList = m.Modal.FullFileList
			}
			m.Modal.FileCursor = 0
		}
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

var (
	baseLogStyle    = lipgloss.NewStyle()
	errorLogStyle   = lipgloss.NewStyle().Bold(true)
	successLogStyle = lipgloss.NewStyle().Bold(true)
	adbStyle        = lipgloss.NewStyle().Bold(true)
	fastbootStyle   = lipgloss.NewStyle().Bold(true)
	actionStyle     = lipgloss.NewStyle().Bold(true)
	doneStyle       = lipgloss.NewStyle().Bold(true)
	failedStyle     = lipgloss.NewStyle().Bold(true)
)

func RenderLogsStr(logs *LogBuffer, width int) string {
	var b strings.Builder
	
	if logs == nil {
		return ""
	}

	// Pre-apply theme colors to the package-level styles once per render (in case theme changed)
	baseLogStyle = baseLogStyle.Foreground(theme.CurrentTheme.Foreground).Width(width).PaddingRight(1)
	errorLogStyle = errorLogStyle.Foreground(theme.CurrentTheme.Error)
	successLogStyle = successLogStyle.Foreground(theme.CurrentTheme.Success)
	adbStyle = adbStyle.Foreground(theme.CurrentTheme.Highlight)
	fastbootStyle = fastbootStyle.Foreground(theme.CurrentTheme.Accent)
	actionStyle = actionStyle.Foreground(theme.CurrentTheme.Warning)
	doneStyle = doneStyle.Foreground(theme.CurrentTheme.Success)
	failedStyle = failedStyle.Foreground(theme.CurrentTheme.Error)

	logs.Iterate(func(l domain.LogEntry) {
		style := baseLogStyle
		text := l.Text
		
		// 1. Level-based styling
		switch l.Level {
		case domain.LogError:
			style = baseLogStyle.Inherit(errorLogStyle)
		case domain.LogSuccess:
			style = baseLogStyle.Inherit(successLogStyle)
		}

		// 2. Keyword Highlighting
		if strings.HasPrefix(text, ">") {
			cmdPart := text
			if strings.Contains(text, "adb") {
				cmdPart = strings.Replace(text, "adb", adbStyle.Render("adb"), 1)
			} else if strings.Contains(text, "fastboot") {
				cmdPart = strings.Replace(text, "fastboot", fastbootStyle.Render("fastboot"), 1)
			}
			
			// Highlight actions
			for _, action := range []string{"flash", "sideload", "wipe-super", "reboot"} {
				if strings.Contains(cmdPart, action) {
					coloredAction := actionStyle.Render(action)
					cmdPart = strings.Replace(cmdPart, action, coloredAction, 1)
				}
			}
			text = cmdPart
		}

		// 3. Status Highlights
		if strings.Contains(text, "[ DONE ]") {
			text = strings.Replace(text, "[ DONE ]", doneStyle.Render("[ DONE ]"), 1)
		} else if strings.Contains(text, "[ FAILED") {
			text = strings.Replace(text, "[ FAILED", failedStyle.Render("[ FAILED"), 1)
		}

		b.WriteString(style.Render(text))
		b.WriteByte('\n')
	})
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
