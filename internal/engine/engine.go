package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/tools"
)

// streamBatchInterval caps how often StreamDelta events are emitted (≤100 ms).
const streamBatchInterval = 100 * time.Millisecond

// executeToolFunc is injectable for tests; defaults to the tools registry.
var executeToolFunc = func(name string, input map[string]interface{}) (string, error) {
	return tools.ExecuteToolWithPermission(name, input)
}

var ensureDBOnce sync.Once

// ensureDB initializes the persistence layer exactly once.
func ensureDB() {
	ensureDBOnce.Do(func() { _ = db.InitDB() })
}

// Engine is the presentation-agnostic conversation/agent loop. It is the ONLY
// owner of the loop; presentations send commands and render events
// (contracts/engine-contract.md).
type Engine struct {
	driver PermissionDriver

	out  chan Event
	cmds chan Command
	done chan struct{}

	mu          sync.Mutex
	sessionID   string
	messages    []api.Message
	client      *api.Client
	autoApprove map[string]bool

	// streamFunc is injectable for tests; defaults to client.StreamRequest.
	streamFunc StreamFunc

	// In-flight stream cancellation.
	streamCancel context.CancelFunc

	// In-flight approval: ID of the tool call currently awaiting a decision.
	pendingToolCall string

	// cancelRequested is set by Cancel; no tool may execute while it is set.
	cancelRequested bool

	busy bool

	// steers holds injected instructions awaiting the in-flight turn restart.
	steers []string

	// queued holds follow-up turns to run FIFO on Idle (FR-010).
	queued []string

	// mode is the execution/approval mode: "auto" (default), "execute", "plan".
	mode string

	// plan tracks the currently pending plan card (FR-012/013).
	plan planState
}

// planState tracks the currently pending plan card (FR-012/013).
type planState struct {
	ID           string
	instructions string
	decided      chan bool
}

// StreamFunc is the provider streaming entry point used by runTurn. The
// default is *api.Client.StreamRequest; tests and presentations that need
// scripted providers can replace it via SetStreamFunc.
type StreamFunc func(
	req api.Request,
	onDelta func(string),
	onToolUse func(api.ToolUse),
	onToolInput func(string, string),
	onUsage func(int, int),
) error

// SetStreamFunc replaces the provider streaming implementation. It exists for
// the conformance/teatest harnesses (005 frontend-contract C-001..C-006, TUI
// teatest); production code never calls it.
func (e *Engine) SetStreamFunc(fn StreamFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.streamFunc = fn
}

// New creates an engine bound to a driver.
func New(driver PermissionDriver) *Engine {
	ensureDB()
	return &Engine{
		driver:          driver,
		out:             make(chan Event, 256),
		cmds:            make(chan Command, 64),
		done:            make(chan struct{}),
		sessionID:       uuid.New().String(),
		messages:        []api.Message{},
		client:          newClientForModel(),
		autoApprove:     make(map[string]bool),
		pendingToolCall: "",
	}
}

// Events returns the presentation-facing event channel.
func (e *Engine) Events() <-chan Event { return e.out }

// Send submits a command to the engine loop (non-blocking).
func (e *Engine) Send(cmd Command) { e.cmds <- cmd }

// Start begins processing commands on a background goroutine.
func (e *Engine) Start() { go e.run() }

// Stop gracefully shuts the loop down.
func (e *Engine) Stop() {
	e.Send(Stop{})
	select {
	case <-e.done:
	case <-time.After(3 * time.Second):
	}
}

// AutoApproveEnabled reports the current auto-approve state for a category.
func (e *Engine) AutoApproveEnabled(category string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.autoApprove[category]
}

// SeedMessages sets the initial conversation context (e.g. welcome/note messages
// previously included by a presentation). Called before the first SendMessage.
func (e *Engine) SeedMessages(messages []api.Message) {
	e.mu.Lock()
	e.messages = append([]api.Message(nil), messages...)
	e.mu.Unlock()
}

// RefreshClient rebuilds the API client from the current config (e.g. after the
// user changes the provider or model).
func (e *Engine) RefreshClient() {
	e.mu.Lock()
	e.client = newClientForModel()
	e.mu.Unlock()
}

// Busy reports whether a turn is in progress.
func (e *Engine) Busy() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.busy
}

// SessionID returns the current session.
func (e *Engine) SessionID() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.sessionID
}

