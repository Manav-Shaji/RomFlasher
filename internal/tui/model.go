package tui

import (
	"flashtool/internal/app"
	"flashtool/internal/core"
	"flashtool/internal/platform"

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

type LogBuffer struct {
	data  []core.LogEntry
	size  int
	head  int
	count int
}

func NewLogBuffer(size int) *LogBuffer {
	return &LogBuffer{
		data: make([]core.LogEntry, size),
		size: size,
	}
}

func (lb *LogBuffer) Add(entry core.LogEntry) {
	if lb.count < lb.size {
		lb.data[lb.count] = entry
		lb.count++
	} else {
		lb.data[lb.head] = entry
		lb.head = (lb.head + 1) % lb.size
	}
}

func (lb *LogBuffer) Len() int {
	return lb.count
}

func (lb *LogBuffer) Iterate(fn func(core.LogEntry)) {
	for i := 0; i < lb.count; i++ {
		idx := (lb.head + i) % lb.size
		fn(lb.data[idx])
	}
}

func (lb *LogBuffer) ReplaceLast(entry core.LogEntry) {
	if lb.count == 0 {
		lb.Add(entry)
		return
	}
	idx := (lb.head + lb.count - 1) % lb.size
	lb.data[idx] = entry
}

func (lb *LogBuffer) Last() (core.LogEntry, bool) {
	if lb.count == 0 {
		return core.LogEntry{}, false
	}
	idx := (lb.head + lb.count - 1) % lb.size
	return lb.data[idx], true
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
	// Core State
	Width, Height int
	Selection     int
	Menu          []MenuItem
	Busy          bool
	ActiveModal   ModalType
	Tick          int // Pulsing animation tick

	App *app.App

	// Device State
	Device platform.DeviceState

	// UI Components
	UI struct {
		Viewport  viewport.Model
		Progress  progress.Model
		TextInput textinput.Model
	}

	// Modal Data
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

		CustomLogs     *LogBuffer
		CustomViewport viewport.Model
		Width          int

		SettingsCursor int
		SettingsInputs []textinput.Model
	}

	// Logs
	Logs        *LogBuffer
	ActiveToast *Toast
	LogsDirty   bool

	// Config
	BaseDir    string
	DevicePath string

	IsInitialized bool
	IsRefreshing  bool
}

/* FACTORY */

func NewModel(app *app.App) AppModel {
	m := AppModel{
		App:        app,
		BaseDir:    app.Config.BaseDir,
		DevicePath: app.Config.DevicePath,
		Device: platform.DeviceState{
			Mode:    platform.ModeDisconnected,
			Serial:  "-",
			Model:   "-",
			Battery: "-",
			Slot:    "-",
			Secure:  "-",
		},
		Logs: NewLogBuffer(500),
	}
	m.Logs.Add(core.LogEntry{Level: core.LogInfo, Text: "SYSTEM INITIALIZED. READY.", Timestamp: time.Now()})
	m.Modal.CustomLogs = NewLogBuffer(500)

	m.Menu = GetDefaultMenu()
	m.SetupUI()
	m.UI.Viewport = viewport.New(0, 0)
	m.UI.Viewport.SetContent(RenderLogsStr(m.Logs, 0))
	return m
}
