package services

import "time"

// AppService exposes app-level metadata to the Wails frontend (005, T002).
// The per-domain services (chat, sessions, git, ...) land here in US3.
type AppService struct{}

func (s *AppService) AppInfo() map[string]string {
	return map[string]string{
		"name":    "LetsGO",
		"version": "0.1.0",
		"started": time.Now().UTC().Format(time.RFC3339),
	}
}