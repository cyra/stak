package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
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
	// Get current working directory and add notes subdirectory
	cwd, _ := os.Getwd()
	notesDir := filepath.Join(cwd, "notes")

	return &Config{
		DataDir:     notesDir,
		LogLevel:    "info",
		Theme:       "default",
		DateFormat:  "2006-01-02",
		AutoSave:    true,
		FuzzySearch: true,
	}
}

func LoadConfig(configPath string) (*Config, error) {
	// Start with defaults
	config := DefaultConfig()

	// If no config path provided, try default locations
	if configPath == "" {
		homeDir, _ := os.UserHomeDir()
		possiblePaths := []string{
			filepath.Join(homeDir, ".stak", "config.yaml"),
			filepath.Join(homeDir, ".config", "stak", "config.yaml"),
			"stak.yaml",
			".stak.yaml",
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	// If config file exists, load it
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
			}
		}
	}

	// Expand relative paths to absolute
	if !filepath.IsAbs(config.DataDir) {
		if abs, err := filepath.Abs(config.DataDir); err == nil {
			config.DataDir = abs
		}
	}

	return config, nil
}

func (c *Config) Save(configPath string) error {
	if configPath == "" {
		homeDir, _ := os.UserHomeDir()
		configDir := filepath.Join(homeDir, ".stak")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		configPath = filepath.Join(configDir, "config.yaml")
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func (c *Config) EnsureDataDir() error {
	return os.MkdirAll(c.DataDir, 0755)
}

func CreateSampleConfig(path string) error {
	cwd, _ := os.Getwd()
	notesDir := filepath.Join(cwd, "notes")

	sampleConfig := &Config{
		DataDir:     notesDir,
		LogLevel:    "info",
		Theme:       "default",
		DateFormat:  "2006-01-02",
		AutoSave:    true,
		FuzzySearch: true,
	}

	return sampleConfig.Save(path)
}