func (e *Engine) run() {
	for cmd := range e.cmds {
		switch c := cmd.(type) {
		case SendMessage:
			e.mu.Lock()
			if e.busy {
				e.mu.Unlock()
				continue
			}
			e.busy = true
			e.mu.Unlock()
			go e.handleSendMessage(c.Text, c.SessionID)
		case SendTask:
			e.mu.Lock()
			if e.busy {
				e.mu.Unlock()
				continue
			}
			e.busy = true
			e.mu.Unlock()
			go e.handleSendMessage(c.Prompt, c.SessionID)
		case Cancel:
			e.handleCancel()
		case SetModelCommand:
			e.handleSetModel(c.Provider, c.Model)
		case RememberCommand:
			e.handleRemember(c.Key, c.Value)
		case ResetMemoryCommand:
			e.handleResetMemory()
		case SlashCommand:
			e.handleSlash(c.Name, c.Args)
		case ApproveTool:
			e.handleApproveTool(c.ToolCallID, true)
		case RejectTool:
			e.handleApproveTool(c.ToolCallID, false)
		case SetAutoApprove:
			e.handleSetAutoApprove(c.Category, c.Enabled)
		case SwitchSession:
			e.handleSwitchSession(c.SessionID)
		case Compact:
			e.handleCompact()
		case Steer:
			e.handleSteer(c.Text, c.SessionID)
		case Queue:
			e.handleQueue(c.Text, c.SessionID)
		case SetPlanMode:
			e.handleSetPlanMode(c.Mode)
		case ApprovePlan:
			e.handlePlanDecision(c.PlanID, true)
		case RejectPlan:
			e.handlePlanDecision(c.PlanID, false)
		case EditPlan:
			e.handlePlanEdit(c.PlanID, c.Instructions)
		case GrantSession:
			e.handleGrantSession(c.Category, c.SessionID, c.Allow)
		case ForkSession:
			e.handleForkSession(c.SourceID, c.MessageID)
		case Stop:
			close(e.done)
			return
		}
	}
}

// handleSendMessage runs one full user turn: persist, stream, tool cycle.
func (e *Engine) handleSendMessage(text, sessionID string) {
	e.mu.Lock()
	if sessionID != "" {
		e.sessionID = sessionID
	}
	text = strings.TrimSpace(text)
	e.cancelRequested = false
	if text == "" {
		e.busy = false
		e.mu.Unlock()
		return
	}
	sid := e.sessionID
	client := e.client
	auto := cloneAutoApprove(e.autoApprove)
	driver := e.driver
	e.mu.Unlock()

	expandedText := expandMentions(text)
	userMsg := api.Message{Role: "user", Content: expandedText}
	e.emit(UserMessageAppended{Message: api.Message{Role: "user", Content: text}})
	_ = db.SaveMessage(sid, "user", text)
	messages := append(e.messages, userMsg)

	for {
		if steers := e.drainSteers(); len(steers) > 0 {
			for _, s := range steers {
				m := api.Message{Role: "user", Content: s}
				messages = append(messages, m)
				e.emit(UserMessageAppended{Message: m})
				_ = db.SaveMessageKind(sid, "user", s, "user:steer")
			}
		}

		result := e.runTurn(sid, messages, client)
		if result.err != nil {
			if result.cancelled {
				// If a steer arrived mid-stream, the turn was cancelled so it can
				// restart with the injected instruction appended to the context.
				if steers := e.drainSteers(); len(steers) > 0 {
					for _, s := range steers {
						m := api.Message{Role: "user", Content: s}
						messages = append(messages, m)
						e.emit(UserMessageAppended{Message: m})
						_ = db.SaveMessageKind(sid, "user", s, "user:steer")
					}
					continue
				}
				e.emit(StreamCancelled{})
				e.emit(Idle{})
				e.finish(sid, messages)
				return
			}
			e.emit(ErrorEvent{Message: result.err.Error(), Recoverable: true})
			e.emit(Idle{})
			e.finish(sid, messages)
			return
		}
		messages = result.messages

		// Plan gate: in plan mode, present a plan card before executing tools.
		if len(result.pending) > 0 && e.planGateActive() {
			planID := "plan_" + uuid.NewString()
			steps := planStepsOf(result.pending)
			files := planFilesOf(result.pending)
			e.emit(PlanProposed{PlanID: planID, Steps: steps, Files: files, Criteria: ""})

			decision, err := e.awaitPlanDecision(planID)
			if err != nil {
				e.emit(PlanDecided{PlanID: planID, Decision: "rejected"})
				rejectMsg := "Plan rejected or abandoned."
				tr := api.ToolResult{ToolUseID: "", ToolName: "plan", Content: rejectMsg, IsError: true}
				messages = appendToolResult(messages, tr)
				_ = db.SaveMessage(sid, "user", contentBlockOf(tr))
				continue
			}
			if !decision {
				e.emit(PlanDecided{PlanID: planID, Decision: "rejected"})
				rejectMsg := "Plan rejected by user."
				tr := api.ToolResult{ToolUseID: "", ToolName: "plan", Content: rejectMsg, IsError: true}
				messages = appendToolResult(messages, tr)
				_ = db.SaveMessage(sid, "user", contentBlockOf(tr))
				continue
			}
			e.emit(PlanDecided{PlanID: planID, Decision: "approved"})
			instruction := e.takePlanInstructions(planID)
			if instruction != "" {
				m := api.Message{Role: "user", Content: "Plan edit: " + instruction}
				messages = append(messages, m)
				_ = db.SaveMessageKind(sid, "user", instruction, "user:plan_instructions")
			}
			messages = append(messages, api.Message{Role: "user", Content: "Plan approved. Proceed with execution."})
		}

		if len(result.pending) == 0 {
			e.emit(Idle{})
			e.finish(sid, messages)
			return
		}
		// Tool cycle: approval per SC-007 (no execution without approval).
		next, done := e.runToolCycle(sid, messages, result.pending, auto, driver)
		if done {
			e.emit(Idle{})
			e.finish(sid, next)
			return
		}
		messages = next
	}
}

