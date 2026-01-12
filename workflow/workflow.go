// Package workflow provides task status management operations for Momentum.
// It offers a reusable interface for transitioning tasks between statuses,
// suitable for both headless and interactive modes.
package workflow

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirsjg/momentum/client"
)

// Workflow provides methods for managing task status transitions.
type Workflow struct {
	client *client.Client
	out    io.Writer
}

// NewWorkflow creates a new Workflow instance with the provided client.
func NewWorkflow(client *client.Client) *Workflow {
	return &Workflow{
		client: client,
		out:    os.Stdout,
	}
}

// SetOutput configures where workflow status messages are written.
// Use io.Discard to silence output (e.g., when a TUI is active).
func (w *Workflow) SetOutput(out io.Writer) {
	w.out = out
}

// StartWorking transitions the specified tasks to "in_progress" status.
// It iterates through all provided task IDs, attempting to update each one.
// If any task fails to update, it continues with the remaining tasks and
// returns an aggregate error describing all failures.
func (w *Workflow) StartWorking(taskIDs []string) error {
	return w.updateTasksStatus(taskIDs, "in_progress", "Starting work on")
}

// MarkComplete transitions the specified tasks to "done" status.
// It iterates through all provided task IDs, attempting to update each one.
// If any task fails to update, it continues with the remaining tasks and
// returns an aggregate error describing all failures.
func (w *Workflow) MarkComplete(taskIDs []string) error {
	return w.updateTasksStatus(taskIDs, "done", "Marking complete")
}

// ResetTask transitions the specified tasks back to "todo" status.
// It iterates through all provided task IDs, attempting to update each one.
// If any task fails to update, it continues with the remaining tasks and
// returns an aggregate error describing all failures.
func (w *Workflow) ResetTask(taskIDs []string) error {
	return w.updateTasksStatus(taskIDs, "todo", "Resetting")
}

// ResetToPlanning transitions the specified tasks back to "planning" status.
// This is typically used when a user stops an agent mid-execution.
// It iterates through all provided task IDs, attempting to update each one.
// If any task fails to update, it continues with the remaining tasks and
// returns an aggregate error describing all failures.
func (w *Workflow) ResetToPlanning(taskIDs []string) error {
	return w.updateTasksStatus(taskIDs, "planning", "Resetting to planning")
}

// updateTasksStatus is the internal method that handles status updates for all tasks.
// It processes each task ID, prints status messages, handles errors gracefully,
// and returns an aggregate error if any updates failed.
func (w *Workflow) updateTasksStatus(taskIDs []string, status, actionVerb string) error {
	if len(taskIDs) == 0 {
		return nil
	}

	var failedTasks []string
	var errorMessages []string

	for _, taskID := range taskIDs {
		w.printf("%s task %s...\n", actionVerb, taskID)

		task, err := w.client.MoveTaskStatus(taskID, status)
		if err != nil {
			w.printf("  Failed to update task %s: %v\n", taskID, err)
			failedTasks = append(failedTasks, taskID)
			errorMessages = append(errorMessages, fmt.Sprintf("task %s: %v", taskID, err))
			continue
		}

		w.printf("  Task %s (%s) -> %s\n", taskID, task.Title, status)
	}

	if len(failedTasks) > 0 {
		return errors.New("failed to update tasks: " + strings.Join(errorMessages, "; "))
	}

	return nil
}

func (w *Workflow) printf(format string, args ...any) {
	if w.out == nil {
		return
	}
	fmt.Fprintf(w.out, format, args...)
}
