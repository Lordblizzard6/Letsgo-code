package engine

import (
	"testing"
	"time"
)

// contractHarness is the headless conformance driver of frontend-contract.md
// §4 (C-001..C-006): a real Engine driven through the contract surface
// (Send/SwitchSession/events) with a fake permission driver and scripted
// provider streams. Every interface (TUI, GUI) must pass these tests.
type contractHarness struct {
	t *testing.T
	e *Engine
}

// newContractHarness boots an engine with an auto-approving driver and a
// scripted stream (deltas + usage), ready to receive contract commands.
func newContractHarness(t *testing.T, script streamScript) *contractHarness {
	t.Helper()
	h := &contractHarness{
		t: t,
		e: newEngine(t, &mockDriver{autoAppr: map[string]bool{}}, script),
	}
	h.e.Start()
	return h
}

// send issues the §1 `send` command.
func (h *contractHarness) send(text string) {
	h.t.Helper()
	h.e.Send(SendMessage{Text: text})
}

// cancel issues the §1 `cancel` command.
func (h *contractHarness) cancel() {
	h.t.Helper()
	h.e.Send(Cancel{})
}

// events drains the bus until pred matches, returning everything seen.
func (h *contractHarness) events(pred func(Event) bool) []Event {
	h.t.Helper()
	return waitForEvents(h.t, h.e, pred, 15*time.Second)
}

// lastOf returns the last event matching pred (or nil if none arrived).
func (h *contractHarness) lastOf(events []Event, pred func(Event) bool) Event {
	for i := len(events) - 1; i >= 0; i-- {
		if pred(events[i]) {
			return events[i]
		}
	}
	return nil
}

// hasEvent reports whether pred matched any event in the collection.
func (h *contractHarness) hasEvent(events []Event, pred func(Event) bool) bool {
	return h.lastOf(events, pred) != nil
}

// isStreamStart/Delta/End/Cancelled/Error are the contract §2 discriminators
// used by the C-001..C-006 conformance tests (and by the TUI/GUI consumers).
func isStreamStart(ev Event) bool { _, ok := ev.(StreamStart); return ok }
func isStreamDelta(ev Event) bool { _, ok := ev.(StreamDelta); return ok }
func isStreamEnd(ev Event) bool   { _, ok := ev.(StreamDone); return ok }
func isStreamCancelled(ev Event) bool {
	_, ok := ev.(StreamCancelled)
	return ok
}
func isStreamError(ev Event) bool { _, ok := ev.(ErrorEvent); return ok }
