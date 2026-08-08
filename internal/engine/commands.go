package engine

// Command is sent by presentations into the engine loop
// (contracts/engine-contract.md).
type Command interface {
	isCommand()
}

// SendMessage starts a user turn; the user message is persisted first.
type SendMessage struct {
	Text      string
	SessionID string
}

// SendTask starts an agent-mode turn (tools enabled, approval flow active).
type SendTask struct {
	Prompt    string
	SessionID string
}

// Cancel aborts the in-flight stream/tool cycle; the last complete message is kept.
type Cancel struct{}

// ApproveTool approves a specific pending tool call, which then executes.
type ApproveTool struct {
	ToolCallID string
}

// RejectTool rejects a specific pending tool call; a rejection is fed back to the
// model and the conversation continues.
type RejectTool struct {
	ToolCallID string
}

// SetAutoApprove updates per-category auto-approval (persisted via internal/config).
type SetAutoApprove struct {
	Category string
	Enabled  bool
}

// SwitchSession flushes current state and resumes the target session.
type SwitchSession struct {
	SessionID string
}

// Compact triggers context compaction.
type Compact struct{}

// Steer injects a user instruction into the in-flight turn (FR-009). Only valid
// between StreamStart and StreamDone/StreamCancelled; ignored in repose.
type Steer struct {
	Text      string
	SessionID string
}

// Queue persists a follow-up user message to run as the next turn when the loop
// hits Idle (FR-010). Valid at any time; executed FIFO.
type Queue struct {
	Text      string
	SessionID string
}

// SetPlanMode switches the approval mode (plan/execute/auto). In "plan" mode
// write-capable tool calls are held until ApprovePlan/RejectPlan.
type SetPlanMode struct {
	Mode string
}

// ApprovePlan approves a reviewed plan card and unblocks execution.
type ApprovePlan struct {
	PlanID string
}

// RejectPlan rejects a plan; the rejection is fed back to the model.
type RejectPlan struct {
	PlanID string
	Reason string
}

// EditPlan updates the instructions of a proposed plan.
type EditPlan struct {
	PlanID       string
	Instructions string
}

// GrantSession persists or revokes a per-{session,category} approval grant.
type GrantSession struct {
	Category  string
	SessionID string
	Allow     bool
}

// ForkSession clones a session up to a given message, creating an independent copy.
type ForkSession struct {
	SourceID  string
	MessageID int64
}

// Stop gracefully shuts down the engine loop.
type Stop struct{}

func (SendMessage) isCommand()    {}
func (SendTask) isCommand()       {}
func (Cancel) isCommand()         {}
func (ApproveTool) isCommand()    {}
func (RejectTool) isCommand()     {}
func (SetAutoApprove) isCommand() {}
func (SwitchSession) isCommand()  {}
func (Compact) isCommand()        {}
func (Steer) isCommand()          {}
func (Queue) isCommand()          {}
func (SetPlanMode) isCommand()    {}
func (ApprovePlan) isCommand()    {}
func (RejectPlan) isCommand()     {}
func (EditPlan) isCommand()       {}
func (GrantSession) isCommand()   {}
func (ForkSession) isCommand()    {}
func (Stop) isCommand()           {}
