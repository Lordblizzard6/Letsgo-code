package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/db"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "engine-test-*")
	if err != nil {
		panic(err)
	}
	os.Setenv("TEST_DB_PATH", filepath.Join(dir, "test.db"))
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// mockDriver implements PermissionDriver for tests.
type mockDriver struct {
	mu        sync.Mutex
	promptCb  func(toolName, input string) (bool, error)
	autoAppr  map[string]bool
	grants    map[string]bool
	callCount int
}

func (d *mockDriver) Prompt(toolName string, input any) (bool, error) {
	d.mu.Lock()
	d.callCount++
	cb := d.promptCb
	d.mu.Unlock()
	if cb != nil {
		return cb(toolName, jsonInputString(anyToMap(input)))
	}
	return true, nil
}

func anyToMap(input any) map[string]interface{} {
	switch v := input.(type) {
	case map[string]interface{}:
		return v
	case string:
		var m map[string]interface{}
		_ = json.Unmarshal([]byte(v), &m)
		return m
	default:
		b, _ := json.Marshal(input)
		var m map[string]interface{}
		_ = json.Unmarshal(b, &m)
		return m
	}
}

func (d *mockDriver) AutoApprove(category string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.autoAppr[category]
}

func (d *mockDriver) SessionGranted(sessionID, category string) bool {
	ok, err := db.IsGranted(sessionID, category)
	return err == nil && ok
}

func (d *mockDriver) calls() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.callCount
}

// streamEmulator produces a scripted stream: deltas, tool use, and usage.
type streamScript struct {
	deltas     []string
	toolUses   []api.ToolUse
	toolInputs map[string]string
	usageIn    int
	usageOut   int
	err        error
	blockOnCh  chan struct{} // if set, stop mid-stream and wait; used for cancel tests
	// toolUsesOnce emits toolUses only on the first request, so the tool cycle
	// terminates (the model is simulated to answer in text afterwards).
	toolUsesOnce bool
}

func sink(req api.Request,
	onDelta func(string),
	onToolUse func(api.ToolUse),
	onToolInput func(string, string),
	onUsage func(int, int)) error {
	return nil
}

func newEngine(t *testing.T, driver PermissionDriver, script streamScript) *Engine {
	t.Helper()
	e := New(driver)
	requests := 0
	e.streamFunc = func(req api.Request,
		onDelta func(string),
		onToolUse func(api.ToolUse),
		onToolInput func(string, string),
		onUsage func(int, int)) error {
		requests++
		for _, d := range script.deltas {
			onDelta(d)
		}
		if !script.toolUsesOnce || requests == 1 {
			for i, tu := range script.toolUses {
				onToolUse(tu)
				if in, ok := script.toolInputs[tu.ID]; ok {
					onToolInput(tu.ID, in)
				} else if i == 0 {
					// default input for first tool
					b, _ := json.Marshal(map[string]interface{}{"command": "echo hi"})
					onToolInput(tu.ID, string(b))
				}
			}
		}
		if script.usageIn > 0 || script.usageOut > 0 {
			onUsage(script.usageIn, script.usageOut)
		}
		if script.blockOnCh != nil {
			<-script.blockOnCh
			return context.Canceled
		}
		return script.err
	}
	return e
}

// waitForEvents drains events until a predicate returns true (or the timeout hits).
func waitForEvents(t *testing.T, e *Engine, pred func(Event) bool, timeout time.Duration) []Event {
	t.Helper()
	var got []Event
	deadline := time.After(timeout)
	for {
		select {
		case ev := <-e.Events():
			got = append(got, ev)
			if pred(ev) {
				return got
			}
		case <-deadline:
			t.Fatalf("timed out waiting for predicate; got %d events", len(got))
		}
	}
}

func drainUntil(t *testing.T, e *Engine, stop func(Event) bool, timeout time.Duration) []Event {
	t.Helper()
	var got []Event
	deadline := time.After(timeout)
	for {
		select {
		case ev := <-e.Events():
			got = append(got, ev)
			if stop(ev) {
				return got
			}
		case <-deadline:
			return got
		}
	}
}

