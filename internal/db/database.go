package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/user/go-claude-code/internal/api"
	_ "modernc.org/sqlite"
)

var DB *sql.DB
var dbMu sync.RWMutex

func InitDB() error {
	dbPath := os.Getenv("TEST_DB_PATH")
	if dbPath == "" {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, ".letsGo", "history.db")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}

	if DB != nil {
		_ = DB.Close()
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			name TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			project_path TEXT,
			is_active BOOLEAN DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT,
			role TEXT,
			content TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			tokens_input INTEGER DEFAULT 0,
			tokens_output INTEGER DEFAULT 0,
			cost_usd REAL DEFAULT 0,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		);`,
		`CREATE TABLE IF NOT EXISTS context_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT,
			file_path TEXT,
			content TEXT,
			added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		);`,
		`CREATE TABLE IF NOT EXISTS session_memory (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE,
			value TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS session_grants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			category TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(session_id, category)
		);`,
		`CREATE TABLE IF NOT EXISTS compact_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT,
			original_messages INTEGER,
			compacted_messages INTEGER,
			summary TEXT,
			compacted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session_timestamp ON messages(session_id, timestamp DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_context_files_session ON context_files(session_id);`,
		`CREATE INDEX IF NOT EXISTS idx_session_grants_session ON session_grants(session_id);`,
		`CREATE INDEX IF NOT EXISTS idx_session_grants_cat ON session_grants(category);`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	// Optional migration: kind column on messages (user:steer/queued/plan_instructions).
	// Safe to re-run; fails silently when the column already exists.
	_, _ = DB.Exec(`ALTER TABLE messages ADD COLUMN kind TEXT NOT NULL DEFAULT 'user'`)

	return nil
}

func SaveMessage(sessionID, role string, content interface{}) error {
	dbMu.Lock()
	defer dbMu.Unlock()
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}
	_, err = DB.Exec("INSERT INTO messages (session_id, role, content) VALUES (?, ?, ?)",
		sessionID, role, string(contentJSON))
	return err
}

func GetHistory(sessionID string) ([]Message, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()
	rows, err := DB.Query("SELECT id, role, content, timestamp, kind FROM messages WHERE session_id = ? ORDER BY timestamp ASC, id ASC", sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var id int64
		var role, contentJSON, kind string
		var timestamp time.Time
		if err := rows.Scan(&id, &role, &contentJSON, &timestamp, &kind); err != nil {
			return nil, err
		}

		var content interface{}
		err = json.Unmarshal([]byte(contentJSON), &content)
		if err != nil {
			content = contentJSON
		}
		messages = append(messages, Message{ID: id, Role: role, Content: content, Timestamp: timestamp, Kind: kind})
	}
	return messages, nil
}

// GetMessages returns a paginated slice of a session's history, oldest first
// (frontend-contract.md §3 `GetMessages(sessionId, limit?, offset?)`; used for
// long transcripts, SC-009). limit<=0 means no limit; offset is the number of
// rows to skip.
func GetMessages(sessionID string, limit, offset int) ([]Message, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()
	q := "SELECT id, role, content, timestamp, kind FROM messages WHERE session_id = ? ORDER BY timestamp ASC, id ASC"
	var args []interface{}
	args = append(args, sessionID)
	if limit > 0 {
		q += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		q += " OFFSET ?"
		args = append(args, offset)
	}
	rows, err := DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var id int64
		var role, contentJSON, kind string
		var timestamp time.Time
		if err := rows.Scan(&id, &role, &contentJSON, &timestamp, &kind); err != nil {
			return nil, err
		}

		var content interface{}
		err = json.Unmarshal([]byte(contentJSON), &content)
		if err != nil {
			content = contentJSON
		}
		messages = append(messages, Message{ID: id, Role: role, Content: content, Timestamp: timestamp, Kind: kind})
	}
	return messages, nil
}

// PruneContext keeps only the last N messages to avoid token overflow
func PruneContext(messages []api.Message, maxMessages int) []api.Message {
	if len(messages) <= maxMessages {
		return messages
	}
	return messages[len(messages)-maxMessages:]
}

// Session Management
func CreateSession(name, projectPath string) (string, error) {
	dbMu.Lock()
	defer dbMu.Unlock()
	sessionID := fmt.Sprintf("sess_%d", time.Now().UnixNano())
	_, err := DB.Exec(
		"INSERT INTO sessions (id, name, project_path, is_active) VALUES (?, ?, ?, ?)",
		sessionID, name, projectPath, true,
	)
	return sessionID, err
}

func GetActiveSession() (*Session, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()
	row := DB.QueryRow(
		"SELECT id, name, created_at, project_path FROM sessions WHERE is_active = 1 ORDER BY updated_at DESC LIMIT 1",
	)
	var s Session
	err := row.Scan(&s.ID, &s.Name, &s.CreatedAt, &s.ProjectPath)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func ListSessions() ([]Session, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()
	rows, err := DB.Query("SELECT id, name, created_at, updated_at, project_path FROM sessions ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt, &s.UpdatedAt, &s.ProjectPath); err != nil {
			continue
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func ResumeSession(sessionID string) error {
	_, err := DB.Exec("UPDATE sessions SET is_active = 0 WHERE is_active = 1")
	if err != nil {
		return err
	}
	_, err = DB.Exec("UPDATE sessions SET is_active = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?", sessionID)
	return err
}

func UpdateSessionTimestamp(sessionID string) error {
	_, err := DB.Exec("UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", sessionID)
	return err
}

// Context Files Management
func AddContextFile(sessionID, filePath, content string) error {
	_, err := DB.Exec(
		"INSERT INTO context_files (session_id, file_path, content) VALUES (?, ?, ?)",
		sessionID, filePath, content,
	)
	return err
}

func RemoveContextFile(sessionID, filePath string) error {
	_, err := DB.Exec(
		"DELETE FROM context_files WHERE session_id = ? AND file_path = ?",
		sessionID, filePath,
	)
	return err
}

func GetContextFiles(sessionID string) ([]ContextFile, error) {
	rows, err := DB.Query(
		"SELECT file_path, content, added_at FROM context_files WHERE session_id = ? ORDER BY added_at DESC",
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []ContextFile
	for rows.Next() {
		var f ContextFile
		if err := rows.Scan(&f.Path, &f.Content, &f.AddedAt); err != nil {
			continue
		}
		files = append(files, f)
	}
	return files, nil
}

// Message represents a message with database ID
type Message struct {
	ID        int64       `json:"id"`
	Role      string      `json:"role"`
	Content   interface{} `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
	Kind      string      `json:"kind,omitempty"`
}

