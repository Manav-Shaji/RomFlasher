package tui

import (
	"flashtool/internal/domain"
	"flashtool/internal/platform/adb"
	"flashtool/internal/engine"
	"flashtool/internal/config"

	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

/* MODAL HANDLER */

func (m AppModel) startSetupFlow() (AppModel, tea.Cmd) {
	m.ActiveModal = ModalFile
	m.Modal.FileTitle = "SELECT MAIN CUSTOM ROMS DIR"
	m.Modal.FileFilter = "" 
	
	m.Modal.FileDir = m.App.Config.BaseDir
	if _, err := os.Stat(m.Modal.FileDir); err != nil {
		m.Modal.FileDir, _ = os.Getwd()
	}

	m.Modal.FullFileList = LoadFiles(m.Modal.FileDir, "")
	m.Modal.FileList = m.Modal.FullFileList
	m.Modal.OnFileSelect = func(path string) tea.Cmd {
		return func() tea.Msg { return SetupMainDirMsg(path) }
	}
	return m, m.UI.TextInput.Focus()
}

func updateModal(m AppModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" && m.IsInitialized {
		m.ActiveModal = ModalNone
		m.UI.TextInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	switch m.ActiveModal {
	case ModalFile:
		oldVal := m.UI.TextInput.Value()
		newTi, tiCmd := m.UI.TextInput.Update(msg)
		m.UI.TextInput = newTi
		cmd = tiCmd

		if m.UI.TextInput.Value() != oldVal {
			return m, tea.Batch(cmd, tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg {
				return SearchDebounceMsg{Query: m.UI.TextInput.Value()}
			}))
		}
	case ModalCustom:
		if !m.Busy {
			newTi, tiCmd := m.UI.TextInput.Update(msg)
			m.UI.TextInput = newTi
			cmd = tiCmd
		}
	}

	switch m.ActiveModal {
	case ModalConfirm:
		switch strings.ToLower(msg.String()) {
		case "y", "enter":
			m.ActiveModal, m.Busy = ModalNone, true
			m.UI.TextInput.Blur()
			if m.Modal.OnConfirm != nil { return m, m.Modal.OnConfirm() }
		case "n", "esc":
			m.ActiveModal = ModalNone
			m.UI.TextInput.Blur()
		}

	case ModalFile:
		switch msg.String() {
		case "up":   if m.Modal.FileCursor > 0 { m.Modal.FileCursor-- }
		case "down": if m.Modal.FileCursor < len(m.Modal.FileList)-1 { m.Modal.FileCursor++ }
		case "enter":
			if len(m.Modal.FileList) == 0 { return m, cmd }
			sel := m.Modal.FileList[m.Modal.FileCursor]
			
			if sel.Name == "[ SELECT THIS FOLDER ]" {
				m.ActiveModal = ModalNone
				m.UI.TextInput.Blur()
				if m.Modal.OnFileSelect != nil { return m, m.Modal.OnFileSelect(m.Modal.FileDir) }
				return m, nil
			}

			if sel.IsDir {
				m.Modal.FileDir = sel.Path
				m.Modal.FileCursor = 0
				m.UI.TextInput.Reset()
				m.Modal.FullFileList = LoadFiles(m.Modal.FileDir, m.Modal.FileFilter)
				m.Modal.FileList = m.Modal.FullFileList
				return m, cmd
			}
			m.ActiveModal = ModalNone
			m.UI.TextInput.Blur()
			if m.Modal.OnFileSelect != nil {
				return m, tea.Batch(cmd, m.Modal.OnFileSelect(sel.Path))
			}
		}

	case ModalCustom:
		if msg.String() == "enter" && !m.Busy {
			val := m.UI.TextInput.Value()
			if val != "" {
				m.ActiveModal = ModalConfirm
				m.Modal.ConfirmMsg = "Run custom command: " + val + "?"
				m.Modal.OnConfirm = func() tea.Cmd {
					m.Modal.CustomLogs = NewLogBuffer(500)
					m.Modal.CustomViewport = viewport.New(0, 0)
					m.Modal.CustomViewport.SetContent("Executing: " + val + "...")
					m.UI.TextInput.Reset()
					return m.App.Engine.ExecuteAsync(func(ctx context.Context) error {
						return m.App.Engine.RunCustomCommand(ctx, val)
					})
				}
				return m, nil
			}
		}

	case ModalHelp:
		m.ActiveModal = ModalNone

	case ModalSettings:
		switch msg.String() {
		case "up", "k", "shift+tab":
			if m.Modal.SettingsCursor > 0 { m.Modal.SettingsCursor-- }
		case "down", "j", "tab":
			if m.Modal.SettingsCursor < 2 { m.Modal.SettingsCursor++ }
		case "enter":
			switch m.Modal.SettingsCursor {
			case 0:
				m.ActiveModal = ModalFile
				m.Modal.FileTitle = "SELECT BASE ROM DIRECTORY"
				m.Modal.FileDir = m.App.Config.BaseDir
				if _, err := os.Stat(m.Modal.FileDir); err != nil { m.Modal.FileDir, _ = os.Getwd() }
				m.Modal.FullFileList = LoadFiles(m.Modal.FileDir, "")
				m.Modal.FileList = m.Modal.FullFileList
				m.Modal.FileCursor = 0
				m.UI.TextInput.Reset()
				m.Modal.OnFileSelect = func(path string) tea.Cmd {
					return func() tea.Msg { return SettingsFolderSelectedMsg{Index: 0, Path: path} }
				}
				return m, m.UI.TextInput.Focus()
			case 1:
				m.ActiveModal = ModalFile
				m.Modal.FileTitle = "SELECT DEVICE FOLDER"
				m.Modal.FileDir = m.App.Config.DevicePath
				if m.Modal.FileDir == "" { m.Modal.FileDir = m.App.Config.BaseDir }
				if _, err := os.Stat(m.Modal.FileDir); err != nil {
					m.Modal.FileDir = m.App.Config.BaseDir
					if _, err := os.Stat(m.Modal.FileDir); err != nil { m.Modal.FileDir, _ = os.Getwd() }
				}
				m.Modal.FullFileList = LoadFiles(m.Modal.FileDir, "")
				m.Modal.FileList = m.Modal.FullFileList
				m.Modal.FileCursor = 0
				m.UI.TextInput.Reset()
				m.Modal.OnFileSelect = func(path string) tea.Cmd {
					return func() tea.Msg { return SettingsFolderSelectedMsg{Index: 1, Path: path} }
				}
				return m, m.UI.TextInput.Focus()
			case 2:
				err := config.SaveConfig(m.App.Config)
				m.BaseDir = m.App.Config.BaseDir
				m.DevicePath = m.App.Config.DevicePath
				
				if err != nil {
					m.ActiveToast = &Toast{Message: "Save Failed", Type: domain.LogError}
				} else {
					m.ActiveToast = &Toast{Message: "Settings Saved", Type: domain.LogSuccess}
				}
				m.ActiveModal = ModalNone
				return m, toastTimeoutCmd(3*time.Second)
			}
		}
	}

	return m, cmd
}

