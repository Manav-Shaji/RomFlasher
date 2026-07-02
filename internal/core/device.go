package core

import (
	"context"
)

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
