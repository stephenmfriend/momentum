package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stevegrehan/momentum/client"
	"github.com/stevegrehan/momentum/selection"
)

var (
	// headless mode flags
	taskID    string
	epicID    string
	projectID string
)

// headlessCmd represents the headless command
var headlessCmd = &cobra.Command{
	Use:   "headless",
	Short: "Run Momentum in headless mode for automation",
	Long: `Run Momentum in headless mode without a user interface.

This mode is designed for automation, scripting, and CI/CD pipelines.
Use flags to specify which project, epic, or task to work with.

If no flags are specified, the newest unblocked todo task across all projects
will be selected automatically.

Examples:
  # Auto-select newest unblocked todo task from any project
  momentum headless

  # Work with a specific project
  momentum headless --project proj-123

  # Work with a specific epic in a project
  momentum headless --epic epic-456

  # Work with a specific task
  momentum headless --task task-789

  # Combine with custom server URL
  momentum --base-url http://flux.example.com:3000 headless --project myproject`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHeadless()
	},
}

func init() {
	rootCmd.AddCommand(headlessCmd)

	// Headless mode specific flags
	headlessCmd.Flags().StringVar(&taskID, "task", "", "Task ID to work with")
	headlessCmd.Flags().StringVar(&epicID, "epic", "", "Epic ID to work with")
	headlessCmd.Flags().StringVar(&projectID, "project", "", "Project ID to work with")
}

// runHeadless executes the headless mode logic
func runHeadless() error {
	fmt.Printf("Running in headless mode...\n")
	fmt.Printf("Connecting to Flux server at: %s\n", GetBaseURL())
	fmt.Println()

	// Create the REST client
	c := client.NewClient(GetBaseURL())

	// Create the selector with the provided filters
	selector := selection.NewSelector(c, projectID, epicID, taskID)

	// Log the selection criteria
	if taskID != "" {
		fmt.Printf("Selection criteria: specific task %s\n", taskID)
	} else if epicID != "" {
		fmt.Printf("Selection criteria: first unblocked todo from epic %s\n", epicID)
	} else if projectID != "" {
		fmt.Printf("Selection criteria: first unblocked todo from project %s\n", projectID)
	} else {
		fmt.Printf("Selection criteria: newest unblocked todo across all projects\n")
	}
	fmt.Println()

	// Select a task
	task, err := selector.SelectTask()
	if err != nil {
		if errors.Is(err, selection.ErrNoTaskAvailable) {
			fmt.Println("No task available matching the selection criteria.")
			return nil
		}
		return fmt.Errorf("failed to select task: %w", err)
	}

	// Print the selected task details
	fmt.Println("Selected task:")
	fmt.Println("==============")
	fmt.Printf("  ID:        %s\n", task.ID)
	fmt.Printf("  Title:     %s\n", task.Title)
	fmt.Printf("  Status:    %s\n", task.Status)
	fmt.Printf("  Blocked:   %t\n", task.Blocked)
	fmt.Printf("  Project:   %s\n", task.ProjectID)
	if task.EpicID != "" {
		fmt.Printf("  Epic:      %s\n", task.EpicID)
	}
	if task.Notes != "" {
		fmt.Printf("  Notes:     %s\n", task.Notes)
	}
	if len(task.DependsOn) > 0 {
		fmt.Printf("  Depends on: %v\n", task.DependsOn)
	}
	fmt.Println()

	// TODO: Wire up status updates in another task
	fmt.Println("Next steps (not yet implemented):")
	fmt.Printf("  - Would move task %s to 'in_progress'\n", task.ID)
	fmt.Printf("  - Would execute task work\n")
	fmt.Printf("  - Would move task %s to 'done' on completion\n", task.ID)

	return nil
}
