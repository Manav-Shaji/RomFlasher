package tui

import (
	"flashtool/internal/core"

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

type MenuItem struct {
	Label  string
	Icon   string
	Desc   string
	Action string
}

type FileItem struct {
	Name  string
	Path  string
	IsDir bool
}

type Toast struct {
	Message string
	Type    core.LogLevel
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
	Device core.DeviceState

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

		CustomLogs     []core.LogEntry
		CustomViewport viewport.Model
		Width          int

		SettingsCursor int
		SettingsInputs []textinput.Model
	}

	// 5. Feedback & Logs
	Logs        []core.LogEntry
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
		Device: core.DeviceState{
			Mode:    core.ModeDisconnected,
			Serial:  "-",
			Model:   "-",
			Battery: "-",
			Slot:    "-",
			Secure:  "-",
		},
		Logs: []core.LogEntry{
			{Level: core.LogInfo, Text: "SYSTEM INITIALIZED. READY.", Timestamp: time.Now()},
		},
	}
	m.Menu = GetDefaultMenu()
	m.SetupUI()
	m.UI.Viewport = viewport.New(0, 0)
	m.UI.Viewport.SetContent(RenderLogsStr(m.Logs, 0))
	return m
}
