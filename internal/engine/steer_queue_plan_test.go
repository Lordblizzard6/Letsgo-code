package engine

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/db"
)

// newSteerEngine builds an engine with a controllable stream func.
func newSteerEngine(t *testing.T, driver PermissionDriver) *Engine {
	t.Helper()
	e := New(driver)
	e.streamFunc = func(req api.Request,
		onDelta func(string),
		onToolUse func(api.ToolUse),
		onToolInput func(string, string),
		onUsage func(int, int)) error {
		onDelta("work")
		return nil
	}
	return e
}

func TestSteerInjectsIntoActiveTurn(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{}}
	e := newSteerEngine(t, driver)
	e.Start()

	e.Send(SendMessage{Text: "first task"})
	waitForEvents(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 3*time.Second)

	// Steer during repose must be rejected with a recoverable error.
	e.Send(Steer{Text: "nudge", SessionID: e.SessionID()})
	got := waitForEvents(t, e, func(ev Event) bool {
		_, ok := ev.(ErrorEvent)
		return ok
	}, 3*time.Second)
	if _, ok := got[len(got)-1].(ErrorEvent); !ok {
		t.Fatalf("expected ErrorEvent for steer in repose, got %T", got[len(got)-1])
	}
}

func TestSteerWithActiveStreamRestartsTurn(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{}}
	e := New(driver)
	block := make(chan struct{})
	requests := 0
	e.streamFunc = func(req api.Request,
		onDelta func(string),
		onToolUse func(api.ToolUse),
		onToolInput func(string, string),
		onUsage func(int, int)) error {
		requests++
		if requests == 1 {
			onDelta("first")
			<-block
			return context.Canceled
		}
		onDelta("second")
		return nil
	}
	e.Start()
	e.Send(SendMessage{Text: "go"})

	// Wait until first stream started.
	waitForEvents(t, e, func(ev Event) bool {
		_, ok := ev.(StreamDelta)
		return ok
	}, 3*time.Second)

	e.Send(Steer{Text: "actually do X", SessionID: e.SessionID()})

	got := drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 4*time.Second)

	foundSteer := false
	for _, ev := range got {
		if _, ok := ev.(SteerQueued); ok {
			foundSteer = true
		}
	}
	if !foundSteer {
		t.Fatalf("expected SteerQueued event; got %d events", len(got))
	}
}

func TestQueueRunsAfterIdle(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{}}
	e := newSteerEngine(t, driver)
	e.Start()

	// Queue while idle: runs as a turn directly.
	e.Send(Queue{Text: "queued task", SessionID: e.SessionID()})
	got := drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 3*time.Second)
	if len(got) == 0 {
		t.Fatal("no events for queued turn")
	}
	if _, ok := got[0].(UserMessageAppended); !ok {
		t.Fatalf("first event must be UserMessageAppended, got %T", got[0])
	}
}

func TestSetPlanModeEmitsModeChanged(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{}}
	e := newSteerEngine(t, driver)
	e.Start()
	e.Send(SetPlanMode{Mode: "plan"})

	got := waitForEvents(t, e, func(ev Event) bool {
		mc, ok := ev.(ModeChanged)
		return ok && mc.Mode == "plan"
	}, 3*time.Second)
	if _, ok := got[len(got)-1].(ModeChanged); !ok {
		t.Fatalf("expected ModeChanged, got %T", got[len(got)-1])
	}
	if e.Mode() != "plan" {
		t.Fatalf("engine mode = %q, want plan", e.Mode())
	}
}

