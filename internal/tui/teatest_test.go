package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	teatest "github.com/charmbracelet/x/exp/teatest"
	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/db"
)

// TestTeaStreamingProgram (T027): a full Bubble Tea program headless with
// teatest reproduces send → streaming → end over the REAL engine (scripted
// provider stream), asserting the view renders the streamed markdown and the
// model returns to "listo" (D5, FR-003).
func TestTeaStreamingProgram(t *testing.T) {
	if db.DB != nil {
		_ = db.DB.Close()
		db.DB = nil
	}
	t.Setenv("TEST_DB_PATH", t.TempDir()+"/tea.db")
	t.Setenv("USERPROFILE", t.TempDir())

	m := NewModel()
	m.engine.SetStreamFunc(func(
		req api.Request,
		onDelta func(string),
		onToolUse func(api.ToolUse),
		onToolInput func(string, string),
		onUsage func(int, int),
	) error {
		onDelta("respuesta ")
		onDelta("**streaming**")
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

	tm := teatest.NewTestModel(t, m)

	// Pump engine events into the program, exactly like RunChat does.
	go func() {
		for ev := range m.engine.Events() {
			tm.Send(engineEventMsg{ev: ev})
		}
	}()

	// Type a message (one rune per key message, as the key handler appends
	// only single-char input) and press Enter → engine.Send(SendMessage{...}).
	for _, r := range []rune("hola teatest") {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "streaming")
	}, teatest.WithDuration(10*time.Second), teatest.WithCheckInterval(50*time.Millisecond))

	// The turn must end: assistant message persisted and model idle again.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		hist, err := db.GetHistory(m.engine.SessionID())
		if err == nil && len(hist) >= 2 && hist[len(hist)-1].Role == "assistant" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	hist, err := db.GetHistory(m.engine.SessionID())
	if err != nil {
		t.Fatalf("T027: history: %v", err)
	}
	if len(hist) < 2 || hist[len(hist)-1].Role != "assistant" {
		t.Fatalf("T027: full turn must end with a persisted assistant message, got %d messages", len(hist))
	}

	_ = tm.Quit()
}