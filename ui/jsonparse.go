package ui

import (
	"encoding/json"
	"fmt"
	"strings"
)

// parseClaudeOutput extracts meaningful text from Claude's stream-json output
func parseClaudeOutput(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// Try to parse as JSON
	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(text), &msg); err != nil {
		// Not JSON, return as-is
		return text
	}

	msgType, _ := msg["type"].(string)

	switch msgType {
	case "assistant":
		// Extract content from assistant message
		if message, ok := msg["message"].(map[string]interface{}); ok {
			if content, ok := message["content"].([]interface{}); ok {
				var texts []string
				for _, c := range content {
					if block, ok := c.(map[string]interface{}); ok {
						if blockType, _ := block["type"].(string); blockType == "text" {
							if t, ok := block["text"].(string); ok && t != "" {
								texts = append(texts, t)
							}
						} else if blockType == "tool_use" {
							if name, ok := block["name"].(string); ok {
								texts = append(texts, fmt.Sprintf("[Tool: %s]", name))
							}
						}
					}
				}
				if len(texts) > 0 {
					return strings.Join(texts, " ")
				}
			}
		}
	case "content_block_delta":
		if delta, ok := msg["delta"].(map[string]interface{}); ok {
			if t, ok := delta["text"].(string); ok && t != "" {
				return t
			}
		}
	case "error":
		if errMsg, ok := msg["error"].(map[string]interface{}); ok {
			if message, ok := errMsg["message"].(string); ok {
				return fmt.Sprintf("[Error: %s]", message)
			}
		}
		return "[Error]"
	}

	// Skip other message types (start, stop, ping, etc.)
	return ""
}
