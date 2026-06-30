package tui

import (
	"flashtool/internal/core"

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
	
	m.Modal.FileDir = m.Config.BaseDir
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
	if m.ActiveModal == ModalFile {
		oldVal := m.UI.TextInput.Value()
		newTi, tiCmd := m.UI.TextInput.Update(msg)
		m.UI.TextInput = newTi
		cmd = tiCmd

		if m.UI.TextInput.Value() != oldVal {
			val := strings.ToLower(m.UI.TextInput.Value())
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
	} else if m.ActiveModal == ModalCustom {
		if !m.Busy {
			newTi, tiCmd := m.UI.TextInput.Update(msg)
			m.UI.TextInput = newTi
			cmd = tiCmd
		}
	}

	switch m.ActiveModal {
	case ModalConfirm:
		key := strings.ToLower(msg.String())
		if key == "y" || key == "enter" {
			m.ActiveModal, m.Busy = ModalNone, true
			m.UI.TextInput.Blur()
			if m.Modal.OnConfirm != nil { return m, m.Modal.OnConfirm() }
		} else if key == "n" || key == "esc" {
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
				m.Busy = true
				m.Modal.CustomLogs = nil
				m.Modal.CustomViewport = viewport.New(0, 0)
				m.Modal.CustomViewport.SetContent("Executing: " + val + "...")
				m.UI.TextInput.Reset()
				return m, core.RunCustomCommand(val)
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
			if m.Modal.SettingsCursor == 0 {
				m.ActiveModal = ModalFile
				m.Modal.FileTitle = "SELECT BASE ROM DIRECTORY"
				m.Modal.FileDir = m.Config.BaseDir
				if _, err := os.Stat(m.Modal.FileDir); err != nil { m.Modal.FileDir, _ = os.Getwd() }
				m.Modal.FullFileList = LoadFiles(m.Modal.FileDir, "")
				m.Modal.FileList = m.Modal.FullFileList
				m.Modal.FileCursor = 0
				m.UI.TextInput.Reset()
				m.Modal.OnFileSelect = func(path string) tea.Cmd {
					return func() tea.Msg { return SettingsFolderSelectedMsg{Index: 0, Path: path} }
				}
				return m, m.UI.TextInput.Focus()
			} else if m.Modal.SettingsCursor == 1 {
				m.ActiveModal = ModalFile
				m.Modal.FileTitle = "SELECT DEVICE FOLDER"
				m.Modal.FileDir = m.Config.DevicePath
				if m.Modal.FileDir == "" { m.Modal.FileDir = m.Config.BaseDir }
				if _, err := os.Stat(m.Modal.FileDir); err != nil {
					m.Modal.FileDir = m.Config.BaseDir
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
			} else if m.Modal.SettingsCursor == 2 {
				err := SaveConfig(m.Config)
				m.BaseDir = m.Config.BaseDir
				m.DevicePath = m.Config.DevicePath
				
				if err != nil {
					m.ActiveToast = &Toast{Message: "Save Failed", Type: core.LogError}
				} else {
					m.ActiveToast = &Toast{Message: "Settings Saved", Type: core.LogSuccess}
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
			core.CancelActiveCommand()
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

func (m AppModel) handlePollMsg(msg core.PollMsg) (tea.Model, tea.Cmd) {
	if m.Busy {
		return m, tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg { return core.PollMsg(t) })
	}
	return m, tea.Batch(func() tea.Msg { return core.CheckDeviceState() }, core.PollDeviceCmd())
}

func (m AppModel) handleDeviceUpdate(msg core.DeviceUpdateMsg) (tea.Model, tea.Cmd) {
	newDevice := core.DeviceState(msg)
	newMode := newDevice.Mode
	var cmds []tea.Cmd

	if m.IsRefreshing {
		m.IsRefreshing = false
		displayMode := formatDeviceMode(newMode)

		if newMode == core.ModeDisconnected {
			m.ActiveToast = &Toast{Message: "Scan Finished: NO DEVICE", Type: core.LogInfo}
		} else {
			m.ActiveToast = &Toast{Message: fmt.Sprintf("Scan Complete: %s", displayMode), Type: core.LogSuccess}
		}
		cmds = append(cmds, toastTimeoutCmd(3*time.Second))
	} else if m.Device.Mode != newMode {
		m.ActiveToast = &Toast{Message: fmt.Sprintf("Mode: %s", newMode), Type: core.LogInfo}
		cmds = append(cmds, toastTimeoutCmd(3*time.Second))
	}
	m.Device = newDevice
	return m, tea.Batch(cmds...)
}

func (m AppModel) handleLogMsg(msg core.LogMsg) (tea.Model, tea.Cmd) {
	line := string(msg)


	// 1. Process Standard Logs
	isOverwrite := strings.HasPrefix(line, "\r")

	cleanLine := strings.TrimPrefix(line, "\r")
	if cleanLine == "" { return m, core.WaitForLogs(core.LogChan) }

	level := parseLogLevel(cleanLine)
	entry := core.LogEntry{Level: level, Text: cleanLine, Timestamp: time.Now()}

	if m.ActiveModal == ModalCustom {
		m.syncCustomLogs(entry, isOverwrite)
	} else {
		m.syncMainLogs(entry, isOverwrite)
	}

	return m, tea.Batch(core.WaitForLogs(core.LogChan))
}

func (m AppModel) handleTaskComplete(msg core.TaskCompleteMsg) (tea.Model, tea.Cmd) {
	m.Busy = false
	status := "[ DONE ]"
	if msg.Err != nil {
		m.ActiveToast = &Toast{Message: "Failed", Type: core.LogError}
		fmt.Print("\a\a")
		status = fmt.Sprintf("[ FAILED: %v ]", msg.Err)
	} else {
		m.ActiveToast = &Toast{Message: "Success", Type: core.LogSuccess}
		fmt.Print("\a")
	}

	entry := core.LogEntry{Text: status, Level: core.LogInfo, Timestamp: time.Now()}
	m.Logs = append(m.Logs, entry)
	
	if m.ActiveModal == ModalCustom {
		m.Modal.CustomLogs = append(m.Modal.CustomLogs, entry)
		m.UI.Viewport.SetContent(RenderLogsStr(m.Modal.CustomLogs, m.Modal.CustomViewport.Width))
	} else {
		m.UI.Viewport.SetContent(RenderLogsStr(m.Logs, m.UI.Viewport.Width))
	}
	
	m.UI.Viewport.GotoBottom()
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
		if msg.Index == 0 {
			m.Config.BaseDir = msg.Path
		} else if msg.Index == 1 {
			m.Config.DevicePath = msg.Path
		}
		m.ActiveModal = ModalSettings
		return m, nil
	}
	return m, nil
}

/* HELPERS */

func formatDeviceMode(m core.DeviceMode) string {
	switch m {
	case core.ModeDevice:       return "ADB DEVICE"
	case core.ModeSideload:     return "ADB SIDELOAD"
	case core.ModeRecovery:     return "RECOVERY"
	case core.ModeFastboot:     return "FASTBOOT"
	case core.ModeDisconnected: return "NOT FOUND"
	default:               return string(m)
	}
}

func parseLogLevel(line string) core.LogLevel {
	line = strings.ToUpper(line)
	if strings.HasPrefix(line, "ERR:") || strings.HasPrefix(line, "ERROR:") || 
	   strings.Contains(line, "FAILED") || strings.Contains(line, "CRITICAL") { 
		return core.LogError 
	}
	if strings.HasPrefix(line, "OK") || strings.Contains(line, "SUCCESS") || 
	   strings.HasPrefix(line, "FINISHED") {
		return core.LogSuccess 
	}
	return core.LogInfo
}

func (m *AppModel) syncCustomLogs(entry core.LogEntry, overwrite bool) {
	if overwrite && len(m.Modal.CustomLogs) > 0 {
		m.Modal.CustomLogs[len(m.Modal.CustomLogs)-1] = entry
	} else {
		m.Modal.CustomLogs = append(m.Modal.CustomLogs, entry)
	}
	if len(m.Modal.CustomLogs) > 100 { m.Modal.CustomLogs = m.Modal.CustomLogs[1:] }
	
	innerW := m.Modal.CustomViewport.Width
	m.Modal.CustomViewport.SetContent(RenderLogsStr(m.Modal.CustomLogs, innerW))
	m.Modal.CustomViewport.GotoBottom()
}

func (m *AppModel) syncMainLogs(entry core.LogEntry, overwrite bool) {
	if overwrite && len(m.Logs) > 0 {
		last := m.Logs[len(m.Logs)-1]
		if !strings.Contains(last.Text, ">") && !strings.Contains(strings.ToUpper(last.Text), "EXECUTION") {
			m.Logs[len(m.Logs)-1] = entry
		} else {
			m.Logs = append(m.Logs, entry)
		}
	} else {
		m.Logs = append(m.Logs, entry)
	}
	if len(m.Logs) > 500 { m.Logs = m.Logs[1:] }

	_, detailW, _, _, logH := m.GetLayoutDimensions()
	m.UI.Viewport.Width = detailW - 2
	m.UI.Viewport.Height = logH
	
	m.UI.Viewport.SetContent(RenderLogsStr(m.Logs, m.UI.Viewport.Width))
	m.UI.Viewport.GotoBottom()
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

		if subDir, ok := m.Config.Folders[sel.Action]; ok && subDir != "" {
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
			m.Modal.OnFileSelect = func(p string) tea.Cmd {
				return func() tea.Msg {
					core.LogChan <- "> fastboot flash recovery " + filepath.Base(p)
					return SetupConfirmMsg{
						Msg: "Flash RECOVERY with: " + filepath.Base(p) + "?",
						Cmd: core.FlashImage("recovery", p),
					}
				}
			}
		case "flash_boot":
			m.Modal.OnFileSelect = func(p string) tea.Cmd {
				return func() tea.Msg {
					core.LogChan <- "> fastboot flash boot " + filepath.Base(p)
					return SetupConfirmMsg{
						Msg: "Flash BOOT with: " + filepath.Base(p) + "?",
						Cmd: core.FlashImage("boot", p),
					}
				}
			}
		case "wipe_super":
			m.Modal.OnFileSelect = func(p string) tea.Cmd {
				return func() tea.Msg {
					core.LogChan <- "> fastboot wipe-super " + filepath.Base(p)
					return SetupConfirmMsg{
						Msg: "WIPE SUPER and Flash: " + filepath.Base(p) + "?",
						Cmd: core.WipeSuper(p),
					}
				}
			}
		case "sideload":
			m.Modal.OnFileSelect = func(p string) tea.Cmd {
				return func() tea.Msg {
					core.LogChan <- "> adb sideload " + filepath.Base(p)
					return SetupConfirmMsg{
						Msg: "core.Sideload: " + filepath.Base(p) + "?",
						Cmd: core.Sideload(p),
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
			core.LogChan <- "> rebooting to system..."
			return core.RebootSystem(m.Device.Mode)
		}
	case "rb_recovery":
		m.ActiveModal, m.Modal.ConfirmMsg = ModalConfirm, "Reboot to Recovery?"
		m.Modal.OnConfirm = func() tea.Cmd {
			core.LogChan <- "> rebooting to recovery..."
			return core.RebootRecovery(m.Device.Mode)
		}
	case "refresh":
		m.IsRefreshing = true
		m.ActiveToast = &Toast{Message: "Scanning for devices...", Type: core.LogInfo}
		return m, tea.Batch(
			func() tea.Msg { return core.CheckDeviceState() },
			core.PollDeviceCmd(),
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