type turnResult struct {
	messages  []api.Message
	pending   []api.ToolUse
	err       error
	cancelled bool
}

// runTurn streams one request and collects deltas + tool uses.
func (e *Engine) runTurn(sid string, messages []api.Message, client *api.Client) turnResult {
	ctx, cancel := context.WithCancel(context.Background())
	e.mu.Lock()
	e.streamCancel = cancel
	e.mu.Unlock()
	defer func() {
		cancel()
		e.mu.Lock()
		e.streamCancel = nil
		e.mu.Unlock()
	}()

	cwd, _ := os.Getwd()
	isPlan := e.planGateActive()
	model := config.AppConfig.Model
	var systemPrompt string
	var toolDefs []api.Tool

	if isPlan {
		if config.AppConfig.PlanModel != "" {
			model = config.AppConfig.PlanModel
		}
		systemPrompt = api.GetPlanSystemPrompt(model, cwd)
		toolDefs = tools.GetPlanToolDefinitions()
	} else {
		systemPrompt = api.GetBuildSystemPrompt(model, cwd)
		toolDefs = tools.GetToolDefinitions()
	}

	pruned := db.PruneContext(messages, 50)
	req := api.Request{
		Model:     model,
		System:    systemPrompt,
		MaxTokens: config.AppConfig.MaxTokens,
		Stream:    true,
		Messages:  pruned,
		Tools:     toolDefs,
	}

	requestID := uuid.NewString()
	e.emit(StreamStart{SessionID: sid, RequestID: requestID})

	var (
		mu         sync.Mutex
		text       strings.Builder
		pending    []api.ToolUse
		toolInputs = map[string]*strings.Builder{}
	)

	// Delta batching: accumulate into a buffer, flush on a ≤100 ms ticker.
	batchBuf := &strings.Builder{}
	flushDone := make(chan struct{})
	flush := func() {
		mu.Lock()
		defer mu.Unlock()
		if batchBuf.Len() > 0 {
			e.emit(StreamDelta{Text: batchBuf.String()})
			batchBuf.Reset()
		}
	}
	go func() {
		t := time.NewTicker(streamBatchInterval)
		defer t.Stop()
		defer close(flushDone)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				flush()
			}
		}
	}()
	// wake channel to flush immediately on stream end.
	wake := make(chan struct{}, 1)
	go func() {
		for range wake {
			flush()
		}
	}()
	defer close(wake)

	stream := client.StreamRequest
	if e.streamFunc != nil {
		stream = e.streamFunc
	}
	err := stream(req,
		func(delta string) {
			mu.Lock()
			text.WriteString(delta)
			batchBuf.WriteString(delta)
			mu.Unlock()
			select {
			case wake <- struct{}{}:
			default:
			}
		},
		func(tu api.ToolUse) {
			mu.Lock()
			if tu.ID != "" && tu.Name != "" {
				pending = append(pending, tu)
				if _, ok := toolInputs[tu.ID]; !ok {
					toolInputs[tu.ID] = &strings.Builder{}
				}
			}
			mu.Unlock()
		},
		func(id, delta string) {
			mu.Lock()
			b, ok := toolInputs[id]
			if !ok {
				b = &strings.Builder{}
				toolInputs[id] = b
			}
			b.WriteString(delta)
			mu.Unlock()
		},
		func(in, out int) {
			if in > 0 || out > 0 {
				e.emit(UsageUpdate{InputTokens: in, OutputTokens: out})
			}
		},
	)

	// Final flush of any remaining batched text.
	select {
	case wake <- struct{}{}:
	default:
		flush()
	}

	if err != nil {
		cancelled := ctx.Err() != nil || errors.Is(err, context.Canceled)
		return turnResult{messages: messages, err: err, cancelled: cancelled}
	}

	// Assemble accumulated streamed tool input arguments into pending tools
	for i := range pending {
		if b, ok := toolInputs[pending[i].ID]; ok && b.Len() > 0 {
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(b.String()), &m); err == nil {
				pending[i].Input = m
			} else {
				pending[i].Input = b.String()
			}
		}
	}

	// Persist the assistant message if we have content or tool calls.
	assistantText := text.String()
	if assistantText != "" || len(pending) > 0 {
		if len(pending) > 0 {
			var blocks []api.ContentBlock
			if assistantText != "" {
				blocks = append(blocks, api.ContentBlock{Type: "text", Text: assistantText})
			}
			for _, tu := range pending {
				tuCopy := tu
				blocks = append(blocks, api.ContentBlock{Type: "tool_use", ToolUse: &tuCopy})
			}
			_ = db.SaveMessage(sid, "assistant", blocks)
			messages = append(messages, api.Message{Role: "assistant", Content: blocks})
		} else {
			_ = db.SaveMessage(sid, "assistant", assistantText)
			messages = append(messages, api.Message{Role: "assistant", Content: assistantText})
		}
		e.emit(StreamDone{MessageID: uuid.NewString()})
	}

	return turnResult{messages: messages, pending: pending}
}

