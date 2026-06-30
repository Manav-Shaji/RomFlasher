package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	fmt.Println("Starting restructure...")

	// 1. Create directories
	dirs := []string{
		"internal/core",
		"internal/tui",
		"internal/tui/theme",
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}

	// 2. Define file moves and their new package
	type move struct {
		oldPath    string
		newPath    string
		newPackage string
	}

	moves := []move{
		{"internal/commands.go", "internal/core/commands.go", "core"},
		{"internal/device.go", "internal/core/device.go", "core"},
		{"internal/device_db.go", "internal/core/device_db.go", "core"},
		
		{"internal/model.go", "internal/tui/model.go", "tui"},
		{"internal/setup.go", "internal/tui/setup.go", "tui"},
		{"internal/update.go", "internal/tui/update.go", "tui"},
		{"internal/view.go", "internal/tui/view.go", "tui"},
		{"internal/view_modals.go", "internal/tui/view_modals.go", "tui"},
		{"internal/view_panels.go", "internal/tui/view_panels.go", "tui"},
		{"internal/handlers.go", "internal/tui/handlers.go", "tui"},

		{"internal/ui/styles.go", "internal/tui/theme/styles.go", "theme"},
		{"internal/ui/theme.go", "internal/tui/theme/theme.go", "theme"},
	}

	for _, m := range moves {
		content, err := ioutil.ReadFile(m.oldPath)
		if err != nil {
			fmt.Printf("Skipping %s: %v\n", m.oldPath, err)
			continue
		}

		strContent := string(content)
		
		// Update package declaration
		if strings.HasPrefix(m.oldPath, "internal/ui/") {
			strContent = strings.Replace(strContent, "package ui", "package theme", 1)
		} else {
			strContent = strings.Replace(strContent, "package internal", "package "+m.newPackage, 1)
		}

		// Update imports
		strContent = strings.ReplaceAll(strContent, "\"flashtool/internal/ui\"", "\"flashtool/internal/tui/theme\"")
		strContent = strings.ReplaceAll(strContent, "ui.", "theme.")
		
		// If in tui, we might need to import core if it uses device stuff
		// Actually they are in same folder right now, so they didn't import each other.
		// If we split them, tui files need to import "flashtool/internal/core".
		// We'll add it if they mention DeviceMode, LogLevel, etc.
		if m.newPackage == "tui" {
			if strings.Contains(strContent, "DeviceMode") || strings.Contains(strContent, "DeviceState") || strings.Contains(strContent, "LogEntry") || strings.Contains(strContent, "LogLevel") {
				// We need to import core. Let's do a naive replacement to add import.
				// Since we don't have a sophisticated AST parser, we'll just inject it.
				if !strings.Contains(strContent, "\"flashtool/internal/core\"") {
					strContent = strings.Replace(strContent, "import (", "import (\n\t\"flashtool/internal/core\"\n", 1)
				}
				
				// And prepend core. to core types
				coreTypes := []string{"DeviceMode", "DeviceState", "LogEntry", "LogLevel", "ModeDisconnected", "ModeFastboot", "ModeDevice", "ModeRecovery", "ModeSideload", "ModeUnauthorized", "ModeOffline", "LogInfo", "LogError", "LogSuccess", "CancelActiveCommand", "RebootSystem", "RebootRecovery", "FlashImage", "WipeSuper", "Sideload", "RunCustomCommand", "PollDeviceCmd", "WaitForLogs", "TaskCompleteMsg", "LogMsg", "LogChan"}
				
				for _, ct := range coreTypes {
					// A very hacky but effective regex-like replacement for simple Go code
					strContent = strings.ReplaceAll(strContent, ct, "core."+ct)
				}
				// Fix double core (e.g. core.core.DeviceMode)
				strContent = strings.ReplaceAll(strContent, "core.core.", "core.")
				// Fix definitions in core package that might have been renamed? No, this is for tui package.
			}
		}

		err = ioutil.WriteFile(m.newPath, []byte(strContent), 0644)
		if err != nil {
			fmt.Printf("Failed to write %s: %v\n", m.newPath, err)
		} else {
			os.Remove(m.oldPath) // Delete old file
			fmt.Printf("Moved %s -> %s\n", m.oldPath, m.newPath)
		}
	}

	// 3. Update main.go
	mainContent, err := ioutil.ReadFile("cmd/flashtool/main.go")
	if err == nil {
		str := string(mainContent)
		str = strings.ReplaceAll(str, "\"flashtool/internal\"", "\"flashtool/internal/tui\"")
		str = strings.ReplaceAll(str, "internal.NewModel()", "tui.NewModel()")
		ioutil.WriteFile("cmd/flashtool/main.go", []byte(str), 0644)
		fmt.Println("Updated cmd/flashtool/main.go")
	}

	// 4. Delete old abandoned files
	os.Remove("internal/ui/model.go")
	os.Remove("internal/ui/ui_test.go")
	os.RemoveAll("internal/ui")

	fmt.Println("Restructure complete!")
}