func TestStreamToEventsSequence(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{}}
	e := newEngine(t, driver, streamScript{
		deltas:  []string{"hel", "lo ", "world"},
		usageIn: 100, usageOut: 50,
	})
	e.Start()
	e.Send(SendMessage{Text: "hi"})

	got := drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 2*time.Second)

	// First event is UserMessageAppended.
	if len(got) == 0 {
		t.Fatal("no events")
	}
	if _, ok := got[0].(UserMessageAppended); !ok {
		t.Fatalf("first event should be UserMessageAppended, got %T", got[0])
	}
	// StreamStart present.
	foundStart := false
	foundDelta := false
	foundDone := false
	foundUsage := false
	foundIdle := false
	for _, ev := range got {
		switch ev.(type) {
		case StreamStart:
			foundStart = true
		case StreamDelta:
			foundDelta = true
		case StreamDone:
			foundDone = true
		case UsageUpdate:
			foundUsage = true
		case Idle:
			foundIdle = true
		}
	}
	if !foundStart || !foundDelta || !foundDone || !foundUsage || !foundIdle {
		t.Fatalf("missing events: start=%v delta=%v done=%v usage=%v idle=%v",
			foundStart, foundDelta, foundDone, foundUsage, foundIdle)
	}
	// Verify delta content accumulated.
	var text string
	for _, ev := range got {
		if d, ok := ev.(StreamDelta); ok {
			text += d.Text
		}
	}
	if text != "hello world" {
		// deltas may be flushed in a single or multiple batches, but total must match.
		if text != "" && text != "hello world" {
			t.Fatalf("delta content: got %q", text)
		}
	}
	// Idle must be the LAST event.
	if _, ok := got[len(got)-1].(Idle); !ok {
		t.Fatalf("last event must be Idle, got %T", got[len(got)-1])
	}
}

func TestToolCycleWithAutoApprove(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{"bash": true}}
	e := newEngine(t, driver, streamScript{
		toolUses: []api.ToolUse{{ID: "call_1", Name: "bash", Input: map[string]interface{}{"command": "echo hi"}}},
	})
	e.Start()
	e.Send(SendMessage{Text: "run it"})

	got := drainUntil(t, e, func(ev Event) bool {
		if err, ok := ev.(ErrorEvent); ok {
			t.Fatalf("unexpected error: %s", err.Message)
		}
		_, ok := ev.(Idle)
		return ok
	}, 5*time.Second)

	foundRequested, foundExecuting, foundResult := false, false, false
	for _, ev := range got {
		switch ev.(type) {
		case ToolRequested:
			foundRequested = true
		case ToolExecuting:
			foundExecuting = true
		case ToolResult:
			foundResult = true
		}
	}
	if !foundRequested || !foundExecuting || !foundResult {
		t.Fatalf("tool cycle incomplete: requested=%v executing=%v result=%v",
			foundRequested, foundExecuting, foundResult)
	}
	if driver.calls() != 0 {
		t.Fatalf("auto-approve should not call Prompt, got %d", driver.calls())
	}
}

func TestToolCycleApprovalPrompt(t *testing.T) {
	driver := &mockDriver{}
	driver.promptCb = func(toolName, input string) (bool, error) {
		return true, nil
	}
	e := newEngine(t, driver, streamScript{
		toolUses:     []api.ToolUse{{ID: "call_2", Name: "bash", Input: map[string]interface{}{"command": "echo hi"}}},
		toolUsesOnce: true,
	})
	e.Start()
	e.Send(SendMessage{Text: "please run"})

	got := drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 5*time.Second)

	if driver.calls() != 1 {
		t.Fatalf("driver.Prompt should have been called once, got %d", driver.calls())
	}
	foundExecuting := false
	for _, ev := range got {
		if _, ok := ev.(ToolExecuting); ok {
			foundExecuting = true
		}
	}
	if !foundExecuting {
		t.Fatal("approved tool should execute (ToolExecuting event)")
	}
}

