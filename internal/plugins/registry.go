package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"runtime"
	"sync"
)

// Plugin represents a loaded plugin
type Plugin struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Path        string                 `json:"path"`
	Type        string                 `json:"type"` // currently only "native" is supported
	Config      map[string]interface{} `json:"config,omitempty"`
	Enabled     bool                   `json:"enabled"`
	Loaded      bool                   `json:"-"`
	Instance    interface{}            `json:"-"`
}

// Registry manages plugins
type Registry struct {
	plugins    map[string]*Plugin
	configPath string
	mu         sync.RWMutex
}

var (
	registryInstance *Registry
	registryOnce     sync.Once
)

// GetRegistry returns the singleton plugin registry
func GetRegistry() *Registry {
	registryOnce.Do(func() {
		registryInstance = &Registry{
			plugins: make(map[string]*Plugin),
		}
		registryInstance.loadConfig()
	})
	return registryInstance
}

// getConfigPath returns the path to the plugins config
func (r *Registry) getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".letsGo", "plugins.json")
}

// getPluginDir returns the plugins directory
func (r *Registry) getPluginDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".letsGo", "plugins")
}

// loadConfig loads plugin configuration
func (r *Registry) loadConfig() {
	// Ensure plugin directory exists
	pluginDir := r.getPluginDir()
	os.MkdirAll(pluginDir, 0755)

	// Load config
	configPath := r.getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		// No config yet
		return
	}

	var plugins map[string]*Plugin
	if err := json.Unmarshal(data, &plugins); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading plugins config: %v\n", err)
		return
	}

	r.plugins = plugins

	// Auto-load enabled plugins
	for name, p := range r.plugins {
		if p.Enabled {
			if err := r.LoadPlugin(name); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load plugin %s: %v\n", name, err)
			}
		}
	}
}

// saveConfig saves plugin configuration
func (r *Registry) saveConfig() error {
	configPath := r.getConfigPath()

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(r.plugins, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// InstallPlugin installs a plugin from a path or URL
func (r *Registry) InstallPlugin(source string, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already installed
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' is already installed", name)
	}

	pluginDir := r.getPluginDir()
	destPath := filepath.Join(pluginDir, name)

	// Determine source type and install
	if isGitURL(source) {
		// Clone git repository
		if err := r.installFromGit(source, destPath); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	} else if isLocalPath(source) {
		// Copy local directory
		if err := copyDir(source, destPath); err != nil {
			return fmt.Errorf("failed to copy plugin: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported source type: %s", source)
	}

	// Read plugin manifest
	manifest, err := r.readManifest(destPath)
	if err != nil {
		// Clean up on error
		os.RemoveAll(destPath)
		return fmt.Errorf("failed to read plugin manifest: %w", err)
	}

	// Create plugin entry
	p := &Plugin{
		Name:        name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Path:        destPath,
		Type:        manifest.Type,
		Config:      make(map[string]interface{}),
		Enabled:     true,
	}

	r.plugins[name] = p

	if err := r.saveConfig(); err != nil {
		return err
	}

	// Try to load the plugin
	if err := r.loadPluginLocked(name); err != nil {
		// Don't fail installation if loading fails
		fmt.Fprintf(os.Stderr, "Warning: plugin installed but failed to load: %v\n", err)
	}

	return nil
}

// installFromGit clones a git repository
func (r *Registry) installFromGit(url, dest string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", url, dest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s", output)
	}
	return nil
}

// UninstallPlugin removes a plugin
func (r *Registry) UninstallPlugin(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	// Unload if loaded
	if p.Loaded {
		r.unloadPluginLocked(name)
	}

	// Remove directory
	if err := os.RemoveAll(p.Path); err != nil {
		return fmt.Errorf("failed to remove plugin files: %w", err)
	}

	delete(r.plugins, name)
	return r.saveConfig()
}

// LoadPlugin loads a plugin
func (r *Registry) LoadPlugin(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.loadPluginLocked(name)
}

func (r *Registry) loadPluginLocked(name string) error {
	p, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	if p.Loaded {
		return nil // Already loaded
	}

	switch p.Type {
	case "native":
		if err := r.loadNativePlugin(p); err != nil {
			return err
		}
	case "wasm":
		if err := r.loadWASMPlugin(p); err != nil {
			return err
		}
	case "script":
		if err := r.loadScriptPlugin(p); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown plugin type: %s", p.Type)
	}

	p.Loaded = true
	return nil
}

