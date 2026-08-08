package gui

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/engine"
	"github.com/user/go-claude-code/internal/tools"
)

// guiPermissionDriver implements engine.PermissionDriver with an INLINE
// approval card shown in the bottom strip above the composer (FR-001, SC-007).
// No modal dialog covers the transcript: the card renders as part of the
// layout and blocks until the user decides via Enter/Esc/Ctrl+Enter.
type guiPermissionDriver struct {
	win fyne.Window

	// timeout overrides the approval timeout; 0 uses engine.ApprovalTimeout.
	timeout time.Duration

	// autoApprove overrides AutoApprove results (used to honor engine state).
	autoApprove map[string]bool

	// onShow/onDismiss are wired by the Controller to mount/unmount cards
	// in the approval zone (approvalcard.go).
	onShow    func(card *approvalCard)
	onDismiss func()

	// sessionID returns the engine's current session; wired by the Controller.
	sessionID func() string

	// mu guards active/pending across the engine goroutine and UI thread.
	mu      sync.Mutex
	active  *approvalCard // current card (tests use this)
	pending []*approvalCard
}

// Prompt shows an inline approval card and blocks until the user decides:
// Enter=approve, Esc=reject, Ctrl+Enter=approve+grant-session. Timeout treats
// the call as rejected with a notice, per engine-contract.md. SC-007: never
// executes the tool.
func (d *guiPermissionDriver) Prompt(toolName string, input any) (bool, error) {
	type decision struct {
		allow bool
		err   error
	}
	ch := make(chan decision, 1)
	var once sync.Once
	decide := func(allow bool, err error) {
		once.Do(func() { ch <- decision{allow: allow, err: err} })
	}

	dismiss := func(card *approvalCard) {
		d.mu.Lock()
		for i, c := range d.pending {
			if c == card {
				d.pending = append(d.pending[:i], d.pending[i+1:]...)
				break
			}
		}
		if d.active == card {
			d.active = nil
		}
		d.mu.Unlock()
		if d.onDismiss != nil {
			d.onDismiss()
		}
	}

	fyne.Do(func() {
		card := newApprovalCard(toolName, formatInput(input))
		d.mu.Lock()
		d.active = card
		d.pending = append(d.pending, card)
		d.mu.Unlock()

		card.onEnter = func() {
			decide(true, nil)
			dismiss(card)
		}
		card.onEsc = func() {
			decide(false, nil)
			dismiss(card)
		}
		card.onGrant = func() {
			// FR-016: grant for this session, then approve.
			if sid := d.currentSessionID(); sid != "" {
				_ = db.GrantCategory(sid, categoryOf(toolName))
			}
			decide(true, nil)
			dismiss(card)
		}
		if d.onShow != nil {
			d.onShow(card)
		}
		card.requestFocus(d.win)
	})

	timeout := d.timeout
	if timeout <= 0 {
		timeout = engine.ApprovalTimeout
	}
	select {
	case d := <-ch:
		return d.allow, d.err
	case <-time.After(timeout):
		return false, engine.ErrApprovalTimedOut
	}
}

// rejectPending rejects every pending approval (FR-022: window close).
func (d *guiPermissionDriver) rejectPending() {
	d.mu.Lock()
	cards := append([]*approvalCard(nil), d.pending...)
	d.pending = nil
	d.active = nil
	d.mu.Unlock()
	for _, c := range cards {
		if c.onEsc != nil {
			c.onEsc()
		}
	}
	if d.onDismiss != nil {
		d.onDismiss()
	}
}

// pendingCount reports how many approval cards are waiting (FR-021).
func (d *guiPermissionDriver) pendingCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.pending)
}

// sessionID returns the engine's current session; wired by the Controller.
func (d *guiPermissionDriver) currentSessionID() string {
	if d.sessionID != nil {
		return d.sessionID()
	}
	return ""
}

// AutoApprove reports whether the category is auto-approved.
func (d *guiPermissionDriver) AutoApprove(category string) bool {
	if d.autoApprove != nil {
		return d.autoApprove[category]
	}
	return autoApproveFromConfig(category)
}

// SessionGranted reports whether the category has a per-session grant (FR-016).
// Persistent denies always win over grants; this only reports the grant AND is
// gated by global opt-out so a hard deny still overrides a session grant
// (deny > session-grant > auto > prompt, engine-contract.md §3).
func (d *guiPermissionDriver) SessionGranted(sessionID, category string) bool {
	if sessionID == "" {
		return false
	}
	if categoryDeniedGlobally(category) {
		return false
	}
	ok, err := db.IsGranted(sessionID, category)
	return err == nil && ok
}

// categoryDeniedGlobally reports whether any managed tool in the category is
// hard-denied (opt-out) in the global permission config. A grant cannot
// resurrect a hard deny (FR-017/security).
func categoryDeniedGlobally(category string) bool {
	for _, tool := range categoryToolNames(category) {
		tp := tools.GetAdvancedPermissionManager().GetToolConfig(tool)
		if tp != nil && tp.Mode == tools.PermissionModeOptOut {
			return true
		}
	}
	return false
}

// categoryToolNames maps the approval category back to the tool names managed
// by the advanced permission manager.
func categoryToolNames(category string) []string {
	switch category {
	case "bash":
		return []string{"bash"}
	case "file-edit":
		return []string{"edit", "write_file"}
	case "web":
		return []string{"web_fetch", "web_search"}
	default:
		return nil
	}
}

// setAutoApprove updates the driver's view of persisted auto-approval state.
func (d *guiPermissionDriver) setAutoApprove(category string, enabled bool) {
	if d.autoApprove == nil {
		d.autoApprove = map[string]bool{}
	}
	d.autoApprove[category] = enabled
}

func autoApproveFromConfig(category string) bool {
	switch category {
	case "bash", "file-edit", "web":
		return config.AppConfig.AutoApprove[category]
	default:
		return false
	}
}

func categoryOf(toolName string) string {
	switch toolName {
	case "edit_file", "write_file", "create_file":
		return "file-edit"
	case "bash", "shell", "run_command":
		return "bash"
	case "web_fetch", "http_get":
		return "web"
	default:
		return "unknown"
	}
}

func formatInput(input any) string {
	b, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", input)
	}
	return string(b)
}
