package core

import (
	"context"
	"testing"
	"flashtool/internal/platform"
)

func TestValidateDeviceForFlash(t *testing.T) {
	fs := &defaultFlashService{}

	tests := []struct {
		name      string
		state     platform.DeviceState
		imgPath   string
		expectErr bool
	}{
		{
			name: "valid fastboot mode",
			state: platform.DeviceState{
				Mode:   platform.ModeFastboot,
				Secure: "NO",
			},
			imgPath:   "non_existent.img",
			expectErr: true, // Should fail on file existence
		},
		{
			name: "invalid mode",
			state: platform.DeviceState{
				Mode: platform.ModeDevice,
			},
			imgPath:   "dummy.img",
			expectErr: true,
		},
		{
			name: "locked bootloader",
			state: platform.DeviceState{
				Mode:   platform.ModeFastboot,
				Secure: "YES",
			},
			imgPath:   "dummy.img",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fs.ValidateDeviceForFlash(context.Background(), tt.state, tt.imgPath)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}
