// Package tui provides an interactive terminal user interface for Momentum.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stevegrehan/momentum/client"
)

// Color palette
var (
	purple    = lipgloss.Color("#7C3AED")
	cyan      = lipgloss.Color("#06B6D4")
	green     = lipgloss.Color("#10B981")
	amber     = lipgloss.Color("#F59E0B")
	red       = lipgloss.Color("#EF4444")
	gray      = lipgloss.Color("#6B7280")
	darkGray  = lipgloss.Color("#374151")
	lightGray = lipgloss.Color("#9CA3AF")
	white     = lipgloss.Color("#F9FAFB")
)

// Styles
var (
	// App container
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	// Header styles
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(purple)

	taglineStyle = lipgloss.NewStyle().
			Foreground(gray).
			Italic(true)

	// Breadcrumb styles
	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			MarginBottom(1)

	breadcrumbActiveStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Bold(true)

	breadcrumbSepStyle = lipgloss.NewStyle().
				Foreground(darkGray)

	// List title styles
	titleStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(purple).
			Padding(0, 1).
			Bold(true)

	titleInactiveStyle = lipgloss.NewStyle().
				Foreground(white).
				Background(darkGray).
				Padding(0, 1)

	// Pane styles
	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(darkGray)

	focusedPaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(purple)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(darkGray).
			Padding(0, 1)

	statusAccentStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Background(darkGray).
				Bold(true)

	// Help text
	helpStyle = lipgloss.NewStyle().
			Foreground(gray)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(cyan)

	// Status indicators
	todoStyle       = lipgloss.NewStyle().Foreground(gray)
	inProgressStyle = lipgloss.NewStyle().Foreground(amber)
	doneStyle       = lipgloss.NewStyle().Foreground(green)
	blockedStyle    = lipgloss.NewStyle().Foreground(red)
	selectedStyle   = lipgloss.NewStyle().Foreground(purple).Bold(true)

	// Empty state
	emptyStyle = lipgloss.NewStyle().
			Foreground(gray).
			Italic(true).
			Align(lipgloss.Center)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)
)

// Pane constants
const (
	PaneProjects = iota
	PaneEpics
	PaneTasks
)

// projectItem implements list.Item
type projectItem struct {
	project    client.Project
	tasksDone  int
	tasksTotal int
}

func (i projectItem) Title() string {
	if i.tasksTotal > 0 {
		return fmt.Sprintf("%s  %s", i.project.Name,
			lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("(%d/%d)", i.tasksDone, i.tasksTotal)))
	}
	return i.project.Name
}
func (i projectItem) Description() string { return i.project.Description }
func (i projectItem) FilterValue() string { return i.project.Name }

// epicItem implements list.Item
type epicItem struct {
	epic client.Epic
}

func (i epicItem) Title() string {
	var icon string
	switch i.epic.Status {
	case "in_progress":
		icon = inProgressStyle.Render("◐")
	case "done":
		icon = doneStyle.Render("●")
	default:
		icon = todoStyle.Render("○")
	}
	return fmt.Sprintf("%s %s", icon, i.epic.Title)
}
func (i epicItem) Description() string { return i.epic.Notes }
func (i epicItem) FilterValue() string { return i.epic.Title }

// taskItem implements list.Item
type taskItem struct {
	task     client.Task
	selected bool
}

func (i taskItem) Title() string {
	var parts []string

	if i.selected {
		parts = append(parts, selectedStyle.Render("◉"))
	}

	switch i.task.Status {
	case "in_progress":
		parts = append(parts, inProgressStyle.Render("▶"))
	case "done":
		parts = append(parts, doneStyle.Render("✓"))
	default:
		parts = append(parts, todoStyle.Render("○"))
	}

	if i.task.Blocked {
		parts = append(parts, blockedStyle.Render("⚠"))
	}

	title := i.task.Title
	if i.task.Status == "done" {
		title = lipgloss.NewStyle().Foreground(gray).Render(title)
	}
	parts = append(parts, title)

	return strings.Join(parts, " ")
}
func (i taskItem) Description() string { return i.task.Notes }
func (i taskItem) FilterValue() string { return i.task.Title }

// StatusFilter for tasks
type StatusFilter int

const (
	FilterAll StatusFilter = iota
	FilterTodo
	FilterInProgress
	FilterDone
)

func (f StatusFilter) Label() string {
	switch f {
	case FilterTodo:
		return "Todo"
	case FilterInProgress:
		return "In Progress"
	case FilterDone:
		return "Done"
	default:
		return "All"
	}
}

// Model is the TUI state
type Model struct {
	client        *client.Client
	spinner       spinner.Model
	projectList   list.Model
	epicList      list.Model
	taskList      list.Model
	focusedPane   int
	selectedTasks map[string]bool
	statusFilter  StatusFilter
	allTasks      []client.Task
	width         int
	height        int
	loading       bool
	err           error
}

