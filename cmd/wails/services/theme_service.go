package services

import "github.com/user/go-claude-code/internal/config"

// ThemeService reads/writes the persisted app theme (T051, gui-contract §4
// Tema). The theme lives in internal/config (FR-007, FR-010): switching here
// is visible to the next engine/config surface without a restart.
type ThemeService struct {
	hub *Hub
}

func NewThemeService(hub *Hub) *ThemeService { return &ThemeService{hub: hub} }

// Get returns the current variant: "dark" or "light".
func (s *ThemeService) Get() string { return config.AppConfig.ThemeVariant }

// Set persists the variant ("dark"/"light") and emits `theme:changed`.
func (s *ThemeService) Set(variant string) (string, error) {
	if variant != "dark" && variant != "light" {
		return config.AppConfig.ThemeVariant, errInvalidTheme
	}
	config.AppConfig.ThemeVariant = variant
	if err := config.SaveConfig(); err != nil {
		return config.AppConfig.ThemeVariant, err
	}
	s.hub.emit("theme:changed", map[string]any{"theme": variant})
	return variant, nil
}

var errInvalidTheme = errorString("theme must be \"dark\" or \"light\"")

type errorString string

func (e errorString) Error() string { return string(e) }