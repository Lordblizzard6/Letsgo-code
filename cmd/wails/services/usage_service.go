package services

import (
	"time"

	"github.com/user/go-claude-code/internal/analytics"
	"github.com/user/go-claude-code/internal/tools"
)

// usageBudgetCap is the reference cap of the usage progress bar
// usage.go used a $25.00 cap so today's cost never pegs the bar).
const usageBudgetCap = 25.0

// UsageService feeds the rail Uso pane (T050, gui-contract §4): KPIs today
// (cost, requests, tokens), a per-model table and the budget bar.
type UsageService struct{}

func NewUsageService() *UsageService { return &UsageService{} }

// Today is the KPI + breakdown payload ("Sin actividad registrada hoy." when
// empty is a frontend concern).
type Today struct {
	CostUSD     float64                        `json:"cost_usd"`
	Requests    int                            `json:"requests"`
	Tokens      int                            `json:"tokens"`
	BudgetCap   float64                        `json:"budget_cap"`
	ByModel     map[string]map[string]any      `json:"by_model"`
}

// GetUsageToday aggregates the analytics session stats (contract §3
// GetUsageToday) plus the cost-tracker breakdown since midnight.
func (s *UsageService) GetUsageToday() Today {
	u := analytics.GetAnalytics().GetUsageToday()
	since := startOfToday()
	stats := tools.GetCostTracker().GetUsageStats(since)

	byModel := make(map[string]map[string]any)
	if raw, ok := stats["by_model"].(map[string]map[string]any); ok {
		byModel = raw
	}

	return Today{
		CostUSD:   u.CostUSD,
		Requests:  u.Requests,
		Tokens:    u.Tokens,
		BudgetCap: usageBudgetCap,
		ByModel:   byModel,
	}
}

func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}