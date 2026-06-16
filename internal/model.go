package internal

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

/* TYPES */

type ModalType int

const (
	ModalNone ModalType = iota
	ModalFile
	ModalConfirm
	ModalHelp
	ModalCustom
	ModalSettings
)

type DeviceMode string

const (
	ModeDisconnected DeviceMode = "DISCONNECTED"
	ModeFastboot     DeviceMode = "FASTBOOT"
	ModeDevice       DeviceMode = "DEVICE"
	ModeRecovery     DeviceMode = "RECOVERY"
	ModeSideload     DeviceMode = "SIDELOAD"
	ModeUnauthorized DeviceMode = "UNAUTHORIZED"
	ModeOffline      DeviceMode = "OFFLINE"
)

type LogLevel string

const (
	LogInfo    LogLevel = "INFO"
	LogError   LogLevel = "ERROR"
	LogSuccess LogLevel = "SUCCESS"
)

/* CONFIG MODELS */

type PickerConfig struct {
	Title  string `json:"title"`
	Filter string `json:"filter"`
	SubDir string `json:"sub_dir"`
}

type MenuConfig struct {
	Label  string        `json:"label"`
	Icon   string        `json:"icon"`
	Desc   string        `json:"desc"`
	Action string        `json:"action"`
	Picker *PickerConfig `json:"picker,omitempty"`
}

type AppConfig struct {
	BaseDir    string            `json:"base_dir"`
	DevicePath string            `json:"device_path,omitempty"`
	Folders    map[string]string `json:"folders"`
}

/* DATA MODELS */

type LogEntry struct {
	Level     LogLevel
	Text      string
	Timestamp time.Time
}

type MenuItem struct {
	Label  string
	Icon   string
	Desc   string
	Action string
}

type DeviceState struct {
	Mode    DeviceMode
	Serial  string
	Model   string
	Battery string
	Slot    string
	Secure  string
}

type FileItem struct {
	Name  string
	Path  string
	IsDir bool
}

type Toast struct {
	Message string
	Type    LogLevel
}

/* APP MODEL */

type AppModel struct {
	// 1. Core State
	Width, Height int
	Selection     int
	Menu          []MenuItem
	Busy          bool
	ActiveModal   ModalType
	Tick          int // Pulsing animation tick

	// 2. Device State
	Device DeviceState

	// 3. UI Components (Standard Bubbles)
	UI struct {
		Viewport  viewport.Model
		Progress  progress.Model
		TextInput textinput.Model
	}

	// 4. Modal / Overlay Data
	Modal struct {
		ConfirmMsg string
		OnConfirm  func() tea.Cmd
		
		FileDir      string
		FileList     []FileItem
		FullFileList []FileItem
		FileCursor   int
		FileTitle    string
		FileFilter   string
		OnFileSelect func(string) tea.Cmd

		CustomLogs     []LogEntry
		CustomViewport viewport.Model
		Width          int

		SettingsCursor int
		SettingsInputs []textinput.Model
	}

	// 5. Feedback & Logs
	Logs        []LogEntry
	ActiveToast *Toast

	// 6. Config & Paths
	Config     AppConfig
	BaseDir    string
	DevicePath string

	IsInitialized bool
	IsRefreshing  bool
}

/* FACTORY */

func NewModel() AppModel {
	m := AppModel{
		Device: DeviceState{
			Mode:    ModeDisconnected,
			Serial:  "-",
			Model:   "-",
			Battery: "-",
			Slot:    "-",
			Secure:  "-",
		},
		Logs: []LogEntry{
			{Level: LogInfo, Text: "SYSTEM INITIALIZED. READY.", Timestamp: time.Now()},
		},
	}
	m.UI.Viewport = viewport.New(0, 0)
	m.UI.Viewport.SetContent(RenderLogsStr(m.Logs, 0))
	return m
}
