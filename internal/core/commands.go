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

type Engine struct {
	LogChan         chan string
	cmdMu           sync.Mutex
	activeCmdCancel context.CancelFunc
}

func NewEngine() *Engine {
	return &Engine{
		LogChan: make(chan string, 200),
	}
}

func (e *Engine) CancelActiveCommand() {
	e.cmdMu.Lock()
	defer e.cmdMu.Unlock()
	if e.activeCmdCancel != nil {
		e.activeCmdCancel()
	}
}

func (e *Engine) WaitForLogs() tea.Cmd {
	return func() tea.Msg { return LogMsg(<-e.LogChan) }
}

func (e *Engine) RunFlashCommand(name string, args ...string) tea.Cmd {
	return func() tea.Msg {
		if _, err := exec.LookPath(name); err != nil {
			localPath := "./" + name
			checkPath := localPath
			if runtime.GOOS == "windows" {
				checkPath += ".exe"
			}
			if _, err := os.Stat(checkPath); err != nil {
				msg := fmt.Sprintf("CRITICAL ERROR: %s not found in PATH or local folder", name)
				e.LogChan <- msg
				return TaskCompleteMsg{Err: fmt.Errorf("%s", msg)}
			}
			name = checkPath
		}

		pwd, _ := os.Getwd()
		displayCmd := name
		displayArgs := args
		if name == "cmd" && len(args) >= 2 && args[0] == "/c" {
			displayCmd = args[1]
			displayArgs = args[2:]
		}

		cmdPrompt := fmt.Sprintf("%s>%s", pwd, displayCmd)
		for _, a := range displayArgs { cmdPrompt += " " + a }
		e.LogChan <- cmdPrompt
		e.LogChan <- "STARTING COMMAND EXECUTION..."

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		
		e.cmdMu.Lock()
		e.activeCmdCancel = cancel
		e.cmdMu.Unlock()

		defer func() {
			e.cmdMu.Lock()
			e.activeCmdCancel = nil
			e.cmdMu.Unlock()
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
			e.LogChan <- fmt.Sprintf("FAILED START: %v", err)
			return TaskCompleteMsg{Err: err}
		}

		e.LogChan <- "STREAMERS ATTACHED. WAITING FOR OUTPUT..."

		var wg sync.WaitGroup
		wg.Add(2)

		stream := func(r io.ReadCloser) {
			defer wg.Done()
			sc := bufio.NewScanner(r)
			sc.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
				if atEOF && len(data) == 0 { return 0, nil, nil }
				for i := 0; i < len(data); i++ {
					if data[i] == '\n' {
						return i + 1, data[0:i], nil
					}
					if data[i] == '\r' {
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
				e.LogChan <- line
			}
			if err := sc.Err(); err != nil { e.LogChan <- fmt.Sprintf("scanner error: %v", err) }
		}

		go stream(stdout)
		go stream(stderr)

		err := cmd.Wait()
		wg.Wait()

		if err != nil {
			e.LogChan <- fmt.Sprintf("EXECUTION ERROR: %v", err)
		} else {
			e.LogChan <- "SUCCESS: Task finished successfully."
		}

		if ctx.Err() == context.Canceled {
			e.LogChan <- "ABORTED: Task was canceled by user."
			return TaskCompleteMsg{Err: fmt.Errorf("task canceled")}
		}

		if ctx.Err() == context.DeadlineExceeded {
			e.LogChan <- "TIMEOUT: Task exceeded the 10-minute limit."
			return TaskCompleteMsg{Err: fmt.Errorf("task timed out after 10m")}
		}

		return TaskCompleteMsg{Err: err}
	}
}

func (e *Engine) RebootSystem(mode DeviceMode) tea.Cmd {
	if mode == ModeFastboot {
		return e.RunFlashCommand("fastboot", "reboot")
	}
	return e.RunFlashCommand("adb", "reboot")
}

func (e *Engine) RebootRecovery(mode DeviceMode) tea.Cmd {
	if mode == ModeFastboot {
		return e.RunFlashCommand("fastboot", "reboot", "recovery")
	}
	return e.RunFlashCommand("adb", "reboot", "recovery")
}

func (e *Engine) FlashImage(part, path string) tea.Cmd {
	return e.RunFlashCommand("fastboot", "flash", part, path)
}

func (e *Engine) WipeSuper(path string) tea.Cmd {
	return e.RunFlashCommand("fastboot", "wipe-super", path)
}

func (e *Engine) Sideload(path string) tea.Cmd {
	return e.RunFlashCommand("adb", "sideload", path)
}

func (e *Engine) RunCustomCommand(cmdStr string) tea.Cmd {
	return e.RunFlashCommand("cmd", "/c", cmdStr)
}
