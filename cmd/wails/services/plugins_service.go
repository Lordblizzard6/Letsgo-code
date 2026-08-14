package services

import (
	"github.com/user/go-claude-code/internal/plugins"
)

// PluginsService exposes the plugin registry for the rail Plugins pane
// (T048, gui-contract §4): list, install/uninstall, load/unload, enable,
// disable.
type PluginsService struct{}

func NewPluginsService() *PluginsService { return &PluginsService{} }

// Plugin is the JSON-safe row for the plugins pane.
type Plugin struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	Loaded      bool   `json:"loaded"`
}

// List returns all registered plugins.
func (s *PluginsService) List() []Plugin {
	r := plugins.GetRegistry()
	all := r.ListPlugins()
	loaded := r.ListLoadedPlugins()
	var out []Plugin
	for _, p := range all {
		out = append(out, Plugin{
			Name:        p.Name,
			Version:     p.Version,
			Description: p.Description,
			Enabled:     p.Enabled,
			Loaded:      contains(loaded, p.Name),
		})
	}
	return out
}

// Install clones the (git) source and registers the plugin.
func (s *PluginsService) Install(source, name string) error {
	return plugins.GetRegistry().InstallPlugin(source, name)
}

// Uninstall removes a plugin.
func (s *PluginsService) Uninstall(name string) error {
	return plugins.GetRegistry().UninstallPlugin(name)
}

// Load loads a plugin into the runtime.
func (s *PluginsService) Load(name string) error { return plugins.GetRegistry().LoadPlugin(name) }

// Enable enables a plugin.
func (s *PluginsService) Enable(name string) error { return plugins.GetRegistry().EnablePlugin(name) }

// Disable disables a plugin.
func (s *PluginsService) Disable(name string) error {
	return plugins.GetRegistry().DisablePlugin(name)
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}