package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sirsjg/momentum/agent"
)

func TestNewModel(t *testing.T) {
	model := NewModel("Test criteria", ExecutionModeAsync, nil)

	if model.criteria != "Test criteria" {
		t.Errorf("expected criteria 'Test criteria', got %q", model.criteria)
	}
	if model.panels == nil {
		t.Error("panels should be initialized")
	}
	if len(model.panels) != 0 {
		t.Errorf("expected 0 panels, got %d", len(model.panels))
	}
	if model.mode != ExecutionModeAsync {
		t.Errorf("expected mode async, got %s", model.mode.String())
	}
	if model.agentUpdates == nil {
		t.Error("agentUpdates channel should be initialized")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"very short max", "hello", 3, "hel"},
		{"max length 4", "hello world", 4, "h..."},
		{"empty string", "", 10, ""},
		{"zero max", "hello", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestAgentPanel_IsRunning(t *testing.T) {
	// Panel with nil runner
	panel := &AgentPanel{}
	if panel.IsRunning() {
		t.Error("panel with nil runner should not be running")
	}

	// Note: Testing with actual runner would require mocking
}

func TestAgentPanel_IsFinished(t *testing.T) {
	// Panel without result
	panel := &AgentPanel{}
	if panel.IsFinished() {
		t.Error("panel without result should not be finished")
	}

	// Panel with result
	panel.Result = &agent.Result{ExitCode: 0}
	if !panel.IsFinished() {
		t.Error("panel with result should be finished")
	}
}

func TestModel_Update_WindowSizeMsg(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}

func TestModel_Update_ListenerConnectedMsg(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	if model.connected {
		t.Error("model should not be connected initially")
	}

	newModel, _ := model.Update(ListenerConnectedMsg{})
	m := newModel.(*Model)

	if !m.connected {
		t.Error("model should be connected after ListenerConnectedMsg")
	}
	if !m.listening {
		t.Error("model should be listening after ListenerConnectedMsg")
	}
}

func TestModel_Update_ListenerErrorMsg(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	err := &testError{msg: "test error"}
	newModel, _ := model.Update(ListenerErrorMsg{Err: err})
	m := newModel.(*Model)

	if m.lastError == nil {
		t.Error("lastError should be set")
	}
	if m.lastError.Error() != "test error" {
		t.Errorf("expected error message 'test error', got %q", m.lastError.Error())
	}
}

func TestModel_Update_AddAgentMsg(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	if len(model.panels) != 0 {
		t.Error("model should have no panels initially")
	}

	newModel, _ := model.Update(AddAgentMsg{
		TaskID:    "task-1",
		TaskTitle: "Test Task",
		AgentName: "Claude",
		Runner:    nil,
	})
	m := newModel.(*Model)

	if len(m.panels) != 1 {
		t.Errorf("expected 1 panel, got %d", len(m.panels))
	}
	if m.panels[0].TaskID != "task-1" {
		t.Errorf("expected task ID 'task-1', got %q", m.panels[0].TaskID)
	}
	if m.panels[0].TaskTitle != "Test Task" {
		t.Errorf("expected task title 'Test Task', got %q", m.panels[0].TaskTitle)
	}
	if !m.panels[0].Focused {
		t.Error("first panel should be focused")
	}
}

func TestModel_Update_AddMultipleAgents(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	newModel, _ := model.Update(AddAgentMsg{TaskID: "task-2", TaskTitle: "Task 2", AgentName: "Claude"})
	m := newModel.(*Model)

	if len(m.panels) != 2 {
		t.Errorf("expected 2 panels, got %d", len(m.panels))
	}
	// First panel should still be focused
	if !m.panels[0].Focused {
		t.Error("first panel should remain focused")
	}
	if m.panels[1].Focused {
		t.Error("second panel should not be focused")
	}
}

func TestModel_Update_AgentOutputMsg(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)
	model.width = 100
	model.height = 50

	// Add a panel first
	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})

	// Send output
	line := agent.OutputLine{
		Text:      `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}`,
		Timestamp: time.Now(),
	}
	newModel, _ := model.Update(AgentOutputMsg{TaskID: "task-1", Line: line})
	m := newModel.(*Model)

	if len(m.panels[0].Output) != 1 {
		t.Errorf("expected 1 output line, got %d", len(m.panels[0].Output))
	}
	if m.panels[0].Output[0].Text != "Hello" {
		t.Errorf("expected parsed output 'Hello', got %q", m.panels[0].Output[0].Text)
	}
}