type SearchMessageResult struct {
	ID        int64     `json:"id"`
	Role      string    `json:"role"`
	Fragment  string    `json:"fragment"`
	Timestamp time.Time `json:"timestamp"`
}

func SearchMessages(sessionID, query string, limit, offset int) ([]SearchMessageResult, int, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	likeQuery := "%" + query + "%"
	var total int
	if err := DB.QueryRow(
		`SELECT COUNT(*) FROM messages WHERE session_id = ? AND LOWER(content) LIKE LOWER(?)`,
		sessionID, likeQuery,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(
		`SELECT id, role, content, timestamp
		 FROM messages
		 WHERE session_id = ? AND LOWER(content) LIKE LOWER(?)
		 ORDER BY timestamp DESC
		 LIMIT ? OFFSET ?`,
		sessionID, likeQuery, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := make([]SearchMessageResult, 0, limit)
	for rows.Next() {
		var (
			id          int64
			role        string
			contentJSON string
			timestamp   time.Time
		)
		if err := rows.Scan(&id, &role, &contentJSON, &timestamp); err != nil {
			return nil, 0, err
		}
		results = append(results, SearchMessageResult{
			ID:        id,
			Role:      role,
			Fragment:  buildSnippet(contentJSON, query, 120),
			Timestamp: timestamp,
		})
	}
	return results, total, nil
}

func buildSnippet(content, query string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 120
	}
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)
	idx := strings.Index(lowerContent, lowerQuery)

	if idx < 0 {
		if len(content) <= maxLen {
			return content
		}
		return content[:maxLen] + "…"
	}

	start := idx - 40
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(content) {
		end = len(content)
	}
	snippet := content[start:end]
	if start > 0 {
		snippet = "…" + snippet
	}
	if end < len(content) {
		snippet += "…"
	}
	return snippet
}

// Session Memory
func SetMemory(key, value string) error {
	_, err := DB.Exec(
		"INSERT INTO session_memory (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = CURRENT_TIMESTAMP",
		key, value, value,
	)
	return err
}

func GetMemory(key string) (string, error) {
	var value string
	err := DB.QueryRow("SELECT value FROM session_memory WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func ListMemory() (map[string]string, error) {
	rows, err := DB.Query("SELECT key, value FROM session_memory ORDER BY updated_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to list memory: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[key] = value
	}
	return result, nil
}

func DeleteMemory(key string) error {
	_, err := DB.Exec("DELETE FROM session_memory WHERE key = ?", key)
	return err
}

func SaveCompactHistory(sessionID string, original, compacted int, summary string) error {
	_, err := DB.Exec(
		"INSERT INTO compact_history (session_id, original_messages, compacted_messages, summary) VALUES (?, ?, ?, ?)",
		sessionID, original, compacted, summary,
	)
	return err
}

func GetLastCompactSummary(sessionID string) (string, error) {
	var summary string
	err := DB.QueryRow(
		"SELECT summary FROM compact_history WHERE session_id = ? ORDER BY compacted_at DESC LIMIT 1",
		sessionID,
	).Scan(&summary)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return summary, err
}

// Data structures
type Session struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ProjectPath string    `json:"project_path"`
	IsActive    bool      `json:"is_active"`
}