// NewModel creates a new TUI model
func NewModel(baseURL string) Model {
	// Spinner setup
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(purple)

	// Custom delegate with accent colors
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(purple).
		BorderLeftForeground(purple)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(gray).
		BorderLeftForeground(purple)

	// Project list
	projectList := list.New([]list.Item{}, delegate, 0, 0)
	projectList.Title = "Projects"
	projectList.Styles.Title = titleStyle
	projectList.SetShowHelp(false)
	projectList.SetFilteringEnabled(true)
	projectList.Styles.NoItems = emptyStyle

	// Epic list
	epicList := list.New([]list.Item{}, delegate, 0, 0)
	epicList.Title = "Epics"
	epicList.Styles.Title = titleInactiveStyle
	epicList.SetShowHelp(false)
	epicList.SetFilteringEnabled(true)
	epicList.Styles.NoItems = emptyStyle

	// Task list
	taskList := list.New([]list.Item{}, delegate, 0, 0)
	taskList.Title = "Tasks"
	taskList.Styles.Title = titleInactiveStyle
	taskList.SetShowHelp(false)
	taskList.SetFilteringEnabled(true)
	taskList.Styles.NoItems = emptyStyle

	return Model{
		client:        client.NewClient(baseURL),
		spinner:       s,
		projectList:   projectList,
		epicList:      epicList,
		taskList:      taskList,
		selectedTasks: make(map[string]bool),
		statusFilter:  FilterAll,
		focusedPane:   PaneProjects,
		loading:       true,
	}
}

// Messages
type projectsLoadedMsg struct {
	projects []client.Project
	stats    map[string]projectStats
	err      error
}

type projectStats struct {
	tasksDone  int
	tasksTotal int
}

type epicsLoadedMsg struct {
	epics []client.Epic
	err   error
}

type tasksLoadedMsg struct {
	tasks []client.Task
	err   error
}

type tasksUpdatedMsg struct {
	err error
}

// Init starts the TUI
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadProjects())
}

func (m Model) loadProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.client.ListProjects()
		if err != nil {
			return projectsLoadedMsg{err: err}
		}

		stats := make(map[string]projectStats)
		for _, p := range projects {
			tasks, err := m.client.ListTasks(p.ID, client.TaskFilters{})
			if err != nil {
				continue
			}
			done := 0
			for _, t := range tasks {
				if t.Status == "done" {
					done++
				}
			}
			stats[p.ID] = projectStats{tasksDone: done, tasksTotal: len(tasks)}
		}

		return projectsLoadedMsg{projects: projects, stats: stats}
	}
}

func (m Model) loadEpics() tea.Cmd {
	if len(m.projectList.Items()) == 0 {
		return nil
	}
	item, ok := m.projectList.SelectedItem().(projectItem)
	if !ok {
		return nil
	}
	projectID := item.project.ID
	return func() tea.Msg {
		epics, err := m.client.ListEpics(projectID)
		return epicsLoadedMsg{epics: epics, err: err}
	}
}

func (m Model) loadTasks() tea.Cmd {
	if len(m.projectList.Items()) == 0 {
		return nil
	}
	item, ok := m.projectList.SelectedItem().(projectItem)
	if !ok {
		return nil
	}
	projectID := item.project.ID
	return func() tea.Msg {
		tasks, err := m.client.ListTasks(projectID, client.TaskFilters{})
		return tasksLoadedMsg{tasks: tasks, err: err}
	}
}

func (m Model) startSelectedTasks() tea.Cmd {
	if len(m.selectedTasks) == 0 {
		return nil
	}
	return func() tea.Msg {
		var lastErr error
		for taskID := range m.selectedTasks {
			_, err := m.client.MoveTaskStatus(taskID, "in_progress")
			if err != nil {
				lastErr = err
			}
		}
		return tasksUpdatedMsg{err: lastErr}
	}
}

// Update handles events
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		paneWidth := (msg.Width - 14) / 3
		paneHeight := msg.Height - 10

		m.projectList.SetSize(paneWidth, paneHeight)
		m.epicList.SetSize(paneWidth, paneHeight)
		m.taskList.SetSize(paneWidth, paneHeight)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.isFiltering() {
			return m.updateFocusedList(msg)
		}
		return m.handleKeyPress(msg)

	case projectsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		items := make([]list.Item, len(msg.projects))
		for i, p := range msg.projects {
			stats := msg.stats[p.ID]
			items[i] = projectItem{
				project:    p,
				tasksDone:  stats.tasksDone,
				tasksTotal: stats.tasksTotal,
			}
		}
		m.projectList.SetItems(items)
		if len(msg.projects) > 0 {
			return m, tea.Batch(m.loadEpics(), m.loadTasks())
		}
		return m, nil

	case epicsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		items := make([]list.Item, len(msg.epics))
		for i, e := range msg.epics {
			items[i] = epicItem{epic: e}
		}
		m.epicList.SetItems(items)
		return m, nil

	case tasksLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.allTasks = msg.tasks
		m.applyTaskFilter()
		return m, nil

	case tasksUpdatedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.selectedTasks = make(map[string]bool)
		return m, m.loadTasks()
	}

	return m.updateFocusedList(msg)
}