// runToolCycle executes pending tool calls in order, each gated by approval.
// Returns (messages, done) where done=true stops the turn.
func (e *Engine) runToolCycle(sid string, messages []api.Message, pending []api.ToolUse, auto map[string]bool, driver PermissionDriver) ([]api.Message, bool) {
	for _, tu := range pending {
		input, err := parseToolInput(tu)
		if err != nil {
			result := fmt.Sprintf("Error parsing tool input: %v", err)
			tr := api.ToolResult{ToolUseID: tu.ID, ToolName: tu.Name, Content: result, IsError: true}
			e.emit(ToolResult{ToolCallID: tu.ID, Name: tu.Name, Content: result, IsError: true})
			messages = appendToolResult(messages, tr)
			_ = db.SaveMessage(sid, "user", contentBlockOf(tr))
			continue
		}

		cat := toolCategory(tu.Name)
		cwd, _ := os.Getwd()
		granted := driver.SessionGranted(sid, cat) || IsProjectGranted(cwd, cat)
		autoApproved := auto[cat] || driver.AutoApprove(cat) || granted
		if autoApproved {
			e.emit(ToolRequested{ToolCallID: tu.ID, Name: tu.Name, Input: input, AutoApproved: true})
		} else {
			e.mu.Lock()
			e.pendingToolCall = tu.ID
			e.mu.Unlock()
			e.emit(ToolRequested{ToolCallID: tu.ID, Name: tu.Name, Input: input, AutoApproved: false})

			allow, perr := promptWithTimeout(driver, tu.Name, jsonInputString(input))
			e.mu.Lock()
			e.pendingToolCall = ""
			cancelled := e.cancelRequested
			e.mu.Unlock()

			if cancelled {
				e.emit(ToolRejected{ToolCallID: tu.ID, Name: tu.Name})
				rejectMsg := "Tool call rejected (cancelled)"
				tr := api.ToolResult{ToolUseID: tu.ID, ToolName: tu.Name, Content: rejectMsg, IsError: true}
				e.emit(ToolResult{ToolCallID: tu.ID, Name: tu.Name, Content: rejectMsg, IsError: true})
				messages = appendToolResult(messages, tr)
				_ = db.SaveMessage(sid, "user", contentBlockOf(tr))
				continue
			}

			if perr == ErrApprovalTimedOut {
				e.emit(ToolTimedOut{ToolCallID: tu.ID, Name: tu.Name})
				e.emit(ToolRejected{ToolCallID: tu.ID, Name: tu.Name})
				rejectMsg := fmt.Sprintf("Tool call rejected (approval timed out): %s", tu.Name)
				tr := api.ToolResult{ToolUseID: tu.ID, ToolName: tu.Name, Content: rejectMsg, IsError: true}
				e.emit(ToolResult{ToolCallID: tu.ID, Name: tu.Name, Content: rejectMsg, IsError: true})
				messages = appendToolResult(messages, tr)
				_ = db.SaveMessage(sid, "user", contentBlockOf(tr))
				continue
			}
			if perr != nil {
				e.emit(ErrorEvent{Message: perr.Error(), Recoverable: false})
			}
			if !allow {
				e.emit(ToolRejected{ToolCallID: tu.ID, Name: tu.Name})
				rejectMsg := fmt.Sprintf("Tool call rejected by user: %s", tu.Name)
				tr := api.ToolResult{ToolUseID: tu.ID, ToolName: tu.Name, Content: rejectMsg, IsError: true}
				e.emit(ToolResult{ToolCallID: tu.ID, Name: tu.Name, Content: rejectMsg, IsError: true})
				messages = appendToolResult(messages, tr)
				_ = db.SaveMessage(sid, "user", contentBlockOf(tr))
				continue
			}
		}

		// SC-007: never execute a tool after a Cancel was requested.
		e.mu.Lock()
		cancelled := e.cancelRequested
		e.mu.Unlock()
		if cancelled {
			e.emit(ToolRejected{ToolCallID: tu.ID, Name: tu.Name})
			rejectMsg := "Tool call rejected (cancelled)"
			tr := api.ToolResult{ToolUseID: tu.ID, ToolName: tu.Name, Content: rejectMsg, IsError: true}
			e.emit(ToolResult{ToolCallID: tu.ID, Name: tu.Name, Content: rejectMsg, IsError: true})
			messages = appendToolResult(messages, tr)
			_ = db.SaveMessage(sid, "user", contentBlockOf(tr))
			continue
		}

		e.emit(ToolExecuting{ToolCallID: tu.ID})
		result, err := executeToolFunc(tu.Name, input)
		isErr := false
		if err != nil {
			result = err.Error()
			isErr = true
		}
		tr := api.ToolResult{ToolUseID: tu.ID, ToolName: tu.Name, Content: result, IsError: isErr}
		e.emit(ToolResult{ToolCallID: tu.ID, Name: tu.Name, Content: result, IsError: isErr})
		messages = appendToolResult(messages, tr)
		_ = db.SaveMessage(sid, "user", contentBlockOf(tr))
	}
	return messages, false
}

