package services

import (
	"os"
	"sync"
	"time"

	"github.com/user/go-claude-code/internal/engine"
)

// ChatService exposes the conversation to the Wails frontend (T043,
// frontend-contract §1). It forwards commands to the engine and re-emits the
// stream as `chat:*` events, batching deltas ~50ms (research.md D3,
// gui-contract §2).
//
// The stream lifecycle mapped here, one Wails event per engine event:
//
//	chat:start / chat:delta* / chat:end | chat:cancelled | chat:error
//	tool:request (driver) / tool:start / tool:end
//	usage:update / mode:changed / task:update / plan:* / chat:cleared / chat:idle
type ChatService struct {
	hub *Hub

	mu     sync.Mutex
	buffer string

	done     chan struct{}
	stopOnce sync.Once
}

// NewChatService wires the service to the hub (engine + emitter).
func NewChatService(hub *Hub) *ChatService {
	return &ChatService{hub: hub, done: make(chan struct{})}
}

// Start launches the engine event pump; it batches outgoing deltas to ~50ms.
func (s *ChatService) Start() { go s.pump() }

// Stop terminates the pump and cancels any in-flight stream (FR-014).
func (s *ChatService) Stop() {
	s.stopOnce.Do(func() { close(s.done) })
	s.hub.Engine.Send(engine.Cancel{})
}

// Send submits a user message to the conversation (contract §1 send).
func (s *ChatService) Send(text string) {
	sessionID := ""
	if s.hub != nil && s.hub.Engine != nil {
		sessionID = s.hub.Engine.SessionID()
	}
	s.SendInSession(text, sessionID)
}

// SendInSession submits a user message ensuring the target session is switched and linked.
func (s *ChatService) SendInSession(text string, sessionID string) {
	if s.hub == nil || s.hub.Engine == nil {
		return
	}
	if sessionID != "" {
		s.hub.Engine.SwitchSession(sessionID)
	}
	s.hub.Engine.Send(engine.SendMessage{
		Text:      text,
		SessionID: s.hub.Engine.SessionID(),
	})
}

// Cancel stops the in-flight stream; partial text is not persisted
// (frontend-contract C-002).
func (s *ChatService) Cancel() { s.hub.Engine.Send(engine.Cancel{}) }

// Approve answers a pending tool approval surfaced via `tool:request`.
func (s *ChatService) Approve(callID string, allow bool) {
	if !s.hub.resolveApproval(callID, allow) {
		s.hub.emit("tool:request:expired", map[string]any{"call_id": callID})
	}
}

// ApproveScope answers a pending tool approval with a scope: "once", "chat", "project", "deny".
func (s *ChatService) ApproveScope(callID string, scope string, toolName string) {
	allow := scope != "deny"
	if scope == "chat" && s.hub.Engine != nil {
		cat := engine.ToolCategory(toolName)
		s.hub.Engine.Send(engine.GrantSession{SessionID: s.hub.Engine.SessionID(), Category: cat, Allow: true})
	} else if scope == "project" && s.hub.Engine != nil {
		cat := engine.ToolCategory(toolName)
		cwd, _ := os.Getwd()
		_ = engine.GrantProject(cwd, cat)
		s.hub.Engine.Send(engine.GrantSession{SessionID: s.hub.Engine.SessionID(), Category: cat, Allow: true})
	}
	s.Approve(callID, allow)
}

// Busy reports whether the engine is mid-turn (feeding the Detener button
// and the account status dot, gui-contract §1).
func (s *ChatService) Busy() bool { return s.hub.Engine.Busy() }

// SessionID is the active conversation id.
func (s *ChatService) SessionID() string { return s.hub.Engine.SessionID() }

func (s *ChatService) pump() {
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case ev, ok := <-s.hub.Engine.Events():
			if !ok {
				s.flush()
				return
			}
			s.dispatch(ev)
		case <-tick.C:
			s.flush()
		case <-s.done:
			return
		}
	}
}

func (s *ChatService) dispatch(ev engine.Event) {
	switch e := ev.(type) {
	case engine.StreamStart:
		s.hub.emit("chat:start", map[string]any{
			"session_id": e.SessionID,
			"request_id": e.RequestID,
		})
	case engine.StreamDelta:
		s.mu.Lock()
		s.buffer += e.Text
		s.mu.Unlock()
	case engine.StreamDone:
		s.flush()
		s.hub.emit("chat:end", map[string]any{"message_id": e.MessageID})
	case engine.StreamCancelled:
		s.flush()
		s.hub.emit("chat:cancelled", map[string]any{})
	case engine.ErrorEvent:
		s.flush()
		s.hub.emit("chat:error", map[string]any{
			"message":     e.Message,
			"recoverable": e.Recoverable,
		})
	case engine.SessionCleared:
		s.flush()
		s.hub.emit("chat:cleared", map[string]any{"session_id": e.SessionID})
	case engine.UserMessageAppended:
		if txt, ok := e.Message.Content.(string); ok {
			s.hub.emit("chat:user", map[string]any{"text": txt})
		}
	case engine.ToolExecuting:
		s.hub.emit("tool:start", map[string]any{
			"call_id": e.ToolCallID,
		})
	case engine.ToolResult:
		s.hub.emit("tool:end", map[string]any{
			"call_id": e.ToolCallID,
			"name":    e.Name,
			"content": e.Content,
			"is_error": e.IsError,
		})
	case engine.ToolRejected:
		s.hub.emit("tool:end", map[string]any{
			"call_id":   e.ToolCallID,
			"name":      e.Name,
			"rejected":  true,
		})
	case engine.ToolTimedOut:
		s.hub.emit("tool:end", map[string]any{
			"call_id":   e.ToolCallID,
			"name":      e.Name,
			"timed_out": true,
		})
	case engine.AgentTaskUpdate:
		s.hub.emit("task:update", map[string]any{
			"task_id": e.TaskID,
			"status":  e.Status,
			"summary": e.Summary,
		})
	case engine.UsageUpdate:
		s.hub.emit("usage:update", map[string]any{
			"input":  e.InputTokens,
			"output": e.OutputTokens,
			"cost":   e.CostUSD,
		})
	case engine.ModeChanged:
		s.hub.emit("mode:changed", map[string]any{"mode": e.Mode})
	case engine.PlanProposed:
		s.hub.emit("plan:proposed", map[string]any{
			"plan_id": e.PlanID,
			"steps":   e.Steps,
			"files":   e.Files,
			"criteria": e.Criteria,
		})
	case engine.PlanDecided:
		s.hub.emit("plan:decided", map[string]any{
			"plan_id": e.PlanID,
			"decision": e.Decision,
		})
	case engine.Idle:
		s.flush()
		s.hub.emit("chat:idle", map[string]any{})
	}
}

// flush emits the accumulated delta buffer as a single batched event.
func (s *ChatService) flush() {
	s.mu.Lock()
	text := s.buffer
	s.buffer = ""
	s.mu.Unlock()
	if text != "" {
		s.hub.emit("chat:delta", map[string]any{"text": text})
	}
}