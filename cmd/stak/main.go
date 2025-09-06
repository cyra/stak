package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"stak/pkg/ui"
)

func main() {
	log.SetLevel(log.InfoLevel)
	log.Info("Starting stak...")

	model := ui.NewModel()
	
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