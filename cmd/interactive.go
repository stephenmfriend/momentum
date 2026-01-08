package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/stevegrehan/momentum/tui"
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start Momentum in interactive TUI mode",
	Long: `Start Momentum in interactive Terminal User Interface (TUI) mode.

This mode provides a full-screen interactive interface for managing your
Flux projects, epics, and tasks. Navigate using keyboard shortcuts and
enjoy a rich visual experience.

Examples:
  # Start interactive mode with default server
  momentum interactive

  # Start interactive mode with custom server URL
  momentum --base-url http://flux.example.com:3000 interactive`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractive()
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

// runInteractive starts the interactive TUI mode
func runInteractive() error {
	// Create the TUI model with the configured base URL
	model := tui.NewModel(GetBaseURL())

	// Create and run the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
