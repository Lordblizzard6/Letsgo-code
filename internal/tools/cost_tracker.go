package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CostEntry tracks API usage costs
type CostEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	CostUSD      float64   `json:"cost_usd"`
	RequestID    string    `json:"request_id"`
}

// CostTracker manages cost tracking
type CostTracker struct {
	Entries []CostEntry `json:"entries"`
	mu      sync.RWMutex
	path    string
}

var (
	costInstance *CostTracker
	costOnce     sync.Once
)

// Cost per 1K tokens (approximate, updated periodically)
var tokenPrices = map[string]struct {
	Input  float64
	Output float64
}{
	"claude-3-5-sonnet-20240620": {Input: 0.003, Output: 0.015},
	"claude-3-opus-20240229":     {Input: 0.015, Output: 0.075},
	"claude-3-haiku-20240307":    {Input: 0.00025, Output: 0.00125},
	"gpt-4o":                     {Input: 0.005, Output: 0.015},
	"gpt-4-turbo":                {Input: 0.01, Output: 0.03},
	"llama-3.3-70b-versatile":    {Input: 0.00059, Output: 0.00079},
	"llama-3.1-8b-instant":       {Input: 0.0001, Output: 0.0002},
}

func GetCostTracker() *CostTracker {
	costOnce.Do(func() {
		home, _ := os.UserHomeDir()
		path := filepath.Join(home, ".letsGo", "costs.json")

		costInstance = &CostTracker{
			Entries: []CostEntry{},
			path:    path,
		}

		costInstance.load()
	})

	return costInstance
}

func (t *CostTracker) load() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, err := os.Stat(t.path); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(t.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &t.Entries)
}

func (t *CostTracker) save() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(t.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(t.Entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(t.path, data, 0644)
}

// RecordUsage records API usage
func (t *CostTracker) RecordUsage(provider, model string, inputTokens, outputTokens int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	cost := t.calculateCost(model, inputTokens, outputTokens)

	entry := CostEntry{
		Timestamp:    time.Now(),
		Provider:     provider,
		Model:        model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		CostUSD:      cost,
		RequestID:    generateRequestID(),
	}

	t.Entries = append(t.Entries, entry)
	return t.save()
}

func (t *CostTracker) calculateCost(model string, inputTokens, outputTokens int) float64 {
	prices, exists := tokenPrices[model]
	if !exists {
		// Default to claude-3.5-sonnet pricing if unknown
		prices = tokenPrices["claude-3-5-sonnet-20240620"]
	}

	inputCost := float64(inputTokens) / 1000 * prices.Input
	outputCost := float64(outputTokens) / 1000 * prices.Output

	return inputCost + outputCost
}

// GetTotalCost returns total cost for a time period
func (t *CostTracker) GetTotalCost(since time.Time) float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	total := 0.0
	for _, entry := range t.Entries {
		if entry.Timestamp.After(since) {
			total += entry.CostUSD
		}
	}

	return total
}

// GetUsageStats returns usage statistics
func (t *CostTracker) GetUsageStats(since time.Time) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := map[string]interface{}{
		"total_requests":      0,
		"total_input_tokens":  0,
		"total_output_tokens": 0,
		"total_tokens":        0,
		"total_cost_usd":      0.0,
		"by_model":            make(map[string]map[string]interface{}),
	}

	modelStats := make(map[string]map[string]interface{})

	for _, entry := range t.Entries {
		if !entry.Timestamp.After(since) {
			continue
		}

		stats["total_requests"] = stats["total_requests"].(int) + 1
		stats["total_input_tokens"] = stats["total_input_tokens"].(int) + entry.InputTokens
		stats["total_output_tokens"] = stats["total_output_tokens"].(int) + entry.OutputTokens
		stats["total_cost_usd"] = stats["total_cost_usd"].(float64) + entry.CostUSD

		if _, exists := modelStats[entry.Model]; !exists {
			modelStats[entry.Model] = map[string]interface{}{
				"requests":      0,
				"input_tokens":  0,
				"output_tokens": 0,
				"cost_usd":      0.0,
			}
		}

		m := modelStats[entry.Model]
		m["requests"] = m["requests"].(int) + 1
		m["input_tokens"] = m["input_tokens"].(int) + entry.InputTokens
		m["output_tokens"] = m["output_tokens"].(int) + entry.OutputTokens
		m["cost_usd"] = m["cost_usd"].(float64) + entry.CostUSD
	}

	stats["total_tokens"] = stats["total_input_tokens"].(int) + stats["total_output_tokens"].(int)
	stats["by_model"] = modelStats

	return stats
}

// GetSessionCost returns cost for current session
func (t *CostTracker) GetSessionCost() float64 {
	// In a full implementation, would track session start time
	// For now, return last 1 hour
	since := time.Now().Add(-1 * time.Hour)
	return t.GetTotalCost(since)
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// FormatCost formats cost as a readable string
func FormatCost(cost float64) string {
	if cost < 0.01 {
		return fmt.Sprintf("$%.4f", cost)
	}
	return fmt.Sprintf("$%.2f", cost)
}

// GetStartOfDay returns the start of today
func GetStartOfDay() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// GetStartOfWeek returns the start of the current week (Sunday)
func GetStartOfWeek() time.Time {
	now := time.Now()
	weekday := int(now.Weekday())
	return now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
}