func TestModel_Update_AgentOutputMsg_SkipsEmptyParsed(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)
	model.width = 100
	model.height = 50

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})

	// Send output that parses to empty (like a ping message)
	line := agent.OutputLine{
		Text:      `{"type":"ping"}`,
		Timestamp: time.Now(),
	}
	newModel, _ := model.Update(AgentOutputMsg{TaskID: "task-1", Line: line})
	m := newModel.(*Model)

	if len(m.panels[0].Output) != 0 {
		t.Errorf("expected 0 output lines for skipped message, got %d", len(m.panels[0].Output))
	}
}

func TestModel_Update_AgentCompletedMsg(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})

	result := agent.Result{ExitCode: 0}
	newModel, _ := model.Update(AgentCompletedMsg{TaskID: "task-1", Result: result})
	m := newModel.(*Model)

	if m.panels[0].Result == nil {
		t.Error("panel should have result set")
	}
	if m.panels[0].Result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", m.panels[0].Result.ExitCode)
	}
	if m.taskCount != 1 {
		t.Errorf("expected task count 1, got %d", m.taskCount)
	}
}

func TestModel_HandleKeyPress_Quit(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	_, cmd := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestModel_HandleKeyPress_Tab(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	// Add two panels
	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	model.Update(AddAgentMsg{TaskID: "task-2", TaskTitle: "Task 2", AgentName: "Claude"})

	if model.focusedPanel != 0 {
		t.Error("first panel should be focused initially")
	}

	// Tab to next panel
	newModel, _ := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyTab})
	m := newModel.(*Model)

	if m.focusedPanel != 1 {
		t.Errorf("expected focused panel 1, got %d", m.focusedPanel)
	}
	if !m.panels[1].Focused {
		t.Error("second panel should be focused")
	}
	if m.panels[0].Focused {
		t.Error("first panel should not be focused")
	}
}

func TestModel_HandleKeyPress_Tab_Wraps(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	model.Update(AddAgentMsg{TaskID: "task-2", TaskTitle: "Task 2", AgentName: "Claude"})
	model.focusedPanel = 1
	model.panels[0].Focused = false
	model.panels[1].Focused = true

	// Tab should wrap to first panel
	newModel, _ := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyTab})
	m := newModel.(*Model)

	if m.focusedPanel != 0 {
		t.Errorf("expected focused panel 0 (wrapped), got %d", m.focusedPanel)
	}
}

func TestModel_HandleKeyPress_ShiftTab(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	model.Update(AddAgentMsg{TaskID: "task-2", TaskTitle: "Task 2", AgentName: "Claude"})

	// Shift+Tab should wrap to last panel
	newModel, _ := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyShiftTab})
	m := newModel.(*Model)

	if m.focusedPanel != 1 {
		t.Errorf("expected focused panel 1, got %d", m.focusedPanel)
	}
}

func TestModel_HandleKeyPress_CloseFinishedPanel(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	model.panels[0].Result = &agent.Result{ExitCode: 0}

	// Press 'x' to close
	newModel, _ := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m := newModel.(*Model)

	if len(m.panels) != 0 {
		t.Errorf("expected 0 panels after closing, got %d", len(m.panels))
	}
}

func TestModel_HandleKeyPress_CannotCloseRunningPanel(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	// Panel has no result, so it's still "running"

	// Press 'x' should not close
	newModel, _ := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m := newModel.(*Model)

	if len(m.panels) != 1 {
		t.Errorf("expected panel to remain, got %d panels", len(m.panels))
	}
}

