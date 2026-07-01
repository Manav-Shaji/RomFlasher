package tui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
)

func GetDefaultMenu() []MenuItem {
	return []MenuItem{
		{Label: "Flash Recovery", Icon: "💾", Desc: "Select & flash recovery.img (FASTBOOT)", Action: "flash_rec"},
		{Label: "Flash Boot", Icon: "👢", Desc: "Select & flash boot.img (FASTBOOT)", Action: "flash_boot"},
		{Label: "Wipe Super", Icon: "🧹", Desc: "Wipe super partition (FASTBOOT)", Action: "wipe_super"},
		{Label: "ADB Sideload", Icon: "📲", Desc: "Sideload ROM ZIP (ADB SIDELOAD)", Action: "sideload"},
		{Label: "Reboot System", Icon: "🔄", Desc: "Reboot to system", Action: "rb_system"},
		{Label: "Reboot Recovery", Icon: "🔄", Desc: "Reboot to recovery", Action: "rb_recovery"},
		{Label: "Refresh Status", Icon: "⚡", Desc: "Re-scan device connection", Action: "refresh"},
		{Label: "Custom Command", Icon: "⌨️", Desc: "Enter & execute arbitrary commands", Action: "custom"},
		{Label: "App Settings", Icon: "⚙️", Desc: "Modify preferences and paths", Action: "settings"},
		{Label: "Help", Icon: "❓", Desc: "Show help and shortcuts", Action: "help"},
		{Label: "Exit", Icon: "❌", Desc: "Exit application", Action: "exit"},
	}
}

func (m *AppModel) SetupUI() {
	m.UI.Progress = progress.New(progress.WithDefaultGradient())
	m.UI.TextInput = textinput.New()
	m.UI.TextInput.Placeholder = "Search files..."
	m.UI.TextInput.CharLimit = 128
	m.UI.TextInput.Width = 40
}
