package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// menuItem implements list.Item
type menuItem struct {
	title       string
	description string
	action      string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.description }
func (i menuItem) FilterValue() string { return i.title }

// Model represents the Bubble Tea state for the dashboard.
type Model struct {
	DeviceCodename string
	DeviceMode     string
	Battery        string
	Unlocked       bool

	Menu   list.Model
	Logs   []string
	Status string

	Width  int
	Height int
}

// NewModel creates an initial UI state.
func NewModel() Model {
	items := []list.Item{
		menuItem{title: "💾 Flash Recovery", description: "Flash recovery.img to recovery partition", action: "flash_recovery"},
		menuItem{title: "👢 Flash Boot", description: "Flash boot.img to boot partition", action: "flash_boot"},
		menuItem{title: "🧹 Wipe Super", description: "Wipe dynamic super partition", action: "wipe_super"},
		menuItem{title: "📲 ADB Sideload", description: "Sideload a ROM zip via ADB", action: "adb_sideload"},
		menuItem{title: "🔄 Reboot System", description: "Reboot device to Android OS", action: "reboot_system"},
		menuItem{title: "🔄 Reboot Recovery", description: "Reboot device to Recovery mode", action: "reboot_recovery"},
		menuItem{title: "⚡ Refresh Status", description: "Scan and refresh device connection", action: "refresh_status"},
		menuItem{title: "⌨️ Custom Command", description: "Execute a raw ADB/Fastboot command", action: "custom_command"},
		menuItem{title: "⚙️ App Settings", description: "Configure VoidFlasher PRIME", action: "app_settings"},
		menuItem{title: "❓ Help", description: "View documentation and help", action: "help"},
		menuItem{title: "❌ Exit", description: "Close VoidFlasher PRIME", action: "exit"},
	}

	m := list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.Title = "Action Menu"
	m.SetShowStatusBar(false)
	m.SetFilteringEnabled(false)

	return Model{
		DeviceCodename: "Scanning...",
		DeviceMode:     "Unknown",
		Battery:        "?",
		Unlocked:       false,
		Menu:           m,
		Logs:           []string{"[SYSTEM] VoidFlasher PRIME initialized.", "[SYSTEM] Awaiting device connection..."},
		Status:         "Ready",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			// Handle menu selection
			selected, ok := m.Menu.SelectedItem().(menuItem)
			if ok {
				m.Logs = append(m.Logs, fmt.Sprintf("[USER] Triggered action: %s", selected.title))
				m.Status = "Executing " + selected.title
				// Here we would normally dispatch a command to the flasher engine
			}
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		
		// Update menu dimensions dynamically
		menuWidth := (m.Width / 2) - 4
		if menuWidth < 30 {
			menuWidth = 30
		}
		m.Menu.SetSize(menuWidth, 12)
	}

	// Route updates to the list model (for arrow keys, scrolling, etc)
	m.Menu, cmd = m.Menu.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.Width == 0 {
		return "Initializing..."
	}

	header := StyleHeader.Width(m.Width - 4).Render("VoidFlasher PRIME | Cyberpunk Edition")

	// Device Info Panel (Left)
	devInfo := fmt.Sprintf("Codename: %s\nMode: %s\nBattery: %s\nUnlocked: %v",
		StyleNeon(m.DeviceCodename), StyleNeon(m.DeviceMode), m.Battery, m.Unlocked)
	devPanelWidth := (m.Width / 2) - 4
	if devPanelWidth < 30 {
		devPanelWidth = 30
	}
	devPanel := StylePanel.Width(devPanelWidth).Height(12).Render("Device Info\n" + devInfo)

	// Interactive Menu (Right)
	menuView := StylePanel.Width(devPanelWidth).Height(12).Render(m.Menu.View())

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, devPanel, menuView)

	// Logs Panel
	logLines := len(m.Logs)
	maxLogLines := m.Height - 20
	if maxLogLines < 1 {
		maxLogLines = 1
	}
	startIdx := 0
	if logLines > maxLogLines {
		startIdx = logLines - maxLogLines
	}

	logText := strings.Join(m.Logs[startIdx:], "\n")
	logPanel := StylePanel.Width(m.Width - 4).Height(maxLogLines + 2).Render("Live Logs:\n" + logText)

	// Status Bar
	statusLine := StyleStatus.Width(m.Width).Render(fmt.Sprintf("Status: %s | Press 'q' to quit", m.Status))

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		topRow,
		logPanel,
		statusLine,
	)
}

func StyleNeon(text string) string {
	return lipgloss.NewStyle().Foreground(ColorNeonCyan).Render(text)
}