/* UPDATE HANDLERS */

func (m AppModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.Width, m.Height = msg.Width, msg.Height
	if !m.IsInitialized && m.ActiveModal == ModalNone {
		return m.startSetupFlow()
	}
	return m, nil
}

func (m AppModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Busy {
		if msg.String() == "ctrl+c" || msg.String() == "esc" { 
			m.App.Engine.CancelActiveCommand()
			return m, nil 
		}
		return m, nil
	}

	if m.ActiveModal != ModalNone {
		return updateModal(m, msg)
	}

	switch msg.String() {
	case "up", "k":   m.Selection = (m.Selection - 1 + len(m.Menu)) % len(m.Menu)
	case "down", "j": m.Selection = (m.Selection + 1) % len(m.Menu)
	case "enter":     return handleMenuSelect(m)
	case "q", "ctrl+c": return m, tea.Quit
	case "h", "?":    m.ActiveModal = ModalHelp
	}
	return m, nil
}

func (m AppModel) handlePollMsg(_ adb.PollMsg) (tea.Model, tea.Cmd) {
	if m.Busy {
		return m, tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg { return adb.PollMsg(t) })
	}
	return m, tea.Batch(func() tea.Msg { return adb.CheckDeviceState() }, adb.PollDeviceCmd())
}