// handleApproveTool approves or rejects the in-flight pending call.
func (e *Engine) handleApproveTool(callID string, approve bool) {
	// The driver Prompt blocks synchronously in runToolCycle; by the time an
	// Approve/Reject command arrives, Prompt has already returned for that call.
	// We track the pending ID to keep SC-007 (approve path only) meaningful for
	// presentations that answer asynchronously.
	e.mu.Lock()
	e.pendingToolCall = ""
	e.mu.Unlock()
	_ = approve
}

func (e *Engine) handleCancel() {
	e.mu.Lock()
	cancel := e.streamCancel
	e.cancelRequested = true
	e.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// handleSetModel implements contract §1 `setModel`: persists the model via
// internal/config, refreshes the client in place and emits ConfigChanged
// (C-004). The interface is never reinitialized.
func (e *Engine) handleSetModel(provider, model string) {
	model = strings.TrimSpace(model)
	if model == "" {
		e.emit(ErrorEvent{Message: "setModel: empty model id", Recoverable: true})
		return
	}
	config.AppConfig.Model = model
	if err := config.SaveConfig(); err != nil {
		e.emit(ErrorEvent{Message: "setModel: " + err.Error(), Recoverable: true})
		return
	}
	e.RefreshClient()
	e.emit(ConfigChanged{Config: map[string]any{
		"model":    model,
		"provider": config.DetectProviderFromModel(model),
	}})
}

// handleRemember implements contract §1 `remember` (session_memory).
func (e *Engine) handleRemember(key, value string) {
	key = strings.TrimSpace(key)
	if key == "" {
		e.emit(ErrorEvent{Message: "remember: empty key", Recoverable: true})
		return
	}
	if err := db.SetMemory(key, value); err != nil {
		e.emit(ErrorEvent{Message: "remember: " + err.Error(), Recoverable: true})
	}
}

// handleResetMemory implements contract §1 `resetMemory`.
func (e *Engine) handleResetMemory() {
	mem, err := db.ListMemory()
	if err != nil {
		e.emit(ErrorEvent{Message: "resetMemory: " + err.Error(), Recoverable: true})
		return
	}
	for k := range mem {
		if err := db.DeleteMemory(k); err != nil {
			e.emit(ErrorEvent{Message: "resetMemory: " + err.Error(), Recoverable: true})
			return
		}
	}
}

// handleSlash dispatches contract §1 `slash:{...}`. Engine-owned names are
// handled here; unknown names get an orientative ErrorEvent for the
// presentation layer to interpret.
func (e *Engine) handleSlash(name string, args []string) {
	switch name {
	case "compact":
		e.handleCompact()
	case "clear":
		e.handleClear()
	default:
		e.emit(ErrorEvent{Message: "slash: unknown command /" + name, Recoverable: true})
	}
}

// handleSteer injects a user instruction into the in-flight turn (FR-009).
// The instruction is persisted as kind=user:steer and the current stream is
// cancelled so the loop restarts with it appended to the context.
func (e *Engine) handleSteer(text, sessionID string) {
	e.mu.Lock()
	if !e.busy {
		e.mu.Unlock()
		e.emit(ErrorEvent{Message: "steer: no active turn", Recoverable: true})
		return
	}
	if sessionID != "" {
		e.sessionID = sessionID
	}
	text = strings.TrimSpace(text)
	sid := e.sessionID
	cancel := e.streamCancel
	e.mu.Unlock()

	if text == "" {
		return
	}
	_ = db.SaveMessageKind(sid, "user", text, "user:steer")
	e.emit(SteerQueued{Text: text})
	// Cancel the current stream so handleSendMessage restarts with the steer.
	if cancel != nil {
		cancel()
	}
}

// handleQueue persists a follow-up message to be run as the next turn (FR-010).
// If a turn is active it is buffered and executed FIFO after Idle.
func (e *Engine) handleQueue(text, sessionID string) {
	e.mu.Lock()
	if sessionID != "" {
		e.sessionID = sessionID
	}
	text = strings.TrimSpace(text)
	sid := e.sessionID
	busy := e.busy
	if busy {
		e.queued = append(e.queued, text)
		e.mu.Unlock()
		_ = db.SaveMessageKind(sid, "user", text, "user:queued")
		return
	}
	e.busy = true
	e.mu.Unlock()
	if text == "" {
		e.finish(sid, e.messages)
		return
	}
	_ = db.SaveMessageKind(sid, "user", text, "user:queued")
	go e.handleSendMessage(text, sid)
}

// handleSetPlanMode toggles the plan/execute/auto mode (FR-012).
func (e *Engine) handleSetPlanMode(mode string) {
	switch mode {
	case "plan", "execute", "auto":
	default:
		e.emit(ErrorEvent{Message: "unknown mode: " + mode, Recoverable: true})
		return
	}
	e.mu.Lock()
	e.mode = mode
	e.mu.Unlock()
	e.emit(ModeChanged{Mode: mode})
}

// handlePlanDecision resolves a pending plan card (FR-013).
func (e *Engine) handlePlanDecision(planID string, approved bool) {
	e.mu.Lock()
	if e.plan.ID != planID {
		e.mu.Unlock()
		return
	}
	ch := e.plan.decided
	e.mu.Unlock()
	if ch != nil {
		select {
		case ch <- approved:
		default:
		}
	}
}

// handlePlanEdit updates the instructions of a proposed plan and approves it
// so a revised run can proceed (FR-013).
func (e *Engine) handlePlanEdit(planID, instructions string) {
	e.mu.Lock()
	if e.plan.ID != planID {
		e.mu.Unlock()
		return
	}
	ch := e.plan.decided
	e.plan.instructions = instructions
	e.mu.Unlock()
	if ch != nil {
		select {
		case ch <- true:
		default:
		}
	}
}

// handleGrantSession persists/revokes a per-session approval grant (FR-016/017).
func (e *Engine) handleGrantSession(category, sessionID string, allow bool) {
	if !validCategory(category) {
		return
	}
	if allow {
		_ = db.GrantCategory(sessionID, category)
	} else {
		_ = db.RevokeCategory(sessionID, category)
	}
	e.emit(GrantsChanged{SessionID: sessionID, Category: category, Allow: allow})
}

// handleForkSession clones the session at messageID and switches to the fork.
func (e *Engine) handleForkSession(sourceID string, messageID int64) {
	newID, err := db.ForkSession(sourceID, messageID)
	if err != nil {
		e.emit(ErrorEvent{Message: "fork: " + err.Error(), Recoverable: true})
		return
	}
	e.handleSwitchSession(newID)
	e.emit(ModeChanged{Mode: e.Mode()})
}

// drainSteers returns and clears any injected steers.
func (e *Engine) drainSteers() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	s := e.steers
	e.steers = nil
	return s
}

