package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Session represents a user session
type Session struct {
	ID           string         `json:"id"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      *time.Time     `json:"end_time,omitempty"`
	Duration     time.Duration  `json:"duration"`
	MessagesSent int            `json:"messages_sent"`
	TokensUsed   int            `json:"tokens_used"`
	Cost         float64        `json:"cost"`
	ToolsUsed    map[string]int `json:"tools_used"`
}

// Stats represents overall statistics
type Stats struct {
	TotalSessions int            `json:"total_sessions"`
	TotalDuration time.Duration  `json:"total_duration"`
	TotalMessages int            `json:"total_messages"`
	TotalTokens   int            `json:"total_tokens"`
	TotalCost     float64        `json:"total_cost"`
	ToolUsage     map[string]int `json:"tool_usage"`
	SessionsByDay map[string]int `json:"sessions_by_day"`
	LastUpdated   time.Time      `json:"last_updated"`
}

// Analytics manages session and usage analytics
type Analytics struct {
	currentSession *Session
	sessions       []Session
	stats          Stats
	dataPath       string
}

var (
	instance *Analytics
)

// GetAnalytics returns the analytics instance
func GetAnalytics() *Analytics {
	if instance == nil {
		instance = &Analytics{
			sessions: []Session{},
			stats: Stats{
				ToolUsage:     make(map[string]int),
				SessionsByDay: make(map[string]int),
			},
		}
		instance.dataPath = instance.getDataPath()
		instance.loadData()
	}
	return instance
}

func (a *Analytics) getDataPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".letsGo", "analytics.json")
}

// loadData loads analytics data from disk
func (a *Analytics) loadData() {
	data, err := os.ReadFile(a.dataPath)
	if err != nil {
		return // No data yet
	}

	var saved struct {
		Sessions []Session `json:"sessions"`
		Stats    Stats     `json:"stats"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		return
	}

	a.sessions = saved.Sessions
	a.stats = saved.Stats
}

// saveData saves analytics data to disk
func (a *Analytics) saveData() error {
	dir := filepath.Dir(a.dataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	saved := struct {
		Sessions []Session `json:"sessions"`
		Stats    Stats     `json:"stats"`
	}{
		Sessions: a.sessions,
		Stats:    a.stats,
	}

	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(a.dataPath, data, 0600)
}

// StartSession starts a new session
func (a *Analytics) StartSession() *Session {
	a.currentSession = &Session{
		ID:        generateSessionID(),
		StartTime: time.Now(),
		ToolsUsed: make(map[string]int),
	}
	return a.currentSession
}

// EndSession ends the current session
func (a *Analytics) EndSession() {
	if a.currentSession == nil {
		return
	}

	now := time.Now()
	a.currentSession.EndTime = &now
	a.currentSession.Duration = now.Sub(a.currentSession.StartTime)

	// Add to sessions list
	a.sessions = append(a.sessions, *a.currentSession)

	// Update stats
	a.stats.TotalSessions++
	a.stats.TotalDuration += a.currentSession.Duration
	a.stats.TotalMessages += a.currentSession.MessagesSent
	a.stats.TotalTokens += a.currentSession.TokensUsed
	a.stats.TotalCost += a.currentSession.Cost

	// Update tool usage
	for tool, count := range a.currentSession.ToolsUsed {
		a.stats.ToolUsage[tool] += count
	}

	// Update sessions by day
	day := a.currentSession.StartTime.Format("2006-01-02")
	a.stats.SessionsByDay[day]++

	a.stats.LastUpdated = now
	a.saveData()

	a.currentSession = nil
}

// RecordMessage records a message sent
func (a *Analytics) RecordMessage() {
	if a.currentSession != nil {
		a.currentSession.MessagesSent++
	}
}

// RecordTokens records token usage
func (a *Analytics) RecordTokens(input, output int, cost float64) {
	if a.currentSession != nil {
		a.currentSession.TokensUsed += input + output
		a.currentSession.Cost += cost
	}
}

// RecordToolUse records a tool usage
func (a *Analytics) RecordToolUse(toolName string) {
	if a.currentSession != nil {
		a.currentSession.ToolsUsed[toolName]++
	}
}

// GetStats returns overall statistics
func (a *Analytics) GetStats() Stats {
	return a.stats
}

// GetCurrentSession returns the current session
func (a *Analytics) GetCurrentSession() *Session {
	return a.currentSession
}

// GetSessions returns all sessions
func (a *Analytics) GetSessions() []Session {
	return a.sessions
}

// GetTopTools returns the most used tools
func (a *Analytics) GetTopTools(limit int) []ToolStat {
	type toolCount struct {
		Name  string
		Count int
	}

	var tools []toolCount
	for name, count := range a.stats.ToolUsage {
		tools = append(tools, toolCount{Name: name, Count: count})
	}

	// Sort by count descending
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Count > tools[j].Count
	})

	if limit > 0 && len(tools) > limit {
		tools = tools[:limit]
	}

	result := make([]ToolStat, len(tools))
	for i, t := range tools {
		result[i] = ToolStat{Name: t.Name, Count: t.Count}
	}

	return result
}

// GetDailyStats returns sessions grouped by day
func (a *Analytics) GetDailyStats() map[string]DailyStat {
	result := make(map[string]DailyStat)

	for _, session := range a.sessions {
		day := session.StartTime.Format("2006-01-02")
		stat := result[day]
		stat.Date = day
		stat.Sessions++
		stat.Messages += session.MessagesSent
		stat.Tokens += session.TokensUsed
		stat.Cost += session.Cost
		result[day] = stat
	}

	return result
}

// ToolStat represents tool usage statistics
type ToolStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// DailyStat represents daily statistics
type DailyStat struct {
	Date     string  `json:"date"`
	Sessions int     `json:"sessions"`
	Messages int     `json:"messages"`
	Tokens   int     `json:"tokens"`
	Cost     float64 `json:"cost"`
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