// loadNativePlugin loads a native Go plugin
func (r *Registry) loadNativePlugin(p *Plugin) error {
	if runtime.GOOS == "windows" {
		// Go plugins not supported on Windows
		return fmt.Errorf("native plugins not supported on Windows")
	}

	pluginPath := filepath.Join(p.Path, "plugin.so")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		// Try to build it
		if err := r.buildNativePlugin(p); err != nil {
			return err
		}
	}

	// Load the plugin
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for exported symbols
	initSym, err := plug.Lookup("Init")
	if err == nil {
		if initFn, ok := initSym.(func() error); ok {
			if err := initFn(); err != nil {
				return fmt.Errorf("plugin init failed: %w", err)
			}
		}
	}

	p.Instance = plug
	return nil
}

// buildNativePlugin builds a native Go plugin
func (r *Registry) buildNativePlugin(p *Plugin) error {
	// This would require Go compiler to be available
	// For now, return error indicating manual build needed
	return fmt.Errorf("plugin needs to be built: cd %s && go build -buildmode=plugin -o plugin.so .", p.Path)
}

// loadWASMPlugin loads a WebAssembly plugin
func (r *Registry) loadWASMPlugin(p *Plugin) error {
	// WASM plugin support would require a WASM runtime
	// For now, return a placeholder
	return fmt.Errorf("WASM plugins not yet implemented")
}

// loadScriptPlugin loads a script-based plugin
func (r *Registry) loadScriptPlugin(p *Plugin) error {
	// Script plugins would be interpreted
	// For now, return a placeholder
	return fmt.Errorf("script plugins not yet implemented")
}

// unloadPluginLocked unloads a plugin
func (r *Registry) unloadPluginLocked(name string) {
	p, exists := r.plugins[name]
	if !exists || !p.Loaded {
		return
	}

	// Call cleanup if available
	if p.Type == "native" && p.Instance != nil {
		if plug, ok := p.Instance.(*plugin.Plugin); ok {
			if cleanupSym, err := plug.Lookup("Cleanup"); err == nil {
				if cleanupFn, ok := cleanupSym.(func()); ok {
					cleanupFn()
				}
			}
		}
	}

	p.Instance = nil
	p.Loaded = false
}

// EnablePlugin enables a plugin
func (r *Registry) EnablePlugin(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	p.Enabled = true
	if err := r.saveConfig(); err != nil {
		return err
	}

	// Try to load
	return r.loadPluginLocked(name)
}

// DisablePlugin disables a plugin
func (r *Registry) DisablePlugin(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	p.Enabled = false
	r.unloadPluginLocked(name)
	return r.saveConfig()
}

// GetPlugin returns a plugin
func (r *Registry) GetPlugin(name string) (*Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	return p, nil
}

// ListPlugins returns all plugins
func (r *Registry) ListPlugins() map[string]*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*Plugin)
	for k, v := range r.plugins {
		result[k] = v
	}
	return result
}

// ListLoadedPlugins returns loaded plugins
func (r *Registry) ListLoadedPlugins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var loaded []string
	for name, p := range r.plugins {
		if p.Loaded {
			loaded = append(loaded, name)
		}
	}
	return loaded
}

// Manifest represents a plugin manifest
type Manifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Type        string `json:"type"` // currently only "native" is supported
	Main        string `json:"main,omitempty"`
	EntryPoint  string `json:"entryPoint,omitempty"`
}

// readManifest reads plugin manifest
func (r *Registry) readManifest(path string) (*Manifest, error) {
	manifestPath := filepath.Join(path, "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	// Validate
	if manifest.Name == "" {
		return nil, fmt.Errorf("manifest missing name")
	}
	if manifest.Type == "" {
		manifest.Type = "native" // Default
	}

	if manifest.Type != "native" {
		return nil, fmt.Errorf("unsupported plugin type '%s': only 'native' plugins are currently supported", manifest.Type)
	}

	return &manifest, nil
}

// Helper functions

func isGitURL(url string) bool {
	return len(url) > 4 && (url[:4] == "git:" || url[:8] == "https://" || url[:7] == "http://")
}

func isLocalPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}
