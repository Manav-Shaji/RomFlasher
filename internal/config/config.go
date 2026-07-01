package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

/* CONFIG MODELS */

type PickerConfig struct {
	Title  string `json:"title" mapstructure:"title"`
	Filter string `json:"filter" mapstructure:"filter"`
	SubDir string `json:"sub_dir" mapstructure:"sub_dir"`
}

type MenuConfig struct {
	Label  string        `json:"label" mapstructure:"label"`
	Icon   string        `json:"icon" mapstructure:"icon"`
	Desc   string        `json:"desc" mapstructure:"desc"`
	Action string        `json:"action" mapstructure:"action"`
	Picker *PickerConfig `json:"picker,omitempty" mapstructure:"picker"`
}

type LogConfig struct {
	Level    string `json:"level" mapstructure:"level"`
	Format   string `json:"format" mapstructure:"format"`
	Dir      string `json:"dir" mapstructure:"dir"`
	Filename string `json:"filename" mapstructure:"filename"`
}

type AppConfig struct {
	BaseDir    string            `json:"base_dir" mapstructure:"base_dir"`
	DevicePath string            `json:"device_path,omitempty" mapstructure:"device_path"`
	Folders    map[string]string `json:"folders" mapstructure:"folders"`
	Log        LogConfig         `json:"log" mapstructure:"log"`
}

// Load loads the AppConfig from the filesystem via Viper
func Load() (*AppConfig, error) {
	v := viper.New()

	// Set Defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.dir", "logs")
	v.SetDefault("log.filename", "app.log")

	pwd, _ := os.Getwd()
	v.SetDefault("base_dir", pwd)
	v.SetDefault("folders", map[string]string{})

	// Setup Env variables (e.g. NEXFORGE_LOG_LEVEL)
	v.SetEnvPrefix("NEXFORGE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Search Paths
	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(home, ".nexforge"))
	}

	// Read Config
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Unmarshal into Struct
	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.Folders == nil {
		cfg.Folders = make(map[string]string)
	}

	return &cfg, nil
}

// SaveConfig saves the AppConfig to the filesystem (legacy support for TUI Settings)
func SaveConfig(cfg *AppConfig) error {
	v := viper.New()
	v.Set("base_dir", cfg.BaseDir)
	v.Set("device_path", cfg.DevicePath)
	v.Set("folders", cfg.Folders)
	v.Set("log", cfg.Log)

	if err := v.WriteConfigAs("config.json"); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}
