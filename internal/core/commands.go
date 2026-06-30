package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type LogMsg string
type ProgressMsg float64
type TaskCompleteMsg struct{ Err error }

var LogChan = make(chan string, 200)

var (
	cmdMu           sync.Mutex
	activeCmdCancel context.CancelFunc
)

func CancelActiveCommand() {
	cmdMu.Lock()
	defer cmdMu.Unlock()
	if activeCmdCancel != nil {
		activeCmdCancel()
	}
}

func WaitForLogs(ch chan string) tea.Cmd {
	return func() tea.Msg { return LogMsg(<-ch) }
}

func RunFlashCommand(name string, args ...string) tea.Cmd {
	return func() tea.Msg {
		// 1. Dependency Check (Check PATH then local folder for portability)
		if _, err := exec.LookPath(name); err != nil {
			// Check if binary exists in current directory
			localPath := "./" + name
			checkPath := localPath
			if runtime.GOOS == "windows" {
				checkPath += ".exe"
			}
			if _, err := os.Stat(checkPath); err != nil {
				msg := fmt.Sprintf("CRITICAL ERROR: %s not found in PATH or local folder", name)
				LogChan <- msg
				return TaskCompleteMsg{Err: fmt.Errorf(msg)}
			}
			// Use local path if found
			name = checkPath
		}

		// Send initial log with the command being run (simulating CMD prompt)
		pwd, _ := os.Getwd()
		
		displayCmd := name
		displayArgs := args
		// If running via cmd /c, show the actual command being passed
		if name == "cmd" && len(args) >= 2 && args[0] == "/c" {
			displayCmd = args[1]
			displayArgs = args[2:]
		}

		cmdPrompt := fmt.Sprintf("%s>%s", pwd, displayCmd)
		for _, a := range displayArgs { cmdPrompt += " " + a }
		LogChan <- cmdPrompt
		LogChan <- "STARTING COMMAND EXECUTION..."

		// 2. Context with Timeout (10 minutes for long flashes)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		
		cmdMu.Lock()
		activeCmdCancel = cancel
		cmdMu.Unlock()

		defer func() {
			cmdMu.Lock()
			activeCmdCancel = nil
			cmdMu.Unlock()
			cancel()
		}()

		var cmd *exec.Cmd
		if name == "adb" && len(args) >= 2 && args[0] == "sideload" {
			newArgs := make([]string, len(args))
			copy(newArgs, args)
			newArgs[1] = filepath.Base(args[1])
			cmd = exec.CommandContext(ctx, name, newArgs...)
			cmd.Dir = filepath.Dir(args[1])
		} else {
			cmd = exec.CommandContext(ctx, name, args...)
		}

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			LogChan <- fmt.Sprintf("FAILED START: %v", err)
			return TaskCompleteMsg{Err: err}
		}

		LogChan <- "STREAMERS ATTACHED. WAITING FOR OUTPUT..."

		var wg sync.WaitGroup
		wg.Add(2)

		stream := func(r io.ReadCloser) {
			defer wg.Done()
			sc := bufio.NewScanner(r)
			
			// Custom split function to handle both \n and \r
			sc.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
				if atEOF && len(data) == 0 { return 0, nil, nil }
				for i := 0; i < len(data); i++ {
					if data[i] == '\n' {
						return i + 1, data[0:i], nil
					}
					if data[i] == '\r' {
						// Mark this token as "terminated by \r" by adding a special character
						// We'll use a hidden prefix in the actual channel
						return i + 1, append([]byte("\r"), data[0:i]...), nil
					}
				}
				if atEOF { return len(data), data, nil }
				return 0, nil, nil
			})

			for sc.Scan() {
				token := sc.Bytes()
				if len(token) == 0 { continue }
				line := string(token)
				LogChan <- line
			}

		}

		go stream(stdout)
		go stream(stderr)

		err := cmd.Wait()
		wg.Wait()

		if err != nil {
			LogChan <- fmt.Sprintf("EXECUTION ERROR: %v", err)
		} else {
			LogChan <- "SUCCESS: Task finished successfully."
		}

		if ctx.Err() == context.Canceled {
			LogChan <- "ABORTED: Task was canceled by user."
			return TaskCompleteMsg{Err: fmt.Errorf("task canceled")}
		}

		if ctx.Err() == context.DeadlineExceeded {
			LogChan <- "TIMEOUT: Task exceeded the 10-minute limit."
			return TaskCompleteMsg{Err: fmt.Errorf("task timed out after 10m")}
		}

		return TaskCompleteMsg{Err: err}
	}
}

func RebootSystem(mode DeviceMode) tea.Cmd {
	if mode == ModeFastboot {
		return RunFlashCommand("fastboot", "reboot")
	}
	return RunFlashCommand("adb", "reboot")
}

func RebootRecovery(mode DeviceMode) tea.Cmd {
	if mode == ModeFastboot {
		return RunFlashCommand("fastboot", "reboot", "recovery")
	}
	return RunFlashCommand("adb", "reboot", "recovery")
}

func FlashImage(part, path string) tea.Cmd {
	return RunFlashCommand("fastboot", "flash", part, path)
}

func WipeSuper(path string) tea.Cmd {
	return RunFlashCommand("fastboot", "wipe-super", path)
}

func Sideload(path string) tea.Cmd {
	return RunFlashCommand("adb", "sideload", path)
}

func RunCustomCommand(cmdStr string) tea.Cmd {
	return RunFlashCommand("cmd", "/c", cmdStr)
}
