package services

import (
	"time"

	"github.com/google/uuid"

	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/engine"
)

// toolbarDriver implements engine.PermissionDriver for the Wails GUI
// (frontend-contract §2 tool:* events). Tool requests are forwarded to the
// frontend as `tool:request`; ChatService.Approve answers them.
type toolbarDriver struct {
	hub *Hub
}

// NewToolbarDriver builds the permission driver bound to the hub's engine.
// The GUI auto-approves categories enabled in config (config:auto_approve).
func NewToolbarDriver(hub *Hub) *toolbarDriver { return &toolbarDriver{hub: hub} }

func (d *toolbarDriver) Prompt(toolName string, input any) (bool, error) {
	callID := uuid.New().String()
	ch := make(chan bool, 1)
	d.hub.registerApproval(callID, ch)
	d.hub.emit("tool:request", map[string]any{
		"call_id": callID,
		"name":    toolName,
		"input":   input,
	})
	select {
	case allow := <-ch:
		return allow, nil
	case <-time.After(engine.ApprovalTimeout):
		d.hub.dropApproval(callID)
		return false, engine.ErrApprovalTimedOut
	}
}

func (d *toolbarDriver) AutoApprove(category string) bool {
	return config.AppConfig.AutoApprove[category]
}

func (d *toolbarDriver) SessionGranted(sessionID, category string) bool {
	allowed, err := db.IsGranted(sessionID, category)
	return err == nil && allowed
}