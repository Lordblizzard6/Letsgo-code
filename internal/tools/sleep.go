package tools

import (
	"fmt"
	"strconv"
	"time"

	"github.com/user/go-claude-code/internal/api"
)

// SleepTool pauses execution for a specified duration
type SleepTool struct{}

func (t *SleepTool) Name() string {
	return "sleep"
}

func (t *SleepTool) Description() string {
	return "Pause execution for a specified duration. Useful for adding delays between operations, rate limiting, or waiting for resources to become available."
}

func (t *SleepTool) Definition() api.Tool {
	return api.Tool{
		Name:        "sleep",
		Description: t.Description(),
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"duration": map[string]interface{}{
					"type":        "number",
					"description": "Duration to sleep in seconds",
					"minimum":     0,
					"maximum":     300,
				},
				"reason": map[string]interface{}{
					"type":        "string",
					"description": "Optional reason for the sleep (for logging)",
				},
			},
			"required": []string{"duration"},
		},
	}
}

func (t *SleepTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	// Get duration
	var duration float64
	switch v := m["duration"].(type) {
	case float64:
		duration = v
	case int:
		duration = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return "", fmt.Errorf("invalid duration: %v", err)
		}
		duration = parsed
	default:
		return "", fmt.Errorf("duration must be a number")
	}

	// Validate range
	if duration < 0 || duration > 300 {
		return "", fmt.Errorf("duration must be between 0 and 300 seconds")
	}

	// Get optional reason
	reason, _ := m["reason"].(string)

	// Sleep
	time.Sleep(time.Duration(duration * float64(time.Second)))

	if reason != "" {
		return fmt.Sprintf("Slept for %.1f seconds (%s)", duration, reason), nil
	}
	return fmt.Sprintf("Slept for %.1f seconds", duration), nil
}
