package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"

	"flashtool/internal/platform"
)

type LogLevel string

const (
	LogInfo    LogLevel = "INFO"
	LogError   LogLevel = "ERROR"
	LogSuccess LogLevel = "SUCCESS"
)

type LogEntry struct {
	Level     LogLevel
	Text      string
	Timestamp time.Time
}

// Executor defines how shell commands are executed.
// This allows the Engine to inject a UI-aware executor that streams logs.
type Executor interface {
	RunCommand(ctx context.Context, name string, args ...string) error
}

type LogMsg string
type ProgressMsg float64
type TaskCompleteMsg struct{ Err error }

// Engine orchestrates background operations, streaming logs and managing cancellations.
type Engine struct {
	LogChan chan string
	logger  *zap.Logger

	FlashService  FlashService
	DeviceService DeviceService

	cmdMu           sync.Mutex
	activeCmdCancel context.CancelFunc
}

// NewEngine creates a new core Engine.
func NewEngine(logger *zap.Logger) *Engine {
	platform.ExtractEmbeddedBinaries()

	e := &Engine{
		LogChan: make(chan string, 200),
		logger:  logger,
	}

	e.FlashService = NewFlashService(e)
	e.DeviceService = NewDeviceService(e)

	return e
}

// CancelActiveCommand allows the UI or CLI to cancel the currently running background command.
func (e *Engine) CancelActiveCommand() {
	e.cmdMu.Lock()
	defer e.cmdMu.Unlock()
	if e.activeCmdCancel != nil {
		e.activeCmdCancel()
		e.logger.Info("User initiated command cancellation")
	}
}

// WaitForLogs creates a tea.Cmd that blocks until a log message is available.
func (e *Engine) WaitForLogs() tea.Cmd {
	return func() tea.Msg { return LogMsg(<-e.LogChan) }
}

// ExecuteAsync wraps a context-aware service action into a Bubble Tea command.
func (e *Engine) ExecuteAsync(action func(context.Context) error) tea.Cmd {
	return func() tea.Msg {
		e.cmdMu.Lock()
		ctx, cancel := context.WithCancel(context.Background())
		e.activeCmdCancel = cancel
		e.cmdMu.Unlock()

		defer func() {
			e.cmdMu.Lock()
			e.activeCmdCancel = nil
			e.cmdMu.Unlock()
			cancel()
		}()

		err := action(ctx)
		return TaskCompleteMsg{Err: err}
	}
}

// RunCommand implements Executor, running raw shell commands and piping output.
func (e *Engine) RunCommand(ctx context.Context, name string, args ...string) error {
	allowedCmds := map[string]bool{"adb": true, "fastboot": true, "cmd": true}
	if !allowedCmds[name] {
		err := fmt.Errorf("security violation: execution of binary '%s' is not permitted", name)
		e.logger.Error("Command rejected by security policy", zap.String("name", name))
		return err
	}

	e.logger.Debug("Resolving command path", zap.String("name", name))
	resolvedPath, err := platform.ResolveCommandPath(name)
	if err != nil {
		msg := fmt.Sprintf("CRITICAL ERROR: %s", err.Error())
		e.LogChan <- msg
		e.logger.Error("Binary not found", zap.String("name", name), zap.Error(err))
		return fmt.Errorf("action failed: %w", err)
	}
	name = resolvedPath

	pwd, _ := os.Getwd()
	displayCmd := name
	displayArgs := args
	if name == "cmd" && len(args) >= 2 && args[0] == "/c" {
		displayCmd = args[1]
		displayArgs = args[2:]
	}

	cmdPrompt := fmt.Sprintf("%s>%s", pwd, displayCmd)
	for _, a := range displayArgs {
		cmdPrompt += " " + a
	}

	e.LogChan <- cmdPrompt
	e.LogChan <- "STARTING COMMAND EXECUTION..."
	e.logger.Info("Executing command", zap.String("command", cmdPrompt))

	// Ensure there is a hard timeout if the context doesn't already have one
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
	}

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
		e.logger.Error("Failed to start command", zap.Error(err))
		return fmt.Errorf("action failed: %w", err)
	}

	e.LogChan <- "STREAMERS ATTACHED. WAITING FOR OUTPUT..."

	var wg sync.WaitGroup
	wg.Add(2)

	stream := func(r io.ReadCloser) {
		defer wg.Done()
		sc := bufio.NewScanner(r)
		sc.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			for i := 0; i < len(data); i++ {
				if data[i] == '\n' || data[i] == '\r' {
					return i + 1, data[0 : i+1], nil
				}
			}
			if atEOF {
				return len(data), append(data, '\n'), nil
			}
			return 0, nil, nil
		})

		for sc.Scan() {
			token := sc.Bytes()
			if len(token) == 0 {
				continue
			}

			isCR := token[len(token)-1] == '\r'
			token = token[:len(token)-1]
			if len(token) == 0 {
				continue
			}

			line := string(token)
			if isCR {
				line = "\r" + line
			}

			select {
			case e.LogChan <- line:
			case <-time.After(50 * time.Millisecond):
				e.logger.Warn("LogChan backpressure, dropping log", zap.String("line", line))
			}
			e.logger.Debug("Process Output", zap.String("line", line))
		}
		if err := sc.Err(); err != nil {
			select {
			case e.LogChan <- fmt.Sprintf("scanner error: %v", err):
			default:
			}
			e.logger.Warn("Scanner error", zap.Error(err))
		}
	}

	go stream(stdout)
	go stream(stderr)

	startExec := time.Now()
	err = cmd.Wait()
	wg.Wait()
	elapsed := time.Since(startExec)

	if err != nil {
		e.LogChan <- fmt.Sprintf("EXECUTION ERROR: %v", err)
		e.logger.Error("Command execution returned error", zap.Error(err), zap.Duration("elapsed", elapsed))
	} else {
		e.LogChan <- fmt.Sprintf("SUCCESS: Task finished successfully in %v.", elapsed.Round(time.Millisecond))
		e.logger.Info("Command executed successfully", zap.Duration("elapsed", elapsed))
	}

	if ctx.Err() == context.Canceled {
		e.LogChan <- "ABORTED: Task was canceled by user."
		e.logger.Warn("Command was canceled via context")
		return fmt.Errorf("task canceled")
	}

	if ctx.Err() == context.DeadlineExceeded {
		e.LogChan <- "TIMEOUT: Task exceeded the time limit."
		e.logger.Warn("Command exceeded timeout limit")
		return fmt.Errorf("task timed out")
	}

	if err != nil {
		return fmt.Errorf("action failed: %w", err)
	}
	return nil
}

// RunCustomCommand is kept for the TUI custom command modal.
func (e *Engine) RunCustomCommand(ctx context.Context, cmdStr string) error {
	return e.RunCommand(ctx, "cmd", "/c", cmdStr)
}
