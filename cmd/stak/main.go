package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"stak/internal/config"
	"stak/pkg/ui"
)

var (
	configPath     = flag.String("config", "", "Path to config file")
	dataDir        = flag.String("dir", "", "Directory to store stak files")
	createConfig   = flag.Bool("create-config", false, "Create sample config file")
	showConfigPath = flag.Bool("show-config", false, "Show current config file location")
	version        = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("stak v1.0.0 - Your intelligent terminal scratchpad")
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override data directory if provided via flag
	if *dataDir != "" {
		if abs, err := filepath.Abs(*dataDir); err == nil {
			cfg.DataDir = abs
		} else {
			cfg.DataDir = *dataDir
		}
	}

	if *createConfig {
		configPath := "stak.yaml"
		if err := config.CreateSampleConfig(configPath); err != nil {
			fmt.Printf("Error creating sample config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Sample config created at %s\n", configPath)
		fmt.Println("Edit this file to customise your stak settings, then run stak again.")
		os.Exit(0)
	}

	if *showConfigPath {
		homeDir, _ := os.UserHomeDir()
		possiblePaths := []string{
			filepath.Join(homeDir, ".stak", "config.yaml"),
			filepath.Join(homeDir, ".config", "stak", "config.yaml"),
			"stak.yaml",
			".stak.yaml",
		}
		
		fmt.Printf("Data directory: %s\n", cfg.DataDir)
		fmt.Println("Config file search order:")
		for i, path := range possiblePaths {
			exists := ""
			if _, err := os.Stat(path); err == nil {
				exists = " (exists)"
			}
			fmt.Printf("  %d. %s%s\n", i+1, path, exists)
		}
		os.Exit(0)
	}

	log.SetLevel(log.InfoLevel)
	log.Info("Starting stak...", "dataDir", cfg.DataDir)

	model := ui.NewModelWithConfig(cfg)
	
	if err := model.Storage().Initialize(); err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}