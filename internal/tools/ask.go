package tools

import (
	"fmt"
	"github.com/user/go-claude-code/internal/api"
)

type AskUserTool struct{}

func (t *AskUserTool) Definition() api.Tool {
	return api.Tool{
		Name:        "ask_user",
		Description: "Ask the user a question to clarify requirements or get permission.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"question": map[string]interface{}{
					"type":        "string",
					"description": "The question to ask the user",
				},
			},
			"required": []string{"question"},
		},
	}
}

func (t *AskUserTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok { return "", fmt.Errorf("invalid input") }
	question, _ := m["question"].(string)

	// Since we are in a TUI, we return the question as a tool result
	// The TUI will display this to the user.
	return fmt.Sprintf("USER QUESTION: %s", question), nil
}
