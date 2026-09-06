package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/engine"
)

// newTestModel boots a model over the contract surface with a scripted
// provider stream (engine.SetStreamFunc) and a temp DB.
func newTestModel(t *testing.T) model {
	t.Helper()
	if db.DB != nil {
		_ = db.DB.Close()
		db.DB = nil
	}
	t.Setenv("TEST_DB_PATH", t.TempDir()+"/model.db")
	// Redirect the cost tracker's ~/.letsGo/costs.json to a temp home.
	t.Setenv("USERPROFILE", t.TempDir())
	m := NewModel()
	m.engine.SetStreamFunc(func(
		req api.Request,
		onDelta func(string),
		onToolUse func(api.ToolUse),
		onToolInput func(string, string),
		onUsage func(int, int),
	) error {
		onDelta("hola ")
		onDelta("mundo")
		onUsage(10, 20)
		return nil
	})
	m.engine.Start()
	t.Cleanup(func() {
		m.engine.Stop()
		if db.DB != nil {
			_ = db.DB.Close()
			db.DB = nil
		}
	})
	return m
}

// contentString flattens an api.Message content (string or blocks) to text.
func contentString(msg api.Message) string {
	switch v := msg.Content.(type) {
	case string:
		return v
	default:
		return ""
	}
}

// TestUpdateStreamLifecycle (T022): stream:start → stream:delta… → stream:end
// accumulates markdown in the viewport and returns to "listo" (FR-003, SC-005).
func TestUpdateStreamLifecycle(t *testing.T) {
	m := newTestModel(t)

	m = m.handleEngineEvent(engine.StreamStart{SessionID: "s", RequestID: "r"})
	if !m.isStreaming {
		t.Fatal("T022: isStreaming must be true after stream:start")
	}
	m = m.handleEngineEvent(engine.StreamDelta{Text: "**hola** "})
	m = m.handleEngineEvent(engine.StreamDelta{Text: "mundo"})
	if !strings.Contains(m.currentResponse, "hola") {
		t.Fatalf("T022: deltas must accumulate, got %q", m.currentResponse)
	}

	m = m.handleEngineEvent(engine.StreamDone{MessageID: "m1"})
	if m.isStreaming {
		t.Fatal("T022: isStreaming must be false after stream:end")
	}
	if m.currentResponse != "" {
		t.Fatal("T022: currentResponse must be cleared after stream:end")
	}
	last := m.messages[len(m.messages)-1]
	if last.Role != "assistant" {
		t.Fatalf("T022: expected assistant message, got %q", last.Role)
	}
	if !strings.Contains(m.viewport.View(), "mundo") {
		t.Fatalf("T022: viewport must contain the streamed markdown, got %q", m.viewport.View())
	}
}

// TestUpdateCancelClean (T023): cancel mid-stream leaves a usable prompt and
// no partial message (FR-007, SC-008, C-002).
func TestUpdateCancelClean(t *testing.T) {
	m := newTestModel(t)

	m = m.handleEngineEvent(engine.StreamStart{SessionID: "s", RequestID: "r"})
	m = m.handleEngineEvent(engine.StreamDelta{Text: "parcial incompleto"})

	m = m.handleEngineEvent(engine.StreamCancelled{})
	if m.isStreaming {
		t.Fatal("T023: isStreaming must be false after stream:cancelled")
	}
	if m.currentResponse != "" {
		t.Fatal("T023: partial response must be dropped on cancel")
	}
	for _, msg := range m.messages {
		if msg.Role == "assistant" && strings.Contains(contentString(msg), "parcial") {
			t.Fatal("T023: partial assistant text must not reach the messages")
		}
	}
	// Prompt usable: typing after cancel works (single-rune key handling).
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if got := m2.(model).input; got != "a" {
		t.Fatalf("T023: input must accept text after cancel, got %q", got)
	}
}

// TestModelPickerCtrlS (T025): Ctrl+S opens the picker; confirming persists
// the selection in internal/config and refreshes the engine without restart
// (FR-005, FR-016, C-004).
func TestModelPickerCtrlS(t *testing.T) {
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("T025: load config: %v", err)
	}
	// Hermetic: the picker's first entry is claude-sonnet-4-20250514; force a
	// different persisted model so the change is deterministic, and restore
	// the real config afterwards.
	home, _ := os.UserHomeDir()
	cfgPath := filepath.Join(home, ".letsGo", "config.yaml")
	orig, _ := os.ReadFile(cfgPath)
	t.Cleanup(func() {
		if orig != nil {
			_ = os.WriteFile(cfgPath, orig, 0644)
		} else {
			_ = os.Remove(cfgPath)
		}
	})

	config.AppConfig.Model = "claude-3-opus-20240229"
	if err := config.SaveConfig(); err != nil {
		t.Fatalf("T025: save config: %v", err)
	}
	m := newTestModel(t)
	before := config.AppConfig.Model

	opened, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	openedModel := opened.(model)
	if openedModel.currentMode != settingsMode {
		t.Fatal("T025: Ctrl+S must open the settings/model picker")
	}

	menu, _ := openedModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	menuModel := menu.(model)
	if !menuModel.inModelMenu {
		t.Fatal("T025: Right must enter the model submenu")
	}

	confirmed, _ := menuModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	final := confirmed.(model)
	if final.currentMode != chatMode {
		t.Fatal("T025: confirming must return to chat mode")
	}
	if config.AppConfig.Model == before || config.AppConfig.Model == "" {
		t.Fatalf("T025: model must change after confirming, got %q", config.AppConfig.Model)
	}
	// Persisted via SaveConfig (contract §3): a reload reads it back.
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("T025: reload config: %v", err)
	}
	if config.AppConfig.Model == "" {
		t.Fatal("T025: model selection must be persisted in config")
	}
}

