package core

import (
	"context"
	"flashtool/internal/config"
	"go.uber.org/zap"
	"os/exec"
	"path/filepath"
	"testing"
)

// Test sideload directory parsing
func TestRunCommand_Sideload_Dir(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.AppConfig{}

	_ = logger
	_ = cfg

	args := []string{"sideload", "/some/dir/rom.zip"}
	name := "adb"

	var cmd *exec.Cmd
	if name == "adb" && len(args) >= 2 && args[0] == "sideload" {
		newArgs := make([]string, len(args))
		copy(newArgs, args)
		newArgs[1] = filepath.Base(args[1])
		cmd = exec.CommandContext(context.Background(), name, newArgs...)
		cmd.Dir = filepath.Dir(args[1])
	} else {
		cmd = exec.CommandContext(context.Background(), name, args...)
	}

	expectedDir := filepath.Dir("/some/dir/rom.zip")
	if cmd.Dir != expectedDir {
		t.Errorf("expected cmd.Dir to be %s, got %s", expectedDir, cmd.Dir)
	}

	if cmd.Args[2] != "rom.zip" {
		t.Errorf("expected cmd.Args[2] to be rom.zip, got %s", cmd.Args[2])
	}
}