// planGateActive reports whether the loop should hold before tool execution (mode==plan).
func (e *Engine) planGateActive() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.mode == "plan"
}

// awaitPlanDecision blocks until the plan card is decided, with the timeout gateway.
func (e *Engine) awaitPlanDecision(planID string) (bool, error) {
	ch := make(chan bool, 1)
	e.mu.Lock()
	e.plan = planState{ID: planID, decided: ch}
	e.mu.Unlock()
	select {
	case approved := <-ch:
		return approved, nil
	case <-time.After(ApprovalTimeout):
		e.mu.Lock()
		e.plan = planState{}
		e.mu.Unlock()
		return false, ErrApprovalTimedOut
	}
}

// Mode returns the current plan/execute/auto mode.
func (e *Engine) Mode() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.mode
}

func (e *Engine) handleSetAutoApprove(category string, enabled bool) {
	if !validCategory(category) {
		return
	}
	e.mu.Lock()
	e.autoApprove[category] = enabled
	e.mu.Unlock()
}

// SwitchSession switches the engine to a session immediately and reloads its history.
func (e *Engine) SwitchSession(sessionID string) {
	e.handleSwitchSession(sessionID)
}

func (e *Engine) handleSwitchSession(sessionID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sessionID = sessionID
	if history, err := db.GetHistory(sessionID); err == nil {
		e.messages = toAPIMessages(history)
	} else {
		e.messages = []api.Message{}
	}
}

