package services

import (
	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/config"
)

// SettingsService exposes GetConfig/SaveConfig of the core contract §3
// (T049). It is the Wails binding of the config surface already consumed by
// the TUI; SaveConfig persists and emits `config:changed` so every surface
// refreshes without restarting (C-004).
type SettingsService struct {
	hub *Hub
}

func NewSettingsService(hub *Hub) *SettingsService {
	return &SettingsService{hub: hub}
}

// GetConfig returns the current application config.
func (s *SettingsService) GetConfig() config.Config { return config.AppConfig }

// SaveConfig merges the partial over the current config, persists it and
// notifies the frontends (frontend-contract §3 SaveConfig).
func (s *SettingsService) SaveConfig(partial map[string]any) (config.Config, error) {
	if err := applyConfig(partial); err != nil {
		return config.AppConfig, err
	}
	if err := config.SaveConfig(); err != nil {
		return config.AppConfig, err
	}
	// Apply to the running engine without restart (FR-016, C-004): the same
	// engine instance keeps the session and rebuilds its API client.
	if s.hub != nil && s.hub.Engine != nil {
		s.hub.Engine.RefreshClient()
	}
	s.hub.emit("config:changed", map[string]any{"config": config.AppConfig})
	return config.AppConfig, nil
}

// DiscoverModels queries models dynamically for the requested provider.
func (s *SettingsService) DiscoverModels(provider string) ([]api.DiscoveredModel, error) {
	apiKey := config.GetAPIKeyForProvider(provider)
	baseURL := config.AppConfig.BaseURL
	if provider == "ollama" && config.AppConfig.OllamaBaseURL != "" {
		baseURL = config.AppConfig.OllamaBaseURL
	}
	return api.DiscoverModels(provider, apiKey, baseURL)
}