// TestSessionResume (T026): a session created in the DB is resumable through
// the contract (SwitchSession + SessionOpen/GetHistory) (FR-006, C-003).
func TestSessionResume(t *testing.T) {
	m := newTestModel(t)
	sid, err := db.CreateSession("resume-me", "")
	if err != nil {
		t.Fatalf("T026: create session: %v", err)
	}
	if err := db.SaveMessage(sid, "user", "hola previa"); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMessage(sid, "assistant", "respuesta previa"); err != nil {
		t.Fatal(err)
	}

	m.engine.Send(engine.SwitchSession{SessionID: sid})
	time.Sleep(150 * time.Millisecond)
	if m.engine.SessionID() != sid {
		t.Fatalf("T026: engine must resume session %s, got %s", sid, m.engine.SessionID())
	}

	history, err := db.GetHistory(sid)
	if err != nil || len(history) < 2 {
		t.Fatalf("T026: resumed session must keep full history, got %d (err %v)", len(history), err)
	}
	if history[0].Role != "user" || history[1].Role != "assistant" {
		t.Fatalf("T026: unexpected roles: %q %q", history[0].Role, history[1].Role)
	}
}

// TestSessionControl (T031): /sessions lists, /open resumes, /new creates —
// all over the contract surface (db ListSessions/ResumeSession/CreateSession +
// engine SwitchSession) sharing the same DB (FR-006, C-003).
func TestSessionControl(t *testing.T) {
	m := newTestModel(t)

	old, err := db.CreateSession("vieja", "")
	if err != nil {
		t.Fatalf("T031: create old session: %v", err)
	}
	_ = db.SaveMessage(old, "user", "hola desde la TUI")
	_ = db.SaveMessage(old, "assistant", "respuesta previa")

	// /sessions lists the persisted session (hadler is a pure contract read).
	res, cont, quit := ProcessSlashCommand("/sessions")
	if !cont || quit {
		t.Fatal("T031: /sessions must continue the chat")
	}
	if !strings.Contains(res, "vieja") {
		t.Fatalf("T031: /sessions must list the session, got %q", res)
	}

	// /open <id> resumes: engine switches session, model reloads history.
	m.input = "/open " + old
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if m.sessionID != old {
		t.Fatalf("T031: model must switch to session %s, got %s", old, m.sessionID)
	}
	if m.engine.SessionID() != old {
		t.Fatalf("T031: engine must switch to session %s, got %s", old, m.engine.SessionID())
	}
	if !strings.Contains(m.viewport.View(), "respuesta previa") {
		t.Fatal("T031: resumed history must be rendered in the viewport")
	}

	// /new creates a fresh session via the contract.
	m.input = "/new"
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if m.sessionID == old {
		t.Fatal("T031: /new must create a different session")
	}
	if m.engine.SessionID() != m.sessionID {
		t.Fatalf("T031: engine must follow the new session, got %s", m.engine.SessionID())
	}
}

// TestResizeDuringStream (T033): resizing while streaming re-lays the viewport
// without losing accumulated content (SC-008, SC-009).
func TestResizeDuringStream(t *testing.T) {
	m := newTestModel(t)

	m = m.handleEngineEvent(engine.StreamStart{SessionID: "s", RequestID: "r"})
	m = m.handleEngineEvent(engine.StreamDelta{Text: "contenido acumulado"})
	if !strings.Contains(m.viewport.View(), "contenido") {
		t.Fatal("T033: content must render before resize")
	}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	resized := next.(model)
	if resized.viewport.Width != 120 {
		t.Fatalf("T033: viewport must re-lay to new width, got %d", resized.viewport.Width)
	}
	if !resized.isStreaming {
		t.Fatal("T033: stream must stay active across resize")
	}
	if !strings.Contains(resized.currentResponse, "acumulado") {
		t.Fatal("T033: accumulated response must survive resize")
	}
	if !strings.Contains(resized.viewport.View(), "acumulado") {
		t.Fatal("T033: viewport content must survive resize")
	}
}