func TestPlanGateHoldsToolsUntilApproved(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{}}
	e := New(driver)
	executed := false
	e.streamFunc = func(req api.Request,
		onDelta func(string),
		onToolUse func(api.ToolUse),
		onToolInput func(string, string),
		onUsage func(int, int)) error {
		onToolUse(api.ToolUse{ID: "call_plan", Name: "write_file", Input: map[string]interface{}{"path": "/tmp/x.txt"}})
		b, _ := json.Marshal(map[string]interface{}{"path": "/tmp/x.txt", "content": "x"})
		onToolInput("call_plan", string(b))
		return nil
	}
	e.streamFunc = func(req api.Request, onDelta func(string), onToolUse func(api.ToolUse), onToolInput func(string, string), onUsage func(int, int)) error {
		onToolUse(api.ToolUse{ID: "call_plan", Name: "write_file", Input: map[string]interface{}{"path": "/tmp/x.txt"}})
		onToolInput("call_plan", `{"path":"/tmp/x.txt","content":"x"}`)
		return nil
	}
	// Intercept tool execution to assert the gate prevented it pre-approval.
	orig := executeToolFunc
	executeToolFunc = func(name string, input map[string]interface{}) (string, error) {
		executed = true
		return "ok", nil
	}
	defer func() { executeToolFunc = orig }()

	e.Start()
	e.Send(SetPlanMode{Mode: "plan"})
	waitForEvents(t, e, func(ev Event) bool {
		_, ok := ev.(ModeChanged)
		return ok
	}, 3*time.Second)

	e.Send(SendMessage{Text: "create a file"})

	got := waitForEvents(t, e, func(ev Event) bool {
		_, ok := ev.(PlanProposed)
		return ok
	}, 4*time.Second)
	if _, ok := got[len(got)-1].(PlanProposed); !ok {
		t.Fatalf("expected PlanProposed, got %T", got[len(got)-1])
	}
	if executed {
		t.Fatal("tool executed before plan approval")
	}

	// Approve the plan: the tool must now execute and the turn completes.
	e.Send(ApprovePlan{PlanID: "plan_does_not_match"}) // wrong id ignored
	planID := ""
	for _, ev := range got {
		if pp, ok := ev.(PlanProposed); ok {
			planID = pp.PlanID
		}
	}
	e.Send(ApprovePlan{PlanID: planID})
	drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 4*time.Second)
	if !executed {
		t.Fatal("tool did not execute after plan approval")
	}
}

func TestPlanRejectDoesNotExecute(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{}}
	e := New(driver)
	executed := false
	e.streamFunc = func(req api.Request, onDelta func(string), onToolUse func(api.ToolUse), onToolInput func(string, string), onUsage func(int, int)) error {
		onToolUse(api.ToolUse{ID: "call_plan", Name: "write_file", Input: map[string]interface{}{"path": "/tmp/x.txt"}})
		onToolInput("call_plan", `{"path":"/tmp/x.txt","content":"x"}`)
		return nil
	}
	orig := executeToolFunc
	executeToolFunc = func(name string, input map[string]interface{}) (string, error) {
		executed = true
		return "ok", nil
	}
	defer func() { executeToolFunc = orig }()

	e.Start()
	e.Send(SetPlanMode{Mode: "plan"})
	waitForEvents(t, e, func(ev Event) bool {
		_, ok := ev.(ModeChanged)
		return ok
	}, 3*time.Second)

	e.Send(SendMessage{Text: "create a file"})
	got := waitForEvents(t, e, func(ev Event) bool {
		_, ok := ev.(PlanProposed)
		return ok
	}, 4*time.Second)
	planID := ""
	for _, ev := range got {
		if pp, ok := ev.(PlanProposed); ok {
			planID = pp.PlanID
		}
	}
	e.Send(RejectPlan{PlanID: planID, Reason: "not needed"})

	drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(PlanDecided)
		return ok
	}, 4*time.Second)
	if executed {
		t.Fatal("tool executed after plan rejection")
	}
}