func toAPIMessages(history []db.Message) []api.Message {
	out := make([]api.Message, 0, len(history))
	for _, m := range history {
		out = append(out, api.Message{Role: m.Role, Content: m.Content})
	}
	return out
}

func (e *Engine) handleCompact() {
	e.mu.Lock()
	sid := e.sessionID
	messages := e.messages
	e.mu.Unlock()
	orig := len(messages)
	compact := db.PruneContext(messages, 25)
	e.mu.Lock()
	e.messages = compact
	e.mu.Unlock()
	_ = db.SaveCompactHistory(sid, orig, len(compact), "compact")
}

// handleClear owns the `slash:{clear}` state change: the engine resets its
// in-memory context and wipes the session's persisted history (messages,
// context files, compact history). Presentations mirror it via SessionCleared
// instead of touching the DB themselves (contract §3).
func (e *Engine) handleClear() {
	e.mu.Lock()
	sid := e.sessionID
	e.messages = []api.Message{}
	e.mu.Unlock()

	_ = db.ClearSessionHistory(sid)
	if _, err := db.DB.Exec("DELETE FROM context_files WHERE session_id = ?", sid); err != nil {
		e.emit(ErrorEvent{Message: "clear: " + err.Error(), Recoverable: true})
		return
	}
	if _, err := db.DB.Exec("DELETE FROM compact_history WHERE session_id = ?", sid); err != nil {
		e.emit(ErrorEvent{Message: "clear: " + err.Error(), Recoverable: true})
		return
	}
	e.emit(SessionCleared{SessionID: sid})
}

func (e *Engine) finish(sid string, messages []api.Message) {
	e.mu.Lock()
	e.messages = messages
	e.busy = false
	e.streamCancel = nil
	e.mu.Unlock()
	e.runNextQueued(sid)
}

// takePlanInstructions returns and clears the edited instructions of a plan.
func (e *Engine) takePlanInstructions(planID string) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.plan.ID != planID {
		return ""
	}
	instr := e.plan.instructions
	e.plan = planState{}
	return instr
}

