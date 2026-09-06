package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/engine"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// SessionsService implements the session surface of the core contract §3
// (T044): create/list/open sessions, shared with the TUI (C-003), and
// paginated transcript access (GetMessages).
type SessionsService struct {
	hub *Hub
}

func NewSessionsService(hub *Hub) *SessionsService {
	return &SessionsService{hub: hub}
}

// GetCurrentProject returns the active project path or empty string.
func (s *SessionsService) GetCurrentProject() string {
	if s.hub != nil {
		return s.hub.GetCurrentProject()
	}
	return ""
}

// SetCurrentProject sets the active project path and changes the working directory.
func (s *SessionsService) SetCurrentProject(path string) error {
	if s.hub != nil {
		return s.hub.SetCurrentProject(path)
	}
	if path != "" {
		return os.Chdir(path)
	}
	return nil
}

// Create starts a new conversation and switches the engine to it. It emits
// `session:list` so every surface sees the new row.
func (s *SessionsService) Create(name, projectPath string) (db.Session, error) {
	if projectPath != "" {
		_ = s.SetCurrentProject(projectPath)
	}
	if name == "" || name == "Sesión" || name == "Nueva sesión" {
		dateStr := time.Now().Format("2006-01-02")
		base := "Sesión"
		if projectPath != "" {
			base = filepath.Base(projectPath)
		}
		name = fmt.Sprintf("[%s] %s", dateStr, base)
	}
	id, err := db.CreateSession(name, projectPath)
	if err != nil {
		return db.Session{}, err
	}
	if s.hub != nil && s.hub.Engine != nil {
		s.hub.Engine.SwitchSession(id)
	}
	s.emitList()
	return s.get(id)
}

// List returns all sessions, newest last (contract §3 SessionList).
func (s *SessionsService) List() ([]db.Session, error) {
	return db.ListSessions()
}

// ListProjects returns distinct project paths registered across sessions.
func (s *SessionsService) ListProjects() ([]string, error) {
	return db.ListProjectPaths()
}

// Open resumes a session: marks it active in the db, synchronizes working dir to the session project,
// and switches the engine, then emits `session:loaded` with the full history (C-003).
func (s *SessionsService) Open(id string) ([]db.Message, error) {
	if err := db.ResumeSession(id); err != nil {
		return nil, err
	}
	if s.hub != nil && s.hub.Engine != nil {
		s.hub.Engine.SwitchSession(id)
	}
	sess, _ := s.get(id)
	if sess.ProjectPath != "" {
		_ = s.SetCurrentProject(sess.ProjectPath)
	}
	history, err := db.GetHistory(id)
	if err != nil {
		return nil, err
	}
	s.hub.emit("session:loaded", map[string]any{
		"session_id":   id,
		"messages":     history,
		"project_path": sess.ProjectPath,
	})
	return history, nil
}

// Rollback removes messages from targetMessageID onwards, halts any active stream,
// reloads remaining history into the engine, and emits session:loaded with the updated transcript.
// Returns the restored user prompt text.
func (s *SessionsService) Rollback(id string, targetMessageID int64) (string, error) {
	if s.hub != nil && s.hub.Engine != nil {
		s.hub.Engine.Send(engine.Cancel{})
	}
	restoredText, err := db.RollbackSession(id, targetMessageID)
	if err != nil {
		return "", err
	}
	if s.hub != nil && s.hub.Engine != nil {
		s.hub.Engine.SwitchSession(id)
	}
	history, err := db.GetHistory(id)
	if err == nil && s.hub != nil {
		s.hub.emit("session:loaded", map[string]any{
			"session_id": id,
			"messages":   history,
		})
	}
	s.emitList()
	return restoredText, nil
}

// GetMessages returns a paginated transcript slice (contract §3 GetMessages).
func (s *SessionsService) GetMessages(id string, limit, offset int) ([]db.Message, error) {
	return db.GetMessages(id, limit, offset)
}

func (s *SessionsService) get(id string) (db.Session, error) {
	sessions, err := db.ListSessions()
	if err != nil {
		return db.Session{}, err
	}
	for _, sess := range sessions {
		if sess.ID == id {
			return sess, nil
		}
	}
	return db.Session{}, errSessionNotFound
}

// errSessionNotFound guards Session rows that vanished between List and get.
var errSessionNotFound = errors.New("session not found")

// Rename updates the name of an existing session.
func (s *SessionsService) Rename(id string, newName string) error {
	if err := db.RenameSession(id, newName); err != nil {
		return err
	}
	s.emitList()
	return nil
}

// Delete removes an existing session and all its messages.
func (s *SessionsService) Delete(id string) error {
	if err := db.DeleteSession(id); err != nil {
		return err
	}
	s.emitList()
	return nil
}

// SelectProjectFolder prompts a native directory picker dialog, updates working dir and session.
func (s *SessionsService) SelectProjectFolder() (string, error) {
	app := application.Get()
	if app == nil {
		return "", errors.New("application not initialized")
	}
	dialog := app.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		CanChooseDirectories: true,
		CanChooseFiles:       false,
		Title:                "Seleccionar Carpeta de Proyecto",
	})
	dir, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if dir != "" {
		_ = s.SetCurrentProject(dir)
		active, _ := db.GetActiveSession()
		if active != nil {
			_ = db.SetSessionProjectPath(active.ID, dir)
			s.emitList()
		}
	}
	return dir, nil
}

// SetSessionProject associates a session with a project folder.
func (s *SessionsService) SetSessionProject(id, projectPath string) error {
	if id != "" {
		if err := db.SetSessionProjectPath(id, projectPath); err != nil {
			return err
		}
	}
	if projectPath != "" {
		_ = s.SetCurrentProject(projectPath)
	}
	s.emitList()
	return nil
}

// emitList refreshes the sessions panel in every surface.
func (s *SessionsService) emitList() {
	sessions, err := db.ListSessions()
	if err != nil {
		return
	}
	s.hub.emit("session:list", sessions)
}