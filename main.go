package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joe-yeomans/ecsd/ecs"
	"github.com/joe-yeomans/ecsd/ui"

	"github.com/aws/aws-sdk-go-v2/config"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading AWS config: %v\n", err)
		os.Exit(1)
	}

	// Get AWS profile from environment
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "default"
	}

	// Create ECS client
	client := ecs.NewClient(cfg, profile)

	// Create and run the TUI
	model := ui.NewModel(client)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