func (m AppModel) handleDeviceUpdate(msg adb.DeviceUpdateMsg) (tea.Model, tea.Cmd) {
	newDevice := domain.DeviceState(msg)
	newMode := newDevice.Mode
	var cmds []tea.Cmd

	if m.IsRefreshing {
		m.IsRefreshing = false
		displayMode := formatDeviceMode(newMode)

		if newMode == domain.ModeDisconnected {
			m.ActiveToast = &Toast{Message: "Scan Finished: NO DEVICE", Type: domain.LogInfo}
		} else {
			m.ActiveToast = &Toast{Message: fmt.Sprintf("Scan Complete: %s", displayMode), Type: domain.LogSuccess}
		}
		cmds = append(cmds, toastTimeoutCmd(3*time.Second))
	} else if m.Device.Mode != newMode {
		m.ActiveToast = &Toast{Message: fmt.Sprintf("Mode: %s", newMode), Type: domain.LogInfo}
		cmds = append(cmds, toastTimeoutCmd(3*time.Second))
	}
	m.Device = newDevice
	return m, tea.Batch(cmds...)
}

func (m AppModel) handleLogMsg(msg engine.LogMsg) (tea.Model, tea.Cmd) {
	line := string(msg)


	// 1. Process Standard Logs
	isOverwrite := strings.HasPrefix(line, "\r")

	cleanLine := strings.TrimPrefix(line, "\r")
	if cleanLine == "" { return m, m.App.Engine.WaitForLogs() }

	level := parseLogLevel(cleanLine)
	entry := domain.LogEntry{Level: level, Text: cleanLine, Timestamp: time.Now()}

	if m.ActiveModal == ModalCustom {
		m.syncCustomLogs(entry, isOverwrite)
	} else {
		m.syncMainLogs(entry, isOverwrite)
	}

	return m, tea.Batch(m.App.Engine.WaitForLogs())
}

func (m AppModel) handleTaskComplete(msg engine.TaskCompleteMsg) (tea.Model, tea.Cmd) {
	m.Busy = false
	status := "[ DONE ]"
	if msg.Err != nil {
		m.ActiveToast = &Toast{Message: "Failed", Type: domain.LogError}
		status = fmt.Sprintf("[ FAILED: %v ]", msg.Err)
	} else {
		m.ActiveToast = &Toast{Message: "Success", Type: domain.LogSuccess}
	}

	entry := domain.LogEntry{Text: status, Level: domain.LogInfo, Timestamp: time.Now()}
	
	if m.ActiveModal == ModalCustom {
		m.Modal.CustomLogs.Add(entry)
	} else {
		m.Logs.Add(entry)
	}
	m.LogsDirty = true
	
	return m, toastTimeoutCmd(3*time.Second)
}

func (m AppModel) handleSetupMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetupMainDirMsg:
		m.Modal.FileDir = string(msg)
		m.ActiveModal = ModalFile
		m.Modal.FileTitle = "SELECT DEVICE FOLDER"
		m.Modal.FileCursor = 0
		m.UI.TextInput.Reset()
		m.Modal.FullFileList = LoadFiles(m.Modal.FileDir, "")
		m.Modal.FileList = m.Modal.FullFileList
		m.Modal.OnFileSelect = func(path string) tea.Cmd {
			return func() tea.Msg { return SetupDeviceDirMsg(path) }
		}
		return m, m.UI.TextInput.Focus()

	case SetupDeviceDirMsg:
		m.DevicePath = string(msg)
		m.IsInitialized = true
		m.ActiveModal = ModalNone
		m.UI.TextInput.Blur()
		return m, nil

	case SetupConfirmMsg:
		m.ActiveModal = ModalConfirm
		m.Modal.ConfirmMsg = msg.Msg
		m.Modal.OnConfirm = func() tea.Cmd { return msg.Cmd }
		return m, nil

	case SettingsFolderSelectedMsg:
		switch msg.Index {
		case 0:
			m.App.Config.BaseDir = msg.Path
		case 1:
			m.App.Config.DevicePath = msg.Path
		}
		m.ActiveModal = ModalSettings
		return m, nil
	}
	return m, nil
}

/* HELPERS */

func formatDeviceMode(m domain.DeviceMode) string {
	switch m {
	case domain.ModeDevice:       return "ADB DEVICE"
	case domain.ModeSideload:     return "ADB SIDELOAD"
	case domain.ModeRecovery:     return "RECOVERY"
	case domain.ModeFastboot:     return "FASTBOOT"
	case domain.ModeDisconnected: return "NOT FOUND"
	default:               return string(m)
	}
}

