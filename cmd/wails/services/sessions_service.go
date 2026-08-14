package services

import (
	"errors"

	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/engine"
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

// Create starts a new conversation and switches the engine to it. It emits
// `session:list` so every surface sees the new row.
func (s *SessionsService) Create(name, projectPath string) (db.Session, error) {
	id, err := db.CreateSession(name, projectPath)
	if err != nil {
		return db.Session{}, err
	}
	s.hub.Engine.Send(engine.SwitchSession{SessionID: id})
	s.emitList()
	return s.get(id)
}

// List returns all sessions, newest last (contract §3 SessionList).
func (s *SessionsService) List() ([]db.Session, error) {
	return db.ListSessions()
}

// Open resumes a session: marks it active in the db and switches the engine,
// then emits `session:loaded` with the full history (C-003).
func (s *SessionsService) Open(id string) ([]db.Message, error) {
	if err := db.ResumeSession(id); err != nil {
		return nil, err
	}
	s.hub.Engine.Send(engine.SwitchSession{SessionID: id})
	history, err := db.GetHistory(id)
	if err != nil {
		return nil, err
	}
	s.hub.emit("session:loaded", map[string]any{
		"session_id": id,
		"messages":   history,
	})
	return history, nil
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

// emitList refreshes the sessions panel in every surface.
func (s *SessionsService) emitList() {
	sessions, err := db.ListSessions()
	if err != nil {
		return
	}
	s.hub.emit("session:list", sessions)
}