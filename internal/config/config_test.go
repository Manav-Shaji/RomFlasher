package config

import (
	"os"
	"testing"
)

func TestAppConfig_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    AppConfig
		expected string // Expected log level after validate
	}{
		{
			name:     "empty config gets defaults",
			input:    AppConfig{},
			expected: "info",
		},
		{
			name: "invalid log level falls back to info",
			input: AppConfig{
				Log: LogConfig{Level: "superdebug"},
			},
			expected: "info",
		},
		{
			name: "valid log level remains unchanged",
			input: AppConfig{
				Log: LogConfig{Level: "warn"},
			},
			expected: "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.input.Log.Level != tt.expected {
				t.Errorf("expected log level %s, got %s", tt.expected, tt.input.Log.Level)
			}
			if tt.input.BaseDir == "" {
				t.Errorf("expected BaseDir to be populated")
			}
		})
	}
	
	// Cleanup any created default log directory
	_ = os.Remove("logs")
}