func parseLogLevel(line string) domain.LogLevel {
	line = strings.ToUpper(line)
	if strings.HasPrefix(line, "ERR:") || strings.HasPrefix(line, "ERROR:") || 
	   strings.Contains(line, "FAILED") || strings.Contains(line, "CRITICAL") { 
		return domain.LogError 
	}
	if strings.HasPrefix(line, "OK") || strings.Contains(line, "SUCCESS") || 
	   strings.HasPrefix(line, "FINISHED") {
		return domain.LogSuccess 
	}
	return domain.LogInfo
}

func (m *AppModel) syncCustomLogs(entry domain.LogEntry, overwrite bool) {
	if overwrite && m.Modal.CustomLogs.Len() > 0 {
		m.Modal.CustomLogs.ReplaceLast(entry)
	} else {
		m.Modal.CustomLogs.Add(entry)
	}
	m.LogsDirty = true
}

func (m *AppModel) syncMainLogs(entry domain.LogEntry, overwrite bool) {
	if overwrite && m.Logs.Len() > 0 {
		last, _ := m.Logs.Last()
		if !strings.Contains(last.Text, ">") && !strings.Contains(strings.ToUpper(last.Text), "EXECUTION") {
			m.Logs.ReplaceLast(entry)
		} else {
			m.Logs.Add(entry)
		}
	} else {
		m.Logs.Add(entry)
	}
	m.LogsDirty = true
}

func (m AppModel) canFlash(isSideload bool) (bool, string) {
	if m.Device.Mode == domain.ModeDisconnected || m.Device.Mode == domain.ModeOffline {
		return false, "Device is disconnected or offline."
	}

	if isSideload {
		if m.Device.Mode != domain.ModeRecovery && m.Device.Mode != domain.ModeDevice && m.Device.Mode != domain.ModeSideload {
			return false, "Device must be in Recovery or Sideload mode."
		}
	} else {
		if m.Device.Mode != domain.ModeFastboot {
			return false, "Device must be in Fastboot mode."
		}
	}

	batt := strings.TrimSpace(m.Device.Battery)
	if strings.HasSuffix(batt, "%") {
		var pct int
		if _, err := fmt.Sscanf(batt, "%d%%", &pct); err == nil && pct < 30 {
			return false, fmt.Sprintf("Battery level too low (%d%% < 30%%).", pct)
		}
	} else if strings.HasSuffix(batt, " mV") {
		var mv int
		if _, err := fmt.Sscanf(batt, "%d mV", &mv); err == nil && mv < 3500 {
			return false, fmt.Sprintf("Battery voltage too low (%dmV < 3500mV).", mv)
		}
	}

	return true, ""
}

