package engine

import "github.com/user/go-claude-code/internal/api"

// Event is emitted by the engine to presentations (contracts/engine-contract.md).
type Event interface {
	isEvent()
}

// UserMessageAppended is emitted when a user message has been persisted.
type UserMessageAppended struct {
	Message api.Message
}

// StreamStart is emitted when a provider stream opens.
type StreamStart struct {
	SessionID string
	RequestID string
}

// StreamDelta carries a batched chunk of assistant text (engine batches ≤100ms).
type StreamDelta struct {
	Text string
}

// StreamDone is emitted when an assistant turn completes and is persisted.
type StreamDone struct {
	MessageID string
}

// StreamCancelled is emitted on user or error cancellation.
type StreamCancelled struct{}

// ToolRequested is emitted when a tool wants execution; approval pending unless
// it was auto-approved.
type ToolRequested struct {
	ToolCallID   string
	Name         string
	Input        map[string]interface{}
	AutoApproved bool
}

// ToolExecuting is emitted once a tool call has been approved and started.
type ToolExecuting struct {
	ToolCallID string
}

// ToolResult carries the outcome of a tool execution.
type ToolResult struct {
	ToolCallID string
	Name       string
	Content    string
	IsError    bool
}

// ToolRejected is emitted when the user rejected a tool call.
type ToolRejected struct {
	ToolCallID string
	Name       string
}

// ToolTimedOut is emitted when an approval prompt exceeded the configured timeout
// and was therefore treated as a rejection.
type ToolTimedOut struct {
	ToolCallID string
	Name       string
}

// AgentTaskUpdate reports parallel sub-agent progress.
type AgentTaskUpdate struct {
	TaskID  string
	Status  string
	Summary string
}

// UsageUpdate reports token/cost usage after a persisted message.
type UsageUpdate struct {
	InputTokens  int
	OutputTokens int
	CostUSD      float64
}

// PlanProposed is emitted when the agent presents a plan card in plan mode.
type PlanProposed struct {
	PlanID   string
	Steps    []PlanStep
	Files    []string
	Criteria string
}

// PlanDecided reports the outcome of a plan card.
type PlanDecided struct {
	PlanID   string
	Decision string // approved | rejected | edited
}

// ModeChanged reports a change of Plan/Execute/Auto mode (FR-012).
type ModeChanged struct {
	Mode string
}

// SteerQueued confirms a steer was injected into the current turn (FR-009).
type SteerQueued struct {
	Text string
}

// QueueRuns signals that a queued follow-up message started its turn (FR-010).
type QueueRuns struct {
	Text string
}

// GrantsChanged reports a session grant being created or revoked (FR-016/017).
type GrantsChanged struct {
	SessionID string
	Category  string
	Allow     bool
}

// PlanStep is a single step of a proposed plan (FR-013).
type PlanStep struct {
	Action   string
	File     string
	Criteria string
}

// ErrorEvent reports a provider/network/failure.
type ErrorEvent struct {
	Message     string
	Recoverable bool
}

// Idle signals the loop has no work; presentations may re-enable input.
type Idle struct{}

// SessionInfo is one row of the sessions panel (contract §3, SessionList).
type SessionInfo struct {
	ID        string
	Name      string
	UpdatedAt string
}

// SessionList is emitted by presentations after SessionList() (contract §2,
// session:list); it feeds the sessions panel of TUI and GUI.
type SessionList struct {
	Sessions []SessionInfo
}

// SessionLoaded is emitted by presentations after SessionOpen(id) (contract
// §2, session:loaded); it carries the full history of the resumed session.
type SessionLoaded struct {
	SessionID string
	Messages  []api.Message
}

// ConfigChanged is emitted by presentations after SaveConfig (contract §2,
// config:changed); consumers refresh model/theme in place without restarting.
type ConfigChanged struct {
	Config map[string]any
}

// SessionCleared is emitted by the engine after the `slash:{clear}` command
// (contract §2, slash:{...}); presentations reset their view of the session.
type SessionCleared struct {
	SessionID string
}

func (UserMessageAppended) isEvent() {}
func (StreamStart) isEvent()         {}
func (StreamDelta) isEvent()         {}
func (StreamDone) isEvent()          {}
func (StreamCancelled) isEvent()     {}
func (ToolRequested) isEvent()       {}
func (ToolExecuting) isEvent()       {}
func (ToolResult) isEvent()          {}
func (ToolRejected) isEvent()        {}
func (ToolTimedOut) isEvent()        {}
func (AgentTaskUpdate) isEvent()     {}
func (UsageUpdate) isEvent()         {}
func (PlanProposed) isEvent()        {}
func (PlanDecided) isEvent()         {}
func (ModeChanged) isEvent()         {}
func (SteerQueued) isEvent()         {}
func (QueueRuns) isEvent()           {}
func (GrantsChanged) isEvent()       {}
func (ErrorEvent) isEvent()          {}
func (Idle) isEvent()                {}
func (SessionList) isEvent()         {}
func (SessionLoaded) isEvent()       {}
func (ConfigChanged) isEvent()       {}
func (SessionCleared) isEvent()      {}
