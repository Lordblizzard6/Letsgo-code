package tools

import (
	"fmt"
	"sync"
	"time"

	"github.com/user/go-claude-code/internal/api"
)

// PlanMode represents the planning mode state
type PlanMode struct {
	IsActive    bool       `json:"is_active"`
	Plan        string     `json:"plan"`
	Steps       []PlanStep `json:"steps"`
	CurrentStep int        `json:"current_step"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PlanStep struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"` // "pending", "in_progress", "completed", "skipped"
	Result      string `json:"result"`
}

var (
	planModeInstance *PlanMode
	planModeOnce     sync.Once
)

func getPlanMode() *PlanMode {
	planModeOnce.Do(func() {
		planModeInstance = &PlanMode{
			IsActive:    false,
			Steps:       []PlanStep{},
			CurrentStep: 0,
		}
	})
	return planModeInstance
}

// EnterPlanModeTool enters plan mode
type EnterPlanModeTool struct{}

func (t *EnterPlanModeTool) Definition() api.Tool {
	return api.Tool{
		Name:        "enter_plan_mode",
		Description: "Enter plan mode. In plan mode, Claude will create a structured plan before executing any changes. Use this for complex tasks where you want to review the approach before execution. While in plan mode, Claude will not execute tools until you approve the plan.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"goal": map[string]interface{}{
					"type":        "string",
					"description": "The high-level goal or objective",
				},
			},
			"required": []string{"goal"},
		},
	}
}

func (t *EnterPlanModeTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input")
	}

	goal, _ := m["goal"].(string)

	plan := getPlanMode()
	plan.IsActive = true
	plan.Plan = goal
	plan.Steps = []PlanStep{}
	plan.CurrentStep = 0
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()

	return fmt.Sprintf("Entered plan mode. Goal: %s\n\nClaude will now create a structured plan before making any changes. Review the plan and approve before execution.", goal), nil
}

// ExitPlanModeTool exits plan mode
type ExitPlanModeTool struct{}

func (t *ExitPlanModeTool) Definition() api.Tool {
	return api.Tool{
		Name:        "exit_plan_mode",
		Description: "Exit plan mode and return to normal operation. Any pending steps will be discarded.",
		InputSchema: map[string]interface{}{
			"type":     "object",
			"properties": map[string]interface{}{},
		},
	}
}

func (t *ExitPlanModeTool) Execute(input interface{}) (string, error) {
	plan := getPlanMode()

	if !plan.IsActive {
		return "Not in plan mode.", nil
	}

	completed := 0
	for _, step := range plan.Steps {
		if step.Status == "completed" {
			completed++
		}
	}

	plan.IsActive = false
	plan.Steps = []PlanStep{}
	plan.CurrentStep = 0

	return fmt.Sprintf("Exited plan mode. Completed %d/%d steps.", completed, len(plan.Steps)), nil
}

// PlanStepAddTool adds a step to the plan
type PlanStepAddTool struct{}

func (t *PlanStepAddTool) Definition() api.Tool {
	return api.Tool{
		Name:        "plan_step_add",
		Description: "Add a step to the current plan. Only valid when in plan mode.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the step",
				},
			},
			"required": []string{"description"},
		},
	}
}

func (t *PlanStepAddTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input")
	}

	plan := getPlanMode()
	if !plan.IsActive {
		return "", fmt.Errorf("not in plan mode")
	}

	description, _ := m["description"].(string)
	if description == "" {
		return "", fmt.Errorf("description is required")
	}

	step := PlanStep{
		ID:          fmt.Sprintf("step_%d", len(plan.Steps)),
		Description: description,
		Status:      "pending",
	}

	plan.Steps = append(plan.Steps, step)
	plan.UpdatedAt = time.Now()

	return fmt.Sprintf("Added step %d: %s", len(plan.Steps), description), nil
}

// PlanShowTool displays the current plan
type PlanShowTool struct{}

