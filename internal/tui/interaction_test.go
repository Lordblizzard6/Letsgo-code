package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIInteractiveApprovalFlow(t *testing.T) {
	m := newTestModel(t)

	// Simulate receiving a tool approval request
	respCh := make(chan bool, 1)
	req := pendingApproval{
		toolName: "bash",
		input:    map[string]interface{}{"command": "go build ."},
		inputStr: "go build .",
		respCh:   respCh,
	}

	updated, _ := m.Update(tuiApprovalMsg{req: req})
	m2 := updated.(model)

	if m2.pendingApproval == nil {
		t.Fatal("expected pendingApproval to be set")
	}

	// Press 'y' to allow
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m3 := updated2.(model)

	if m3.pendingApproval != nil {
		t.Fatal("expected pendingApproval to be cleared after 'y'")
	}

	select {
	case allowed := <-respCh:
		if !allowed {
			t.Fatal("expected allow == true after 'y'")
		}
	default:
		t.Fatal("expected approval response sent to channel")
	}
}

func TestTUITextareaUTF8Support(t *testing.T) {
	m := newTestModel(t)

	// Spanish input with accents, ñ and symbols
	spanishText := "¿Cómo estás? Edición en español: año y éxito"
	for _, r := range []rune(spanishText) {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(model)
	}

	if m.input != spanishText {
		t.Fatalf("expected UTF-8 input %q, got %q", spanishText, m.input)
	}
	if m.textarea.Value() != spanishText {
		t.Fatalf("expected textarea value %q, got %q", spanishText, m.textarea.Value())
	}
}

func TestTUIModeToggle(t *testing.T) {
	m := newTestModel(t)

	if m.agentMode != "build" {
		t.Fatalf("default agentMode must be build, got %s", m.agentMode)
	}

	// Press Tab with empty textarea to toggle to plan mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(model)

	if m.agentMode != "plan" {
		t.Fatalf("expected agentMode plan after Tab, got %s", m.agentMode)
	}

	// Press Tab again to toggle back to build mode
	updated2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated2.(model)

	if m.agentMode != "build" {
		t.Fatalf("expected agentMode build after second Tab, got %s", m.agentMode)
	}
}

func TestTUIOverlaySwitching(t *testing.T) {
	m := newTestModel(t)

	// Ctrl+D toggles git review mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(model)
	if m.currentMode != gitMode {
		t.Fatalf("expected currentMode == gitMode after Ctrl+D, got %v", m.currentMode)
	}

	// Esc returns to chatMode
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.currentMode != chatMode {
		t.Fatalf("expected currentMode == chatMode after Esc, got %v", m.currentMode)
	}

	// Ctrl+B toggles activityMode
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	m = updated.(model)
	if m.currentMode != activityMode {
		t.Fatalf("expected currentMode == activityMode after Ctrl+B, got %v", m.currentMode)
	}

	// Ctrl+B again returns to chatMode
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	m = updated.(model)
	if m.currentMode != chatMode {
		t.Fatalf("expected currentMode == chatMode after Ctrl+B, got %v", m.currentMode)
	}
}

func TestTUIAutocompletePopup(t *testing.T) {
	m := newTestModel(t)

	// Type "/"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(model)

	if !m.autocomplete.Active {
		t.Fatal("expected autocomplete.Active == true after typing '/'")
	}
	if len(m.autocomplete.Items) == 0 {
		t.Fatal("expected autocomplete items to be populated")
	}

	// Press Tab to complete first suggestion
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(model)

	if m.autocomplete.Active {
		t.Fatal("expected autocomplete to be dismissed after Tab completion")
	}
	if m.textarea.Value() == "/" {
		t.Fatal("expected textarea to contain completed command, got '/'")
	}
}

func TestTUIGitReviewView(t *testing.T) {
	summary := TUIGitSummary{
		Branch:           "main",
		Clean:            false,
		UncommittedCount: 1,
		Files: []TUIGitFile{
			{
				Path:      "main.go",
				Name:      "main.go",
				Dir:       ".",
				Status:    "M",
				Staged:    false,
				Additions: 5,
				Deletions: 2,
			},
		},
		SelectedFile: "main.go",
		DiffContent:  "+package main\n-package old",
	}

	view := RenderGitReview(summary, 0, false, "", 80, 24)
	if !strings.Contains(view, "GIT REVIEW") {
		t.Fatalf("expected view to contain GIT REVIEW header, got %s", view)
	}
	if !strings.Contains(view, "main.go") {
		t.Fatalf("expected view to contain file main.go, got %s", view)
	}
}
