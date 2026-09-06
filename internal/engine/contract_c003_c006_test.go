package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
)

// TestConformanceC003 — C-003: a session created by one consumer (TUI harness)
// is resumable by another (GUI harness) with full history
// (frontend-contract.md §4, §3 SessionOpen/GetHistory).
func TestConformanceC003(t *testing.T) {
	h := newContractHarness(t, streamScript{
		deltas: []string{"texto "},
	})
	defer h.e.Stop()

	// TUI harness: SessionCreate (contract §3) then chat in that session.
	tuiSession, err := db.CreateSession("TUI session", "")
	if err != nil {
		t.Fatalf("C-003: create session: %v", err)
	}
	h.e.Send(SendMessage{Text: "crea historial", SessionID: tuiSession})
	waitForEvents(t, h.e, isStreamEnd, 10*time.Second)

	// A second consumer opens the same session through the contract surface:
	// ResumeSession (internal/db) + SwitchSession (engine command) and reads
	// the full history via GetHistory (contract §3).
	guiEngine := newEngine(t, &mockDriver{autoAppr: map[string]bool{}}, streamScript{
		deltas: []string{"otro"},
	})
	defer guiEngine.Stop()
	if err := db.ResumeSession(tuiSession); err != nil {
		t.Fatalf("C-003: resume session: %v", err)
	}
	guiEngine.Start()
	guiEngine.Send(SwitchSession{SessionID: tuiSession})
	time.Sleep(150 * time.Millisecond)

	history, err := db.GetHistory(tuiSession)
	if err != nil {
		t.Fatalf("C-003: get history: %v", err)
	}
	if len(history) < 2 {
		t.Fatalf("C-003: expected full history (user+assistant), got %d messages", len(history))
	}
	if history[0].Role != "user" || history[1].Role != "assistant" {
		t.Fatalf("C-003: unexpected roles in resumed history: %q, %q", history[0].Role, history[1].Role)
	}
	if s, err := db.GetActiveSession(); err != nil || s == nil || s.ID != tuiSession {
		t.Fatalf("C-003: active session should be the resumed one (got %v, err %v)", s, err)
	}
}

// TestConformanceC004 — C-004: setModel/SaveConfig refreshes the engine in
// place without reinitializing the interface (frontend-contract.md §4, §3).
// The presentation layer (TUI/GUI) is responsible for emitting the
// config:changed event after saving; the engine keeps streaming afterwards.
func TestConformanceC004(t *testing.T) {
	h := newContractHarness(t, streamScript{
		deltas: []string{"a", "b"},
	})
	defer h.e.Stop()

	before := h.e.SessionID()
	prev := config.AppConfig.Model
	config.AppConfig.Model = "claude-sonnet-4-test"
	t.Cleanup(func() { config.AppConfig.Model = prev })

	// Contract path: SaveConfig (internal/config) + RefreshClient (engine) —
	// the same engine instance keeps running, no reinit of the consumer.
	h.e.RefreshClient()

	h.send("tras refresh")
	got := waitForEvents(t, h.e, isStreamEnd, 10*time.Second)
	if !h.hasEvent(got, isStreamStart) {
		t.Fatal("C-004: engine did not stream after refresh")
	}
	if h.e.SessionID() != before {
		t.Fatal("C-004: refresh must not reset the session")
	}
}

// TestConformanceC005 — C-005: a stream:error leaves the history intact
// (frontend-contract.md §4): no partial assistant message persisted and the
// pre-error messages are untouched.
func TestConformanceC005(t *testing.T) {
	boom := &errString{}
	h := newContractHarness(t, streamScript{
		deltas: []string{"parcial"},
		err:    boom,
	})
	defer h.e.Stop()

	h.send("mensaje que falla")
	got := waitForEvents(t, h.e, isStreamError, 10*time.Second)
	if !h.hasEvent(got, isStreamError) {
		t.Fatal("C-005: expected stream:error")
	}
	er := h.lastOf(got, isStreamError).(ErrorEvent)
	if er.Message == "" {
		t.Fatal("C-005: stream:error must carry an orientative message")
	}

	history, err := db.GetHistory(h.e.SessionID())
	if err != nil {
		t.Fatalf("C-005: read history: %v", err)
	}
	for _, m := range history {
		if m.Role == "assistant" {
			t.Fatalf("C-005: no assistant message may persist after stream:error, got %q", m.Content)
		}
	}
}

