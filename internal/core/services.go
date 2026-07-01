package core

import (
	"context"
)

// Executor defines how shell commands are executed.
// This allows the Engine to inject a UI-aware executor that streams logs.
type Executor interface {
	RunCommand(ctx context.Context, name string, args ...string) error
}

// FlashService defines operations for flashing devices.
type FlashService interface {
	FlashImage(ctx context.Context, partition, path string) error
	WipeSuper(ctx context.Context, path string) error
	Sideload(ctx context.Context, path string) error
}

// DeviceService defines operations for managing device state.
type DeviceService interface {
	RebootSystem(ctx context.Context, currentMode string) error
	RebootRecovery(ctx context.Context, currentMode string) error
}

type defaultDeviceService struct {
	exec Executor
}

// NewDeviceService creates a new DeviceService.
func NewDeviceService(exec Executor) DeviceService {
	return &defaultDeviceService{exec: exec}
}

func (s *defaultDeviceService) RebootSystem(ctx context.Context, currentMode string) error {
	if currentMode == "FASTBOOT" {
		return s.exec.RunCommand(ctx, "fastboot", "reboot")
	}
	return s.exec.RunCommand(ctx, "adb", "reboot")
}

func (s *defaultDeviceService) RebootRecovery(ctx context.Context, currentMode string) error {
	if currentMode == "FASTBOOT" {
		// Some bootloaders do not support fastboot reboot recovery
		// But this is the standard syntax for those that do
		return s.exec.RunCommand(ctx, "fastboot", "reboot", "recovery")
	}
	return s.exec.RunCommand(ctx, "adb", "reboot", "recovery")
}

type defaultFlashService struct {
	exec Executor
}

// NewFlashService creates a new FlashService using the provided executor.
func NewFlashService(exec Executor) FlashService {
	return &defaultFlashService{exec: exec}
}

func (s *defaultFlashService) FlashImage(ctx context.Context, partition, path string) error {
	return s.exec.RunCommand(ctx, "fastboot", "flash", partition, path)
}

func (s *defaultFlashService) WipeSuper(ctx context.Context, path string) error {
	return s.exec.RunCommand(ctx, "fastboot", "wipe-super", path)
}

func (s *defaultFlashService) Sideload(ctx context.Context, path string) error {
	// For ADB sideload, we typically need to run it from the directory of the file,
	return s.exec.RunCommand(ctx, "adb", "sideload", path)
}