func (m Model) isFiltering() bool {
	return m.projectList.FilterState() == list.Filtering ||
		m.epicList.FilterState() == list.Filtering ||
		m.taskList.FilterState() == list.Filtering
}

func (m Model) updateFocusedList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focusedPane {
	case PaneProjects:
		m.projectList, cmd = m.projectList.Update(msg)
	case PaneEpics:
		m.epicList, cmd = m.epicList.Update(msg)
	case PaneTasks:
		m.taskList, cmd = m.taskList.Update(msg)
	}
	return m, cmd
}

func (m *Model) applyTaskFilter() {
	// Get selected epic ID (if any)
	var selectedEpicID string
	if item, ok := m.epicList.SelectedItem().(epicItem); ok {
		selectedEpicID = item.epic.ID
	}

	var filtered []client.Task
	for _, t := range m.allTasks {
		// Filter by epic - only show tasks belonging to the selected epic
		if selectedEpicID != "" && t.EpicID != selectedEpicID {
			continue
		}

		// Filter by status
		switch m.statusFilter {
		case FilterAll:
			filtered = append(filtered, t)
		case FilterTodo:
			if t.Status == "todo" {
				filtered = append(filtered, t)
			}
		case FilterInProgress:
			if t.Status == "in_progress" {
				filtered = append(filtered, t)
			}
		case FilterDone:
			if t.Status == "done" {
				filtered = append(filtered, t)
			}
		}
	}

	items := make([]list.Item, len(filtered))
	for i, t := range filtered {
		items[i] = taskItem{task: t, selected: m.selectedTasks[t.ID]}
	}
	m.taskList.SetItems(items)

	title := "Tasks"
	if m.statusFilter != FilterAll {
		title = fmt.Sprintf("Tasks [%s]", m.statusFilter.Label())
	}
	m.taskList.Title = title
}