func (t *PlanShowTool) Definition() api.Tool {
	return api.Tool{
		Name:        "plan_show",
		Description: "Display the current plan with all steps and their status.",
		InputSchema: map[string]interface{}{
			"type":     "object",
			"properties": map[string]interface{}{},
		},
	}
}

func (t *PlanShowTool) Execute(input interface{}) (string, error) {
	plan := getPlanMode()

	if !plan.IsActive {
		return "Not in plan mode.", nil
	}

	var result string
	result = fmt.Sprintf("📋 PLAN: %s\n\n", plan.Plan)

	if len(plan.Steps) == 0 {
		result += "No steps defined yet.\n"
	} else {
		result += "Steps:\n"
		for i, step := range plan.Steps {
			status := "[ ]"
			switch step.Status {
			case "in_progress":
				status = "[~]"
			case "completed":
				status = "[x]"
			case "skipped":
				status = "[-]"
			}
			result += fmt.Sprintf("  %s %d. %s\n", status, i+1, step.Description)
		}
	}

	result += fmt.Sprintf("\nProgress: %d/%d steps completed\n", countCompleted(plan.Steps), len(plan.Steps))

	return result, nil
}

// PlanApproveTool approves and starts executing the plan
type PlanApproveTool struct{}

func (t *PlanApproveTool) Definition() api.Tool {
	return api.Tool{
		Name:        "plan_approve",
		Description: "Approve the plan and begin execution. After approval, Claude will execute the plan steps.",
		InputSchema: map[string]interface{}{
			"type":     "object",
			"properties": map[string]interface{}{},
		},
	}
}

func (t *PlanApproveTool) Execute(input interface{}) (string, error) {
	plan := getPlanMode()

	if !plan.IsActive {
		return "", fmt.Errorf("not in plan mode")
	}

	if len(plan.Steps) == 0 {
		return "", fmt.Errorf("no steps in plan")
	}

	// Mark first step as in_progress
	if len(plan.Steps) > 0 {
		plan.Steps[0].Status = "in_progress"
	}

	return fmt.Sprintf("Plan approved. Starting execution of %d steps...", len(plan.Steps)), nil
}

// PlanStepCompleteTool marks a step as completed
type PlanStepCompleteTool struct{}

func (t *PlanStepCompleteTool) Definition() api.Tool {
	return api.Tool{
		Name:        "plan_step_complete",
		Description: "Mark the current plan step as completed and advance to the next step.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"result": map[string]interface{}{
					"type":        "string",
					"description": "Result or outcome of the step",
				},
			},
		},
	}
}

func (t *PlanStepCompleteTool) Execute(input interface{}) (string, error) {
	m, _ := input.(map[string]interface{})

	plan := getPlanMode()
	if !plan.IsActive {
		return "", fmt.Errorf("not in plan mode")
	}

	if plan.CurrentStep >= len(plan.Steps) {
		return "", fmt.Errorf("no more steps")
	}

	// Mark current step as completed
	plan.Steps[plan.CurrentStep].Status = "completed"
	if result, ok := m["result"].(string); ok {
		plan.Steps[plan.CurrentStep].Result = result
	}

	// Advance to next step
	plan.CurrentStep++
	if plan.CurrentStep < len(plan.Steps) {
		plan.Steps[plan.CurrentStep].Status = "in_progress"
	}

	plan.UpdatedAt = time.Now()

	if plan.CurrentStep >= len(plan.Steps) {
		return "All plan steps completed! Exiting plan mode.", nil
	}

	return fmt.Sprintf("Step %d completed. Now working on step %d: %s",
		plan.CurrentStep, plan.CurrentStep+1, plan.Steps[plan.CurrentStep].Description), nil
}

func countCompleted(steps []PlanStep) int {
	count := 0
	for _, step := range steps {
		if step.Status == "completed" {
			count++
		}
	}
	return count
}

// IsPlanModeActive returns whether plan mode is currently active
func IsPlanModeActive() bool {
	return getPlanMode().IsActive
}