func handleMenuSelect(m AppModel) (tea.Model, tea.Cmd) {
	sel := m.Menu[m.Selection]

	var filter string
	switch sel.Action {
	case "flash_rec", "flash_boot", "wipe_super": filter = ".img"
	case "sideload": filter = ".zip"
	}

	if filter != "" {
		m.ActiveModal = ModalFile
		m.Modal.FileTitle = "SELECT " + strings.ToUpper(strings.Split(sel.Label, " ")[1]) + " FILE"
		m.Modal.FileFilter = filter
		
		m.Modal.FileDir = m.DevicePath
		if m.Modal.FileDir == "" { m.Modal.FileDir = m.BaseDir }
		if m.Modal.FileDir == "" { m.Modal.FileDir, _ = os.Getwd() }

		if subDir, ok := m.App.Config.Folders[sel.Action]; ok && subDir != "" {
			target := filepath.Join(m.Modal.FileDir, subDir)
			if stat, err := os.Stat(target); err == nil && stat.IsDir() {
				m.Modal.FileDir = target
			}
		}

		m.Modal.FullFileList = LoadFiles(m.Modal.FileDir, m.Modal.FileFilter)
		m.Modal.FileList = m.Modal.FullFileList
		m.Modal.FileCursor = 0
		
		switch sel.Action {
		case "flash_rec":
			if ok, msg := m.canFlash(false); !ok {
				m.ActiveModal = ModalNone
				m.ActiveToast = &Toast{Message: msg, Type: domain.LogError}
				return m, toastTimeoutCmd(3 * time.Second)
			}
			m.Modal.OnFileSelect = func(p string) tea.Cmd {
				return func() tea.Msg {
					m.App.Engine.LogChan <- "> fastboot flash recovery " + filepath.Base(p)
					return SetupConfirmMsg{
						Msg: "Flash RECOVERY with: " + filepath.Base(p) + "?",
						Cmd: m.App.Engine.ExecuteAsync(func(ctx context.Context) error {
							return m.App.Engine.FlashService.FlashImage(ctx, "recovery", p)
						}),
					}
				}
			}
		case "flash_boot":
			if ok, msg := m.canFlash(false); !ok {
				m.ActiveModal = ModalNone
				m.ActiveToast = &Toast{Message: msg, Type: domain.LogError}
				return m, toastTimeoutCmd(3 * time.Second)
			}
			m.Modal.OnFileSelect = func(p string) tea.Cmd {
				return func() tea.Msg {
					m.App.Engine.LogChan <- "> fastboot flash boot " + filepath.Base(p)
					return SetupConfirmMsg{
						Msg: "Flash BOOT with: " + filepath.Base(p) + "?",
						Cmd: m.App.Engine.ExecuteAsync(func(ctx context.Context) error {
							return m.App.Engine.FlashService.FlashImage(ctx, "boot", p)
						}),
					}
				}
			}
		case "wipe_super":
			if ok, msg := m.canFlash(false); !ok {
				m.ActiveModal = ModalNone
				m.ActiveToast = &Toast{Message: msg, Type: domain.LogError}
				return m, toastTimeoutCmd(3 * time.Second)
			}
			m.Modal.OnFileSelect = func(p string) tea.Cmd {
				return func() tea.Msg {
					m.App.Engine.LogChan <- "> fastboot wipe-super " + filepath.Base(p)
					return SetupConfirmMsg{
						Msg: "WIPE SUPER and Flash: " + filepath.Base(p) + "?",
						Cmd: m.App.Engine.ExecuteAsync(func(ctx context.Context) error {
							return m.App.Engine.FlashService.WipeSuper(ctx, p)
						}),
					}
				}
			}
		case "sideload":
			if ok, msg := m.canFlash(true); !ok {
				m.ActiveModal = ModalNone
				m.ActiveToast = &Toast{Message: msg, Type: domain.LogError}
				return m, toastTimeoutCmd(3 * time.Second)
			}
			m.Modal.OnFileSelect = func(p string) tea.Cmd {
				return func() tea.Msg {
					m.App.Engine.LogChan <- "> adb sideload " + filepath.Base(p)
					return SetupConfirmMsg{
						Msg: "Sideload: " + filepath.Base(p) + "?",
						Cmd: m.App.Engine.ExecuteAsync(func(ctx context.Context) error {
							return m.App.Engine.FlashService.Sideload(ctx, p)
						}),
					}
				}
			}
		}
		return m, m.UI.TextInput.Focus()
	}

	switch sel.Action {
	case "rb_system":
		m.ActiveModal, m.Modal.ConfirmMsg = ModalConfirm, "Reboot to System?"
		m.Modal.OnConfirm = func() tea.Cmd {
			m.App.Engine.LogChan <- "> rebooting to system..."
			return m.App.Engine.ExecuteAsync(func(ctx context.Context) error {
				return m.App.Engine.DeviceService.RebootSystem(ctx, string(m.Device.Mode))
			})
		}
	case "rb_recovery":
		m.ActiveModal, m.Modal.ConfirmMsg = ModalConfirm, "Reboot to Recovery?"
		m.Modal.OnConfirm = func() tea.Cmd {
			m.App.Engine.LogChan <- "> rebooting to recovery..."
			return m.App.Engine.ExecuteAsync(func(ctx context.Context) error {
				return m.App.Engine.DeviceService.RebootRecovery(ctx, string(m.Device.Mode))
			})
		}
	case "refresh":
		m.IsRefreshing = true
		m.ActiveToast = &Toast{Message: "Scanning for devices...", Type: domain.LogInfo}
		return m, tea.Batch(
			func() tea.Msg { return adb.CheckDeviceState() },
			adb.PollDeviceCmd(),
			toastTimeoutCmd(2*time.Second),
		)
	case "help":
		m.ActiveModal = ModalHelp
	case "custom":
		m.ActiveModal = ModalCustom
		m.UI.TextInput.Reset()
		m.UI.TextInput.Width = 50
		m.UI.TextInput.Placeholder = "Type command here..."
		return m, m.UI.TextInput.Focus()
	case "settings":
		m.ActiveModal = ModalSettings
		m.Modal.SettingsCursor = 0
		return m, nil
	case "exit":
		return m, tea.Quit
	}
	return m, nil
}
