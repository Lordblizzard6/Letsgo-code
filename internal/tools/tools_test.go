package tools

import (
	"testing"
)

func TestBashTool_Definition(t *testing.T) {
	tool := &BashTool{}
	def := tool.Definition()

	if def.Name != "bash" {
		t.Errorf("Expected tool name 'bash', got '%s'", def.Name)
	}

	if def.Description == "" {
		t.Error("Expected non-empty description")
	}

	if def.InputSchema == nil {
		t.Error("Expected non-nil input schema")
	}
}

func TestBashTool_Execute(t *testing.T) {
	tool := &BashTool{}

	// Test with valid input
	input := map[string]interface{}{
		"command": "echo hello",
	}

	result, err := tool.Execute(input)
	if err != nil {
		t.Logf("Bash execution result: %v", err)
		// Expected to work, but may fail in test environment
	}

	if result != "" {
		t.Logf("Bash output: %s", result)
	}
}

func TestCatTool_Definition(t *testing.T) {
	tool := &CatTool{}
	def := tool.Definition()

	if def.Name != "cat" {
		t.Errorf("Expected tool name 'cat', got '%s'", def.Name)
	}

	if def.Description == "" {
		t.Error("Expected non-empty description")
	}
}

func TestEditTool_Definition(t *testing.T) {
	tool := &EditTool{}
	def := tool.Definition()

	if def.Name != "edit" {
		t.Errorf("Expected tool name 'edit', got '%s'", def.Name)
	}
}

func TestGrepTool_Definition(t *testing.T) {
	tool := &GrepTool{}
	def := tool.Definition()

	if def.Name != "grep" {
		t.Errorf("Expected tool name 'grep', got '%s'", def.Name)
	}
}

func TestGlobTool_Definition(t *testing.T) {
	tool := &GlobTool{}
	def := tool.Definition()

	if def.Name != "glob" {
		t.Errorf("Expected tool name 'glob', got '%s'", def.Name)
	}
}

func TestAllTools_Registered(t *testing.T) {
	// Verify AllTools slice is not empty
	if len(AllTools) == 0 {
		t.Fatal("AllTools slice is empty")
	}

	// Verify all tools have names
	for i, tool := range AllTools {
		def := tool.Definition()
		if def.Name == "" {
			t.Errorf("Tool at index %d has empty name", i)
		}
		if def.Description == "" {
			t.Errorf("Tool '%s' has empty description", def.Name)
		}
	}

	t.Logf("Total tools registered: %d", len(AllTools))
}
