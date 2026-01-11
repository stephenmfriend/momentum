package ui

import (
	"strings"
	"testing"
)

func TestParseClaudeOutput_EmptyInput(t *testing.T) {
	result := parseClaudeOutput("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestParseClaudeOutput_WhitespaceOnly(t *testing.T) {
	result := parseClaudeOutput("   \n\t  ")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestParseClaudeOutput_PlainText(t *testing.T) {
	result := parseClaudeOutput("Hello, this is plain text")
	if result != "Hello, this is plain text" {
		t.Errorf("expected plain text returned as-is, got %q", result)
	}
}

func TestParseClaudeOutput_InvalidJSON(t *testing.T) {
	result := parseClaudeOutput("{not valid json")
	if result != "{not valid json" {
		t.Errorf("expected invalid JSON returned as-is, got %q", result)
	}
}

func TestParseClaudeOutput_AssistantMessage_TextBlock(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello world"}]}}`
	result := parseClaudeOutput(input)
	if result != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", result)
	}
}

func TestParseClaudeOutput_AssistantMessage_MultipleTextBlocks(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"},{"type":"text","text":"world"}]}}`
	result := parseClaudeOutput(input)
	if result != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", result)
	}
}

func TestParseClaudeOutput_AssistantMessage_ToolUse(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"read_file"}]}}`
	result := parseClaudeOutput(input)
	if result != "[Tool: read_file]" {
		t.Errorf("expected '[Tool: read_file]', got %q", result)
	}
}

func TestParseClaudeOutput_AssistantMessage_MixedContent(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Reading file"},{"type":"tool_use","name":"read_file"}]}}`
	result := parseClaudeOutput(input)
	if result != "Reading file [Tool: read_file]" {
		t.Errorf("expected 'Reading file [Tool: read_file]', got %q", result)
	}
}

func TestParseClaudeOutput_AssistantMessage_EmptyContent(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[]}}`
	result := parseClaudeOutput(input)
	if result != "" {
		t.Errorf("expected empty string for empty content, got %q", result)
	}
}

func TestParseClaudeOutput_AssistantMessage_EmptyTextBlock(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":""}]}}`
	result := parseClaudeOutput(input)
	if result != "" {
		t.Errorf("expected empty string for empty text block, got %q", result)
	}
}

func TestParseClaudeOutput_ContentBlockDelta(t *testing.T) {
	input := `{"type":"content_block_delta","delta":{"text":"streaming text"}}`
	result := parseClaudeOutput(input)
	if result != "streaming text" {
		t.Errorf("expected 'streaming text', got %q", result)
	}
}

func TestParseClaudeOutput_ContentBlockDelta_EmptyText(t *testing.T) {
	input := `{"type":"content_block_delta","delta":{"text":""}}`
	result := parseClaudeOutput(input)
	if result != "" {
		t.Errorf("expected empty string for empty delta text, got %q", result)
	}
}

func TestParseClaudeOutput_ContentBlockDelta_NoDelta(t *testing.T) {
	input := `{"type":"content_block_delta"}`
	result := parseClaudeOutput(input)
	if result != "" {
		t.Errorf("expected empty string for missing delta, got %q", result)
	}
}

func TestParseClaudeOutput_Error_WithMessage(t *testing.T) {
	input := `{"type":"error","error":{"message":"Something went wrong"}}`
	result := parseClaudeOutput(input)
	if result != "[Error: Something went wrong]" {
		t.Errorf("expected '[Error: Something went wrong]', got %q", result)
	}
}

func TestParseClaudeOutput_Error_NoMessage(t *testing.T) {
	input := `{"type":"error","error":{}}`
	result := parseClaudeOutput(input)
	if result != "[Error]" {
		t.Errorf("expected '[Error]', got %q", result)
	}
}

func TestParseClaudeOutput_Error_NoErrorObject(t *testing.T) {
	input := `{"type":"error"}`
	result := parseClaudeOutput(input)
	if result != "[Error]" {
		t.Errorf("expected '[Error]', got %q", result)
	}
}

func TestParseClaudeOutput_SkippedTypes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"start", `{"type":"start"}`},
		{"stop", `{"type":"stop"}`},
		{"ping", `{"type":"ping"}`},
		{"message_start", `{"type":"message_start"}`},
		{"message_stop", `{"type":"message_stop"}`},
		{"content_block_start", `{"type":"content_block_start"}`},
		{"content_block_stop", `{"type":"content_block_stop"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseClaudeOutput(tt.input)
			if result != "" {
				t.Errorf("expected empty string for type %s, got %q", tt.name, result)
			}
		})
	}
}

func TestParseClaudeOutput_UnknownType(t *testing.T) {
	input := `{"type":"unknown_type","data":"something"}`
	result := parseClaudeOutput(input)
	if result != "" {
		t.Errorf("expected empty string for unknown type, got %q", result)
	}
}

func TestParseClaudeOutput_JSONWithNoType(t *testing.T) {
	input := `{"message":"no type field"}`
	result := parseClaudeOutput(input)
	if result != "" {
		t.Errorf("expected empty string for JSON without type, got %q", result)
	}
}

func TestParseClaudeOutput_AssistantMessage_NoMessage(t *testing.T) {
	input := `{"type":"assistant"}`
	result := parseClaudeOutput(input)
	if result != "" {
		t.Errorf("expected empty string for assistant without message, got %q", result)
	}
}

func TestParseClaudeOutput_AssistantMessage_NoContent(t *testing.T) {
	input := `{"type":"assistant","message":{}}`
	result := parseClaudeOutput(input)
	if result != "" {
		t.Errorf("expected empty string for message without content, got %q", result)
	}
}

func TestParseClaudeOutput_ToolUse_NoName(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"tool_use"}]}}`
	result := parseClaudeOutput(input)
	if result != "" {
		t.Errorf("expected empty string for tool_use without name, got %q", result)
	}
}

func TestParseClaudeOutput_SpecialCharacters(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello \"world\" with <tags> & stuff"}]}}`
	result := parseClaudeOutput(input)
	if !strings.Contains(result, "Hello") || !strings.Contains(result, "world") {
		t.Errorf("expected special characters to be preserved, got %q", result)
	}
}

func TestParseClaudeOutput_Multiline(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Line 1\nLine 2\nLine 3"}]}}`
	result := parseClaudeOutput(input)
	if !strings.Contains(result, "Line 1") || !strings.Contains(result, "Line 2") {
		t.Errorf("expected multiline text to be preserved, got %q", result)
	}
}
