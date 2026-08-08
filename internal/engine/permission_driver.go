package engine

import (
	"errors"
	"time"
)

// ApprovalTimeout is how long the engine waits for a driver decision before
// treating a pending tool approval as rejected.
var ApprovalTimeout = 5 * time.Minute

// ErrApprovalTimedOut is returned by the driver wrapper when the configured
// approval timeout elapses without a user decision.
var ErrApprovalTimedOut = errors.New("approval timed out")

// PermissionDriver is implemented by presentations (contracts/engine-contract.md).
type PermissionDriver interface {
	// Prompt is called when a tool call needs approval. It blocks until the user
	// decides. Implementations MUST NOT execute the tool; they only decide.
	Prompt(toolName string, input any) (allow bool, err error)
	// AutoApprove reports whether the given category is currently auto-approved
	// (file-edit, bash, web).
	AutoApprove(category string) bool
	// SessionGranted reports whether the given category has a per-session grant
	// for the session (FR-016). Implementations MUST still honor persistent
	// denies over session grants (deny > session-grant > auto > prompt).
	SessionGranted(sessionID, category string) bool
}

// AutoApproveCategories are the categories an approval decision can target.
var AutoApproveCategories = map[string]bool{
	"file-edit": true,
	"bash":      true,
	"web":       true,
}

// toolCategory maps a tool name to its approval category.
func toolCategory(toolName string) string {
	switch {
	case toolName == "bash" || toolName == "powershell":
		return "bash"
	case toolName == "edit" || toolName == "write_file" || toolName == "notebook_edit":
		return "file-edit"
	case toolName == "web_search" || toolName == "web_fetch" || toolName == "web_browser":
		return "web"
	default:
		return "other"
	}
}

// promptWithTimeout runs the driver's Prompt under the engine approval timeout.
func promptWithTimeout(driver PermissionDriver, toolName, input string) (bool, error) {
	type result struct {
		allow bool
		err   error
	}
	done := make(chan result, 1)
	go func() {
		allow, err := driver.Prompt(toolName, input)
		done <- result{allow: allow, err: err}
	}()
	select {
	case r := <-done:
		return r.allow, r.err
	case <-time.After(ApprovalTimeout):
		return false, ErrApprovalTimedOut
	}
}
