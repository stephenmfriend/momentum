// Package selection provides task selection logic for the Momentum headless mode.
package selection

import (
	"errors"
	"fmt"
	"sort"

	"github.com/stevegrehan/momentum/client"
)

// ErrNoTaskAvailable is returned when no suitable task can be found.
var ErrNoTaskAvailable = errors.New("no task available matching the selection criteria")

// Selector handles task selection logic for headless mode.
// It supports filtering by project, epic, or specific task ID.
type Selector struct {
	client    *client.Client
	projectID string
	epicID    string
	taskID    string
}

// NewSelector creates a new Selector with the given filters.
// All filter parameters are optional - pass empty strings if not needed.
func NewSelector(c *client.Client, projectID, epicID, taskID string) *Selector {
	return &Selector{
		client:    c,
		projectID: projectID,
		epicID:    epicID,
		taskID:    taskID,
	}
}

// SelectTask selects a task based on the configured filters.
// The selection logic follows this priority:
//  1. If taskID is provided, fetch that specific task
//  2. If epicID is provided, get the first unblocked todo task from that epic
//  3. If projectID is provided, get the first unblocked todo task from that project
//  4. If nothing is provided, get the newest unblocked todo task across ALL projects
//
// Within each scope, tasks are prioritized by:
//   - Unblocked tasks (blocked=false) come first
//   - Tasks with status "todo" are preferred
//   - Newer tasks (by ID, assuming lexicographic order reflects creation time) come first
func (s *Selector) SelectTask() (*client.Task, error) {
	// Case 1: Specific task ID provided
	if s.taskID != "" {
		return s.fetchSpecificTask()
	}

	// Case 2: Epic ID provided - get tasks from that epic's project filtered by epic
	if s.epicID != "" {
		return s.selectFromEpic()
	}

	// Case 3: Project ID provided - get tasks from that project
	if s.projectID != "" {
		return s.selectFromProject(s.projectID)
	}

	// Case 4: No filters - search across all projects
	return s.selectFromAllProjects()
}

// fetchSpecificTask fetches a task by its ID.
// Since the client doesn't have a GetTask method, we need to find it
// by listing tasks from all projects.
func (s *Selector) fetchSpecificTask() (*client.Task, error) {
	// We need to find the task across all projects since we don't know which project it belongs to
	projects, err := s.client.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	for _, project := range projects {
		tasks, err := s.client.ListTasks(project.ID, client.TaskFilters{})
		if err != nil {
			// Log but continue searching other projects
			continue
		}

		for i := range tasks {
			if tasks[i].ID == s.taskID {
				return &tasks[i], nil
			}
		}
	}

	return nil, fmt.Errorf("task %s not found: %w", s.taskID, ErrNoTaskAvailable)
}

// selectFromEpic selects the best task from the specified epic.
func (s *Selector) selectFromEpic() (*client.Task, error) {
	// First, we need to find which project this epic belongs to
	projects, err := s.client.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Find the project containing the epic
	var targetProjectID string
	for _, project := range projects {
		epics, err := s.client.ListEpics(project.ID)
		if err != nil {
			continue
		}

		for _, epic := range epics {
			if epic.ID == s.epicID {
				targetProjectID = project.ID
				break
			}
		}
		if targetProjectID != "" {
			break
		}
	}

	if targetProjectID == "" {
		return nil, fmt.Errorf("epic %s not found: %w", s.epicID, ErrNoTaskAvailable)
	}

	// Get tasks filtered by epic
	filters := client.TaskFilters{
		EpicID: client.StringPtr(s.epicID),
	}
	tasks, err := s.client.ListTasks(targetProjectID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks for epic %s: %w", s.epicID, err)
	}

	return s.selectBestTask(tasks)
}

// selectFromProject selects the best task from the specified project.
func (s *Selector) selectFromProject(projectID string) (*client.Task, error) {
	tasks, err := s.client.ListTasks(projectID, client.TaskFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks for project %s: %w", projectID, err)
	}

	return s.selectBestTask(tasks)
}

// selectFromAllProjects selects the best task across all projects.
func (s *Selector) selectFromAllProjects() (*client.Task, error) {
	projects, err := s.client.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found: %w", ErrNoTaskAvailable)
	}

	var allTasks []client.Task
	for _, project := range projects {
		tasks, err := s.client.ListTasks(project.ID, client.TaskFilters{})
		if err != nil {
			// Log but continue with other projects
			continue
		}
		allTasks = append(allTasks, tasks...)
	}

	return s.selectBestTask(allTasks)
}

// selectBestTask selects the best task from a list based on priority:
// 1. Unblocked tasks come first
// 2. Tasks with status "todo" are preferred
// 3. Newer tasks (by ID, reverse lexicographic order) come first
func (s *Selector) selectBestTask(tasks []client.Task) (*client.Task, error) {
	if len(tasks) == 0 {
		return nil, ErrNoTaskAvailable
	}

	// Filter and sort tasks
	candidates := filterAndSortTasks(tasks)

	if len(candidates) == 0 {
		return nil, ErrNoTaskAvailable
	}

	return &candidates[0], nil
}

// filterAndSortTasks filters tasks to only include unblocked todos
// and sorts them by priority (newer first).
func filterAndSortTasks(tasks []client.Task) []client.Task {
	var candidates []client.Task

	// First pass: collect unblocked todo tasks
	for _, task := range tasks {
		if !task.Blocked && task.Status == "todo" {
			candidates = append(candidates, task)
		}
	}

	// If no unblocked todos, try unblocked tasks of any status
	if len(candidates) == 0 {
		for _, task := range tasks {
			if !task.Blocked {
				candidates = append(candidates, task)
			}
		}
	}

	// If still no candidates, try any todo tasks (even blocked)
	if len(candidates) == 0 {
		for _, task := range tasks {
			if task.Status == "todo" {
				candidates = append(candidates, task)
			}
		}
	}

	// Last resort: take any task
	if len(candidates) == 0 {
		candidates = tasks
	}

	// Sort by ID descending (assuming newer tasks have "larger" IDs)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].ID > candidates[j].ID
	})

	return candidates
}
