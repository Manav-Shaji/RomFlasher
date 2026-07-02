package core

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"flashtool/internal/platform"
)

// FlashService defines operations for flashing devices.
type FlashService interface {
	ValidateDeviceForFlash(ctx context.Context, state platform.DeviceState, imgPath string) error
	ValidateDeviceForSideload(ctx context.Context, state platform.DeviceState, imgPath string) error
	FlashImage(ctx context.Context, partition, path string) error
	WipeSuper(ctx context.Context, path string) error
	Sideload(ctx context.Context, path string) error
}

type defaultFlashService struct {
	exec Executor
}

// NewFlashService creates a new FlashService using the provided executor.
func NewFlashService(exec Executor) FlashService {
	return &defaultFlashService{exec: exec}
}

func (s *defaultFlashService) ValidateDeviceForFlash(ctx context.Context, state platform.DeviceState, imgPath string) error {
	if state.Mode != platform.ModeFastboot {
		return fmt.Errorf("%w: expected FASTBOOT mode, got %s", ErrDeviceDisconnected, state.Mode)
	}
	if state.Secure != "NO" && state.Secure != "-" {
		return fmt.Errorf("%w: cannot flash while secure/locked (status: %s)", ErrBootloaderLocked, state.Secure)
	}

	// Basic battery check if available and valid
	if state.Battery != "" && state.Battery != "-" {
		if battStr := strings.TrimSuffix(state.Battery, " mV"); battStr != state.Battery {
			if mv, err := strconv.Atoi(battStr); err == nil && mv < 3500 { // Below ~3.5V is very low
				return fmt.Errorf("%w: battery at %d mV is too low", ErrBatteryLow, mv)
			}
		}
	}

	// Verify image exists
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrFlashFailed, "image file does not exist")
	}

	return nil
}

func (s *defaultFlashService) ValidateDeviceForSideload(ctx context.Context, state platform.DeviceState, imgPath string) error {
	if state.Mode != platform.ModeRecovery && state.Mode != platform.ModeSideload && state.Mode != platform.ModeDevice {
		return fmt.Errorf("%w: expected RECOVERY/SIDELOAD mode, got %s", ErrDeviceDisconnected, state.Mode)
	}
	
	// Verify zip exists
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrFlashFailed, "zip file does not exist")
	}
	
	if !strings.HasSuffix(strings.ToLower(imgPath), ".zip") {
		return fmt.Errorf("%w: expected .zip file for sideload", ErrFlashFailed)
	}

	return nil
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