// runNextQueued starts the next queued turn FIFO when the loop is idle (FR-010).
func (e *Engine) runNextQueued(sid string) {
	e.mu.Lock()
	if e.busy || len(e.queued) == 0 {
		e.mu.Unlock()
		return
	}
	next := e.queued[0]
	e.queued = e.queued[1:]
	e.busy = true
	e.mu.Unlock()
	e.emit(QueueRuns{Text: next})
	go e.handleSendMessage(next, sid)
}

func (e *Engine) emit(ev Event) {
	select {
	case e.out <- ev:
	case <-e.done:
	}
}

// ---- helpers ----

func newClientForModel() *api.Client {
	provider := config.DetectProviderFromModel(config.AppConfig.Model)
	apiKey := config.GetAPIKeyForProvider(provider)
	var baseURL string
	switch provider {
	case "anthropic":
		baseURL = "https://api.anthropic.com/v1/messages"
	case "openai":
		baseURL = "https://api.openai.com/v1/chat/completions"
	case "groq":
		baseURL = "https://api.groq.com/openai/v1/chat/completions"
	case "openrouter":
		baseURL = "https://openrouter.ai/api/v1/chat/completions"
	case "ollama":
		baseURL = "http://localhost:11434/api/chat"
	default:
		baseURL = config.AppConfig.BaseURL
	}
	return api.NewClient(apiKey, baseURL)
}

func cloneAutoApprove(m map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func parseToolInput(tu api.ToolUse) (map[string]interface{}, error) {
	switch in := tu.Input.(type) {
	case map[string]interface{}:
		return in, nil
	case string:
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(in), &m); err != nil {
			return nil, err
		}
		return m, nil
	default:
		b, err := json.Marshal(tu.Input)
		if err != nil {
			return nil, err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		return m, nil
	}
}

func jsonInputString(input map[string]interface{}) string {
	b, err := json.Marshal(input)
	if err != nil {
		return fmt.Sprintf("%v", input)
	}
	return string(b)
}

func appendToolResult(messages []api.Message, tr api.ToolResult) []api.Message {
	return append(messages, api.Message{Role: "user", Content: []api.ContentBlock{
		{Type: "tool_result", ToolResult: &tr},
	}})
}

func contentBlockOf(tr api.ToolResult) []api.ContentBlock {
	t := tr
	return []api.ContentBlock{{Type: "tool_result", ToolResult: &t}}
}

func validCategory(category string) bool {
	switch category {
	case "bash", "file-edit", "web":
		return true
	}
	return false
}

// planStepsOf derives plan steps from pending tool calls (FR-013).
func planStepsOf(pending []api.ToolUse) []PlanStep {
	steps := make([]PlanStep, 0, len(pending))
	for _, tu := range pending {
		s := PlanStep{Action: tu.Name}
		if input, err := parseToolInput(tu); err == nil {
			if p, ok := input["file"].(string); ok {
				s.File = p
			}
			if p, ok := input["path"].(string); ok {
				s.File = p
			}
			if cmd, ok := input["command"].(string); ok {
				s.Action = s.Action + " " + cmd
			}
		}
		s.Criteria = "User must review this step before it executes."
		steps = append(steps, s)
	}
	return steps
}

// planFilesOf extracts affected file paths from pending tool calls.
func planFilesOf(pending []api.ToolUse) []string {
	var files []string
	for _, tu := range pending {
		input, err := parseToolInput(tu)
		if err != nil {
			continue
		}
		for _, key := range []string{"file", "path"} {
			if p, ok := input[key].(string); ok {
				files = append(files, p)
			}
		}
	}
	return files
}

// expandMentions replaces @path references in user text with formatted file contents.
func expandMentions(text string) string {
	if !strings.Contains(text, "@") {
		return text
	}
	words := strings.Fields(text)
	var out []string
	for _, word := range words {
		if strings.HasPrefix(word, "@") && len(word) > 1 {
			target := strings.Trim(word[1:], `",':;`)
			// Clean path
			cleaned := filepath.Clean(target)
			if info, err := os.Stat(cleaned); err == nil && !info.IsDir() {
				if data, err := os.ReadFile(cleaned); err == nil {
					content := string(data)
					if len(content) > 4000 {
						content = content[:4000] + "\n... [truncated]"
					}
					out = append(out, fmt.Sprintf("\n📄 %s:\n```\n%s\n```\n", target, content))
					continue
				}
			}
		}
		out = append(out, word)
	}
	return strings.Join(out, " ")
}