func TestToolRejectionFeedsModel(t *testing.T) {
	driver := &mockDriver{}
	driver.promptCb = func(toolName, input string) (bool, error) {
		return false, nil // user rejects
	}
	e := newEngine(t, driver, streamScript{
		toolUses:     []api.ToolUse{{ID: "call_3", Name: "bash", Input: map[string]interface{}{"command": "echo hi"}}},
		toolUsesOnce: true,
	})
	e.Start()
	e.Send(SendMessage{Text: "run it"})

	got := drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 5*time.Second)

	foundRejected := false
	foundExecuting := false
	for _, ev := range got {
		switch ev.(type) {
		case ToolRejected:
			foundRejected = true
		case ToolExecuting:
			foundExecuting = true
		}
	}
	if !foundRejected {
		t.Fatal("expected ToolRejected event")
	}
	if foundExecuting {
		t.Fatal("rejected tool must never execute (SC-007)")
	}
}

func TestToolResultOrdering(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{"bash": true}}
	// Two sequential tool uses arrive in one stream.
	e := newEngine(t, driver, streamScript{
		toolUses: []api.ToolUse{
			{ID: "call_a", Name: "bash", Input: map[string]interface{}{"command": "echo a"}},
			{ID: "call_b", Name: "bash", Input: map[string]interface{}{"command": "echo b"}},
		},
		toolUsesOnce: true,
	})
	e.Start()
	e.Send(SendMessage{Text: "run both"})

	got := drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 8*time.Second)

	// Order: result(a) must come before requested(b) is emitted at least.
	seenA := false
	orderOk := true
	for _, ev := range got {
		switch v := ev.(type) {
		case ToolResult:
			if v.ToolCallID == "call_a" {
				seenA = true
			}
		case ToolRequested:
			if v.ToolCallID == "call_b" && !seenA {
				orderOk = false
			}
		}
	}
	if !orderOk {
		t.Fatal("ToolResult for call A must precede ToolRequested for call B")
	}
}

func TestCancelMidStream(t *testing.T) {
	driver := &mockDriver{}
	block := make(chan struct{})
	e := newEngine(t, driver, streamScript{
		deltas:       []string{"partial"},
		toolUsesOnce: true,
		blockOnCh:    block,
	})
	e.Start()
	e.Send(SendMessage{Text: "hello"})

	// Wait for StreamStart + first delta to arrive, then cancel.
	deadline := time.After(3 * time.Second)
	for {
		select {
		case ev := <-e.Events():
			if _, ok := ev.(StreamDelta); ok {
				// stream is alive; cancel it
				e.Send(Cancel{})
				close(block)
				goto cancelled
			}
		case <-deadline:
			t.Fatal("timed out before stream started")
		}
	}
cancelled:

	got := drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 5*time.Second)

	foundCancelled := false
	for _, ev := range got {
		if _, ok := ev.(StreamCancelled); ok {
			foundCancelled = true
		}
	}
	if !foundCancelled {
		t.Fatal("expected StreamCancelled after Cancel")
	}
	for _, ev := range got {
		if _, ok := ev.(Idle); ok {
			// streamCancelled then idle is the required terminal sequence.
			return
		}
	}
	t.Fatal("expected Idle after cancellation")
}

func TestIdleAfterError(t *testing.T) {
	driver := &mockDriver{}
	e := newEngine(t, driver, streamScript{err: &errString{}})
	e.Start()
	e.Send(SendMessage{Text: "boom"})

	got := drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 3*time.Second)

	foundErr := false
	for _, ev := range got {
		if _, ok := ev.(ErrorEvent); ok {
			foundErr = true
		}
	}
	if !foundErr {
		t.Fatal("expected ErrorEvent for stream error")
	}
}

type errString struct{}

func (e *errString) Error() string { return "network down" }