// TestConformanceC006 — C-006: interfaces only consume the contract surface;
// they must not reach into the core outside it (frontend-contract.md §4).
// The GUI (cmd/wails) is forbidden from importing the TUI or the Fyne GUI,
// and the TUI (internal/tui) must not import cmd/wails either.
func TestConformanceC006(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Fatal("C-006: go.mod not found")
		}
		root = parent
	}

	walk := func(dir string, out *[]string) {
		filepath.WalkDir(filepath.Join(root, dir), func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && strings.HasSuffix(path, ".go") {
				b, _ := os.ReadFile(path)
				*out = append(*out, string(b))
			}
			return nil
		})
	}

	var guiSrc []string
	walk("cmd/wails", &guiSrc)
	for _, src := range guiSrc {
		if strings.Contains(src, `"github.com/user/go-claude-code/internal/tui"`) ||
			strings.Contains(src, `"github.com/user/go-claude-code/internal/gui"`) {
			t.Fatal("C-006: cmd/wails must not import internal/tui or internal/gui")
		}
	}

	var tuiSrc []string
	walk("internal/tui", &tuiSrc)
	for _, src := range tuiSrc {
		if strings.Contains(src, `"github.com/user/go-claude-code/cmd/wails"`) {
			t.Fatal("C-006: internal/tui must not import cmd/wails")
		}
	}
}

// TestConformanceC007 — C-007: the `slash:{clear}` command is engine-owned:
// it wipes the in-memory context AND the persisted history (messages,
// context_files, compact_history) and emits SessionCleared so presentations
// mirror the state without touching the DB (frontend-contract.md §2/§3).
func TestConformanceC007(t *testing.T) {
	h := newContractHarness(t, streamScript{
		deltas: []string{"texto "},
	})
	defer h.e.Stop()

	sid := h.e.SessionID()
	h.send("primer turno")
	if got := h.events(isStreamEnd); !h.hasEvent(got, isStreamEnd) {
		t.Fatal("C-007: first turn must complete before clearing")
	}
	if hist, _ := db.GetHistory(sid); len(hist) < 2 {
		t.Fatalf("C-007: expected persisted history before clear, got %d", len(hist))
	}

	// Contract path: the presentation sends the slash command; the engine
	// owns the state change and announces it.
	h.e.Send(SlashCommand{Name: "clear"})
	got := h.events(isEventOfType(SessionCleared{}))
	if !h.hasEvent(got, isEventOfType(SessionCleared{})) {
		t.Fatal("C-007: expected SessionCleared after slash:{clear}")
	}
	sc := h.lastOf(got, isEventOfType(SessionCleared{})).(SessionCleared)
	if sc.SessionID != sid {
		t.Fatalf("C-007: SessionCleared must carry the session id, got %s", sc.SessionID)
	}

	hist, err := db.GetHistory(sid)
	if err != nil {
		t.Fatalf("C-007: read history: %v", err)
	}
	if len(hist) != 0 {
		t.Fatalf("C-007: persisted history must be empty after clear, got %d messages", len(hist))
	}

	// The in-memory context is reset too: the next turn has no stale context.
	h.send("después de limpiar")
	got2 := h.events(isStreamEnd)
	if !h.hasEvent(got2, isStreamEnd) {
		t.Fatal("C-007: engine must keep streaming after clear")
	}
}

func isEventOfType(target Event) func(Event) bool {
	return func(ev Event) bool {
		return fmt.Sprintf("%T", ev) == fmt.Sprintf("%T", target)
	}
}