func (m *Model) updateTitleStyles() {
	m.projectList.Styles.Title = titleInactiveStyle
	m.epicList.Styles.Title = titleInactiveStyle
	m.taskList.Styles.Title = titleInactiveStyle

	switch m.focusedPane {
	case PaneProjects:
		m.projectList.Styles.Title = titleStyle
	case PaneEpics:
		m.epicList.Styles.Title = titleStyle
	case PaneTasks:
		m.taskList.Styles.Title = titleStyle
	}
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "tab", "l":
		m.focusedPane = (m.focusedPane + 1) % 3
		m.updateTitleStyles()
		return m, nil

	case "shift+tab", "h":
		m.focusedPane = (m.focusedPane + 2) % 3
		m.updateTitleStyles()
		return m, nil

	case " ":
		if m.focusedPane == PaneTasks && len(m.taskList.Items()) > 0 {
			if item, ok := m.taskList.SelectedItem().(taskItem); ok {
				taskID := item.task.ID
				if m.selectedTasks[taskID] {
					delete(m.selectedTasks, taskID)
				} else {
					m.selectedTasks[taskID] = true
				}
				m.applyTaskFilter()
			}
		}
		return m, nil

	case "enter":
		switch m.focusedPane {
		case PaneTasks:
			if len(m.selectedTasks) > 0 {
				return m, m.startSelectedTasks()
			}
		case PaneProjects:
			return m, tea.Batch(m.loadEpics(), m.loadTasks())
		}
		return m, nil

	case "r":
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.loadProjects())

	case "f":
		m.statusFilter = (m.statusFilter + 1) % 4
		m.applyTaskFilter()
		return m, nil
	}

	var cmd tea.Cmd
	switch m.focusedPane {
	case PaneProjects:
		m.projectList, cmd = m.projectList.Update(msg)
		if msg.String() == "j" || msg.String() == "k" || msg.String() == "up" || msg.String() == "down" {
			return m, tea.Batch(cmd, m.loadEpics(), m.loadTasks())
		}
	case PaneEpics:
		m.epicList, cmd = m.epicList.Update(msg)
		if msg.String() == "j" || msg.String() == "k" || msg.String() == "up" || msg.String() == "down" {
			m.applyTaskFilter()
		}
	case PaneTasks:
		m.taskList, cmd = m.taskList.Update(msg)
	}

	return m, cmd
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Header
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		logoStyle.Render("⚡ MOMENTUM"),
		"  ",
		taglineStyle.Render("keep the board moving"),
	)
	b.WriteString(header)
	b.WriteString("\n")

	// Breadcrumb
	breadcrumb := m.renderBreadcrumb()
	b.WriteString(breadcrumb)
	b.WriteString("\n")

	// Error
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	// Loading state
	if m.loading {
		loading := lipgloss.NewStyle().
			Width(m.width - 4).
			Align(lipgloss.Center).
			Padding(4, 0).
			Render(m.spinner.View() + "  Loading projects...")
		b.WriteString(loading)
		return appStyle.Render(b.String())
	}

	// Panes
	paneWidth := (m.width - 14) / 3

	projectPane := paneStyle
	epicPane := paneStyle
	taskPane := paneStyle

	switch m.focusedPane {
	case PaneProjects:
		projectPane = focusedPaneStyle
	case PaneEpics:
		epicPane = focusedPaneStyle
	case PaneTasks:
		taskPane = focusedPaneStyle
	}

	// Empty state messages
	projectView := m.projectList.View()
	if len(m.projectList.Items()) == 0 && !m.loading {
		projectView = emptyStyle.Width(paneWidth - 4).Render("\n\nNo projects yet\n\nCreate one in Flux!")
	}

	epicView := m.epicList.View()
	if len(m.epicList.Items()) == 0 {
		epicView = emptyStyle.Width(paneWidth - 4).Render("\n\nNo epics\n\nSelect a project")
	}

	taskView := m.taskList.View()
	if len(m.taskList.Items()) == 0 {
		msg := "\n\nNo tasks"
		if m.statusFilter != FilterAll {
			msg = fmt.Sprintf("\n\nNo %s tasks\n\nPress f to change filter", m.statusFilter.Label())
		}
		taskView = emptyStyle.Width(paneWidth - 4).Render(msg)
	}

	panes := lipgloss.JoinHorizontal(lipgloss.Top,
		projectPane.Width(paneWidth).Render(projectView),
		epicPane.Width(paneWidth).Render(epicView),
		taskPane.Width(paneWidth).Render(taskView),
	)
	b.WriteString(panes)
	b.WriteString("\n")

	// Status bar
	var statusParts []string
	if m.statusFilter != FilterAll {
		statusParts = append(statusParts, fmt.Sprintf("Filter: %s", m.statusFilter.Label()))
	}
	if len(m.selectedTasks) > 0 {
		statusParts = append(statusParts, statusAccentStyle.Render(fmt.Sprintf("%d selected", len(m.selectedTasks))))
	}
	if len(m.selectedTasks) > 0 {
		statusParts = append(statusParts, "Press Enter to start working")
	}

	statusText := "Ready"
	if len(statusParts) > 0 {
		statusText = strings.Join(statusParts, "  •  ")
	}
	b.WriteString(statusBarStyle.Width(m.width - 4).Render(statusText))
	b.WriteString("\n")

	// Help
	help := helpKeyStyle.Render("↑↓") + helpStyle.Render(" nav  ") +
		helpKeyStyle.Render("Tab") + helpStyle.Render(" pane  ") +
		helpKeyStyle.Render("Space") + helpStyle.Render(" select  ") +
		helpKeyStyle.Render("/") + helpStyle.Render(" search  ") +
		helpKeyStyle.Render("f") + helpStyle.Render(" filter  ") +
		helpKeyStyle.Render("r") + helpStyle.Render(" refresh  ") +
		helpKeyStyle.Render("q") + helpStyle.Render(" quit")
	b.WriteString(help)

	return appStyle.Render(b.String())
}

func (m Model) renderBreadcrumb() string {
	var parts []string

	// Current project
	if item, ok := m.projectList.SelectedItem().(projectItem); ok {
		if m.focusedPane == PaneProjects {
			parts = append(parts, breadcrumbActiveStyle.Render(item.project.Name))
		} else {
			parts = append(parts, breadcrumbStyle.Render(item.project.Name))
		}
	}

	// Current epic (if selected and in epics/tasks pane)
	if m.focusedPane >= PaneEpics {
		if item, ok := m.epicList.SelectedItem().(epicItem); ok {
			if m.focusedPane == PaneEpics {
				parts = append(parts, breadcrumbActiveStyle.Render(item.epic.Title))
			} else {
				parts = append(parts, breadcrumbStyle.Render(item.epic.Title))
			}
		}
	}

	if len(parts) == 0 {
		return breadcrumbStyle.Render("Select a project to get started")
	}

	sep := breadcrumbSepStyle.Render(" › ")
	return strings.Join(parts, sep)
}
