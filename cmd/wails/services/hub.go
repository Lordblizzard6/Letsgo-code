package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/user/go-claude-code/internal/engine"
)

// Emitter delivers a named event to the Wails frontend (research.md D3,
// frontend-contract §2). Production wiring uses the Wails Event bus
// (application.Get().Event.Emit); headless service tests inject a fake.
type Emitter func(name string, data any)

// Hub is the shared state of the bound Wails services (005, T043-T052).
// It owns the single engine instance and the event emitter; presentation
// tests may construct a Hub without a Wails application and override Emit.
type Hub struct {
	Engine *engine.Engine
	Emit   Emitter

	mu               sync.Mutex
	pendingApprovals map[string]chan bool
	currentProject   string
}

// NewHub creates a hub with a no-op emitter; main.go replaces Emit with the
// Wails Event bus after application.New (T053).
func NewHub() *Hub {
	return &Hub{
		Emit:             func(name string, data any) {},
		pendingApprovals: make(map[string]chan bool),
	}
}

// SetEngine attaches the engine. The driver is created before the engine (the
// engine needs the driver), so the hub starts empty and is wired afterwards.
func (h *Hub) SetEngine(e *engine.Engine) { h.Engine = e }

// registerApproval stores the channel a tool approval awaits (gui_driver).
func (h *Hub) registerApproval(callID string, ch chan bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pendingApprovals[callID] = ch
}

// resolveApproval answers a pending tool approval by call ID (T043 approve).
func (h *Hub) resolveApproval(callID string, allow bool) bool {
	h.mu.Lock()
	ch, ok := h.pendingApprovals[callID]
	delete(h.pendingApprovals, callID)
	h.mu.Unlock()
	if !ok {
		return false
	}
	ch <- allow
	return true
}

// dropApproval removes a timed-out approval without deciding it (gui_driver).
func (h *Hub) dropApproval(callID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.pendingApprovals, callID)
}

// emit is the thread-safe emit shorthand used by services.
func (h *Hub) emit(name string, data any) {
	if h.Emit == nil {
		return
	}
	h.Emit(name, data)
}

// SetCurrentProject changes the working directory and tracks the active project path.
func (h *Hub) SetCurrentProject(path string) error {
	var finalPath string
	if path != "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		info, err := os.Stat(abs)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("not a directory: %s", abs)
		}
		if err := os.Chdir(abs); err != nil {
			return err
		}
		finalPath = abs
	}

	h.mu.Lock()
	h.currentProject = finalPath
	h.mu.Unlock()

	name := ""
	if finalPath != "" {
		name = filepath.Base(finalPath)
	}
	h.emit("project:changed", map[string]any{
		"project_path": finalPath,
		"name":         name,
	})
	return nil
}

// GetCurrentProject returns the active project path, or "" if none.
func (h *Hub) GetCurrentProject() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.currentProject
}