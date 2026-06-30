package config

import (
	"encoding/json"
	"os"
)

/* CONFIG MODELS */

type PickerConfig struct {
	Title  string `json:"title"`
	Filter string `json:"filter"`
	SubDir string `json:"sub_dir"`
}

type MenuConfig struct {
	Label  string        `json:"label"`
	Icon   string        `json:"icon"`
	Desc   string        `json:"desc"`
	Action string        `json:"action"`
	Picker *PickerConfig `json:"picker,omitempty"`
}

type AppConfig struct {
	BaseDir    string            `json:"base_dir"`
	DevicePath string            `json:"device_path,omitempty"`
	Folders    map[string]string `json:"folders"`
}

const configPath = "config.json"

// LoadConfig loads the AppConfig from the filesystem or returns an empty one
func LoadConfig() AppConfig {
	var cfg AppConfig
	data, err := os.ReadFile(configPath)
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	if cfg.Folders == nil {
		cfg.Folders = make(map[string]string)
	}
	return cfg
}

// SaveConfig saves the AppConfig to the filesystem
func SaveConfig(cfg AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
