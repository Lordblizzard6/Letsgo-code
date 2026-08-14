package engine

import (
	"testing"
	"time"

	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/db"
)

// TestConformanceC001 — C-001: `send` emits stream:start → stream:delta… →
// stream:end with a persisted message (frontend-contract.md §4).
func TestConformanceC001(t *testing.T) {
	h := newContractHarness(t, streamScript{
		deltas: []string{"hola ", "LetsGO"},
	})
	defer h.e.Stop()

	h.send("hola core")

	got := h.events(func(ev Event) bool {
		_, ok := ev.(StreamDone)
		return ok
	})

	if !h.hasEvent(got, isStreamStart) {
		t.Fatal("C-001: expected stream:start before stream:end")
	}
	deltas := 0
	for _, ev := range got {
		if d, ok := ev.(StreamDelta); ok {
			deltas += len(d.Text)
		}
	}
	if deltas == 0 {
		t.Fatal("C-001: expected stream:delta chunks before stream:end")
	}
	end := h.lastOf(got, isStreamEnd).(StreamDone)
	if end.MessageID == "" {
		t.Fatal("C-001: stream:end must carry the persisted message id")
	}

	msgs, err := db.GetHistory(h.e.SessionID())
	if err != nil {
		t.Fatalf("C-001: reading persisted history: %v", err)
	}
	last := msgs[len(msgs)-1]
	if last.Role != "assistant" {
		t.Fatalf("C-001: last persisted message should be assistant, got %q", last.Role)
	}
	if last.Content == nil {
		t.Fatal("C-001: persisted assistant message has no content")
	}
}

// TestConformanceC002 — C-002: `cancel` mid-stream does not persist partial
// text and returns to a usable state (frontend-contract.md §4).
func TestConformanceC002(t *testing.T) {
	block := make(chan struct{})
	h := newContractHarness(t, streamScript{
		deltas:      []string{"parcial", "incompleto"},
		blockOnCh:   block,
		toolUsesOnce: true,
	})
	defer h.e.Stop()

	h.send("mensaje largo")

	// Wait for the stream to be alive (first delta), then cancel while blocked.
	waitForEvents(t, h.e, isStreamDelta, 3*time.Second)
	h.cancel()
	close(block)

	// Stream is cancelled mid-turn; the turn ends with Idle.
	got := drainUntil(t, h.e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 5*time.Second)

	if !h.hasEvent(got, isStreamCancelled) {
		t.Fatal("C-002: expected stream:cancelled after cancel")
	}

	msgs, err := db.GetHistory(h.e.SessionID())
	if err != nil {
		t.Fatalf("C-002: reading persisted history: %v", err)
	}
	for _, m := range msgs {
		if m.Role == "assistant" && containsPartial(m, "parcial") {
			t.Fatal("C-002: partial assistant text must not be persisted")
		}
	}

	h.send("siguiente")
	// The blockOnCh cancel fired once; give the follow-up turn a plain script
	// (the shared emulator would otherwise return context.Canceled again).
	h.e.streamFunc = func(req api.Request,
		onDelta func(string), onToolUse func(api.ToolUse),
		onToolInput func(string, string), onUsage func(int, int)) error {
		onDelta("respuesta completa")
		return nil
	}
	again := drainUntil(t, h.e, isStreamEnd, 10*time.Second)
	if !h.hasEvent(again, isStreamEnd) {
		t.Fatal("C-002: engine must be usable after cancel (send → stream:end)")
	}
}

func containsPartial(m db.Message, fragment string) bool {
	switch v := m.Content.(type) {
	case string:
		return len(v) >= len(fragment) && v[:len(fragment)] == fragment
	default:
		return false
	}
}
