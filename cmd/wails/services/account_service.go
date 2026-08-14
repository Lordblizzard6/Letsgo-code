package services

import (
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/tools"
)

// AccountService feeds the rail avatar flyout (T052, gui-contract §1): status
// dot (online/busy), provider/model and today's usage summary, plus
// sign-out (keys cleared).
type AccountService struct {
	hub *Hub
}

func NewAccountService(hub *Hub) *AccountService { return &AccountService{hub: hub} }

// Status is the flyout payload.
type Status struct {
	Online   bool    `json:"online"`
	Busy     bool    `json:"busy"`
	Provider string  `json:"provider"`
	Model    string  `json:"model"`
	CostUSD  float64 `json:"cost_usd"`
	Requests int     `json:"requests"`
}

// GetStatus assembles the current account/status snapshot.
func (s *AccountService) GetStatus() Status {
	busy := s.hub.Engine != nil && s.hub.Engine.Busy()
	provider := config.DetectProviderFromModel(config.AppConfig.Model)
	model := config.AppConfig.Model
	if model == "" {
		model = "(sin configurar)"
	}
	stats := tools.GetCostTracker().GetUsageStats(startOfToday())
	cost, _ := stats["total_cost_usd"].(float64)
	requests, _ := stats["total_requests"].(int)
	return Status{
		Online:   true,
		Busy:     busy,
		Provider: provider,
		Model:    model,
		CostUSD:  cost,
		Requests: requests,
	}
}

// SignOut clears all API keys and persists (the "Cerrar sesión" action).
func (s *AccountService) SignOut() error {
	config.AppConfig.APIKey = ""
	config.AppConfig.AnthropicAPIKey = ""
	config.AppConfig.OpenAIAPIKey = ""
	config.AppConfig.GroqAPIKey = ""
	config.AppConfig.OpenRouterAPIKey = ""
	return config.SaveConfig()
}