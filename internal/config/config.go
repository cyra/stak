package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DataDir     string `yaml:"data_dir"`
	LogLevel    string `yaml:"log_level"`
	Theme       string `yaml:"theme"`
	DateFormat  string `yaml:"date_format"`
	AutoSave    bool   `yaml:"auto_save"`
	FuzzySearch bool   `yaml:"fuzzy_search"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		DataDir:     filepath.Join(homeDir, ".stak"),
		LogLevel:    "info",
		Theme:       "default",
		DateFormat:  "2006-01-02",
		AutoSave:    true,
		FuzzySearch: true,
	}
}

func (c *Config) EnsureDataDir() error {
	return os.MkdirAll(c.DataDir, 0755)
}