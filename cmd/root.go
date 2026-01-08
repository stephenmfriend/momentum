package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// baseURL is the Flux server base URL
	baseURL string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "momentum",
	Short: "Momentum - A TUI client for Flux project management",
	Long: `Momentum is a terminal user interface for interacting with the Flux
project management system. It provides both interactive TUI mode and headless
mode for automation and scripting.

Because once the board starts moving, it shouldn't stop.

Examples:
  # Start interactive TUI mode
  momentum interactive

  # Run in headless mode with a specific task
  momentum headless --project myproject --task task-123

  # Use a custom Flux server URL
  momentum --base-url http://flux.example.com:3000 interactive`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "http://localhost:3000", "Flux server base URL")
}

// GetBaseURL returns the configured base URL for the Flux server
func GetBaseURL() string {
	return baseURL
}

// exitWithError prints an error message to stderr and exits with code 1
func exitWithError(msg string) {
	fmt.Fprintln(os.Stderr, "Error:", msg)
	os.Exit(1)
}