func TestModel_HandleKeyPress_ScrollUp(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)
	model.height = 50

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	model.panels[0].ScrollPos = 10

	newModel, _ := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyUp})
	m := newModel.(*Model)

	if m.panels[0].ScrollPos != 7 {
		t.Errorf("expected scroll pos 7, got %d", m.panels[0].ScrollPos)
	}
}

func TestModel_HandleKeyPress_ScrollUp_AtTop(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	model.panels[0].ScrollPos = 1

	newModel, _ := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyUp})
	m := newModel.(*Model)

	if m.panels[0].ScrollPos != 0 {
		t.Errorf("expected scroll pos 0 (clamped), got %d", m.panels[0].ScrollPos)
	}
}

func TestModel_HandleKeyPress_ScrollDown(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)
	model.height = 50

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	// Add some output lines
	for i := 0; i < 100; i++ {
		model.panels[0].Output = append(model.panels[0].Output, agent.OutputLine{Text: "line"})
	}

	newModel, _ := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown})
	m := newModel.(*Model)

	if m.panels[0].ScrollPos != 3 {
		t.Errorf("expected scroll pos 3, got %d", m.panels[0].ScrollPos)
	}
}

func TestModel_SetListening(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	model.SetListening(true)
	if !model.listening {
		t.Error("expected listening to be true")
	}

	model.SetListening(false)
	if model.listening {
		t.Error("expected listening to be false")
	}
}

func TestModel_SetConnected(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	model.SetConnected(true)
	if !model.connected {
		t.Error("expected connected to be true")
	}

	model.SetConnected(false)
	if model.connected {
		t.Error("expected connected to be false")
	}
}

func TestModel_SetError(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	err := &testError{msg: "test error"}
	model.SetError(err)
	if model.lastError == nil {
		t.Error("expected error to be set")
	}

	model.SetError(nil)
	if model.lastError != nil {
		t.Error("expected error to be nil")
	}
}

func TestModel_GetOpenPanelCount(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	if model.GetOpenPanelCount() != 0 {
		t.Error("expected 0 panels initially")
	}

	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	if model.GetOpenPanelCount() != 1 {
		t.Error("expected 1 panel")
	}

	model.Update(AddAgentMsg{TaskID: "task-2", TaskTitle: "Task 2", AgentName: "Claude"})
	if model.GetOpenPanelCount() != 2 {
		t.Error("expected 2 panels")
	}
}

func TestModel_HasRunningAgents(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	if model.HasRunningAgents() {
		t.Error("expected no running agents initially")
	}

	// Add a panel without a result (still running concept - though Runner is nil)
	model.Update(AddAgentMsg{TaskID: "task-1", TaskTitle: "Task 1", AgentName: "Claude"})
	// Since Runner is nil, IsRunning() returns false
	if model.HasRunningAgents() {
		t.Error("expected no running agents when runner is nil")
	}
}

func TestModel_GetUpdateChannel(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	ch := model.GetUpdateChannel()
	if ch == nil {
		t.Error("expected non-nil channel")
	}
}

func TestModel_AddAgent(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)

	id := model.AddAgent("task-1", "Task 1", "Claude", nil)

	if id == "" {
		t.Error("expected non-empty ID")
	}
	if len(model.panels) != 1 {
		t.Error("expected 1 panel")
	}
	if model.panels[0].ID != id {
		t.Error("panel ID should match returned ID")
	}
}

func TestModel_View_EmptyWidth(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)
	model.width = 0

	result := model.View()
	if result != "" {
		t.Error("expected empty view when width is 0")
	}
}

func TestModel_GetPanelHeight(t *testing.T) {
	model := NewModel("test", ExecutionModeAsync, nil)
	model.height = 50

	height := model.getPanelHeight()
	if height != 42 { // 50 - 8
		t.Errorf("expected panel height 42, got %d", height)
	}
}

// Helper error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
