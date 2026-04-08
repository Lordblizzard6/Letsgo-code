package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/go-claude-code/internal/api"
)

// Skill represents a reusable skill/prompt
type Skill struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Prompt      string           `json:"prompt"`
	Examples    []SkillExample   `json:"examples"`
	Parameters  []SkillParameter `json:"parameters"`
}

type SkillExample struct {
	Input   string `json:"input"`
	Output  string `json:"output"`
	Context string `json:"context"`
}

type SkillParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// SkillTool manages and invokes skills
type SkillTool struct{}

func (t *SkillTool) Definition() api.Tool {
	return api.Tool{
		Name:        "skill",
		Description: "Use a pre-defined skill or create a new one. Skills are reusable prompts that encapsulate common patterns or workflows. This tool can list available skills, use a skill, or create a new skill.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Action to perform: 'list', 'use', 'create'",
					"enum":        []string{"list", "use", "create"},
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Skill name (for use or create)",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Skill prompt content (for create)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Skill description (for create)",
				},
				"input": map[string]interface{}{
					"type":        "string",
					"description": "Input to the skill (for use)",
				},
			},
			"required": []string{"action"},
		},
	}
}

func (t *SkillTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input")
	}

	action, _ := m["action"].(string)
	if action == "" {
		return "", fmt.Errorf("action is required")
	}

	switch action {
	case "list":
		return t.listSkills()
	case "use":
		return t.useSkill(m)
	case "create":
		return t.createSkill(m)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *SkillTool) listSkills() (string, error) {
	skillsDir := getSkillsDir()

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		// No skills directory yet
		return "No skills defined yet. Use 'create' to add skills.", nil
	}

	var skills []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			name := strings.TrimSuffix(entry.Name(), ".md")
			skills = append(skills, name)
		}
	}

	if len(skills) == 0 {
		return "No skills defined yet. Use 'create' to add skills.", nil
	}

	return fmt.Sprintf("Available skills:\n%s", strings.Join(skills, "\n")), nil
}

func (t *SkillTool) useSkill(params map[string]interface{}) (string, error) {
	name, _ := params["name"].(string)
	if name == "" {
		return "", fmt.Errorf("name is required for use action")
	}

	skill, err := t.loadSkill(name)
	if err != nil {
		return "", fmt.Errorf("skill not found: %s", name)
	}

	input, _ := params["input"].(string)

	// Return the skill with the input applied
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Using skill: %s\n", skill.Name))
	result.WriteString(fmt.Sprintf("Description: %s\n\n", skill.Description))
	result.WriteString("Prompt:\n")
	result.WriteString(skill.Prompt)
	result.WriteString("\n\n")

	if input != "" {
		result.WriteString(fmt.Sprintf("Input: %s\n", input))
	}

	return result.String(), nil
}

func (t *SkillTool) createSkill(params map[string]interface{}) (string, error) {
	name, _ := params["name"].(string)
	if name == "" {
		return "", fmt.Errorf("name is required for create action")
	}

	description, _ := params["description"].(string)
	prompt, _ := params["prompt"].(string)

	if prompt == "" {
		return "", fmt.Errorf("prompt is required for create action")
	}

	skill := Skill{
		Name:        name,
		Description: description,
		Prompt:      prompt,
	}

	if err := t.saveSkill(skill); err != nil {
		return "", err
	}

	return fmt.Sprintf("Created skill: %s", name), nil
}

func (t *SkillTool) loadSkill(name string) (*Skill, error) {
	skillsDir := getSkillsDir()
	skillPath := filepath.Join(skillsDir, name+".md")

	content, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, err
	}

	// Parse markdown file
	skill := &Skill{
		Name: name,
	}

	lines := strings.Split(string(content), "\n")
	inPrompt := false
	var promptLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			// Title line - contains name
			continue
		}
		if strings.HasPrefix(line, "Description: ") {
			skill.Description = strings.TrimPrefix(line, "Description: ")
			continue
		}
		if line == "## Prompt" {
			inPrompt = true
			continue
		}
		if inPrompt {
			if strings.HasPrefix(line, "## ") {
				break
			}
			promptLines = append(promptLines, line)
		}
	}

	skill.Prompt = strings.TrimSpace(strings.Join(promptLines, "\n"))

	return skill, nil
}

func (t *SkillTool) saveSkill(skill Skill) error {
	skillsDir := getSkillsDir()
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return err
	}

	skillPath := filepath.Join(skillsDir, skill.Name+".md")

	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s\n\n", skill.Name))
	content.WriteString(fmt.Sprintf("Description: %s\n\n", skill.Description))
	content.WriteString("## Prompt\n\n")
	content.WriteString(skill.Prompt)
	content.WriteString("\n")

	return os.WriteFile(skillPath, []byte(content.String()), 0644)
}

func getSkillsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".letsGo", "skills")
}

// GetBuiltinSkills returns default skills
func GetBuiltinSkills() []Skill {
	return []Skill{
		{
			Name:        "code-review",
			Description: "Perform a thorough code review",
			Prompt: `Review the provided code for:
1. Correctness and potential bugs
2. Code style and best practices
3. Performance considerations
4. Security issues
5. Maintainability

Provide specific, actionable feedback with line references.`,
		},
		{
			Name:        "refactor",
			Description: "Refactor code for better structure",
			Prompt: `Refactor the provided code to improve:
1. Readability
2. Maintainability
3. Testability
4. Performance (if applicable)

Preserve all existing functionality while improving code quality.`,
		},
		{
			Name:        "explain",
			Description: "Explain how code works",
			Prompt: `Explain the provided code in detail:
1. What it does at a high level
2. How it works step by step
3. Key algorithms or patterns used
4. Edge cases and error handling
5. Potential improvements

Make the explanation accessible but technically accurate.`,
		},
		{
			Name:        "test",
			Description: "Generate tests for code",
			Prompt: `Generate comprehensive tests for the provided code:
1. Happy path tests
2. Edge case tests
3. Error condition tests
4. Boundary tests

Use best testing practices for the language/framework.`,
		},
	}
}