type ContextFile struct {
	Path    string    `json:"path"`
	Content string    `json:"content"`
	AddedAt time.Time `json:"added_at"`
}

// RenameSession updates the name of a session
func RenameSession(sessionID, newName string) error {
	_, err := DB.Exec("UPDATE sessions SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", newName, sessionID)
	return err
}

// DeleteSession removes a session and all its data
func DeleteSession(sessionID string) error {
	// Delete messages first
	_, err := DB.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	if err != nil {
		return err
	}
	// Delete context files
	_, err = DB.Exec("DELETE FROM context_files WHERE session_id = ?", sessionID)
	if err != nil {
		return err
	}
	// Delete compact history
	_, err = DB.Exec("DELETE FROM compact_history WHERE session_id = ?", sessionID)
	if err != nil {
		return err
	}
	// Delete session
	_, err = DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// DeleteMessage removes a specific message by its database ID
func DeleteMessage(messageID int64) error {
	_, err := DB.Exec("DELETE FROM messages WHERE id = ?", messageID)
	return err
}

// ClearSessionHistory removes all messages from a session
func ClearSessionHistory(sessionID string) error {
	_, err := DB.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	return err
}

// SetActiveSession is an alias for ResumeSession for API compatibility
func SetActiveSession(sessionID string) error {
	return ResumeSession(sessionID)
}

// SessionGrant is a per-session approval rule (category → allowed) (FR-016/017).
type SessionGrant struct {
	SessionID string
	Category  string
	CreatedAt time.Time
}

// GrantCategory persists an approval grant for (sessionID, category) (upsert).
func GrantCategory(sessionID, category string) error {
	dbMu.Lock()
	defer dbMu.Unlock()
	_, err := DB.Exec(
		`INSERT INTO session_grants (session_id, category) VALUES (?, ?)
		 ON CONFLICT(session_id, category) DO UPDATE SET created_at = CURRENT_TIMESTAMP`,
		sessionID, category)
	return err
}

// RevokeCategory removes an approval grant for (sessionID, category).
func RevokeCategory(sessionID, category string) error {
	dbMu.Lock()
	defer dbMu.Unlock()
	_, err := DB.Exec(`DELETE FROM session_grants WHERE session_id = ? AND category = ?`, sessionID, category)
	return err
}

// ListGrants returns all grants for a session (FR-017).
func ListGrants(sessionID string) ([]SessionGrant, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()
	rows, err := DB.Query(
		`SELECT session_id, category, created_at FROM session_grants WHERE session_id = ? ORDER BY created_at DESC`,
		sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []SessionGrant
	for rows.Next() {
		var g SessionGrant
		if err := rows.Scan(&g.SessionID, &g.Category, &g.CreatedAt); err != nil {
			continue
		}
		grants = append(grants, g)
	}
	return grants, nil
}

// IsGranted reports whether (sessionID, category) has an active grant.
func IsGranted(sessionID, category string) (bool, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()
	var n int
	err := DB.QueryRow(
		`SELECT COUNT(*) FROM session_grants WHERE session_id = ? AND category = ?`,
		sessionID, category).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// SaveMessageKind persists a message with an explicit kind (user:steer/queued/plan_instructions).
func SaveMessageKind(sessionID, role, content string, kind string) error {
	dbMu.Lock()
	defer dbMu.Unlock()
	_, err := DB.Exec(
		`INSERT INTO messages (session_id, role, content, kind) VALUES (?, ?, ?, ?)`,
		sessionID, role, content, kind)
	return err
}

// ForkSession creates an independent copy of a session containing all messages
// up to and including messageID. The original session is never modified; grants
// and memory are NOT copied (data-model.md §3).
func ForkSession(sourceID string, messageID int64) (string, error) {
	dbMu.Lock()
	defer dbMu.Unlock()

	var name string
	err := DB.QueryRow(`SELECT name FROM sessions WHERE id = ?`, sourceID).Scan(&name)
	if err != nil {
		return "", err
	}

	newID := fmt.Sprintf("sess_%d", time.Now().UnixNano())
	if _, err := DB.Exec(
		`INSERT INTO sessions (id, name, created_at, updated_at, project_path, is_active)
		 VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, (SELECT project_path FROM sessions WHERE id = ?), 1)`,
		newID, "Fork de "+name, sourceID); err != nil {
		return "", err
	}

	_, err = DB.Exec(
		`INSERT INTO messages (session_id, role, content, timestamp, tokens_input, tokens_output, cost_usd, kind)
		 SELECT ?, role, content, timestamp, tokens_input, tokens_output, cost_usd, kind
		 FROM messages
		 WHERE session_id = ? AND id <= ?
		 ORDER BY id ASC`,
		newID, sourceID, messageID)
	if err != nil {
		return "", err
	}

	return newID, nil
}