func TestSessionGrantSkipsPrompt(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{}, grants: map[string]bool{}}
	e := New(driver)
	e.streamFunc = func(req api.Request, onDelta func(string), onToolUse func(api.ToolUse), onToolInput func(string, string), onUsage func(int, int)) error {
		onToolUse(api.ToolUse{ID: "call_1", Name: "bash", Input: map[string]interface{}{"command": "echo hi"}})
		onToolInput("call_1", `{"command":"echo hi"}`)
		return nil
	}
	orig := executeToolFunc
	executeToolFunc = func(name string, input map[string]interface{}) (string, error) { return "ok", nil }
	defer func() { executeToolFunc = orig }()

	e.Start()
	e.Send(GrantSession{Category: "bash", SessionID: e.SessionID(), Allow: true})
	waitForEvents(t, e, func(ev Event) bool {
		_, ok := ev.(GrantsChanged)
		return ok
	}, 3*time.Second)

	e.Send(SendMessage{Text: "run it"})
	drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 4*time.Second)

	if driver.calls() != 0 {
		t.Fatalf("expected 0 prompt calls with session grant, got %d", driver.calls())
	}
}

func TestGrantRevokeRestoresPrompt(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{}, grants: map[string]bool{}}
	e := New(driver)
	reqs := 0
	e.streamFunc = func(req api.Request, onDelta func(string), onToolUse func(api.ToolUse), onToolInput func(string, string), onUsage func(int, int)) error {
		reqs++
		onToolUse(api.ToolUse{ID: "call_1", Name: "bash", Input: map[string]interface{}{"command": "echo hi"}})
		onToolInput("call_1", `{"command":"echo hi"}`)
		return nil
	}
	orig := executeToolFunc
	executeToolFunc = func(name string, input map[string]interface{}) (string, error) { return "ok", nil }
	defer func() { executeToolFunc = orig }()
	driver.promptCb = func(toolName, input string) (bool, error) { return true, nil }

	e.Start()
	// Grant, use, revoke, use again.
	e.Send(GrantSession{Category: "bash", SessionID: e.SessionID(), Allow: true})
	waitForEvents(t, e, func(ev Event) bool {
		_, ok := ev.(GrantsChanged)
		return ok
	}, 3*time.Second)
	e.Send(SendMessage{Text: "run once"})
	drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 4*time.Second)
	if driver.calls() != 0 {
		t.Fatalf("prompt called while grant active: %d", driver.calls())
	}

	e.Send(GrantSession{Category: "bash", SessionID: e.SessionID(), Allow: false})
	waitForEvents(t, e, func(ev Event) bool {
		gc, ok := ev.(GrantsChanged)
		return ok && !gc.Allow
	}, 3*time.Second)
	e.Send(SendMessage{Text: "run twice"})
	drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 4*time.Second)
	if driver.calls() == 0 {
		t.Fatal("prompt not restored after grant revoke")
	}
}

func TestForkSessionSwitchesEngine(t *testing.T) {
	driver := &mockDriver{autoAppr: map[string]bool{}}
	e := newSteerEngine(t, driver)
	e.Start()

	sid, err := db.CreateSession("fork-src", "/tmp")
	if err != nil {
		t.Fatal(err)
	}
	e.Send(SendMessage{Text: "hello world", SessionID: sid})
	drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(Idle)
		return ok
	}, 3*time.Second)
	if e.SessionID() != sid {
		t.Fatalf("engine session = %q, want %q", e.SessionID(), sid)
	}

	hist, err := db.GetHistory(sid)
	if err != nil || len(hist) == 0 {
		t.Fatalf("no history before fork: %v, len=%d", err, len(hist))
	}
	targetID := hist[len(hist)-1].ID

	e.Send(ForkSession{SourceID: sid, MessageID: targetID})
	drainUntil(t, e, func(ev Event) bool {
		_, ok := ev.(ModeChanged)
		return ok
	}, 3*time.Second)

	if e.SessionID() == sid {
		t.Fatal("engine did not switch to forked session")
	}
	// The fork must have history (the original "hello world").
	hist, err = db.GetHistory(e.SessionID())
	if err != nil || len(hist) == 0 {
		t.Fatalf("fork has no history: %v, len=%d", err, len(hist))
	}
}
