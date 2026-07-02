package core

import (
	"errors"
	"fmt"
)

var (
	ErrDeviceDisconnected = errors.New("device disconnected")
	ErrBootloaderLocked   = errors.New("bootloader is locked")
	ErrHashMismatch       = errors.New("hash mismatch")
	ErrUnsupportedDevice  = errors.New("unsupported device")
	ErrFlashFailed        = errors.New("flash failed")
	ErrFastboot           = errors.New("fastboot error")
	ErrBatteryLow         = errors.New("battery level too low")
)

// FlashError wraps an underlying error with additional context.
type FlashError struct {
	Partition string
	Err       error
}

func (e *FlashError) Error() string {
	return fmt.Sprintf("failed to flash partition '%s': %v", e.Partition, e.Err)
}

func (e *FlashError) Unwrap() error {
	return e.Err
}
